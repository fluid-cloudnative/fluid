/*
Copyright 2021 The Fluid Author.

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

import corev1 "k8s.io/api/core/v1"

// InjectPreferredSchedulingTerms inject the preferredSchedulingTerms into a pod
func InjectPreferredSchedulingTerms(preferredSchedulingTerms []corev1.PreferredSchedulingTerm, pod *corev1.Pod) {
	if len(preferredSchedulingTerms) == 0 {
		return
	}

	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &corev1.Affinity{}
	}

	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = preferredSchedulingTerms
	} else {
		pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
			append(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				preferredSchedulingTerms...)
	}
}

// InjectRequiredSchedulingTerms inject the NodeSelectorTerms into a pod
func InjectNodeSelectorTerms(requiredSchedulingTerms []corev1.NodeSelectorTerm, pod *corev1.Pod) {
	if len(requiredSchedulingTerms) == 0 {
		return
	}

	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &corev1.Affinity{}
	}

	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}

	if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = requiredSchedulingTerms
	} else {
		for i := 0; i < len(requiredSchedulingTerms); i++ {
			pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
				append(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions, requiredSchedulingTerms[i].MatchExpressions...)
		}
	}

}

func InjectMountPropagation(datasetNames []string, pod *corev1.Pod) {
	propagation := corev1.MountPropagationHostToContainer
	mountNames := make([]string, 0)
	for _, mount := range pod.Spec.Volumes {
		if mount.PersistentVolumeClaim != nil && ContainsString(datasetNames, mount.PersistentVolumeClaim.ClaimName) {
			mountNames = append(mountNames, mount.Name)
		}
	}
	for i, cn := range pod.Spec.Containers {
		for j, mount := range cn.VolumeMounts {
			if ContainsString(mountNames, mount.Name) && mount.MountPropagation == nil {
				pod.Spec.Containers[i].VolumeMounts[j].MountPropagation = &propagation
			}
		}
	}
}
