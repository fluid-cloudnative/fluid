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
func TestIsPersistentVolumeExist(t *testing.T) {

	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{Name: "notCreatedByFluid"},
		Spec:       v1.PersistentVolumeSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "createdByFluid", Annotations: common.ExpectedFluidAnnotations},
		Spec:       v1.PersistentVolumeSpec{},
	}}

	testPVs := []runtime.Object{}

	for _, pv := range testPVInputs {
		testPVs = append(testPVs, pv.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)

	type args struct {
		name        string
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
				annotations: map[string]string{},
			},
			want: false,
		},
		{
			name: "volume is not created by fluid",
			args: args{
				name:        "notCreatedByFluid",
				annotations: map[string]string{},
			},
			want: false,
		},
		{
			name: "volume is created by fluid",
			args: args{
				name:        "createdByFluid",
				annotations: common.ExpectedFluidAnnotations,
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
			if got, _ := IsPersistentVolumeExist(client, tt.args.name, tt.args.annotations); got != tt.want {
				t.Errorf("testcase %v IsPersistentVolumeExist() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

}

func TestDeletePersistentVolume(t *testing.T) {
	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{Name: "found"},
		Spec:       v1.PersistentVolumeSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "bbb", Annotations: common.ExpectedFluidAnnotations},
		Spec:       v1.PersistentVolumeSpec{},
	}}

	testPVs := []runtime.Object{}

	for _, pv := range testPVInputs {
		testPVs = append(testPVs, pv.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testPVs...)

	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		err  error
	}{
		{
			name: "volume doesn't exist",
			args: args{
				name: "notfound",
			},
			err: nil,
		},
		{
			name: "volume exists",
			args: args{
				name: "found",
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeletePersistentVolume(client, tt.args.name); err != tt.err {
				t.Errorf("testcase %v DeletePersistentVolume() = %v, want %v", tt.name, err, tt.err)
			}
		})
	}
}

func TestGetPVCsFromPod(t *testing.T) {

	volumeName := "pvc"
	pod := v1.Pod{}
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "hostpath",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/tmp/data",
			},
		},
	})

	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: volumeName,
				ReadOnly:  true,
			},
		},
	})

	pvcNames := GetPVCsFromPod(pod)

	if len(pvcNames) != 1 || pvcNames[0].Name != volumeName {
		t.Errorf("the result of GetPVCsFromPod is not right, %v", pvcNames)
	}
}

func TestGetPvcMountPods(t *testing.T) {
	namespace := "test"
	volumeName := "found"
	testPodInputs := []*v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "found"},
		Spec:       v1.PodSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "bbb", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName,
							ReadOnly:  true,
						}},
				},
			},
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
			name: "pod list is null",
			args: args{
				name:      "not found",
				namespace: namespace,
			},
			length: 0,
		},
		{
			name: "pod list is not empty",
			args: args{
				name:      "found",
				namespace: namespace,
			},
			length: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if pods, _ := GetPvcMountPods(client, tt.args.name, tt.args.namespace); len(pods) != tt.length {
				t.Errorf("testcase %v GetPvcMountPods() = %v, want %v", tt.name, pods, tt.length)
			}
		})
	}
}

func TestShouldDeleteDataset(t *testing.T) {
	namespace := "test"
	volumeName := "found"
	testPodInputs := []*v1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "found"},
		Spec:       v1.PodSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "bbb", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: volumeName,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: volumeName,
							ReadOnly:  true,
						}},
				},
			},
		},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "ccc", Namespace: namespace},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "runningDataset",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "runningDataset",
							ReadOnly:  true,
						}},
				},
			},
		}, Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	}}

	testObjects := []runtime.Object{}

	for _, pod := range testPodInputs {
		testObjects = append(testObjects, pod.DeepCopy())
	}

	testPVInputs := []*v1.PersistentVolume{{
		ObjectMeta: metav1.ObjectMeta{Name: "found"},
		Spec:       v1.PersistentVolumeSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "bbb",
			Annotations: common.ExpectedFluidAnnotations},
		Spec: v1.PersistentVolumeSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "runningDataset",
			Annotations: common.ExpectedFluidAnnotations},
		Spec: v1.PersistentVolumeSpec{},
	}}

	for _, pv := range testPVInputs {
		testObjects = append(testObjects, pv.DeepCopy())
	}

	testPVCInputs := []*v1.PersistentVolumeClaim{{
		ObjectMeta: metav1.ObjectMeta{Name: volumeName,
			Namespace: namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "runningDataset",
			Annotations: common.ExpectedFluidAnnotations,
			Namespace:   namespace},
		Spec: v1.PersistentVolumeClaimSpec{},
	}}

	for _, pvc := range testPVCInputs {
		testObjects = append(testObjects, pvc.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjects...)
	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name      string
		args      args
		errReturn bool
	}{
		{
			name: "pvc doesn't exist",
			args: args{
				name:      "notfound",
				namespace: namespace,
			},
			errReturn: false,
		},
		{
			name: "pvc exists and no pod on it",
			args: args{
				name:      "found",
				namespace: namespace,
			},
			errReturn: false,
		}, {
			name: "pvc exists and complete pod on it",
			args: args{
				name:      "completeDataset",
				namespace: namespace,
			},
			errReturn: false,
		}, {
			name: "pvc exists and running pod on it",
			args: args{
				name:      "runningDataset",
				namespace: namespace,
			},
			errReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ShouldDeleteDataset(client, tt.args.name, tt.args.namespace); (err != nil) != tt.errReturn {
				t.Errorf("testcase %v ShouldDeleteDataset() = %v, want err=%v", tt.name, err, tt.errReturn)
			}
		})
	}
}
