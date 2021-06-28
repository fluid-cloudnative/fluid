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

package requirenodewithfuse

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
)

func TestGetRequiredSchedulingTermWithGlobalMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Test case 1: Global fuse with selector enable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{"test1": "test1"})
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
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})
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

func TestGetRequiredSchedulingTermWithDefaultMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.Tieredstore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfo.SetupFuseDeployMode(false, map[string]string{})
	terms, _ := getRequiredSchedulingTerm(runtimeInfo)

	expectTerms := corev1.NodeSelectorTerm{
		MatchExpressions: []corev1.NodeSelectorRequirement{
			{
				Key:      "fluid.io/s-fluid-test",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"true"},
			},
		},
	}

	if !reflect.DeepEqual(terms, expectTerms) {
		t.Errorf("getRequiredSchedulingTerm failure, want:%v, got:%v", expectTerms, terms)
	}
}

func TestMutate(t *testing.T) {

}
