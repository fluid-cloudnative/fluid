/*
Copyright 2021 The Fluid Authors.

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

package ctrl

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/zapr"
	openkruise "github.com/openkruise/kruise/apis/apps/v1beta1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *Helper) checkWorkerAffinity(workers *cacheworkerset.CacheWorkerSet) (found bool) {
	zapLogger, _ := zap.NewProduction()
	logger := zapr.NewLogger(zapLogger)

	logger.Info("ENTER--+++++++++++++++++++++++--checkWorkerAffinity") // 进入函数时记录日志
	if workers == nil {
		logger.Info("checkWorkerAffinity workers is nil")
	}
	logger.Info("checkWorkerAffinity workers is not nil") // 进入函数时记录日志
	if workers.GetAffinity() == nil {
		return
	}

	if workers.GetNodeAffinity() == nil {
		return
	}

	if len(workers.GetNodeAffinityPreferredDuringSchedulingIgnoredDuringExecution()) == 0 {
		return
	}

	for _, preferred := range workers.GetNodeAffinityPreferredDuringSchedulingIgnoredDuringExecution() {
		for _, term := range preferred.Preference.MatchExpressions {
			if term.Key == e.runtimeInfo.GetFuseLabelName() {
				found = true
				return
			}
		}
	}

	return
}
func (e *Helper) checkWorkerAffinityForAts(workers *openkruise.StatefulSet) (found bool) {

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

func (e *Helper) BuildCacheWorkersAffinity(workers *cacheworkerset.CacheWorkerSet) (workersToUpdate *cacheworkerset.CacheWorkerSet, err error) {
	returnV, err := e.BuildWorkersAffinity(workers)
	return returnV, err
}

func (e *Helper) BuildWorkersAffinityForAsts(workers *openkruise.StatefulSet) (workersToUpdate *openkruise.StatefulSet, err error) {
	workersToUpdate = workers.DeepCopy()
	if e.checkWorkerAffinityForAts(workersToUpdate) {
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
	}

	return
}

// BuildWorkersAffinity builds workers affinity if it doesn't have
func (e *Helper) BuildWorkersAffinity(workers *cacheworkerset.CacheWorkerSet) (workersToUpdate *cacheworkerset.CacheWorkerSet, err error) {
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

	if workersToUpdate.GetAffinity() == nil {
		affinity := &corev1.Affinity{}
		workersToUpdate.SetAffinity(affinity)
		dataset, err := utils.GetDataset(e.client, name, namespace)
		if err != nil {
			return workersToUpdate, err
		}
		// 1. Set pod anti affinity(required) for same dataset (Current using port conflict for scheduling, no need to do)

		// 2. Set pod anti affinity for the different dataset
		if dataset.IsExclusiveMode() {
			podAntiAffinity := &corev1.PodAntiAffinity{
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
			workersToUpdate.SetPodAntiAffinity(podAntiAffinity)

		} else {
			podAntiAffinity := &corev1.PodAntiAffinity{
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
			workersToUpdate.SetPodAntiAffinity(podAntiAffinity)

			// TODO: remove this when EFC is ready for spread-first scheduling policy
			// Currently EFC prefers binpack-first scheduling policy to spread-first scheduling policy. Set PreferredDuringSchedulingIgnoredDuringExecution to empty
			// to avoid using spread-first scheduling policy
			if e.runtimeInfo.GetRuntimeType() == common.EFCRuntime {
				preferredDuringSchedulingIgnoredDuringExecution := []corev1.WeightedPodAffinityTerm{}
				workersToUpdate.SetPodAntiAffinityPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)

			}
		}

		// 3. Prefer to locate on the node which already has fuse on it
		if workersToUpdate.GetNodeAffinity() == nil {
			NodeAffinity := &corev1.NodeAffinity{}
			workersToUpdate.SetNodeAffinity(NodeAffinity)

		}

		if len(workersToUpdate.GetNodeAffinityPreferredDuringSchedulingIgnoredDuringExecution()) == 0 {

			PreferredDuringSchedulingIgnoredDuringExecution := []corev1.PreferredSchedulingTerm{}
			workersToUpdate.SetNodeAffinityPreferredDuringSchedulingIgnoredDuringExecution(PreferredDuringSchedulingIgnoredDuringExecution)

		}

		//workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
		//	append(workersToUpdate.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		//		corev1.PreferredSchedulingTerm{
		//			Weight: 100,
		//			Preference: corev1.NodeSelectorTerm{
		//				MatchExpressions: []corev1.NodeSelectorRequirement{
		//					{
		//						Key:      e.runtimeInfo.GetFuseLabelName(),
		//						Operator: corev1.NodeSelectorOpIn,
		//						Values:   []string{"true"},
		//					},
		//				},
		//			},
		//		})
		PreferredDuringSchedulingIgnoredDuringExecution := corev1.PreferredSchedulingTerm{
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
		}
		workersToUpdate.AppendNodeAffinityPreferredDuringSchedulingIgnoredDuringExecution(PreferredDuringSchedulingIgnoredDuringExecution)

		// 3. set node affinity if possible
		if dataset.Spec.NodeAffinity != nil {
			if dataset.Spec.NodeAffinity.Required != nil {
				workersToUpdate.SetNodeAffinityRequired(dataset.Spec.NodeAffinity.Required)

			}
		}
	}

	return
}
