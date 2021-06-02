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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Use fake client because of it will be maintained in the long term
// due to https://github.com/kubernetes-sigs/controller-runtime/pull/1101
func TestIsPersistentVolumeClaimExist(t *testing.T) {

	namespace := "default"
	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{Name: "notCreatedByFluid",
			Namespace: namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "createdByFluid",
			Annotations: common.ExpectedFluidAnnotations,
			Namespace:   namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}}

	testPVCs := []runtime.Object{}

	for _, pvc := range testPVCInputs {
		testPVCs = append(testPVCs, pvc.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPVCs...)

	type args struct {
		name        string
		namespace   string
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "volume doesn't exist",
			args: args{
				name:        "notExist",
				namespace:   namespace,
				annotations: map[string]string{},
			},
			want: false,
		},
		{
			name: "volume is not created by fluid",
			args: args{
				name:        "notCreatedByFluid",
				namespace:   namespace,
				annotations: map[string]string{},
			},
			want: false,
		},
		{
			name: "volume is created by fluid",
			args: args{
				name:        "createdByFluid",
				namespace:   namespace,
				annotations: common.ExpectedFluidAnnotations,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := IsPersistentVolumeClaimExist(client, tt.args.name, tt.args.namespace, tt.args.annotations); got != tt.want {
				t.Errorf("testcase %v IsPersistentVolumeClaimExist() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

}

func TestDeletePersistentVolumeClaim(t *testing.T) {
	namespace := "default"
	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{Name: "aaa",
			Namespace: namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "bbb",
			Annotations: common.ExpectedFluidAnnotations,
			Namespace:   namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}}

	testPVCs := []runtime.Object{}

	for _, pvc := range testPVCInputs {
		testPVCs = append(testPVCs, pvc.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPVCs...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		namespace string
		args      args
		err       error
	}{
		{
			name: "volume doesn't exist",
			args: args{
				name:      "notfound",
				namespace: namespace,
			},
			err: nil,
		},
		{
			name: "volume exists",
			args: args{
				name:      "found",
				namespace: namespace,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeletePersistentVolumeClaim(client, tt.args.name, tt.args.namespace); err != tt.err {
				t.Errorf("testcase %v DeletePersistentVolumeClaim() = %v, want %v", tt.name, err, tt.err)
			}
		})
	}

}
