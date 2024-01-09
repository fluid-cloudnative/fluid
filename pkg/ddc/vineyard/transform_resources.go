/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"context"
	"fmt"
	"reflect"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
)

func (e *VineyardEngine) transformResourcesForMaster(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) {

	if runtime == nil {
		return
	}
	if len(runtime.Spec.Master.Resources.Limits) > 0 || len(runtime.Spec.Master.Resources.Requests) > 0 {
		value.Master.Resources = utils.TransformRequirementsToResources(runtime.Spec.Master.Resources)
	}

}

// transformResourcesForWorker transform the resources for worker
// The tieredStore's memory size take precedence over the memory request and limit size.
// If the tieredStore's memory size is not found, the error will be returned.
// E,g.
// tieredStore:
//   - levels:
//     mediumtype: memory
//     size: 20Gi
//
// request:
//
//	memory: 10Gi
//
// limit:
//
//	memory: 15Gi
//
// After transform, the resources will be:
//
// request:
//
//	# tiered store memory size + 500Mi
//	memory: 20.5Gi
//
// limit:
//
//	# tiered store memory size + 500Mi
//	memory: 20.5Gi
func (e *VineyardEngine) transformResourcesForWorker(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) error {

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		e.Log.Error(err, "failed to transformResourcesForWorker")
		return err
	}
	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	e.Log.Info("transformResourcesForWorker", "storageMap", storageMap)

	// mem set request
	needUpdated := false
	var needSetMem resource.Quantity

	// vineyard itself need 500Mi memory to run
	vineyardMem, err := resource.ParseQuantity("500Mi")
	if err != nil {
		e.Log.Error(err, "failed to parse 500Mi to resource.Quantity")
		return err
	}
	isValid := false
	for key, requirement := range storageMap {
		if key == common.MemoryCacheStore {
			isValid = true
			req := requirement.DeepCopy()
			// if req mem > request mem && req mem < request mem - 500Mi
			standardMem := *runtime.Spec.Worker.Resources.Requests.Memory()
			standardMem.Sub(vineyardMem)

			if runtime.Spec.Worker.Resources.Requests == nil ||
				runtime.Spec.Worker.Resources.Requests.Memory() == nil ||
				runtime.Spec.Worker.Resources.Requests.Memory().IsZero() ||
				req.Cmp(*runtime.Spec.Worker.Resources.Requests.Memory()) > 0 ||
				req.Cmp(*runtime.Spec.Worker.Resources.Requests.Memory()) < 0 && req.Cmp(standardMem) > 0 {
				needUpdated = true
				needSetMem.Add(req)
				needSetMem.Add(vineyardMem)
			}

			if needUpdated || runtime.Spec.Worker.Resources.Limits == nil ||
				runtime.Spec.Worker.Resources.Limits.Memory() == nil ||
				runtime.Spec.Worker.Resources.Limits.Memory().IsZero() ||
				req.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) > 0 ||
				req.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) < 0 && req.Cmp(standardMem) > 0 {
				needUpdated = true

			}

		}

	}
	if !isValid {
		err = fmt.Errorf("the tierdStore's memory size is not found")
		return err
	}
	//for worker
	value.Worker.Resources = utils.TransformRequirementsToResources(runtime.Spec.Worker.Resources)

	if needUpdated {
		if value.Worker.Resources.Requests == nil {
			value.Worker.Resources.Requests = make(common.ResourceList)
		}
		value.Worker.Resources.Requests[corev1.ResourceMemory] = needSetMem.String()
		if value.Worker.Resources.Limits == nil {
			value.Worker.Resources.Limits = make(common.ResourceList)
		}
		// if limits mem > request mem + 500Mi, set the original limits mem
		if !runtime.Spec.Worker.Resources.Limits.Memory().IsZero() &&
			needSetMem.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) < 0 {
			value.Worker.Resources.Limits[corev1.ResourceMemory] = runtime.Spec.Worker.Resources.Limits.Memory().String()
		} else {
			value.Worker.Resources.Limits[corev1.ResourceMemory] = needSetMem.String()
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			runtime, err := e.getRuntime()
			if err != nil {
				return err
			}
			runtimeToUpdate := runtime.DeepCopy()
			if len(runtimeToUpdate.Spec.Worker.Resources.Requests) == 0 {
				runtimeToUpdate.Spec.Worker.Resources.Requests = make(corev1.ResourceList)
			}
			runtimeToUpdate.Spec.Worker.Resources.Requests[corev1.ResourceMemory] = needSetMem
			if !reflect.DeepEqual(runtimeToUpdate, runtime) {
				err = e.Client.Update(context.TODO(), runtimeToUpdate)
				if err != nil {
					if apierrors.IsConflict(err) {
						time.Sleep(3 * time.Second)
					}
					return err
				}
				time.Sleep(1 * time.Second)
			}

			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func (e *VineyardEngine) transformResourcesForFuse(runtime *datav1alpha1.VineyardRuntime, value *Vineyard) {

	if runtime.Spec.Fuse.Resources.Limits == nil {
		e.Log.Info("skip setting memory limit")
		return
	}

	if _, found := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]; !found {
		e.Log.Info("skip setting memory limit")
		return
	}

	value.Fuse.Resources = utils.TransformRequirementsToResources(runtime.Spec.Fuse.Resources)

}
