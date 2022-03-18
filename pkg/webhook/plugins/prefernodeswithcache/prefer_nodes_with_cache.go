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

package prefernodeswithcache

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a mounted dataset.
   If the runtime is in fuse mode, they should prefer nodes with the mounted dataset on them.

*/

const NAME = "PreferNodesWithCache"

type PreferNodesWithCache struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *PreferNodesWithCache {
	return &PreferNodesWithCache{
		client: c,
		name:   NAME,
	}
}

func (p *PreferNodesWithCache) GetName() string {
	return p.name
}

func (p *PreferNodesWithCache) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}
	var preferredSchedulingTerms []corev1.PreferredSchedulingTerm
	for _, runtimeInfo := range runtimeInfos {

		preferredSchedulingTerm, err := getPreferredSchedulingTerm(runtimeInfo)
		if err != nil {
			return shouldStop, err
		}
		if preferredSchedulingTerm != nil {
			preferredSchedulingTerms = append(preferredSchedulingTerms, *preferredSchedulingTerm)
		}

	}
	utils.InjectPreferredSchedulingTerms(preferredSchedulingTerms, pod)

	return
}

func getPreferredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface) (preferredSchedulingTerm *corev1.PreferredSchedulingTerm, err error) {
	preferredSchedulingTerm = nil

	if runtimeInfo == nil {
		err = fmt.Errorf("RuntimeInfo is nil")
		return
	}

	isGlobalMode, _ := runtimeInfo.GetFuseDeployMode()
	if isGlobalMode {
		preferredSchedulingTerm = &corev1.PreferredSchedulingTerm{
			Weight: 100,
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      runtimeInfo.GetCommonLabelName(),
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		}
	}

	return
}
