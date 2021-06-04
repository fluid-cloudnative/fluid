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

package kubeclient

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Use fake client because of it will be maintained in the long term
// due to https://github.com/kubernetes-sigs/controller-runtime/pull/1101
func TestIsNamespaceExist(t *testing.T) {

	testNamespaceInputs := []*v1.Namespace{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1"},
		Spec:       v1.NamespaceSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "test2"},
		Spec:       v1.NamespaceSpec{},
	}}

	testNamespaces := []runtime.Object{}

	for _, ns := range testNamespaceInputs {
		testNamespaces = append(testNamespaces, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testNamespaces...)

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "volume doesn't exist",
			args: args{
				name: "notExist",
			},
		},
		{
			name: "volume is not created by fluid",
			args: args{
				name: "test1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsureNamespace(client, tt.args.name); err != nil {
				t.Errorf("testcase %v EnsureNamespace()'s err is %v", tt.name, err)
			}
		})
	}

}
