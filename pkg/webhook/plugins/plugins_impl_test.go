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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/nodeaffinitywithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Plugins Implementation", func() {
	var (
		tieredLocality = `
preferred:
- name: fluid.io/node
  weight: 100
required:
- fluid.io/node
`
		jindoRuntime *datav1alpha1.JindoRuntime
		schema       *runtime.Scheme
		c            client.Client
	)

	BeforeEach(func() {
		jindoRuntime = &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "default",
			},
		}
		schema = runtime.NewScheme()
		Expect(corev1.AddToScheme(schema)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(schema)).To(Succeed())
		c = fake.NewFakeClientWithScheme(schema, jindoRuntime)
	})

	It("should test all plugins for 3 turns", func() {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		var pod corev1.Pod

		var nilRuntimeInfos map[string]base.RuntimeInfoInterface = map[string]base.RuntimeInfoInterface{}
		runtimeInfo, err := base.BuildRuntimeInfo("hbase", "default", "jindo")
		Expect(err).NotTo(HaveOccurred())
		runtimeInfo.SetFuseNodeSelector(map[string]string{})
		runtimeInfos := map[string]base.RuntimeInfoInterface{"hbase": runtimeInfo}

		for i := 0; i < 3; i++ {
			lenNodePrefer := r.Intn(3) + 1
			lenNodeRequire := r.Intn(3) + 1
			lenPodPrefer := r.Intn(3) + 1
			lenPodAntiPrefer := r.Intn(3) + 1
			lenPodRequire := r.Intn(3) + 1
			lenPodAntiRequire := r.Intn(3) + 1

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

			// preferNodesWithoutCache
			plugin, err := prefernodeswithoutcache.NewPlugin(c, "")
			Expect(err).NotTo(HaveOccurred())
			_, err = plugin.Mutate(&pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenPodPrefer))
			Expect(len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenPodRequire))
			Expect(len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenPodAntiRequire))
			Expect(len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenPodAntiPrefer))
			Expect(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).NotTo(BeNil())
			Expect(len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms)).To(Equal(lenNodeRequire))
			Expect(len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenNodePrefer))

			_, err = plugin.Mutate(&pod, nilRuntimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenNodePrefer + 1))

			// nodeaffinitywithcache
			plugin, err = nodeaffinitywithcache.NewPlugin(c, tieredLocality)
			Expect(err).NotTo(HaveOccurred())
			_, err = plugin.Mutate(&pod, nilRuntimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenNodePrefer + 1))

			_, err = plugin.Mutate(&pod, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(lenNodePrefer + 2))
		}
	})

	DescribeTable("GetRegistryHandler", func(tt struct {
		name          string
		want          want
		newPluginErr  bool
		pluginProfile string
	}) {
		schema := runtime.NewScheme()
		Expect(corev1.AddToScheme(schema)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(schema)).To(Succeed())
		var clientWithScheme client.Client

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(tt.pluginProfile), nil
		}
		patch := gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
		defer patch.Reset()

		err := RegisterMutatingHandlers(clientWithScheme)
		if tt.newPluginErr {
			Expect(err).To(HaveOccurred())
			return
		} else {
			Expect(err).NotTo(HaveOccurred())
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
		Expect(reflect.DeepEqual(got, tt.want)).To(BeTrue())
	},
		Entry("existing correct configmap", struct {
			name          string
			want          want
			newPluginErr  bool
			pluginProfile string
		}{
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
		}),
		Entry("existing wrong configmap", struct {
			name          string
			want          want
			newPluginErr  bool
			pluginProfile string
		}{
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
		}),
	)
})

type want struct {
	podWithDatasetHandlerNames []string
	nodeWithCacheArgs          *nodeaffinitywithcache.TieredLocality

	podWithoutDatasetHandlerSize           int
	serverlessPodWithDatasetHandlerSize    int
	serverlessPodWithoutDatasetHandlerSize int
}
