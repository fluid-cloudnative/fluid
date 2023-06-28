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
	"errors"
	"fmt"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

// getPhysicalDatasetRuntimeStatus get the runtime status of the physical dataset
func (e *ReferenceDatasetEngine) getPhysicalDatasetRuntimeStatus() (status *datav1alpha1.RuntimeStatus, err error) {
	physicalRuntimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil {
		return status, err
	}

	// if physicalRuntimeInfo is nil and no err, the runtime is deleting.
	if physicalRuntimeInfo == nil {
		return nil, nil
	}

	return base.GetRuntimeStatus(e.Client, physicalRuntimeInfo.GetRuntimeType(),
		physicalRuntimeInfo.GetName(), physicalRuntimeInfo.GetNamespace())
}

// getRuntime get the current runtime
func (e *ReferenceDatasetEngine) getRuntime() (*datav1alpha1.ThinRuntime, error) {
	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.ThinRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

func (e *ReferenceDatasetEngine) getRuntimeInfo() (base.RuntimeInfoInterface, error) {
	if e.runtimeInfo != nil {
		return e.runtimeInfo, nil
	}

	runtime, err := e.getRuntime()
	if err != nil {
		return e.runtimeInfo, err
	}

	e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, runtime.Spec.TieredStore, base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)))
	if err != nil {
		return e.runtimeInfo, err
	}

	// Setup Fuse Deploy Mode
	e.runtimeInfo.SetupFuseDeployMode(true, runtime.Spec.Fuse.NodeSelector)

	// Ignore the deprecated common labels and PersistentVolumes, use physical runtime

	// Setup with Dataset Info
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info("Dataset is notfound", "name", e.name, "namespace", e.namespace)
			return e.runtimeInfo, nil
		}

		e.Log.Info("Failed to get dataset when get runtimeInfo")
		return e.runtimeInfo, err
	}

	// set exclusive mode
	// TODO: how to handle the exclusive mode ?
	e.runtimeInfo.SetupWithDataset(dataset)

	e.Log.Info("Setup with dataset done", "exclusive", e.runtimeInfo.IsExclusive())

	return e.runtimeInfo, nil
}

// getPhysicalRuntimeInfo get physicalRuntimeInfo from dataset.
// If could not get dataset, getPhysicalRuntimeInfo try to get physicalRuntimeInfo from runtime status.
func (e *ReferenceDatasetEngine) getPhysicalRuntimeInfo() (base.RuntimeInfoInterface, error) {
	// If already have physicalRuntimeInfo, return it directly
	if e.physicalRuntimeInfo != nil {
		return e.physicalRuntimeInfo, nil
	}

	var physicalNameSpacedNames []types.NamespacedName

	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil && utils.IgnoreNotFound(err) != nil {
		// return if it is not a not-found error
		return e.physicalRuntimeInfo, err
	}

	if dataset != nil {
		// get physicalRuntimeInfo from dataset
		physicalNameSpacedNames = base.GetPhysicalDatasetFromMounts(dataset.Spec.Mounts)
	} else {
		// try to get physicalRuntimeInfo from runtime status
		runtime, err := e.getRuntime()
		if err != nil && utils.IgnoreNotFound(err) != nil {
			// return if it is not a not-found error
			return e.physicalRuntimeInfo, err
		}
		if len(runtime.Status.Mounts) != 0 {
			physicalNameSpacedNames = base.GetPhysicalDatasetFromMounts(runtime.Status.Mounts)
		}
	}

	if len(physicalNameSpacedNames) == 0 {
		// dataset is nil and len(runtime.Status.Mounts) is 0, return a not-found error
		return e.physicalRuntimeInfo, fmt.Errorf("%d: %s: %w", http.StatusNotFound, metav1.StatusReasonNotFound,
			errors.New("can't get physical runtime info from either dataset or runtime"))
	}
	if len(physicalNameSpacedNames) > 1 {
		return e.physicalRuntimeInfo, fmt.Errorf("ThinEngine with no profile name can only handle dataset only mounting one dataset but get %v", len(physicalNameSpacedNames))
	}
	namespacedName := physicalNameSpacedNames[0]

	physicalRuntimeInfo, err := base.GetRuntimeInfo(e.Client, namespacedName.Name, namespacedName.Namespace)
	if err != nil {
		return e.physicalRuntimeInfo, err
	}

	e.physicalRuntimeInfo = physicalRuntimeInfo

	return e.physicalRuntimeInfo, nil
}
