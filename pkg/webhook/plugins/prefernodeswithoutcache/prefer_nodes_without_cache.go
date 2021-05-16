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

package prefernodeswithoutcache

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods without a mounted dataset.
   They should prefer nods without cache workers on them.

*/

const NAME = "PreferNodesWithoutCache"

type PreferNodesWithoutCache struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *PreferNodesWithoutCache {
	return &PreferNodesWithoutCache{
		client: c,
		name:   NAME,
	}
}

var weightedPodAffinityTerm = corev1.WeightedPodAffinityTerm{
	Weight: 50,
	PodAffinityTerm: corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"alluxio-worker", "jindofs-worker"},
				},
			},
		},
		TopologyKey: "kubernetes.io/hostname",
	},
}

func (p *PreferNodesWithoutCache) GetName() string {
	return p.name
}

func (p *PreferNodesWithoutCache) InjectAffinity(pod *corev1.Pod, runtimeInfos []base.RuntimeInfoInterface) (finish bool) {
	if len(runtimeInfos) != 0 {
		return
	}

	// if the pod has no mounted dataset, no need to call other plugins
	finish = true
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.PodAntiAffinity != nil {
			pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution =
				append(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
					weightedPodAffinityTerm)
		} else {
			pod.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					weightedPodAffinityTerm,
				},
			}
		}
	} else {
		pod.Spec.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					weightedPodAffinityTerm,
				},
			},
		}
	}
	return
}
