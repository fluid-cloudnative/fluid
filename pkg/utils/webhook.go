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
