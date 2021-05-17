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

package plugins

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestPods(t *testing.T) {
	pod := corev1.Pod{}
	var c client.Client

	var plugin AffinityInterface

	var nilRuntimeInfos []base.RuntimeInfoInterface
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "default", "jindo", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Error("fail to build runtimeInfo because of err", err)
	}
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})
	runtimeInfo.SetDeprecatedNodeLabel(false)
	runtimeInfos := append(nilRuntimeInfos, runtimeInfo)
	runtimeInfo, err = base.BuildRuntimeInfo("spark", "default", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Error("fail to build runtimeInfo because of err", err)
	}
	runtimeInfo.SetupFuseDeployMode(false, map[string]string{})
	runtimeInfo.SetDeprecatedNodeLabel(false)
	runtimeInfos = append(runtimeInfos, runtimeInfo)

	pod.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchFields: []corev1.NodeSelectorRequirement{},
					},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
				{
					Weight: 50,
					Preference: corev1.NodeSelectorTerm{
						MatchFields: []corev1.NodeSelectorRequirement{},
					},
				},
			},
		},
		PodAffinity: &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{},
					},
				},
			},
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{},
				},
			},
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 50,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{},
					},
				},
			},
		},
	}

	// test of plugin preferNodesWithoutCache
	plugin = prefernodeswithoutcache.NewPlugin(c)
	pluginName := plugin.GetName()
	plugin.InjectAffinity(&pod, runtimeInfos)
	if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 {
		t.Errorf("the plugin %v should only inject into PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
	}
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.Errorf("the plugin %v should only inject into PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
	} else {
		if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
			t.Errorf("the plugin %v should only inject into PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
	}
	if len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 {
		t.Errorf("the plugin %v should not inject PodAntiAffinity prefer of pod with dataset mounted", pluginName)
	}

	plugin.InjectAffinity(&pod, nilRuntimeInfos)
	if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 {
		t.Errorf("the plugin %v should only inject into PodAntiAffinity prefer", pluginName)
	}
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.Errorf("the plugin %v should only inject into PodAntiAffinity prefer", pluginName)
	} else {
		if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
			t.Errorf("the plugin %v should only inject into PodAntiAffinity prefer", pluginName)
		}
	}
	if len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 2 {
		t.Errorf("the plugin %v should inject PodAntiAffinity prefer of pod without dataset mounted", pluginName)
	}

	// test of plugin preferNodesWithCache
	plugin = prefernodeswithcache.NewPlugin(c)
	pluginName = plugin.GetName()
	plugin.InjectAffinity(&pod, nilRuntimeInfos)
	if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 2 {
		t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
	}
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
	} else {
		if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
	}
	if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 {
		t.Errorf("the plugin %v should not inject NodeAffinity prefer of pod without dataset mounted", pluginName)
	}

	plugin.InjectAffinity(&pod, runtimeInfos)
	if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 1 ||
		len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 2 {
		t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
	}
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
	} else {
		if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
	}
	if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 2 {
		t.Errorf("the plugin %v should not inject NodeAffinity prefer of pod without dataset mounted", pluginName)
	}

}
