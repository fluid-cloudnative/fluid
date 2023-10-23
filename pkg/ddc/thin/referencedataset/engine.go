/*
  Copyright 2022 The Fluid Authors.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package referencedataset

import (
	"context"
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

const (
	syncRetryDurationEnv     string = "FLUID_SYNC_RETRY_DURATION"
	defaultSyncRetryDuration        = 5 * time.Second
)

// Use compiler to check if the struct implements all the interface
var _ base.Engine = (*ReferenceDatasetEngine)(nil)

// ReferenceDatasetEngine is used for handling datasets mounting another dataset.
// We use `virtual` dataset/runtime to represent the reference dataset/runtime itself,
// and use `physical` dataset/runtime to represent the dataset/runtime is mounted by virtual dataset.
type ReferenceDatasetEngine struct {
	Id string
	client.Client
	Log logr.Logger

	name      string
	namespace string

	syncRetryDuration time.Duration
	timeOfLastSync    time.Time

	runtimeType string
	// use getRuntimeInfo instead of directly use this field
	runtimeInfo base.RuntimeInfoInterface
	// physical dataset corresponding runtimeInfo, use getPhysicalRuntimeInfo instead of directly use this field.
	physicalRuntimeInfo base.RuntimeInfoInterface
}

func (e *ReferenceDatasetEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *v1alpha1.OperationStatus, operation dataoperation.OperationReconcilerInterface) (ctrl.Result, error) {
	object := operation.GetObject()
	// reference thin engine not support data operation
	err := errors.NewNotSupported(
		schema.GroupResource{
			Group:    object.GetObjectKind().GroupVersionKind().Group,
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		}, "ThinRuntime")
	ctx.Log.Error(err, "ThinRuntime for reference dataset does not support data operations")
	ctx.Recorder.Event(object, v1.EventTypeWarning, common.DataOperationNotSupport, "thinEngine for reference dataset does not support data operations")
	return utils.NoRequeue()
}

// ID returns the id of the engine
func (e *ReferenceDatasetEngine) ID() string {
	return e.Id
}

// BuildReferenceDatasetThinEngine build engine for handling virtual dataset
func BuildReferenceDatasetThinEngine(id string, ctx cruntime.ReconcileRequestContext) (base.Engine, error) {
	// currently, ctx has no dataset
	engine := &ReferenceDatasetEngine{
		Id:          id,
		Client:      ctx.Client,
		name:        ctx.Name,
		namespace:   ctx.Namespace,
		runtimeType: ctx.RuntimeType,
	}
	engine.Log = ctx.Log.WithValues("virtual engine", ctx.RuntimeType).WithValues("id", id)

	// check if support the dataset mount format
	err := engine.checkDatasetMountSupport()
	if err != nil {
		return nil, err
	}

	// Build and setup runtime info
	_, err = engine.getRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get runtime info", ctx.Name)
	}

	// set sync duration
	duration, err := getSyncRetryDuration()
	if err != nil {
		engine.Log.Error(err, "Failed to parse syncRetryDurationEnv: FLUID_SYNC_RETRY_DURATION, use the default setting")
	}
	if duration != nil {
		engine.syncRetryDuration = *duration
	} else {
		engine.syncRetryDuration = defaultSyncRetryDuration
	}
	engine.timeOfLastSync = time.Now().Add(-engine.syncRetryDuration)
	engine.Log.Info("Set the syncRetryDuration", "syncRetryDuration", engine.syncRetryDuration)

	// get the physicalRuntimeInfo
	_, err = engine.getPhysicalRuntimeInfo()
	if err != nil {
		// return err if the runtime is running or error is not not-found
		if utils.IgnoreNotFound(err) != nil || ctx.Runtime.GetDeletionTimestamp().IsZero() {
			return nil, fmt.Errorf("engine %s failed to get physical dataset's runtime info", ctx.Name)
		}
	}

	return engine, err
}

func (e *ReferenceDatasetEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	// 1. get the physical datasets according the virtual dataset
	dataset := ctx.Dataset

	physicalDatasetNameSpacedNames := base.GetPhysicalDatasetFromMounts(dataset.Spec.Mounts)
	if len(physicalDatasetNameSpacedNames) != 1 {
		return false, fmt.Errorf("ThinEngine can only handle dataset only mounting one dataset")
	}
	namespacedName := physicalDatasetNameSpacedNames[0]

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		physicalDataset, err := utils.GetDataset(ctx.Client, namespacedName.Name, namespacedName.Namespace)
		if err != nil {
			return err
		}

		// 2. get runtime according to dataset status runtimes field
		runtimes := physicalDataset.Status.Runtimes
		if len(runtimes) == 0 {
			return fmt.Errorf("mounting dataset is not bound to a runtime yet")
		}

		// 3. add this dataset to physical dataset DatasetRef field
		datasetRefName := base.GetDatasetRefName(dataset.Name, dataset.Namespace)
		if !utils.ContainsString(physicalDataset.Status.DatasetRef, datasetRefName) {
			newDataset := physicalDataset.DeepCopy()
			newDataset.Status.DatasetRef = append(newDataset.Status.DatasetRef, datasetRefName)
			err := e.Client.Status().Update(context.TODO(), newDataset)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	// config map is for the fuse sidecar container
	runtimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil {
		return false, err
	}

	err = copyFuseDaemonSetForRefDataset(e.Client, dataset, runtimeInfo)
	if err != nil {
		return false, err
	}

	err = e.createConfigMapForRefDataset(e.Client, dataset, runtimeInfo)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Shutdown and clean up the engine
func (e *ReferenceDatasetEngine) Shutdown() (err error) {
	// 1. delete this dataset in physical dataset DatasetRef field
	datasetRefName := base.GetDatasetRefName(e.name, e.namespace)

	physicalRuntimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil && utils.IgnoreNotFound(err) != nil {
		return err
	}

	if physicalRuntimeInfo != nil {
		physicalDataset, err := utils.GetDataset(e.Client, physicalRuntimeInfo.GetName(), physicalRuntimeInfo.GetNamespace())
		if err != nil {
			return err
		}

		if utils.ContainsString(physicalDataset.Status.DatasetRef, datasetRefName) {
			newDataset := physicalDataset.DeepCopy()
			newDataset.Status.DatasetRef = utils.RemoveString(newDataset.Status.DatasetRef, datasetRefName)
			err := e.Client.Status().Update(context.TODO(), newDataset)
			if err != nil {
				return err
			}
		}
	} else {
		e.Log.Info("physicalRuntimeInfo is not found, can't update physical dataset datasetRef", "name", e.name, "namespace", e.namespace)
	}
	return
}

func (e *ReferenceDatasetEngine) checkDatasetMountSupport() error {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		// not found dataset error indicates the runtime is deleting, pass checkDatasetMountSupport
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("The dataset is not found, pass checkDatasetMountSupport because runtime is deleting")
			return nil
		} else {
			return err
		}
	}

	physicalDatasetNamespacedName := base.GetPhysicalDatasetFromMounts(dataset.Spec.Mounts)
	physicalSize := len(physicalDatasetNamespacedName)

	// currently only support dataset mounting only one dataset
	if physicalSize > 1 || physicalSize == 0 {
		return fmt.Errorf("ThinRuntime can only handle dataset only mounting one dataset")
	}

	// currently not support both 'dataset://' and other mount schema in one dataset
	if len(physicalDatasetNamespacedName) != len(dataset.Spec.Mounts) {
		return fmt.Errorf("dataset with 'dataset://' mount point can not has other mount format")
	}

	physicalDataset, err := utils.GetDataset(e.Client, physicalDatasetNamespacedName[0].Name, physicalDatasetNamespacedName[0].Namespace)
	if err != nil {
		return fmt.Errorf("failed to create reference dataset due to %v", err)
	}

	// currently not support physical dataset mounting another dataset
	if len(base.GetPhysicalDatasetFromMounts(physicalDataset.Spec.Mounts)) != 0 {
		return fmt.Errorf("ThinRuntime with no profile name can only handle dataset only mounting one dataset")
	}

	return nil
}
