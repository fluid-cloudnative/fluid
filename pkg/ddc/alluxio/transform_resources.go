/*

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

package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"strconv"
)

func (e *AlluxioEngine) transformResourcesForMaster(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {

	if runtime == nil {
		return
	}
	if len(runtime.Spec.Master.Resources.Limits) > 0 || len(runtime.Spec.Master.Resources.Requests) > 0 {
		value.Master.Resources = utils.TransformRequirementsToResources(runtime.Spec.Master.Resources)
	}
	if len(runtime.Spec.JobMaster.Resources.Limits) > 0 || len(runtime.Spec.JobMaster.Resources.Requests) > 0 {
		value.JobMaster.Resources = utils.TransformRequirementsToResources(runtime.Spec.JobMaster.Resources)
	}

}

func (e *AlluxioEngine) transformResourcesForWorker(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {

	//if runtime.Spec.Worker.Resources.Limits == nil {
	//	e.Log.Info("skip setting memory limit")
	//	return
	//}
	//
	//if _, found := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]; !found {
	//	e.Log.Info("skip setting memory limit")
	//	return
	//}

	// transform memory resource for worker
	value.Worker.Resources = utils.TransformRequirementsToResources(runtime.Spec.Worker.Resources)

	var memResSet bool
	if len(value.Worker.Resources.Requests) > 0 {
		if _, exists := value.Worker.Resources.Requests[corev1.ResourceMemory]; exists {
			memResSet = true
		}
	}

	if len(value.Worker.Resources.Limits) > 0 {
		if _, exists := value.Worker.Resources.Limits[corev1.ResourceMemory]; exists {
			memResSet = true
		}
	}

	if !memResSet {
		//todo(xuzhihao) MEM tieredstore quota + Alluxio worker JVM usage
	} else {
		e.Log.Info("User already set memory request or limit. Skipping setting it.")
	}

	// transform disk resource for worker
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		e.Log.Error(err, "failed to get runtime info when transforming resources for worker")
	}

	devQuotaMap := tieredstore.GetDeviceQuotaMap(runtimeInfo)
	e.Log.Info("GetDeviceQuotaMap", "devQuotaMap", devQuotaMap)
	for dev, quota := range devQuotaMap {
		// Fluid uses GiB as the resource unit
		if value.Worker.Resources.Limits == nil {
			value.Worker.Resources.Limits = common.ResourceList{}
		}
		value.Worker.Resources.Limits[corev1.ResourceName(dev)] = strconv.FormatInt(quota.ScaledValue(resource.Giga), 10)
	}

	// transform resource for job worker
	value.JobWorker.Resources = utils.TransformRequirementsToResources(runtime.Spec.JobWorker.Resources)

	// for job worker
	//if len(runtime.Spec.JobWorker.Resources.Limits) > 0 || len(runtime.Spec.JobWorker.Resources.Requests) > 0 {
	//	value.JobWorker.Resources = utils.TransformRequirementsToResources(runtime.Spec.JobWorker.Resources)
	//}

	//runtimeInfo, err := e.getRuntimeInfo()
	//if err != nil {
	//	e.Log.Error(err, "failed to transformResourcesForWorker")
	//}
	//storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)
	//
	//e.Log.Info("transformResourcesForWorker", "storageMap", storageMap)
	//
	//// TODO(iluoeli): it should be xmx + direct memory
	//memLimit := resource.MustParse("20Gi")
	//if quantity, exists := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]; exists && !quantity.IsZero() {
	//	memLimit = quantity
	//}
	//
	//for key, requirement := range storageMap {
	//	if value.Worker.Resources.Limits == nil {
	//		value.Worker.Resources.Limits = make(common.ResourceList)
	//	}
	//	if key == common.MemoryCacheStore {
	//		req := requirement.DeepCopy()
	//
	//		memLimit.Add(req)
	//
	//		e.Log.Info("update the requirement for memory", "requirement", memLimit)
	//
	//	}
	//	// } else if key == common.DiskCacheStore {
	//	// 	req := requirement.DeepCopy()
	//
	//	// 	e.Log.Info("update the requiremnet for disk", "requirement", req)
	//
	//	// 	value.Worker.Resources.Limits[corev1.ResourceEphemeralStorage] = req.String()
	//	// }
	//}
	//
	//value.Worker.Resources.Limits[corev1.ResourceMemory] = memLimit.String()
}

func (e *AlluxioEngine) transformResourcesForFuse(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {

	if runtime.Spec.Fuse.Resources.Limits == nil {
		e.Log.Info("skip setting memory limit")
		return
	}

	if _, found := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]; !found {
		e.Log.Info("skip setting memory limit")
		return
	}

	value.Fuse.Resources = utils.TransformRequirementsToResources(runtime.Spec.Fuse.Resources)

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		e.Log.Error(err, "failed to transformResourcesForFuse")
	}
	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	e.Log.Info("transformFuse", "storageMap", storageMap)

	// TODO(iluoeli): it should be xmx + direct memory
	memLimit := resource.MustParse("50Gi")
	if quantity, exists := runtime.Spec.Fuse.Resources.Limits[corev1.ResourceMemory]; exists && !quantity.IsZero() {
		memLimit = quantity
	}

	for key, requirement := range storageMap {
		if value.Fuse.Resources.Limits == nil {
			value.Fuse.Resources.Limits = make(common.ResourceList)
		}
		if key == common.MemoryCacheStore {
			req := requirement.DeepCopy()

			memLimit.Add(req)

			e.Log.Info("update the requiremnet for memory", "requirement", memLimit)

		}
		// } else if key == common.DiskCacheStore {
		// 	req := requirement.DeepCopy()
		// 	e.Log.Info("update the requiremnet for disk", "requirement", req)
		// 	value.Fuse.Resources.Limits[corev1.ResourceEphemeralStorage] = req.String()
		// }
	}
	if value.Fuse.Resources.Limits != nil {
		value.Fuse.Resources.Limits[corev1.ResourceMemory] = memLimit.String()
	}

}

func (e *AlluxioEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *Alluxio) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}
