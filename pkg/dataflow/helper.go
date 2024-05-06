/*
Copyright 2024 The Fluid Authors.

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

package dataflow

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenerateNodeAffinity(reader client.Reader, pod *corev1.Pod) (*corev1.NodeAffinity, error) {
	if pod == nil {
		return nil, nil
	}
	nodeName := pod.Spec.NodeName
	if len(nodeName) == 0 {
		return nil, nil
	}

	node, err := kubeclient.GetNode(reader, nodeName)
	if err != nil {
		return nil, fmt.Errorf("error to get node %s: %v", nodeName, err)
	}

	// node name
	nodeAffinity := &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      common.K8sNodeNameLabelKey,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{nodeName},
						},
					},
				},
			},
		},
	}

	// region
	region, exist := node.Labels[common.K8sRegionLabelKey]
	if exist {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
			append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions,
				corev1.NodeSelectorRequirement{
					Key:      common.K8sRegionLabelKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{region},
				})
	}
	// zone
	zone, exist := node.Labels[common.K8sZoneLabelKey]
	if exist {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
			append(nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions,
				corev1.NodeSelectorRequirement{
					Key:      common.K8sZoneLabelKey,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{zone},
				})
	}

	// customized labels
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		fillCustomizedNodeAffinity(pod.Spec.Affinity.NodeAffinity, nodeAffinity, node)
	}

	return nodeAffinity, nil
}

func fillCustomizedNodeAffinity(podNodeAffinity *corev1.NodeAffinity, dstNodeAffinity *corev1.NodeAffinity, node *corev1.Node) {
	// prefer
	for _, term := range podNodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		for _, expression := range term.Preference.MatchExpressions {
			// use the actually value in the node. Transform In, NotIn, Exists, DoesNotExist. Gt, and Lt to In.
			value, exist := node.Labels[expression.Key]
			if exist {
				dstNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
					append(dstNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions,
						corev1.NodeSelectorRequirement{
							Key:      expression.Key,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{value},
						})
			}
		}
	}

	if podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return
	}

	// require
	for _, term := range podNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, expression := range term.MatchExpressions {
			// use the actually value in the node. Transform In, NotIn, Exists, DoesNotExist. Gt, and Lt to In.
			value, exist := node.Labels[expression.Key]
			if exist {
				dstNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions =
					append(dstNodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions,
						corev1.NodeSelectorRequirement{
							Key:      expression.Key,
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{value},
						})
			}
		}
	}
}
