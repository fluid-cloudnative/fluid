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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestPlugin(t *testing.T) {
	var (
		client client.Client
	)
	plugin := NewPlugin(client)
	if plugin.GetName() != Name {
		t.Errorf("GetName expect %v, got %v", Name, plugin.GetName())
	}
}

func TestGetPreferredSchedulingTermWithGlobalMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Test case 1: Global fuse with selector enable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{"test1": "test1"})
	term, _ := getPreferredSchedulingTerm(runtimeInfo, 100)

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

	if !reflect.DeepEqual(*term, expectTerm) {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", expectTerm, term)
	}

	// Test case 2: Global fuse with selector disable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})
	term, _ = getPreferredSchedulingTerm(runtimeInfo, 100)

	if !reflect.DeepEqual(*term, expectTerm) {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", expectTerm, term)
	}

	// Test case 3: runtime Info is nil to handle the error path
	_, err = getPreferredSchedulingTerm(nil, 100)
	if err == nil {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", nil, err)
	}
}

func TestMutateOnlyRequired(t *testing.T) {
	var (
		client   client.Client
		schedPod *corev1.Pod
	)

	plugin := NewPlugin(client)
	runtimeInfo, err := base.BuildRuntimeInfo("test10-ds", "fluid", "alluxio", datav1alpha1.TieredStore{})
	// enable preferred scheduling
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	schedPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset.test10-ds.sched": "required",
			},
		},
	}

	// labeled dataset not exist, no err
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
	if err != nil {
		t.Errorf("expect error is nil")
	}
	// reset injected scheduling terms
	schedPod.Spec = corev1.PodSpec{}

	// labeled dataset exist with nil value, return err
	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": nil})
	if err == nil {
		t.Errorf("expect error is not nil")
	}
	// reset injected scheduling terms
	schedPod.Spec = corev1.PodSpec{}

	_, err = plugin.Mutate(schedPod, map[string]base.RuntimeInfoInterface{"test10-ds": runtimeInfo})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if len(schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
		t.Errorf("fail to mutate pod, not add node affinity")
	}

	if schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		t.Errorf("fail to mutate pod, not need to add preferred scheduling term")
	}
	// reset injected scheduling terms
	schedPod.Spec = corev1.PodSpec{}
}

func TestMutateOnlyPrefer(t *testing.T) {
	var (
		client client.Client
		pod    *corev1.Pod
	)

	plugin := NewPlugin(client)
	if plugin.GetName() != Name {
		t.Errorf("GetName expect %v, got %v", Name, plugin.GetName())
	}

	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": runtimeInfo})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if shouldStop {
		t.Errorf("expect shouldStop as false, but got %v", shouldStop)
	}

	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"pvcName": nil})
	if err == nil {
		t.Errorf("expect error is nil")
	}

}

func TestMutateBothRequiredAndPrefer(t *testing.T) {
	var (
		client   client.Client
		schedPod *corev1.Pod
	)

	plugin := NewPlugin(client)
	runtimeInfo, err := base.BuildRuntimeInfo("test10-ds", "fluid", "alluxio", datav1alpha1.TieredStore{})
	// set global true to enable prefer
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	schedPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"fluid.io/dataset.test10-ds.sched": "required",
				"fluid.io/dataset.no_exist.sched":  "required",
			},
		},
	}
	runtimeInfos := map[string]base.RuntimeInfoInterface{
		"test10-ds":           runtimeInfo,
		"prefer_dataset_name": runtimeInfo,
	}
	_, err = plugin.Mutate(schedPod, runtimeInfos)

	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if len(schedPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != 1 {
		t.Errorf("fail to mutate pod, not add node required scheduling term")
	}

	if len(schedPod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 {
		t.Errorf("fail to mutate pod, not add node preferred scheduling term")
	}

	if len(runtimeInfos) != 2 {
		t.Errorf("mutate should not modify the parameter runtimeInfo")
	}
}

func TestTieredLocality(t *testing.T) {
	tieredConfigMap := &corev1.ConfigMap{
		Data: map[string]string{
			"tieredLocality": "",
		},
	}

	schema := runtime.NewScheme()
	_ = v1.AddToScheme(schema)
	client := fake.NewFakeClientWithScheme(schema, tieredConfigMap)

	runtimeInfo, _ := base.BuildRuntimeInfo("test10-ds", "fluid", "alluxio", datav1alpha1.TieredStore{})
	// set global true to enable prefer
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})

	type args struct {
		plugin       *NodeAffinityWithCache
		pod          *corev1.Pod
		runtimeInfos map[string]base.RuntimeInfoInterface
	}

	tests := []struct {
		name string
		args args
	}{
		{
			args: args{
				plugin: NewPlugin(client),
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"fluid.io/dataset.test10-ds.sched": "required",
							"fluid.io/dataset.no_exist.sched":  "required",
						},
					},
				},
				runtimeInfos: map[string]base.RuntimeInfoInterface{
					"test10-ds":           runtimeInfo,
					"prefer_dataset_name": runtimeInfo,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.args.plugin.Mutate(tt.args.pod, tt.args.runtimeInfos)
			if err != nil {
				t.Errorf("should not have error")
			}
		})
	}
}
