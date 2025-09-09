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

package utils

import (
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("utils")
}

// InjectNodeSelectorRequirements injects(not append) a node selector term to affinity‘s nodeAffinity.
func InjectNodeSelectorRequirements(matchExpressions []corev1.NodeSelectorRequirement, affinity *corev1.Affinity) *corev1.Affinity {
	result := affinity

	if len(matchExpressions) == 0 {
		return result
	}

	if affinity == nil {
		result = &corev1.Affinity{}
	}

	if result.NodeAffinity == nil {
		result.NodeAffinity = &corev1.NodeAffinity{}
	}
	if result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}
	// no element
	if result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{
			{
				MatchExpressions: matchExpressions,
			},
		}
		return result
	}
	// has element, inject term's match expressions to each element
	for _, nodeSelectorTerm := range result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		nodeSelectorTerm.MatchExpressions = append(nodeSelectorTerm.MatchExpressions, matchExpressions...)
	}

	return result
}

func InjectPreferredSchedulingTermsToAffinity(terms []corev1.PreferredSchedulingTerm, affinity *corev1.Affinity) *corev1.Affinity {
	result := affinity
	if len(terms) == 0 {
		return result
	}
	if affinity == nil {
		result = &corev1.Affinity{}
	}

	if result.NodeAffinity == nil {
		result.NodeAffinity = &corev1.NodeAffinity{}
	}

	result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
		append(result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, terms...)

	return result
}
