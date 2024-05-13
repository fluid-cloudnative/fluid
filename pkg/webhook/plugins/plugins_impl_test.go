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

package plugins

import (
	"math/rand"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/nodeaffinitywithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	corev1 "k8s.io/api/core/v1"
)

func TestPods(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	tieredLocality := `
preferred:
- name: fluid.io/node
  weight: 100
required:
- fluid.io/node
`
	jindoRuntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "default",
		},
	}

	schema := runtime.NewScheme()
	_ = corev1.AddToScheme(schema)
	_ = datav1alpha1.AddToScheme(schema)
	var (
		pod               corev1.Pod
		c                 = fake.NewFakeClientWithScheme(schema, jindoRuntime)
		plugin            api.MutatingHandler
		pluginName        string
		lenNodePrefer     int
		lenNodeRequire    int
		lenPodPrefer      int
		lenPodAntiPrefer  int
		lenPodRequire     int
		lenPodAntiRequire int
	)

	// build slice of RuntimeInfos
	var nilRuntimeInfos map[string]base.RuntimeInfoInterface = map[string]base.RuntimeInfoInterface{}
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "default", "jindo", datav1alpha1.TieredStore{})
	if err != nil {
		t.Error("fail to build runtimeInfo because of err", err)
	}
	runtimeInfo.SetupFuseDeployMode(map[string]string{})
	runtimeInfo.SetDeprecatedNodeLabel(false)
	// runtimeInfos := append(nilRuntimeInfos, runtimeInfo)
	runtimeInfos := map[string]base.RuntimeInfoInterface{"hbase": runtimeInfo}

	// test all plugins for 3 turns
	for i := 0; i < 3; i++ {
		lenNodePrefer = r.Intn(3) + 1
		lenNodeRequire = r.Intn(3) + 1
		lenPodPrefer = r.Intn(3) + 1
		lenPodAntiPrefer = r.Intn(3) + 1
		lenPodRequire = r.Intn(3) + 1
		lenPodAntiRequire = r.Intn(3) + 1

		// build affinity of pod
		pod.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: make([]corev1.NodeSelectorTerm, lenNodeRequire),
				},
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.PreferredSchedulingTerm, lenNodePrefer),
			},
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  make([]corev1.PodAffinityTerm, lenPodRequire),
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.WeightedPodAffinityTerm, lenPodPrefer),
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  make([]corev1.PodAffinityTerm, lenPodAntiRequire),
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.WeightedPodAffinityTerm, lenPodAntiPrefer),
			},
		}

		// test of plugin preferNodesWithoutCache
		plugin, err = prefernodeswithoutcache.NewPlugin(c, "")
		if err != nil {
			t.Error("new plugin occurs error", err)
		}
		pluginName = plugin.GetName()
		_, err = plugin.Mutate(&pod, runtimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer {
			t.Errorf("the plugin %v should exit and call other plugins if the pod has mounted datasets", pluginName)
		}

		_, err = plugin.Mutate(&pod, nilRuntimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+1 {
			t.Errorf("the plugin %v inject wrong terms when the pod has no mounted datasets", pluginName)
		}

		// test of plugin preferNodesWithCache
		plugin, err = nodeaffinitywithcache.NewPlugin(c, tieredLocality)
		if err != nil {
			t.Error("new plugin occurs error", err)
		}
		pluginName = plugin.GetName()
		_, err = plugin.Mutate(&pod, nilRuntimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+1 {
			t.Errorf("the plugin %v should exit and call other plugins if the pod has no mounted datasets", pluginName)
		}

		_, err = plugin.Mutate(&pod, runtimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}
		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+2 {
			t.Errorf("the plugin %v inject wrong terms when the pod has mounted datasets", pluginName)
		}

	}

}

func TestGetRegistryHandler(t *testing.T) {
	type want struct {
		podWithDatasetHandlerNames []string
		nodeWithCacheArgs          *nodeaffinitywithcache.TieredLocality

		podWithoutDatasetHandlerSize           int
		serverlessPodWithDatasetHandlerSize    int
		serverlessPodWithoutDatasetHandlerSize int
	}

	tests := []struct {
		name          string
		want          want
		newPluginErr  bool
		pluginProfile string
	}{
		{
			name: "existing correct configmap",
			pluginProfile: `
plugins:
  serverful:
    withDataset:
    - RequireNodeWithFuse
    - NodeAffinityWithCache
    - MountPropagationInjector
    withoutDataset:
    - PreferNodesWithoutCache
  serverless:
    withDataset:
    - FuseSidecar
    withoutDataset:
    - PreferNodesWithoutCache
pluginConfig:
  - name: NodeAffinityWithCache
    args: |
      preferred:
      # fluid existed node affinity, the name can not be modified.
      - name: fluid.io/node
        weight: 100
      # runtime worker's zone label name, can be changed according to k8s environment.
      - name: topology.kubernetes.io/zone
        weight: 50
      # runtime worker's region label name, can be changed according to k8s environment.
      - name: topology.kubernetes.io/region
        weight: 10
      required:
      - fluid.io/node
`,
			want: want{
				podWithDatasetHandlerNames: []string{
					"RequireNodeWithFuse", "NodeAffinityWithCache", "MountPropagationInjector",
				},
				nodeWithCacheArgs: &nodeaffinitywithcache.TieredLocality{
					Preferred: []nodeaffinitywithcache.Preferred{
						{
							Name:   "fluid.io/node",
							Weight: 100,
						},
						{
							Name:   "topology.kubernetes.io/zone",
							Weight: 50,
						},
						{
							Name:   "topology.kubernetes.io/region",
							Weight: 10,
						},
					},
					Required: []string{
						"fluid.io/node",
					},
				},
				podWithoutDatasetHandlerSize:           1,
				serverlessPodWithDatasetHandlerSize:    1,
				serverlessPodWithoutDatasetHandlerSize: 1,
			},
			newPluginErr: false,
		},
		{
			name: "existing wrong configmap",
			pluginProfile: `
plugins:
  serverful:
    withDataset:
    - NotExistPlugin
    - RequireNodeWithFuse
    - NodeAffinityWithCache
    - MountPropagationInjector
    withoutDataset:
    - PreferNodesWithoutCache
  serverless:
    withDataset:
    - FuseSidecar
    withoutDataset:
    - PreferNodesWithoutCache
pluginConfig:
  - name: NodeAffinityWithCache
    args: |
      preferred:
      # fluid existed node affinity, the name can not be modified.
      - name: fluid.io/node
        weight: 100
      # runtime worker's zone label name, can be changed according to k8s environment.
      - name: topology.kubernetes.io/zone
        weight: 50
      # runtime worker's region label name, can be changed according to k8s environment.
      - name: topology.kubernetes.io/region
        weight: 10
      required:
      - fluid.io/node
`,
			want:         want{},
			newPluginErr: true,
		},
	}

	schema := runtime.NewScheme()
	_ = corev1.AddToScheme(schema)
	_ = datav1alpha1.AddToScheme(schema)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var clientWithScheme client.Client

			pluginProfile := tt.pluginProfile
			mockReadFile := func(content string) ([]byte, error) {
				return []byte(pluginProfile), nil
			}
			patch := gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
			defer patch.Reset()

			err := RegisterMutatingHandlers(clientWithScheme)

			if (err != nil) != tt.newPluginErr {
				t.Errorf("new plugin error = %v, wantErr %v", err, tt.newPluginErr)
			}

			if tt.newPluginErr {
				return
			}

			plugins := GetRegistryHandler()
			got := want{
				podWithoutDatasetHandlerSize:           len(plugins.GetPodWithoutDatasetHandler()),
				serverlessPodWithoutDatasetHandlerSize: len(plugins.GetServerlessPodWithoutDatasetHandler()),
				serverlessPodWithDatasetHandlerSize:    len(plugins.GetServerlessPodWithDatasetHandler()),
			}
			for _, handler := range plugins.GetPodWithDatasetHandler() {
				got.podWithDatasetHandlerNames = append(got.podWithDatasetHandlerNames, handler.GetName())
				if handler.GetName() == nodeaffinitywithcache.Name {
					cacheHandler := handler.(*nodeaffinitywithcache.NodeAffinityWithCache)
					got.nodeWithCacheArgs = cacheHandler.GetTieredLocality()
				}
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRegistryHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
