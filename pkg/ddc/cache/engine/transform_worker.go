/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *CacheEngine) transformWorker(dataset *datav1alpha1.Dataset, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass,
	commonConfig *CacheRuntimeComponentCommonConfig, value *common.CacheRuntimeValue) error {
	if runtimeClass.Topology == nil || runtimeClass.Topology.Worker == nil || runtime.Spec.Worker.Disabled {
		value.Worker = &common.CacheRuntimeComponentValue{Enabled: false}
		return nil
	}

	runtimeWorker := runtime.Spec.Worker
	componentDefinition := runtimeClass.Topology.Worker

	// Initialize component value with common fields
	var err error
	value.Worker, err = e.initComponentValue(common.ComponentTypeWorker, componentDefinition, commonConfig.Owner, runtimeWorker.Replicas)
	if err != nil {
		return err
	}

	// TODO: TieredStore handling

	// transform container related config, currently only modify the first container
	e.transformComponentPodTemplate(runtimeWorker.RuntimeComponentCommonSpec, dataset, value.Worker)

	// make sure affinity not nil
	if value.Worker.PodTemplateSpec.Spec.Affinity == nil {
		value.Worker.PodTemplateSpec.Spec.Affinity = &corev1.Affinity{}
	}

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	// dataset.Spec.NodeAffinity only affects worker (cache) pods
	e.buildWorkerAffinity(value.Worker.PodTemplateSpec.Spec.Affinity, dataset, runtimeInfo)

	// inject stateful set pod match labels for workers
	value.Worker.MatchLabels = map[string]string{
		common.LabelAnnotationDataset:          runtimeInfo.GetOwnerDatasetUID(),
		common.LabelAnnotationDatasetPlacement: (string)(runtimeInfo.GetPlacementModeWithDefault(datav1alpha1.ExclusiveMode)),
	}

	// transform all volume-related configurations
	err = e.transformVolumes(runtime.Spec.Volumes, runtime.Spec.Worker.VolumeMounts, dataset, componentDefinition, commonConfig, true, &value.Worker.PodTemplateSpec.Spec)
	if err != nil {
		return err
	}

	return nil
}

// buildWorkerAffinity builds affinity for worker pods, refer to Helper.BuildWorkerAffinity
func (e *CacheEngine) buildWorkerAffinity(affinity *corev1.Affinity, dataset *datav1alpha1.Dataset, runtimeInfo base.RuntimeInfoInterface) {
	// 1. Set pod anti affinity(required) for same dataset (Current using port conflict for scheduling, no need to do)

	// 2. Set pod anti affinity for the different dataset
	if affinity.PodAntiAffinity == nil {
		// Ensure PodAntiAffinity exists
		affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
	}

	if dataset.IsExclusiveMode() {
		affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      common.LabelAnnotationDataset,
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
				TopologyKey: common.K8sNodeNameLabelKey,
			},
		)
	} else {
		// Append to PreferredDuringSchedulingIgnoredDuringExecution
		affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			corev1.WeightedPodAffinityTerm{
				// The default weight is 50
				Weight: 50,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      common.LabelAnnotationDataset,
								Operator: metav1.LabelSelectorOpExists,
							},
						},
					},
					TopologyKey: common.K8sNodeNameLabelKey,
				},
			},
		)
		// Append to RequiredDuringSchedulingIgnoredDuringExecution
		affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      common.LabelAnnotationDatasetPlacement,
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{string(datav1alpha1.ExclusiveMode)},
						},
					},
				},
				TopologyKey: common.K8sNodeNameLabelKey,
			},
		)
	}

	// 3. Prefer to locate on the node which already has fuse on it
	if affinity.NodeAffinity == nil {
		affinity.NodeAffinity = &corev1.NodeAffinity{}
	}

	affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
		affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		corev1.PreferredSchedulingTerm{
			Weight: 100,
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      runtimeInfo.GetFuseLabelName(),
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		})

	// append dataset node affinity
	datasetNodeAffinity := dataset.Spec.NodeAffinity
	if datasetNodeAffinity == nil || datasetNodeAffinity.Required == nil || len(datasetNodeAffinity.Required.NodeSelectorTerms) == 0 {
		return
	}

	// Ensure NodeAffinity exists in result
	if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = datasetNodeAffinity.Required
		return
	}

	// Merge node selector terms from both
	affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(
		affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		datasetNodeAffinity.Required.NodeSelectorTerms...)
}
