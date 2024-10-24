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

package requirenodewithfuse

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetRequiredSchedulingTermWithGlobalMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Test case 1: Global fuse with selector enable
	runtimeInfo.SetFuseNodeSelector(map[string]string{"test1": "test1"})
	terms, _ := getRequiredSchedulingTerm(runtimeInfo)

	expectTerms := corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      "test1",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"test1"},
			},
		},
	}

	if !reflect.DeepEqual(terms, expectTerms) {
		t.Errorf("getRequiredSchedulingTerm failure, want:%v, got:%v", expectTerms, terms)
	}

	// Test case 2: Global fuse with selector disable
	runtimeInfo.SetFuseNodeSelector(map[string]string{})
	terms, _ = getRequiredSchedulingTerm(runtimeInfo)
	expectTerms = corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{}}

	if !reflect.DeepEqual(terms, expectTerms) {
		t.Errorf("getRequiredSchedulingTerm failure, want:%v, got:%v", expectTerms, terms)
	}

	// Test case 3: runtime Info is nil to handle the error path
	_, err = getRequiredSchedulingTerm(nil)
	if err == nil {
		t.Errorf("getRequiredSchedulingTerm failure, want:%v, got:%v", nil, err)
	}
}

func TestMutate(t *testing.T) {
	var (
		client client.Client
		pod    *corev1.Pod
	)

	plugin, err := NewPlugin(client, "")
	if err != nil {
		t.Error("new plugin occurs error", err)
	}
	if plugin.GetName() != Name {
		t.Errorf("GetName expect %v, got %v", Name, plugin.GetName())
	}

	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio")
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
		t.Errorf("expect error is not nil")
	}

}
