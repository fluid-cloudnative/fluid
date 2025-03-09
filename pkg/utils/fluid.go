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

package utils

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

// IsPodManagedByFluid checks if the given Pod is managed by Fluid.
func IsPodManagedByFluid(pod *corev1.Pod) bool {
	fluidPodLabels := []string{common.AlluxioRuntime,
		// For jindo, Data Operation / Runtime pods use jindofs, jindofsx, jindocache as the app label value.
		common.JindoChartName,
		common.JindoFSxEngineImpl,
		common.JindoCacheEngineImpl,
		common.GooseFSRuntime,
		common.JuiceFSRuntime,
		common.ThinRuntime,
		common.EFCRuntime,
		common.CacheRuntime}

	// Runtime Pod and DataOperation Pod both have the App label.
	for _, label := range fluidPodLabels {
		if pod.Labels[common.App] == label {
			return true
		}
	}
	return false
}
