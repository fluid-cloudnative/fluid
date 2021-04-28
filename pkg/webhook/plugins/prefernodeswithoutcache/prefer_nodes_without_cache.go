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
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods without a mounted dataset.
   They should prefer nods without cache workers on them.

*/

var setupLog = ctrl.Log.WithName("PreferNodesWithoutCache")

type PreferNodesWithoutCache struct {
	client client.Client
}

func NewPlugin(c client.Client) *PreferNodesWithoutCache {
	return &PreferNodesWithoutCache{
		client: c,
	}
}

var Plugin = PreferNodesWithoutCache{}

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

func (p *PreferNodesWithoutCache) NodePrefer(corev1.Pod) (preferredSchedulingTerms []corev1.PreferredSchedulingTerm) {
	return
}

func (p *PreferNodesWithoutCache) PodPrefer(corev1.Pod) (weightedPodAffinityTerms []corev1.WeightedPodAffinityTerm) {
	return
}

func (p *PreferNodesWithoutCache) PodNotPrefer(pod corev1.Pod) (weightedPodAffinityTerms []corev1.WeightedPodAffinityTerm) {
	pvcNames := kubeclient.GetPVCNamesFromPod(&pod)

	datasetMountPod := false
	for _, pvcName := range pvcNames {
		find, err := kubeclient.IsDatasetPVC(p.client, pvcName, pod.Namespace)
		if err != nil {
			setupLog.Error(err, "unable to get pvc")
			continue
		}
		if find {
			datasetMountPod = true
		}
	}
	if !datasetMountPod {
		weightedPodAffinityTerms = append(weightedPodAffinityTerms, weightedPodAffinityTerm)
	}
	return
}
