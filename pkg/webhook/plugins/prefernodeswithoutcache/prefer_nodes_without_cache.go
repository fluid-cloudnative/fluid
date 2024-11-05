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

package prefernodeswithoutcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
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

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	return &PreferNodesWithoutCache{
		client: c,
		name:   Name,
	}, nil
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
