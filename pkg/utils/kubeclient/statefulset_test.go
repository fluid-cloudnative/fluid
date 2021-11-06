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

package kubeclient

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetStatefulSet(t *testing.T) {
	namespace := "default"
	testStsInputs := []*appsv1.StatefulSet{{
		ObjectMeta: metav1.ObjectMeta{Name: "exist",
			Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{},
	}}

	testStatefulSets := []runtime.Object{}

	for _, sts := range testStsInputs {
		testStatefulSets = append(testStatefulSets, sts.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testStatefulSets...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name     string
		args     args
		want     bool
		hasError bool
	}{
		{
			name: "statefulset doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want:     false,
			hasError: true,
		}, {
			name: "statefulset exists",
			args: args{
				name:      "exist",
				namespace: namespace,
			},
			want:     true,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStatefulSet(client, tt.args.name, tt.args.namespace)

			if tt.hasError {
				if err == nil {
					t.Errorf("testcase %v GetStatefulSet()  expects there is error, but no error", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("testcase %v GetStatefulSet()  expects there is not error, but got error %v", tt.name, err)
				}
			}

			if tt.want != (tt.args.name == got.Name) {
				t.Errorf("testcase %v GetStatefulSet()  got statefulset name %v, the want name of statefulset is %v", tt.name, got.Name, tt.args.name)
			}

			// t.Errorf("testcase %v IsPersistentVolumeClaimExist() = %v, want %v", tt.name, got, tt.want)

		})
	}
}
