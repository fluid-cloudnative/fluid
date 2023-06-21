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

// ReferenceDatasetEngine is used for handling datasets mounting another dataset
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
	// mounted dataset corresponding runtimeInfo,  use getMountedRuntimeInfo instead of directly use this field
	mountedRuntimeInfo base.RuntimeInfoInterface
}

func (e *ReferenceDatasetEngine) Operate(ctx cruntime.ReconcileRequestContext, object client.Object, opStatus *v1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
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

	// get the mountedRuntimeInfo
	_, err = engine.getMountedRuntimeInfo()
	if err != nil {
		return nil, fmt.Errorf("engine %s failed to get mounted dataset's runtime info", ctx.Name)
	}

	return engine, err
}

func (e *ReferenceDatasetEngine) Setup(ctx cruntime.ReconcileRequestContext) (ready bool, err error) {
	// 1. get the physical datasets according the virtual dataset
	dataset := ctx.Dataset

	physicalNameSpacedNames := base.GetMountedDatasetNamespacedName(dataset.Spec.Mounts)
	if len(physicalNameSpacedNames) != 1 {
		return false, fmt.Errorf("ThinEngine can only handle dataset only mounting one dataset")
	}
	namespacedName := physicalNameSpacedNames[0]

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		mountedDataset, err := utils.GetDataset(ctx.Client, namespacedName.Name, namespacedName.Namespace)
		if err != nil {
			return err
		}

		// 2. get runtime according to dataset status runtimes field
		runtimes := mountedDataset.Status.Runtimes
		if len(runtimes) == 0 {
			return fmt.Errorf("mounting dataset is not bound to a runtime yet")
		}

		// 3. add this dataset to mounted dataset DatasetRef field
		datasetRefName := base.GetDatasetRefName(dataset.Name, dataset.Namespace)
		if !utils.ContainsString(mountedDataset.Status.DatasetRef, datasetRefName) {
			newDataset := mountedDataset.DeepCopy()
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
	runtimeInfo, err := e.getMountedRuntimeInfo()
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
	// 1. delete this dataset to mounted dataset DatasetRef field
	datasetRefName := base.GetDatasetRefName(e.name, e.namespace)

	mountedRuntimeInfo, err := e.getMountedRuntimeInfo()
	if err != nil {
		return err
	}

	if mountedRuntimeInfo != nil {
		mountedDataset, err := utils.GetDataset(e.Client, mountedRuntimeInfo.GetName(), mountedRuntimeInfo.GetNamespace())
		if err != nil {
			return err
		}

		if utils.ContainsString(mountedDataset.Status.DatasetRef, datasetRefName) {
			newDataset := mountedDataset.DeepCopy()
			newDataset.Status.DatasetRef = utils.RemoveString(newDataset.Status.DatasetRef, datasetRefName)
			err := e.Client.Status().Update(context.TODO(), newDataset)
			if err != nil {
				return err
			}
		}
	}
	return
}

func (e *ReferenceDatasetEngine) checkDatasetMountSupport() error {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		// if not found dataset, it indicates the runtime is deleting, pass checkDatasetMountSupport
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("Not found dataset, runtime is deleting")
			return nil
		} else {
			return err
		}
	}

	mountedNamespacedName := base.GetMountedDatasetNamespacedName(dataset.Spec.Mounts)
	mountedSize := len(mountedNamespacedName)

	// currently only support dataset mounting only one dataset
	if mountedSize > 1 || mountedSize == 0 {
		return fmt.Errorf("ThinRuntime can only handle dataset only mounting one dataset")
	}

	// currently not support both 'dataset://' and other mount schema in one dataset
	if len(mountedNamespacedName) != len(dataset.Spec.Mounts) {
		return fmt.Errorf("dataset with 'dataset://' mount point can not has other mount format")
	}

	mountedDataset, err := utils.GetDataset(e.Client, mountedNamespacedName[0].Name, mountedNamespacedName[0].Namespace)
	if err != nil {
		return fmt.Errorf("failed to create reference dataset due to %v", err)
	}

	// currently not support mounted dataset mounting another dataset
	if len(base.GetMountedDatasetNamespacedName(mountedDataset.Spec.Mounts)) != 0 {
		return fmt.Errorf("ThinRuntime with no profile name can only handle dataset only mounting one dataset")
	}

	return nil
}
