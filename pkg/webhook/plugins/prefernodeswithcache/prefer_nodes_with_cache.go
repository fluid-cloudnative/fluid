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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods with a mounted dataset.
   If the runtime is in fuse mode, they should prefer nods with the mounted dataset on them.

*/

var setupLog = ctrl.Log.WithName("PreferNodesWithCache")

type PreferNodesWithCache struct {
	client client.Client
}

func NewPlugin(c client.Client) *PreferNodesWithCache {
	return &PreferNodesWithCache{
		client: c,
	}
}

func (p *PreferNodesWithCache) NodePrefer(pod corev1.Pod) (preferredSchedulingTerms []corev1.PreferredSchedulingTerm) {
	pvcNames := kubeclient.GetPVCNamesFromPod(&pod)
	for _, pvcName := range pvcNames {
		setupLog.Info(pvcName)
		isDatasetPVC, err := kubeclient.IsDatasetPVC(p.client, pvcName, pod.Namespace)
		if err != nil {
			setupLog.Error(err, "unable to check pvc, will ignore it", "pvc", pvcName)
			continue
		}
		if isDatasetPVC {
			global, err := utils.IsFuseGlobal(p.client, pvcName, pod.Namespace)
			if err != nil {
				setupLog.Error(err, "unable to get alluxioRuntime, will ignore it", "alluxioRuntime", pvcName)
				continue
			}
			if global {
				preferredSchedulingTerm := getPreferredSchedulingTerm(pvcName, pod.Namespace)
				preferredSchedulingTerms = append(preferredSchedulingTerms, preferredSchedulingTerm)
			}
		}
	}
	return
}

func (p *PreferNodesWithCache) PodPrefer(corev1.Pod) (weightedPodAffinityTerms []corev1.WeightedPodAffinityTerm) {
	return
}

func (p *PreferNodesWithCache) PodNotPrefer(corev1.Pod) (weightedPodAffinityTerms []corev1.WeightedPodAffinityTerm) {
	return
}

func getPreferredSchedulingTerm(name, namespace string) corev1.PreferredSchedulingTerm {
	return corev1.PreferredSchedulingTerm{
		Weight: 50,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      "fluid.io/s-" + namespace + "-" + name,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"true"},
				},
			},
		},
	}
}
