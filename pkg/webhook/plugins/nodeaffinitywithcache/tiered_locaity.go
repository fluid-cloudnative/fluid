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

package nodeaffinitywithcache

import (
	corev1 "k8s.io/api/core/v1"
)

type Preferred struct {
	Name   string `yaml:"name"`
	Weight int32  `yaml:"weight"`
}

type TieredLocality struct {
	Preferred []Preferred `yaml:"preferred"`
	Required  []string    `yaml:"required"`
}

func (t *TieredLocality) getPreferredAsMap() map[string]int32 {
	localityMap := map[string]int32{}
	for _, preferred := range t.Preferred {
		localityMap[preferred.Name] = preferred.Weight
	}
	return localityMap
}

func (t *TieredLocality) getTieredLocalityNames() (names map[string]bool) {
	names = map[string]bool{}
	for _, preferred := range t.Preferred {
		names[preferred.Name] = true
	}
	for _, required := range t.Required {
		names[required] = true
	}
	return
}

// hasRepeatedLocality whether pod specified the same label as tiered locality
func (t *TieredLocality) hasRepeatedLocality(pod *corev1.Pod) bool {
	localityKeys := t.getTieredLocalityNames()

	if pod.Spec.Affinity == nil && pod.Spec.Affinity.NodeAffinity == nil && pod.Spec.NodeSelector == nil {
		return false
	}

	for _, term := range pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution {
		for _, expression := range term.Preference.MatchExpressions {
			if _, ok := localityKeys[expression.Key]; ok {
				return true
			}
		}
	}

	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		return false
	}
	for _, terms := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
		for _, expression := range terms.MatchExpressions {
			if _, ok := localityKeys[expression.Key]; ok {
				return true
			}
		}
	}

	for name := range pod.Spec.NodeSelector {
		if _, ok := localityKeys[name]; ok {
			return true
		}
	}
	return false
}
