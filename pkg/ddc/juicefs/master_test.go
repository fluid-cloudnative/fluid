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

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestJuiceFSEngine_ShouldSetupMaster(t *testing.T) {
	juicefsruntimeInputs := []datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test0",
				Namespace: "fluid",
			},
			Status: datav1alpha1.JuiceFSRuntimeStatus{
				WorkerPhase: datav1alpha1.RuntimePhaseNotReady,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "fluid",
			},
			Status: datav1alpha1.JuiceFSRuntimeStatus{
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
	daemonSetInputs := []v1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-worker",
				Namespace: "fluid",
			},
			Status: v1.DaemonSetStatus{
				NumberReady: 1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
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
			Log:       log.NullLogger{},
		},
	}

	var testCases = []struct {
		engine          JuiceFSEngine
		wantedDaemonSet v1.DaemonSet
	}{
		{
			engine:          engines[0],
			wantedDaemonSet: daemonSetInputs[0],
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
