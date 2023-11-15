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

package requirenodewithfuse

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"

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

func NewPlugin(c client.Client, args string) api.MutatingHandler {
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
