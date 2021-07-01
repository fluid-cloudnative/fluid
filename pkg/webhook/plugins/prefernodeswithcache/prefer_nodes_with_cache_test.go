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
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetPreferredSchedulingTermWithGlobalMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Test case 1: Global fuse with selector enable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{"test1": "test1"})
	term, _ := getPreferredSchedulingTerm(runtimeInfo)

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
	term, _ = getPreferredSchedulingTerm(runtimeInfo)

	if !reflect.DeepEqual(*term, expectTerm) {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", expectTerm, term)
	}

	// Test case 3: runtime Info is nil to handle the error path
	_, err = getPreferredSchedulingTerm(nil)
	if err == nil {
		t.Errorf("getPreferredSchedulingTerm failure, want:%v, got:%v", nil, err)
	}
}

func TestGetPreferredSchedulingTermWithDefaultMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfo.SetupFuseDeployMode(false, map[string]string{})
	term, _ := getPreferredSchedulingTerm(runtimeInfo)

	if term != nil {
		t.Errorf("getPreferredSchedulingTerm failure, want:nil, got:%v", term)
	}
}

func TestMutate(t *testing.T) {
	var (
		client client.Client
		pod    *corev1.Pod
	)

	plugin := NewPlugin(client)
	if plugin.GetName() != NAME {
		t.Errorf("GetName expect %v, got %v", NAME, plugin.GetName())
	}

	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	shouldStop, err := plugin.Mutate(pod, []base.RuntimeInfoInterface{runtimeInfo})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if shouldStop {
		t.Errorf("expect shouldStop as false, but got %v", shouldStop)
	}

	_, err = plugin.Mutate(pod, []base.RuntimeInfoInterface{})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	_, err = plugin.Mutate(pod, []base.RuntimeInfoInterface{nil})
	if err == nil {
		t.Errorf("expect error is not nil")
	}

}
