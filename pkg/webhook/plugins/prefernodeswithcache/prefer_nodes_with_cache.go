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

package prefernodeswithcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
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

func (p *PreferNodesWithCache) InjectAffinity(pod *corev1.Pod, runtimeInfos []base.RuntimeInfoInterface) (finish bool) {
	if len(runtimeInfos) == 0 {
		return
	}
	var preferredSchedulingTerms []corev1.PreferredSchedulingTerm
	for _, runtimeInfo := range runtimeInfos {
		// if runtime in global mode, inject a new PreferredSchedulingTerm
		global, _ := runtimeInfo.GetFuseDeployMode()
		if global {
			preferredSchedulingTerm := getPreferredSchedulingTerm(runtimeInfo)
			preferredSchedulingTerms = append(preferredSchedulingTerms, preferredSchedulingTerm)
		}
	}
	if len(preferredSchedulingTerms) == 0 {
		return
	}
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
				append(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					preferredSchedulingTerms...)
		} else {
			pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTerms,
			}
		}
	} else {
		pod.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: preferredSchedulingTerms,
			},
		}
	}
	return
}

func getPreferredSchedulingTerm(runtimeInfo base.RuntimeInfoInterface) corev1.PreferredSchedulingTerm {
	return corev1.PreferredSchedulingTerm{
		Weight: 50,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      runtimeInfo.GetCommonLabelname(),
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	}
}
