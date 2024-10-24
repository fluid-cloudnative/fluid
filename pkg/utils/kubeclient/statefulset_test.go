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

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

func TestScaleStatefulSet(t *testing.T) {
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
		want     int32
		hasError bool
	}{
		{
			name: "statefulset doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want:     0,
			hasError: true,
		}, {
			name: "statefulset exists",
			args: args{
				name:      "exist",
				namespace: namespace,
			},
			want:     2,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ScaleStatefulSet(client, tt.args.name, tt.args.namespace, 2)

			if (err != nil) != tt.hasError {
				t.Errorf("testcase ScaleStatefulSet() error = %v, wantErr %v", err, tt.hasError)
			}

			if err != nil {
				return
			}

			got, err := GetStatefulSet(client, tt.args.name, tt.args.namespace)
			if err != nil {
				t.Errorf("testcase %v ScaleStatefulSet() expects there is not error, but got error %v", tt.name, err)
			}
			if tt.want != *got.Spec.Replicas {
				t.Errorf("testcase %v GetStatefulSet()  got statefulset name %v, the want name of statefulset is %v", tt.name, got.Name, tt.args.name)
			}
		})
	}
}

func TestGetPhaseFromStatefulset(t *testing.T) {
	namespace := "default"
	tests := []struct {
		name     string
		args     appsv1.StatefulSet
		replicas int32
		want     datav1alpha1.RuntimePhase
	}{
		{
			name: "notReady",
			args: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "notReady",
					Namespace: namespace},
				Spec: appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 0,
				},
			},
			replicas: 1,
			want:     datav1alpha1.RuntimePhaseNotReady,
		}, {
			name: "partialReady",
			args: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "partialReady",
					Namespace: namespace},
				Spec: appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 1,
				},
			},
			replicas: 3,
			want:     datav1alpha1.RuntimePhasePartialReady,
		}, {
			name: "ready-1",
			args: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "ready",
					Namespace: namespace},
				Spec: appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 1,
				},
			},
			replicas: 1,
			want:     datav1alpha1.RuntimePhaseReady,
		}, {
			name: "ready-1",
			args: appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "ready",
					Namespace: namespace},
				Spec: appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 2,
				},
			},
			replicas: 2,
			want:     datav1alpha1.RuntimePhaseReady,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			result := GetPhaseFromStatefulset(tt.replicas, tt.args)

			if result != tt.want {
				t.Errorf("testcase %v GetPhaseFromStatefulset= %v, the expect is %v", tt.name, result, tt.want)
			}

			// t.Errorf("testcase %v IsPersistentVolumeClaimExist() = %v, want %v", tt.name, got, tt.want)

		})
	}

}

func TestGetUnavailablePodNamesForStatefulSet(t *testing.T) {
	namespace := "default"
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "unavailableSts",
			Namespace: namespace},
		Spec: appsv1.StatefulSetSpec{},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{Name: "unavailableSts-0",
					Namespace: namespace},
				Spec: corev1.PodSpec{},
			},
		},
	}

	// testRuntimes := []runtime.Object{}
	// testRuntimes = append(testRuntimes, sts.DeepCopy())
	// testRuntimes = append(testRuntimes, podList)

	// for _, pod := range podList.Items {
	// 	testRuntimes = append(testRuntimes, &pod)
	// }

	client := fake.NewFakeClientWithScheme(testScheme, podList, sts.DeepCopy())
	selector, _ := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	podNames, err := GetUnavailablePodNamesForStatefulSet(client, sts, selector)

	if err != nil {
		t.Errorf("failed due to %v", err)
	}

	if len(podNames) > 0 {
		t.Errorf("failed due to pod name %v", podNames)
	}

	// expectPodNames := []types.NamespacedName{
	// 	{
	// 		// Namespace: namespace,
	// 		// Name:      "unavailableSts-0",
	// 	},
	// }

	// if !reflect.DeepEqual(podNames, expectPodNames) {
	// 	t.Errorf("The Pod names %v and expected name %v are not equal.", podNames, expectPodNames)
	// }

}
