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

func InjectNodeSelectorTermsToAffinity(terms []v1.NodeSelectorTerm, affinity *v1.Affinity) *v1.Affinity {
	result := affinity
	if affinity == nil {
		result = &v1.Affinity{}
	}

	if result.NodeAffinity == nil {
		result.NodeAffinity = &v1.NodeAffinity{}
	}
	if result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &v1.NodeSelector{}
	}
	result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms =
		append(result.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, terms...)

	return result
}

func InjectPreferredSchedulingTermsToAffinity(terms []v1.PreferredSchedulingTerm, affinity *v1.Affinity) *v1.Affinity {
	result := affinity
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
