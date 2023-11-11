/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
