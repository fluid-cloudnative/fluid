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
