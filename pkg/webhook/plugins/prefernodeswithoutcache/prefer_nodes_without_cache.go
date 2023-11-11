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

package prefernodeswithoutcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods without a mounted dataset.
   They should prefer nods without cache workers on them.

*/

const Name = "PreferNodesWithoutCache"

type PreferNodesWithoutCache struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *PreferNodesWithoutCache {
	return &PreferNodesWithoutCache{
		client: c,
		name:   Name,
	}
}

func (p *PreferNodesWithoutCache) GetName() string {
	return p.name
}

func (p *PreferNodesWithoutCache) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted dataset, no need to call other plugins
	shouldStop = true

	// if the pod has mounted datasets, should exit and call other plugins
	if len(runtimeInfos) != 0 {
		// err = fmt.Errorf("runtimeInfos for PreferNodesWithoutCache is not empty, %v", runtimeInfos)
		return
	}

	preferredSchedulingTerms := []corev1.PreferredSchedulingTerm{
		getPreferredSchedulingTermForPodWithoutCache(),
	}

	utils.InjectPreferredSchedulingTerms(preferredSchedulingTerms, pod)

	return
}

func getPreferredSchedulingTermForPodWithoutCache() corev1.PreferredSchedulingTerm {
	return corev1.PreferredSchedulingTerm{
		Weight: 100,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      common.GetDatasetNumLabelName(),
					Operator: corev1.NodeSelectorOpDoesNotExist,
				},
			},
		},
	}
}
