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

package nodeaffinitywithcache

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const simpleTieredLocality = `
preferred:
- name: fluid.io/node
  weight: 100
required:
- fluid.io/node
`

var testScheme = func() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(s)
	_ = corev1.AddToScheme(s)
	return s
}()

func newTestAlluxioRuntime() *datav1alpha1.AlluxioRuntime {
	return &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio-runtime",
			Namespace: "fluid-test",
		},
	}
}

func TestNewPluginAndGetName(t *testing.T) {
	var cl client.Client
	plugin, err := NewPlugin(cl, "")
	require.NoError(t, err)
	assert.Equal(t, Name, plugin.GetName())
}

func TestGetPreferredSchedulingTerm(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
	require.NoError(t, err)

	runtimeInfo.SetFuseNodeSelector(map[string]string{"test1": "test1"})
	term := getPreferredSchedulingTerm(100, runtimeInfo.GetCommonLabelName())

	expectTerm := corev1.PreferredSchedulingTerm{
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
	assert.Equal(t, expectTerm, term)

	// same result when selector is empty
	runtimeInfo.SetFuseNodeSelector(map[string]string{})
	term = getPreferredSchedulingTerm(100, runtimeInfo.GetCommonLabelName())
	assert.Equal(t, expectTerm, term)
}

func TestMutateOnlyRequired(t *testing.T) {
	alluxioRuntime := newTestAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	schedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset.test10-ds.sched": "required",
			},
		},
	}

	plugin, err := NewPlugin(cl, simpleTieredLocality)
	require.NoError(t, err)

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	// pvcName does not match any sched label — no required injection
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
	require.NoError(t, err)
	schedPod.Spec = corev1.PodSpec{} // reset

	// nil runtimeInfo — pod affinity stays nil
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": nil})
	require.NoError(t, err)
	assert.Nil(t, schedPod.Spec.Affinity)
	schedPod.Spec = corev1.PodSpec{} // reset

	// matching dataset name with valid runtimeInfo — required terms injected
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": runtimeInfo})
	require.NoError(t, err)
	require.NotNil(t, schedPod.Spec.Affinity)
	require.NotNil(t, schedPod.Spec.Affinity.NodeAffinity)
	require.NotNil(t, schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	assert.Len(t, schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 1)
	assert.Nil(t, schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)
}

func TestMutateOnlyPrefer(t *testing.T) {
	alluxioRuntime := newTestAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	plugin, err := NewPlugin(cl, simpleTieredLocality)
	require.NoError(t, err)
	assert.Equal(t, Name, plugin.GetName())

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	// pod has no sched label so runtime goes to preferred path — no error
	shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
	require.NoError(t, err)
	assert.False(t, shouldStop)

	// empty runtimeInfos map — early return
	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
	require.NoError(t, err)

	// nil runtimeInfo — no error
	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
	require.NoError(t, err)
}

func TestMutateBothRequiredAndPrefer(t *testing.T) {
	alluxioRuntime := newTestAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	schedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
				"fluid.io/dataset.no_exist.sched":                    "required",
			},
		},
	}

	plugin, err := NewPlugin(cl, simpleTieredLocality)
	require.NoError(t, err)

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	runtimeInfos := map[string]base.RuntimeInfoInterface{
		alluxioRuntime.Name:   runtimeInfo,
		"prefer_dataset_name": runtimeInfo,
	}
	_, err = plugin.Mutate(schedPod, runtimeInfos)
	require.NoError(t, err)
	require.NotNil(t, schedPod.Spec.Affinity)
	require.NotNil(t, schedPod.Spec.Affinity.NodeAffinity)
	require.NotNil(t, schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	assert.Len(t, schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 1)
	assert.Len(t, schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
	assert.Len(t, runtimeInfos, 2)
}

const customizedTieredLocality = `
preferred:
- name: fluid.io/fuse
  weight: 100
- name: fluid.io/node
  weight: 100
- name: topology.kubernetes.io/rack
  weight: 50
- name: topology.kubernetes.io/zone
  weight: 10
required:
- fluid.io/node
`

func newCacheAlluxioRuntime() *datav1alpha1.AlluxioRuntime {
	return &datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "alluxio-runtime",
			Namespace: "fluid-test",
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "topology.kubernetes.io/rack",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"rack-a"},
								},
								{
									Key:      "topology.kubernetes.io/zone",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"zone-a"},
								},
							},
						},
					},
				},
			},
		},
	}
}

func TestTieredLocalityMutatePodWithDatasetSched(t *testing.T) {
	alluxioRuntime := newCacheAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	plugin, err := NewPlugin(cl, customizedTieredLocality)
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset." + alluxioRuntime.Name + ".sched": "required",
			},
		},
	}
	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
		alluxioRuntime.Name: runtimeInfo,
	})
	require.NoError(t, err)
	require.NotNil(t, pod.Spec.Affinity)
	require.NotNil(t, pod.Spec.Affinity.NodeAffinity)
	require.NotNil(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution)
	assert.Len(t, pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, 1)
}

func TestTieredLocalityMutatePodWithPreferredTerms(t *testing.T) {
	alluxioRuntime := newCacheAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	plugin, err := NewPlugin(cl, customizedTieredLocality)
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
		alluxioRuntime.Name: runtimeInfo,
	})
	require.NoError(t, err)
	require.NotNil(t, pod.Spec.Affinity)
	require.NotNil(t, pod.Spec.Affinity.NodeAffinity)
	assert.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 4)
}

func TestTieredLocalitySkipMutateWhenPodAlreadyHasPreferred(t *testing.T) {
	alluxioRuntime := newCacheAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	plugin, err := NewPlugin(cl, customizedTieredLocality)
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: corev1.PodSpec{
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 100,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "topology.kubernetes.io/rack",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"rack-a"},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
		alluxioRuntime.Name: runtimeInfo,
	})
	require.NoError(t, err)
	assert.Len(t, pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, 1)
}

func TestTieredLocalitySkipMutateWhenPluginArgEmpty(t *testing.T) {
	alluxioRuntime := newCacheAlluxioRuntime()

	cl := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

	runtimeInfo, err := base.BuildRuntimeInfo(alluxioRuntime.Name, alluxioRuntime.Namespace, "alluxio")
	require.NoError(t, err)
	runtimeInfo.SetFuseNodeSelector(map[string]string{})

	plugin, err := NewPlugin(cl, "")
	require.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: corev1.PodSpec{},
	}
	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{
		alluxioRuntime.Name: runtimeInfo,
	})
	require.NoError(t, err)
	assert.Nil(t, pod.Spec.Affinity)
}
