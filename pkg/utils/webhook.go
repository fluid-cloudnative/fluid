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
		hasRequiredConstraints := false
		for i := 0; i < len(requiredSchedulingTerms); i++ {
			if len(requiredSchedulingTerms[i].MatchExpressions) != 0 || len(requiredSchedulingTerms[i].MatchFields) != 0 {
				hasRequiredConstraints = true
				break
			}
		}
		if !hasRequiredConstraints {
			return
		}

		existingTerms := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		combinedTerms := make([]corev1.NodeSelectorTerm, 0, len(existingTerms)*len(requiredSchedulingTerms))
		hasExistingConstraints := false
		for i := 0; i < len(existingTerms); i++ {
			if len(existingTerms[i].MatchExpressions) != 0 || len(existingTerms[i].MatchFields) != 0 {
				hasExistingConstraints = true
				break
			}
		}
		if !hasExistingConstraints {
			filteredTerms := make([]corev1.NodeSelectorTerm, 0, len(requiredSchedulingTerms))
			for i := 0; i < len(requiredSchedulingTerms); i++ {
				if len(requiredSchedulingTerms[i].MatchExpressions) == 0 && len(requiredSchedulingTerms[i].MatchFields) == 0 {
					continue
				}
				filteredTerms = append(filteredTerms, requiredSchedulingTerms[i])
			}
			pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = filteredTerms
			return
		}
		for i := 0; i < len(existingTerms); i++ {
			if len(existingTerms[i].MatchExpressions) == 0 && len(existingTerms[i].MatchFields) == 0 {
				continue
			}
			for j := 0; j < len(requiredSchedulingTerms); j++ {
				if len(requiredSchedulingTerms[j].MatchExpressions) == 0 && len(requiredSchedulingTerms[j].MatchFields) == 0 {
					continue
				}
				combinedTerm := corev1.NodeSelectorTerm{
					MatchExpressions: append(append([]corev1.NodeSelectorRequirement{}, existingTerms[i].MatchExpressions...), requiredSchedulingTerms[j].MatchExpressions...),
					MatchFields:      append(append([]corev1.NodeSelectorRequirement{}, existingTerms[i].MatchFields...), requiredSchedulingTerms[j].MatchFields...),
				}
				combinedTerms = append(combinedTerms, combinedTerm)
			}
		}
		if len(combinedTerms) == 0 {
			return
		}
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = combinedTerms
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
