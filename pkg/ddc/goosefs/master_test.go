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

package goosefs

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
				ReadyReplicas: 1,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

	goosefsRuntimeInputs := []datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Master: datav1alpha1.GooseFSCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Master: datav1alpha1.GooseFSCompTemplateSpec{
					Replicas: 2,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Master: datav1alpha1.GooseFSCompTemplateSpec{
					Replicas: 1,
				},
				APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
					Enabled: true,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				APIGatewayStatus: &datav1alpha1.APIGatewayStatus{
					Endpoint: "test-endpoint",
				},
			},
		},
	}
	for _, goosefsRuntime := range goosefsRuntimeInputs {
		testObjs = append(testObjs, goosefsRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
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
		engine         GooseFSEngine
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
	}

	for _, test := range testCases {
		if ready, _ := test.engine.CheckMasterReady(); ready != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
		if !test.expectedResult {
			continue
		}
		goosefsRuntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get runtime %v", err)
			return
		}
		if len(goosefsRuntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime conditions")
			return
		}
	}
}

func TestShouldSetupMaster(t *testing.T) {
	goosefsRuntimeInputs := []datav1alpha1.GooseFSRuntime{
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
	for _, goosefsRuntime := range goosefsRuntimeInputs {
		testObjs = append(testObjs, goosefsRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
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
		engine         GooseFSEngine
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

	goosefsRuntimeInputs := []datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, goosefsRuntime := range goosefsRuntimeInputs {
		testObjs = append(testObjs, goosefsRuntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
		{
			name:      "spark",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine                GooseFSEngine
		expectedSelector      string
		expectedConfigMapName string
	}{
		{
			engine:                engines[0],
			expectedConfigMapName: "spark--values",
			expectedSelector:      "app=goosefs,release=spark,role=goosefs-worker",
		},
	}

	for _, test := range testCases {
		if err := test.engine.SetupMaster(); err != nil {
			t.Errorf("fail to exec the func with error %v", err)
			return
		}
		goosefsRuntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime")
			return
		}
		if goosefsRuntime.Status.Selector != test.expectedSelector || goosefsRuntime.Status.ValueFileConfigmap != test.expectedConfigMapName ||
			len(goosefsRuntime.Status.Conditions) == 0 {
			t.Errorf("fail to update the runtime")
			return
		}
	}
}
