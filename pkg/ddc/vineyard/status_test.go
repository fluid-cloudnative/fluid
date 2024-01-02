/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
)

func TestCheckAndUpdateRuntimeStatus(t *testing.T) {

	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	var deprecatedWorkerInputs = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated-worker",
				Namespace: "fluid",
			},
		},
	}

	var workerInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 2,
			},
		},
	}

	runtimeInputs := []*datav1alpha1.VineyardRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Replicas: 2,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 3,
				CurrentMasterNumberScheduled: 3,
				CurrentFuseNumberScheduled:   3,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 3,
				DesiredFuseNumberScheduled:   2,
				Conditions: []datav1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", v1.ConditionTrue),
					utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", v1.ConditionTrue),
				},
				WorkerPhase: "NotReady",
				FusePhase:   "NotReady",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Replicas: 2,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				WorkerPhase:                  "NotReady",
				FusePhase:                    "NotReady",
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.VineyardRuntimeSpec{
				Worker: datav1alpha1.VineyardCompTemplateSpec{
					Replicas: 2,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				WorkerPhase:                  "NotReady",
				FusePhase:                    "NotReady",
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

	for _, deprecatedWorkerInput := range deprecatedWorkerInputs {
		objs = append(objs, deprecatedWorkerInput.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	testCases := []struct {
		testName   string
		name       string
		namespace  string
		isErr      bool
		deprecated bool
		wanted     bool
	}{
		{testName: "deprecated",
			name:       "deprecated",
			namespace:  "fluid",
			deprecated: true,
		}, {
			testName:  "hadoop",
			name:      "hadoop",
			namespace: "fluid",
			wanted:    true,
		},
	}

	for _, testCase := range testCases {
		engine := newVineyardEngineREP(fakeClient, testCase.name, testCase.namespace)

		patch := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
			func(_ *VineyardEngine) ([]string, error) {
				return []string{"instances_memory_usage_bytes 1234", "instances_memory_usage_bytes 4321"}, nil
			})
		defer patch.Reset()

		ready, err := engine.CheckAndUpdateRuntimeStatus()
		if err != nil || ready != testCase.wanted {
			t.Errorf("testcase %s Failed due to %v", testCase.testName, err)
		}
	}
}
