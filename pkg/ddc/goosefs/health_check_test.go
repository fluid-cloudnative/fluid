/*
Copyright 2022 The Fluid Authors.

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

package goosefs

import (
	"context"
	"reflect"
	"testing"

	"k8s.io/client-go/tools/record"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilpointer "k8s.io/utils/pointer"
)

func TestCheckRuntimeHealthy(t *testing.T) {
	var statefulsetInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 3,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-worker",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:        1,
				ReadyReplicas:   1,
				CurrentReplicas: 1,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

	var daemonSetInputs = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 0,
				NumberReady:       1,
				NumberAvailable:   1,
			},
		},
	}
	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
	}

	var goosefsRuntimeInputs = []datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Master: datav1alpha1.GooseFSCompTemplateSpec{
					Replicas: 1,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
	}
	for _, goosefsRuntime := range goosefsRuntimeInputs {
		testObjs = append(testObjs, goosefsRuntime.DeepCopy())
	}

	var datasetInputs = []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
			Status: datav1alpha1.DatasetStatus{
				HCFSStatus: &datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			},
		},
	}
	for _, dataset := range datasetInputs {
		testObjs = append(testObjs, dataset.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime:   &goosefsRuntimeInputs[0],
		},
	}

	var testCase = []struct {
		engine                             GooseFSEngine
		expectedErrorNil                   bool
		expectedMasterPhase                datav1alpha1.RuntimePhase
		expectedWorkerPhase                datav1alpha1.RuntimePhase
		expectedRuntimeWorkerNumberReady   int32
		expectedRuntimeWorkerAvailable     int32
		expectedRuntimeFuseNumberReady     int32
		expectedRuntimeFuseNumberAvailable int32
		expectedDataset                    datav1alpha1.Dataset
	}{
		{
			engine:                             engines[0],
			expectedErrorNil:                   true,
			expectedMasterPhase:                datav1alpha1.RuntimePhaseReady,
			expectedWorkerPhase:                "",
			expectedRuntimeWorkerNumberReady:   1,
			expectedRuntimeWorkerAvailable:     1,
			expectedRuntimeFuseNumberReady:     1,
			expectedRuntimeFuseNumberAvailable: 1,
			expectedDataset: datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
				},
			},
		},
	}
	for _, test := range testCase {
		err := test.engine.CheckRuntimeHealthy()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the checkMasterHealthy function with err %v", err)
			return
		}
		if test.expectedErrorNil == false {
			continue
		}

		goosefsRuntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}
		if goosefsRuntime.Status.MasterPhase != test.expectedMasterPhase {
			t.Errorf("fail to update the runtime status, get %s, expect %s", goosefsRuntime.Status.MasterPhase, test.expectedMasterPhase)
			return
		}
		if goosefsRuntime.Status.WorkerNumberReady != test.expectedRuntimeWorkerNumberReady ||
			goosefsRuntime.Status.WorkerNumberAvailable != test.expectedRuntimeWorkerAvailable {
			t.Errorf("fail to update the runtime")
			return
		}
		if goosefsRuntime.Status.FuseNumberReady != test.expectedRuntimeFuseNumberReady ||
			goosefsRuntime.Status.FuseNumberAvailable != test.expectedRuntimeFuseNumberAvailable {
			t.Errorf("fail to update the runtime")
			return
		}

		_, cond := utils.GetRuntimeCondition(goosefsRuntime.Status.Conditions, datav1alpha1.RuntimeMasterReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
		_, cond = utils.GetRuntimeCondition(goosefsRuntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
		_, cond = utils.GetRuntimeCondition(goosefsRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}

		var datasets datav1alpha1.DatasetList
		err = client.List(context.TODO(), &datasets)
		if err != nil {
			t.Errorf("fail to list the datasets with error %v", err)
			return
		}
		if !reflect.DeepEqual(datasets.Items[0].Status.Phase, test.expectedDataset.Status.Phase) ||
			!reflect.DeepEqual(datasets.Items[0].Status.CacheStates, test.expectedDataset.Status.CacheStates) ||
			!reflect.DeepEqual(datasets.Items[0].Status.HCFSStatus, test.expectedDataset.Status.HCFSStatus) {
			t.Errorf("fail to exec the function with error %v", err)
			return
		}
	}
}

func TestCheckMasterHealthy(t *testing.T) {
	var statefulsetInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-master",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 3,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, statefulset := range statefulsetInputs {
		testObjs = append(testObjs, statefulset.DeepCopy())
	}

	var goosefsruntimeInputs = []datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, goosefsruntimeInput := range goosefsruntimeInputs {
		testObjs = append(testObjs, goosefsruntimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			},
		},
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "spark",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			},
		},
	}

	var testCase = []struct {
		engine              GooseFSEngine
		expectedErrorNil    bool
		expectedMasterPhase datav1alpha1.RuntimePhase
	}{
		{
			engine:              engines[0],
			expectedErrorNil:    false,
			expectedMasterPhase: "",
		},
		{
			engine:              engines[1],
			expectedErrorNil:    true,
			expectedMasterPhase: datav1alpha1.RuntimePhaseReady,
		},
	}

	for _, test := range testCase {
		err := test.engine.checkMasterHealthy()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the checkMasterHealthy function with err %v", err)
			return
		}

		if test.expectedErrorNil == false {
			continue
		}
		goosefsruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if goosefsruntime.Status.MasterPhase != test.expectedMasterPhase {
			t.Errorf("fail to update the runtime status, get %s, expect %s", goosefsruntime.Status.MasterPhase, test.expectedMasterPhase)
			return
		}

		_, cond := utils.GetRuntimeCondition(goosefsruntime.Status.Conditions, datav1alpha1.RuntimeMasterReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
	}
}

func TestCheckWorkersHealthy(t *testing.T) {
	var statefulSetInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-worker",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:        1,
				ReadyReplicas:   0,
				CurrentReplicas: 1,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(2),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-worker",
				Namespace: "fluid",
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:        1,
				ReadyReplicas:   1,
				CurrentReplicas: 1,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
		},
	}

	var daemonSetInputs = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated-worker",
				Namespace: "fluid",
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, statefulSet := range statefulSetInputs {
		testObjs = append(testObjs, statefulSet.DeepCopy())
	}

	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
	}

	var goosefsruntimeInputs = []datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated",
				Namespace: "fluid",
			},
		},
	}
	for _, goosefsruntimeInput := range goosefsruntimeInputs {
		testObjs = append(testObjs, goosefsruntimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			},
		},
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "spark",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			},
		},
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "deprecated",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "deprecated",
					Namespace: "fluid",
				},
			},
			Recorder: record.NewFakeRecorder(1),
		},
	}

	var testCase = []struct {
		engine                           GooseFSEngine
		expectedWorkerPhase              datav1alpha1.RuntimePhase
		expectedErrorNil                 bool
		expectedRuntimeWorkerNumberReady int32
		expectedRuntimeWorkerAvailable   int32
	}{
		{
			engine:                           engines[0],
			expectedWorkerPhase:              datav1alpha1.RuntimePhaseNotReady,
			expectedErrorNil:                 false,
			expectedRuntimeWorkerNumberReady: 0,
			expectedRuntimeWorkerAvailable:   1,
		},
		{
			engine:                           engines[1],
			expectedWorkerPhase:              "",
			expectedErrorNil:                 true,
			expectedRuntimeWorkerNumberReady: 1,
			expectedRuntimeWorkerAvailable:   1,
		},
	}

	for _, test := range testCase {
		err := test.engine.checkWorkersHealthy()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the checkMasterHealthy function with err %v", err)
			return
		}

		goosefsruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if goosefsruntime.Status.WorkerNumberReady != test.expectedRuntimeWorkerNumberReady ||
			goosefsruntime.Status.WorkerNumberAvailable != test.expectedRuntimeWorkerAvailable {
			t.Errorf("fail to update the runtime")
			return
		}

		_, cond := utils.GetRuntimeCondition(goosefsruntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
	}

	var testCaseWithDeprecatedRuntime = []struct {
		engine           GooseFSEngine
		expectedErrorNil bool
	}{
		{
			engine:           engines[2],
			expectedErrorNil: true,
		},
	}

	for _, test := range testCaseWithDeprecatedRuntime {
		runtimeBefore, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		err = test.engine.checkWorkersHealthy()
		if test.expectedErrorNil != (err == nil) {
			t.Errorf("fail to exec the checkMasterHealthy function with err %v", err)
			return
		}

		runtimeAfter, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if !reflect.DeepEqual(runtimeBefore, runtimeAfter) {
			t.Error("Runtime should remain unchanged")
			return
		}
	}
}

func TestCheckFuseHealthy(t *testing.T) {
	var daemonSetInputs = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 1,
				NumberReady:       1,
				NumberAvailable:   1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark-fuse",
				Namespace: "fluid",
			},
			Status: appsv1.DaemonSetStatus{
				NumberUnavailable: 0,
				NumberReady:       1,
				NumberAvailable:   1,
			},
		},
	}

	testObjs := []runtime.Object{}
	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
	}

	var goosefsruntimeInputs = []datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "spark",
				Namespace: "fluid",
			},
		},
	}
	for _, goosefsruntimeInput := range goosefsruntimeInputs {
		testObjs = append(testObjs, goosefsruntimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []GooseFSEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
			},
		},
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "spark",
			runtime: &datav1alpha1.GooseFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			},
		},
	}

	var testCase = []struct {
		engine                             GooseFSEngine
		expectedWorkerPhase                datav1alpha1.RuntimePhase
		expectedErrorNil                   bool
		expectedRuntimeFuseNumberReady     int32
		expectedRuntimeFuseNumberAvailable int32
	}{
		{
			engine:                             engines[0],
			expectedWorkerPhase:                datav1alpha1.RuntimePhaseNotReady,
			expectedErrorNil:                   false,
			expectedRuntimeFuseNumberReady:     1,
			expectedRuntimeFuseNumberAvailable: 1,
		},
		{
			engine:                             engines[1],
			expectedWorkerPhase:                "",
			expectedErrorNil:                   true,
			expectedRuntimeFuseNumberReady:     1,
			expectedRuntimeFuseNumberAvailable: 1,
		},
	}

	for _, test := range testCase {
		err := test.engine.checkFuseHealthy()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the checkMasterHealthy function with err %v", err)
			return
		}

		goosefsruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if goosefsruntime.Status.FuseNumberReady != test.expectedRuntimeFuseNumberReady ||
			goosefsruntime.Status.FuseNumberAvailable != test.expectedRuntimeFuseNumberAvailable {
			t.Errorf("fail to update the runtime")
			return
		}

		_, cond := utils.GetRuntimeCondition(goosefsruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
	}
}
