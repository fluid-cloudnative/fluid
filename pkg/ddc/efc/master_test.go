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
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCheckMasterReady(t *testing.T) {
	statefulsetInputs := []v1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

	EFCRuntimeInputs := []datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.EFCRuntimeSpec{
				Master: datav1alpha1.EFCCompTemplateSpec{
					Disabled: true,
				},
			},
		},
	}
	for _, EFCRuntime := range EFCRuntimeInputs {
		testObjs = append(testObjs, EFCRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []EFCEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine         EFCEngine
		expectedResult bool
	}{
		{
			engine:         engines[0],
			expectedResult: true,
		},
		{
			engine:         engines[1],
			expectedResult: false,
		},
		{
			engine:         engines[2],
			expectedResult: true,
		},
	}

	for _, test := range testCases {
		if ready, _ := test.engine.CheckMasterReady(); ready != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
		if !test.expectedResult {
			continue
		}
		EFCRuntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get runtime %v", err)
			return
		}
		if len(EFCRuntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime conditions")
			return
		}
	}
}

func TestShouldSetupMaster(t *testing.T) {
	EFCRuntimeInputs := []datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseNotReady,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				MasterPhase: datav1alpha1.RuntimePhaseNone,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, EFCRuntime := range EFCRuntimeInputs {
		testObjs = append(testObjs, EFCRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []EFCEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
		},
		{
			name:      "hbase",
			namespace: "fluid",
			Client:    client,
		},
	}

	var testCases = []struct {
		engine         EFCEngine
		expectedResult bool
	}{
		{
			engine:         engines[0],
			expectedResult: false,
		},
		{
			engine:         engines[1],
			expectedResult: true,
		},
	}

	for _, test := range testCases {
		if should, _ := test.engine.ShouldSetupMaster(); should != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
	}
}

func TestSetupMaster(t *testing.T) {
	statefulSetInputs := []v1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, statefulSet := range statefulSetInputs {
		testObjs = append(testObjs, statefulSet.DeepCopy())
	}

	EFCRuntimeInputs := []datav1alpha1.EFCRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, EFCRuntime := range EFCRuntimeInputs {
		testObjs = append(testObjs, EFCRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []EFCEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine                EFCEngine
		expectedSelector      string
		expectedConfigMapName string
	}{
		{
			engine:                engines[0],
			expectedConfigMapName: engines[0].getConfigmapName(),
			expectedSelector:      engines[0].getWorkerSelectors(),
		},
	}

	for _, test := range testCases {
		_ = test.engine.SetupMaster()
		EFCRuntime, _ := test.engine.getRuntime()
		if len(EFCRuntime.Status.Conditions) == 0 || EFCRuntime.Status.Selector != test.expectedSelector || EFCRuntime.Status.ValueFileConfigmap != test.expectedConfigMapName {
			t.Errorf("fail to update the runtime")
			return
		}
	}
}
