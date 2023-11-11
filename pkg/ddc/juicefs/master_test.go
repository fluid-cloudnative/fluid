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

package juicefs

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestJuiceFSEngine_ShouldSetupMaster(t *testing.T) {
	juicefsruntimeInputs := []datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "fluid",
			},
			Status: datav1alpha1.RuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhaseNone,
			},
		},
	}
	testObjs := []runtime.Object{}
	for _, juicefsruntime := range juicefsruntimeInputs {
		testObjs = append(testObjs, juicefsruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JuiceFSEngine{
		{
			name:      "test0",
			namespace: "fluid",
			Client:    client,
		},
		{
			name:      "test1",
			namespace: "fluid",
			Client:    client,
		},
	}

	var testCases = []struct {
		engine         JuiceFSEngine
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

func TestJuiceFSEngine_SetupMaster(t *testing.T) {
	stsInputs := []v1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-worker",
				Namespace: "fluid",
			},
			Status: v1.StatefulSetStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, sts := range stsInputs {
		testObjs = append(testObjs, sts.DeepCopy())
	}

	juicefsruntimeInputs := []datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
		},
	}
	for _, juicefsruntime := range juicefsruntimeInputs {
		testObjs = append(testObjs, juicefsruntime.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JuiceFSEngine{
		{
			name:      "test",
			namespace: "fluid",
			Client:    client,
			Log:       fake.NullLogger(),
		},
	}

	var testCases = []struct {
		engine            JuiceFSEngine
		wantedStatefulSet v1.StatefulSet
	}{
		{
			engine:            engines[0],
			wantedStatefulSet: stsInputs[0],
		},
	}

	for _, test := range testCases {
		if err := test.engine.SetupMaster(); err != nil {
			t.Errorf("fail to exec the func with error %v", err)
			return
		}
		juicefsruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime")
			return
		}
		if juicefsruntime.Status.WorkerPhase == datav1alpha1.RuntimePhaseNone {
			t.Errorf("fail to update the runtime")
			return
		}
	}
}

func TestJuiceFSEngine_CheckMasterReady(t *testing.T) {
	type fields struct {
		runtime     *datav1alpha1.JuiceFSRuntime
		name        string
		namespace   string
		runtimeType string
		runtimeInfo base.RuntimeInfoInterface
	}
	tests := []struct {
		name      string
		fields    fields
		wantReady bool
		wantErr   bool
	}{
		{
			name:      "test",
			fields:    fields{},
			wantReady: true,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				runtime:     tt.fields.runtime,
				name:        tt.fields.name,
				namespace:   tt.fields.namespace,
				runtimeType: tt.fields.runtimeType,
				runtimeInfo: tt.fields.runtimeInfo,
			}
			gotReady, err := j.CheckMasterReady()
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckMasterReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotReady != tt.wantReady {
				t.Errorf("CheckMasterReady() gotReady = %v, want %v", gotReady, tt.wantReady)
			}
		})
	}
}
