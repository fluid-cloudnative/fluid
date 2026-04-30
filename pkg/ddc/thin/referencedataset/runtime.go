/*
  Copyright 2023 The Fluid Authors.

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
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// getPhysicalDatasetRuntimeStatus get the runtime status of the physical dataset
// Note: This function only supports DDC-based runtimes (Alluxio, Jindo, etc.)
// CacheRuntime is not supported because its status structure is incompatible with ThinRuntime.
func (e *ReferenceDatasetEngine) getPhysicalDatasetRuntimeStatus() (status *datav1alpha1.RuntimeStatus, err error) {
	physicalRuntimeInfo, err := e.getPhysicalRuntimeInfo()
	if err != nil {
		return status, err
	}

	// if physicalRuntimeInfo is nil and no err, the runtime is deleting.
	if physicalRuntimeInfo == nil {
		return nil, nil
	}

	return getRuntimeStatus(e.Client, physicalRuntimeInfo.GetRuntimeType(),
		physicalRuntimeInfo.GetName(), physicalRuntimeInfo.GetNamespace())
}

// getRuntimeStatus gets the runtime status according to the runtime type, name, and namespace.
// This is a private function used only within the thin runtime package for status synchronization.
// It returns the raw RuntimeStatus which is incompatible with CacheRuntime's CacheRuntimeStatus.
func getRuntimeStatus(client client.Client, runtimeType, name, namespace string) (status *datav1alpha1.RuntimeStatus, err error) {
	switch runtimeType {
	case common.AlluxioRuntime:
		runtime, err := utils.GetAlluxioRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.JindoRuntime:
		runtime, err := utils.GetJindoRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.GooseFSRuntime:
		runtime, err := utils.GetGooseFSRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.JuiceFSRuntime:
		runtime, err := utils.GetJuiceFSRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.EFCRuntime:
		runtime, err := utils.GetEFCRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.ThinRuntime:
		runtime, err := utils.GetThinRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	case common.VineyardRuntime:
		runtime, err := utils.GetVineyardRuntime(client, name, namespace)
		if err != nil {
			return status, err
		}
		return &runtime.Status, nil
	default:
		err = fmt.Errorf("%s is not supported as physical runtime for ThinRuntime with reference dataset", runtimeType)
		return nil, err
	}
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
	if e.runtimeInfo == nil {
		runtime, err := e.getRuntime()
		if err != nil {
			return e.runtimeInfo, err
		}

		opts := []base.RuntimeInfoOption{
			base.WithTieredStore(runtime.Spec.TieredStore),
			base.WithMetadataList(base.GetMetadataListFromAnnotation(runtime)),
			base.WithAnnotations(runtime.Annotations),
		}

		e.runtimeInfo, err = base.BuildRuntimeInfo(e.name, e.namespace, e.runtimeType, opts...)
		if err != nil {
			return e.runtimeInfo, err
		}

		// Setup Fuse Deploy Mode
		e.runtimeInfo.SetFuseNodeSelector(runtime.Spec.Fuse.NodeSelector)
	}

	// Handling information of bound dataset. XXXEngine.getRuntimeInfo() might be called before the runtime is bound to a dataset,
	// so here we must lazily set dataset-related information once we found there's one bound dataset.
	if len(e.runtimeInfo.GetOwnerDatasetUID()) == 0 {
		runtime, err := e.getRuntime()
		if err != nil {
			return nil, err
		}

		uid, err := base.GetOwnerDatasetUIDFromRuntimeMeta(runtime.ObjectMeta)
		if err != nil {
			return nil, err
		}

		if len(uid) > 0 {
			e.runtimeInfo.SetOwnerDatasetUID(uid)
		}
	}

	if !e.runtimeInfo.IsPlacementModeSet() {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if utils.IgnoreNotFound(err) != nil {
			return nil, err
		}

		if dataset != nil {
			e.runtimeInfo.SetupWithDataset(dataset)
		}
	}

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
		if err != nil {
			return e.physicalRuntimeInfo, err
		}
		if len(runtime.Status.Mounts) != 0 {
			physicalNameSpacedNames = base.GetPhysicalDatasetFromMounts(runtime.Status.Mounts)
		}
	}

	if len(physicalNameSpacedNames) == 0 {
		// dataset is nil and len(runtime.Status.Mounts) is 0, return a not-found error
		return e.physicalRuntimeInfo, &k8serrors.StatusError{
			ErrStatus: metav1.Status{
				Reason:  metav1.StatusReasonNotFound,
				Code:    http.StatusNotFound,
				Message: "can't get physical runtime info from either dataset or runtime",
			},
		}
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
