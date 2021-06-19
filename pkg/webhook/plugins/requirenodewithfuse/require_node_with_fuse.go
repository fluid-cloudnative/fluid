/*

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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a  dataset.
   They should require nods with fuse.
*/

const NAME = "RequireNodeWithFuse"

type RequireNodeWithFuse struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *RequireNodeWithFuse {
	return &RequireNodeWithFuse{
		client: c,
		name:   NAME,
	}
}

func (p *RequireNodeWithFuse) GetName() string {
	return p.name
}

func (p *RequireNodeWithFuse) Mutate(pod *corev1.Pod, runtimeInfos []base.RuntimeInfoInterface) (shouldStop bool) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}

	requiredSchedulingTerms := []corev1.NodeSelectorTerm{}

	for _, runtime := range runtimeInfos {
		requiredSchedulingTerms = append(requiredSchedulingTerms, getRequiredSchedulingTerm(runtime))
	}

	utils.InjectNodeSelectorTerms(requiredSchedulingTerms, pod)

	return
}

func getRequiredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface) corev1.NodeSelectorTerm {
	return corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      runtimeInfo.GetCommonLabelName(),
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"true"},
			},
		},
	}
}
