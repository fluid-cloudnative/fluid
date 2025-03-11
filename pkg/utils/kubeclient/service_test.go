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

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Use fake client because of it will be maintained in the long term
// due to https://github.com/kubernetes-sigs/controller-runtime/pull/1101
func TestGetServiceByName(t *testing.T) {
	namespace := "default"
	testServiceInputs := []*v1.Service{{
		ObjectMeta: metav1.ObjectMeta{Name: "svc1"},
		Spec:       v1.ServiceSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "svc2", Annotations: common.GetExpectedFluidAnnotations()},
		Spec:       v1.ServiceSpec{},
	}}

	testServices := []runtime.Object{}

	for _, pv := range testServiceInputs {
		testServices = append(testServices, pv.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testServices...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "service doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
		},
		{
			name: "service is not created by fluid",
			args: args{
				name:      "svc1",
				namespace: namespace,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetServiceByName(client, tt.args.name, tt.args.namespace); err != nil {
				t.Errorf("testcase %v GetServiceByName() got error %v", tt.name, err)
			}
		})
	}

}
