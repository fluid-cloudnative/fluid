/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package goosefs

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (e *GooseFSEngine) transformResourcesForMaster(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {

	if runtime == nil {
		return
	}

	if len(runtime.Spec.Master.Resources.Limits) > 0 || len(runtime.Spec.Master.Resources.Requests) > 0 {
		value.Master.Resources = utils.TransformRequirementsToResources(runtime.Spec.Master.Resources)
	}
	if len(runtime.Spec.JobMaster.Resources.Limits) > 0 || len(runtime.Spec.JobMaster.Resources.Requests) > 0 {
		value.JobMaster.Resources = utils.TransformRequirementsToResources(runtime.Spec.JobMaster.Resources)
	}
	if len(runtime.Spec.Master.Resources.Limits) == 0 && len(runtime.Spec.Master.Resources.Requests) == 0 {
		return
	}

	value.Master.Resources = utils.TransformRequirementsToResources(runtime.Spec.Master.Resources)
}

func (e *GooseFSEngine) transformResourcesForWorker(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {

	if runtime.Spec.Worker.Resources.Limits == nil {
		e.Log.Info("skip setting memory limit")
		return
	}

	if _, found := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]; !found {
		e.Log.Info("skip setting memory limit")
		return
	}

	value.Worker.Resources = utils.TransformRequirementsToResources(runtime.Spec.Worker.Resources)

	// for job worker
	if len(runtime.Spec.JobWorker.Resources.Limits) > 0 || len(runtime.Spec.JobWorker.Resources.Requests) > 0 {
		value.JobWorker.Resources = utils.TransformRequirementsToResources(runtime.Spec.JobWorker.Resources)
	}

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		e.Log.Error(err, "failed to transformResourcesForWorker")
	}
	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	e.Log.Info("transformResourcesForWorker", "storageMap", storageMap)

	// TODO(iluoeli): it should be xmx + direct memory
	memLimit := resource.MustParse("20Gi")
	if quantity, exists := runtime.Spec.Worker.Resources.Limits[corev1.ResourceMemory]; exists && !quantity.IsZero() {
		memLimit = quantity
	}

	for key, requirement := range storageMap {
		if value.Worker.Resources.Limits == nil {
			value.Worker.Resources.Limits = make(common.ResourceList)
		}
		if key == common.MemoryCacheStore {
			req := requirement.DeepCopy()

			memLimit.Add(req)

			e.Log.Info("update the requirement for memory", "requirement", memLimit)

		}
		// } else if key == common.DiskCacheStore {
		// 	req := requirement.DeepCopy()

		// 	e.Log.Info("update the requiremnet for disk", "requirement", req)

		// 	value.Worker.Resources.Limits[corev1.ResourceEphemeralStorage] = req.String()
		// }
	}

	value.Worker.Resources.Limits[corev1.ResourceMemory] = memLimit.String()
}

func (e *GooseFSEngine) transformResourcesForFuse(runtime *datav1alpha1.GooseFSRuntime, value *GooseFS) {

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

func (e *GooseFSEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *GooseFS) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}
