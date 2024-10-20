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

import v1 "k8s.io/api/core/v1"

// InjectNodeSelectorRequirements injects(not append) a node selector term to affinity‘s nodeAffinity.
func InjectNodeSelectorRequirements(matchExpressions []v1.NodeSelectorRequirement, affinity *v1.Affinity) *v1.Affinity {
	result := affinity

	if len(matchExpressions) == 0 {
		return result
	}

	if affinity == nil {
		result = &v1.Affinity{}
	}

	if result.NodeAffinity == nil {
		result.NodeAffinity = &v1.NodeAffinity{}
	}
	if result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
	}
	// no element
	if result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms == nil {
		result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []v1.NodeSelectorTerm{
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

func InjectPreferredSchedulingTermsToAffinity(terms []v1.PreferredSchedulingTerm, affinity *v1.Affinity) *v1.Affinity {
	result := affinity
	if len(terms) == 0 {
		return result
	}
	if affinity == nil {
		result = &v1.Affinity{}
	}

	if result.NodeAffinity == nil {
		result.NodeAffinity = &v1.NodeAffinity{}
	}

	result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
		append(result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, terms...)

	return result
}
