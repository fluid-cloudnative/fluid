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

package requirenodewithfuse

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a  dataset.
   They should require nods with fuse.
*/

const Name = "RequireNodeWithFuse"

type RequireNodeWithFuse struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *RequireNodeWithFuse {
	return &RequireNodeWithFuse{
		client: c,
		name:   Name,
	}
}

func (p *RequireNodeWithFuse) GetName() string {
	return p.name
}

// Mutate mutates the pod based on runtimeInfo, this action shouldn't stop other handler
func (p *RequireNodeWithFuse) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}

	requiredSchedulingTerms := []corev1.NodeSelectorTerm{}

	for _, runtime := range runtimeInfos {
		term, err := getRequiredSchedulingTerm(runtime)
		if err != nil {
			return true, fmt.Errorf("should stop mutating pod %s in namespace %s due to %v",
				pod.Name,
				pod.Namespace,
				err)
		}

		if len(term.MatchExpressions) > 0 {
			requiredSchedulingTerms = append(requiredSchedulingTerms, term)
		}
	}

	if len(requiredSchedulingTerms) > 0 {
		utils.InjectNodeSelectorTerms(requiredSchedulingTerms, pod)
	}

	return
}

func getRequiredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface) (requiredSchedulingTerm corev1.NodeSelectorTerm, err error) {
	requiredSchedulingTerm = corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{},
	}

	if runtimeInfo == nil {
		err = fmt.Errorf("RuntimeInfo is nil")
		return
	}

	isGlobalMode, selectors := runtimeInfo.GetFuseDeployMode()
	if isGlobalMode {
		for key, value := range selectors {
			requiredSchedulingTerm.MatchExpressions = append(requiredSchedulingTerm.MatchExpressions, corev1.NodeSelectorRequirement{
				Key:      key,
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{value},
			})
		}
	} else {
		requiredSchedulingTerm = corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      runtimeInfo.GetCommonLabelName(),
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		}
	}

	return
}
