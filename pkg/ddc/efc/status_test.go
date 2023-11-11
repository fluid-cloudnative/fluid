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

package efc

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
)

func TestCheckAndUpdateRuntimeStatus(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ready-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "master-not-ready-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-partial-ready-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-not-ready-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	var workerInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ready-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "master-not-ready-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-partial-ready-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-not-ready-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 0,
			},
		},
	}

	runtimeInputs := []*datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ready",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Replicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "master-not-ready",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Replicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-partial-ready",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Replicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "worker-not-ready",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Replicas: 2,
			},
		},
	}

	objs := []runtime.Object{}
	for _, masterInput := range masterInputs {
		objs = append(objs, masterInput.DeepCopy())
	}

	for _, workerInput := range workerInputs {
		objs = append(objs, workerInput.DeepCopy())
	}

	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	testCases := []struct {
		testName  string
		name      string
		namespace string
		isErr     bool
		wanted    bool
	}{
		{
			testName:  "ready",
			name:      "ready",
			namespace: "fluid",
			isErr:     false,
			wanted:    true,
		},
		{
			testName:  "master-not-ready",
			name:      "master-not-ready",
			namespace: "fluid",
			wanted:    false,
		},
		{
			testName:  "worker-partial-ready",
			name:      "worker-partial-ready",
			namespace: "fluid",
			wanted:    true,
		},
		{
			testName:  "worker-not-ready",
			name:      "worker-not-ready",
			namespace: "fluid",
			wanted:    false,
		},
	}

	for _, testCase := range testCases {
		engine := newEFCEngineREP(fakeClient, testCase.name, testCase.namespace)
		ready, err := engine.CheckAndUpdateRuntimeStatus()
		if err != nil || ready != testCase.wanted {
			t.Errorf("testcase %s Failed due to %v", testCase.testName, err)
		}
	}
}
