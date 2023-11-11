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

package utils

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// IsPodManagedByFluid checks if the given Pod is managed by Fluid.
func IsPodManagedByFluid(pod *corev1.Pod) bool {
	fluidPodLabels := []string{common.AlluxioRuntime,
		common.JindoChartName,
		common.GooseFSRuntime,
		common.JuiceFSRuntime,
		common.ThinRuntime,
		common.EFCRuntime}
	if _, ok := pod.Labels[common.PodRoleType]; ok && pod.Labels[common.PodRoleType] == common.DataloadPod {
		return true
	}
	for _, label := range fluidPodLabels {
		if pod.Labels[common.App] == label {
			return true
		}
	}
	return false
}
