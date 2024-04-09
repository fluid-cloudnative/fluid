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
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GenerateNodeLabels(c client.Client, pod *corev1.Pod) (map[string]string, error) {
	var result = map[string]string{}
	if pod == nil {
		return result, nil
	}
	nodeName := pod.Spec.NodeName
	if len(nodeName) == 0 {
		return result, nil
	}
	// node name
	result[common.K8sNodeNameLabelKey] = nodeName

	var node corev1.Node
	err := c.Get(context.TODO(), types.NamespacedName{Name: nodeName}, &node)
	if err != nil {
		return result, fmt.Errorf("error to get node %s: %v", nodeName, err)
	}
	// region
	region, exist := node.Labels[common.K8sRegionLabelKey]
	if exist {
		result[common.K8sRegionLabelKey] = region
	}
	// zone
	zone, exist := node.Labels[common.K8sZoneLabelKey]
	if exist {
		result[common.K8sZoneLabelKey] = zone
	}

	// customized labels
	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
		fillCustomizedNodeLabels(pod.Spec.Affinity.NodeAffinity, result, &node)
	}

	return result, nil
}

func fillCustomizedNodeLabels(nodeAffinity *corev1.NodeAffinity, result map[string]string, node *corev1.Node) {
	// prefer
	for _, term := range nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		for _, expression := range term.Preference.MatchExpressions {
			// use the actually value in the node.
			value, exist := node.Labels[expression.Key]
			if exist {
				result[expression.Key] = value
			}
		}
	}

	if nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return
	}

	// require
	for _, term := range nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, expression := range term.MatchExpressions {
			// use the actually value in the node.
			value, exist := node.Labels[expression.Key]
			if exist {
				result[expression.Key] = value
			}
		}
	}
}
