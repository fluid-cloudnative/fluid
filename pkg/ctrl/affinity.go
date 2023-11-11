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

package ctrl

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Helper) checkWorkerAffinity(workers *appsv1.StatefulSet) (found bool) {

	if workers.Spec.Template.Spec.Affinity == nil {
		return
	}

	if workers.Spec.Template.Spec.Affinity.NodeAffinity == nil {
		return
	}

	if len(workers.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
		return
	}

	for _, preferred := range workers.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		for _, term := range preferred.Preference.MatchExpressions {
			if term.Key == e.runtimeInfo.GetFuseLabelName() {
				found = true
				return
			}
		}
	}

	return
}

// BuildWorkersAffinity builds workers affinity if it doesn't have
func (e *Helper) BuildWorkersAffinity(workers *appsv1.StatefulSet) (workersToUpdate *appsv1.StatefulSet, err error) {
	// TODO: for now, runtime affinity can't be set by user, so we can assume the affinity is nil in the first time.
	// We need to enhance it in future
	workersToUpdate = workers.DeepCopy()
	if e.checkWorkerAffinity(workersToUpdate) {
		return
	}
	var (
		name      = e.runtimeInfo.GetName()
		namespace = e.runtimeInfo.GetNamespace()
	)

	if workersToUpdate.Spec.Template.Spec.Affinity == nil {
		workersToUpdate.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		dataset, err := utils.GetDataset(e.client, name, namespace)
		if err != nil {
			return workersToUpdate, err
		}
		// 1. Set pod anti affinity(required) for same dataset (Current using port conflict for scheduling, no need to do)

		// 2. Set pod anti affinity for the different dataset
		if dataset.IsExclusiveMode() {
			workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset",
									Operator: metav1.LabelSelectorOpExists,
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}
		} else {
			workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						// The default weight is 50
						Weight: 50,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "fluid.io/dataset",
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "fluid.io/dataset-placement",
									Operator: metav1.LabelSelectorOpIn,
									Values:   []string{string(datav1alpha1.ExclusiveMode)},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			}

			// TODO: remove this when EFC is ready for spread-first scheduling policy
			// Currently EFC prefers binpack-first scheduling policy to spread-first scheduling policy. Set PreferredDuringSchedulingIgnoredDuringExecution to empty
			// to avoid using spread-first scheduling policy
			if e.runtimeInfo.GetRuntimeType() == common.EFCRuntime {
				workersToUpdate.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{}
			}
		}

		// 3. Prefer to locate on the node which already has fuse on it
		if workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity == nil {
			workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
		}

		if len(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0 {
			workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{}
		}

		workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
			append(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
				corev1.PreferredSchedulingTerm{
					Weight: 100,
					Preference: corev1.NodeSelectorTerm{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      e.runtimeInfo.GetFuseLabelName(),
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"true"},
							},
						},
					},
				})

		// 3. set node affinity if possible
		if dataset.Spec.NodeAffinity != nil {
			if dataset.Spec.NodeAffinity.Required != nil {
				workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution =
					dataset.Spec.NodeAffinity.Required
			}
		}

		err = e.client.Update(context.TODO(), workersToUpdate)
		if err != nil {
			return workersToUpdate, err
		}

	}

	return
}
