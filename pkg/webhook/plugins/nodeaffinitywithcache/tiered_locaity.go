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

	if pod.Spec.Affinity != nil && pod.Spec.Affinity.NodeAffinity != nil {
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
	}

	for name := range pod.Spec.NodeSelector {
		if _, ok := localityKeys[name]; ok {
			return true
		}
	}
	return false
}
