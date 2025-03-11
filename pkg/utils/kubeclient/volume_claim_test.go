/*
Copyright 2023 The Fluid Authors.

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
func TestIsPersistentVolumeClaimExist(t *testing.T) {

	namespace := "default"
	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{Name: "notCreatedByFluid",
			Namespace: namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "createdByFluid",
			Annotations: common.GetExpectedFluidAnnotations(),
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
				annotations: common.GetExpectedFluidAnnotations(),
			},
			want: true,
		}, {
			name: "volume is not created by fluid 2",
			args: args{
				name: "notCreatedByFluid2",
				annotations: map[string]string{
					"test1": "test1",
				},
			},
			want: false,
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
			Annotations: common.GetExpectedFluidAnnotations(),
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

func TestGetPvcMountNodes(t *testing.T) {
	namespace := "test"
	volumeName1 := "found"
	volumeName2 := "found1"
	testPodInputs := []*v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "found"},
		Spec:       v1.PodSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "bbb", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName1,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName1,
							ReadOnly:  true,
						}},
				},
			},
			NodeName: "node1",
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "ccc", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName1,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName1,
							ReadOnly:  true,
						}},
				},
			},
			NodeName: "node2",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "ddd", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName1,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName1,
							ReadOnly:  true,
						}},
				},
			},
			NodeName: "node3",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "eee", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName2,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName2,
							ReadOnly:  true,
						}},
				},
			},
			NodeName: "node4",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "fff", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName2,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName2,
							ReadOnly:  true,
						}},
				},
			},
			NodeName: "",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "hhh", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName2,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName1,
							ReadOnly:  true,
						}},
				},
			},
			NodeName: "node3",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}}

	testPods := []runtime.Object{}

	for _, pod := range testPodInputs {
		testPods = append(testPods, pod.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPods...)
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name   string
		args   args
		length int
	}{
		{
			name: "node list empty",
			args: args{
				name:      "not found",
				namespace: namespace,
			},
			length: 0,
		},
		{
			name: "node list is 1",
			args: args{
				name:      volumeName2,
				namespace: namespace,
			},
			length: 1,
		}, {
			name: "node list is 2",
			args: args{
				name:      volumeName1,
				namespace: namespace,
			},
			length: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if pvcMountNodes, _ := GetPvcMountNodes(client, tt.args.name, tt.args.namespace); len(pvcMountNodes) != tt.length {
				t.Errorf("testcase %v GetPvcMountPods() = %v, want %v", tt.name, pvcMountNodes, tt.length)
			}
		})
	}
}

func TestRemoveProtectionFinalizer(t *testing.T) {
	namespace := "default"
	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{Name: "hasNoFinalizer",
			Namespace: namespace},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "hasNoFinalizer",
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "hasFinalizer",
			Annotations: common.GetExpectedFluidAnnotations(),
			Namespace:   namespace,
			Finalizers:  []string{persistentVolumeClaimProtectionFinalizerName}},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "hasFinalizer",
		},
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
		args      args
		wantError bool
	}{
		{
			name: "volumeClaim doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			wantError: true,
		},
		{
			name: "volumeClaim is not created by fluid",
			args: args{
				name:      "notCreatedByFluid",
				namespace: namespace,
			},
			wantError: true,
		},
		{
			name: "volumeClaim is created by fluid",
			args: args{
				name:      "hasNoFinalizer",
				namespace: namespace,
			},
			wantError: false,
		}, {
			name: "volumeClaim is not created by fluid 2",
			args: args{
				name:      "hasFinalizer",
				namespace: namespace,
			},
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RemoveProtectionFinalizer(client, tt.args.name, tt.args.namespace)
			got := err != nil

			if got != tt.wantError {
				t.Errorf("testcase %v RemoveProtectionFinalizer() for %v in %v = %v, err = %v", tt.name,
					tt.args.name,
					tt.args.namespace,
					got,
					err)
			}

		})
	}

}
