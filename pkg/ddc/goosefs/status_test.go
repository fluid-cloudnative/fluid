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

package goosefs

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
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testStatusNamespace       = "fluid"
	testStatusRuntimeHadoop   = "hadoop"
	testStatusRuntimeHbase    = "hbase"
	testStatusRuntimeNoWorker = "no-worker"
	testStatusRuntimeNoMaster = "no-master"
	testStatusPhaseNotReady   = "NotReady"
	testStatusUfsTotal        = "19.07MiB"
)

func TestCheckAndUpdateRuntimeStatus(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeHadoop + "-master",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeHbase + "-master",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
			},
		},
	}

	var workerInputs = []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeHadoop + "-worker",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](3),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      3,
				ReadyReplicas: 2,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeHbase + "-worker",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](2),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 2,
			},
		},
	}

	runtimeInputs := []*datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeHadoop,
				Namespace: testStatusNamespace,
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Replicas: 2,
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
				WorkerPhase: testStatusPhaseNotReady,
				FusePhase:   testStatusPhaseNotReady,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeHbase,
				Namespace: testStatusNamespace,
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Replicas: 2,
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				Conditions: []datav1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", v1.ConditionTrue),
					utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", v1.ConditionTrue),
				},
				WorkerPhase: testStatusPhaseNotReady,
				FusePhase:   testStatusPhaseNotReady,
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
			testName:  "test master and worker ready",
			name:      testStatusRuntimeHadoop,
			namespace: testStatusNamespace,
			isErr:     false,
			wanted:    true,
		},
		{
			testName:  "test master not ready",
			name:      testStatusRuntimeHbase,
			namespace: testStatusNamespace,
			isErr:     false,
			wanted:    false,
		},
	}

	for _, testCase := range testCases {
		engine := newGooseFSEngineREP(fakeClient, testCase.name, testCase.namespace)

		patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
			func(_ *GooseFSEngine) (string, error) {
				summary := mockGooseFSReportSummary()
				return summary, nil
			})
		defer patch1.Reset()

		patch2 := ApplyFunc(utils.GetDataset,
			func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
				d := &datav1alpha1.Dataset{
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: testStatusUfsTotal,
					},
				}
				return d, nil
			})
		defer patch2.Reset()

		patch3 := ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
			func(_ *GooseFSEngine) cacheHitStates {
				return cacheHitStates{
					bytesReadLocal:  20310917,
					bytesReadUfsAll: 32243712,
				}
			})
		defer patch3.Reset()

		ready, err := engine.CheckAndUpdateRuntimeStatus()
		hasErr := err != nil
		if hasErr != testCase.isErr {
			t.Errorf("testcase %s failed: expected isErr=%v, got error=%v", testCase.testName, testCase.isErr, err)
		}
		if ready != testCase.wanted {
			t.Errorf("testcase %s failed: expected ready=%v, got ready=%v", testCase.testName, testCase.wanted, ready)
		}
	}
}

func TestCheckAndUpdateRuntimeStatusWithNoWorker(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeNoWorker + "-master",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	runtimeInputs := []*datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeNoWorker,
				Namespace: testStatusNamespace,
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Replicas: 2,
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				WorkerPhase:                  testStatusPhaseNotReady,
				FusePhase:                    testStatusPhaseNotReady,
			},
		},
	}

	objs := []runtime.Object{}
	for _, masterInput := range masterInputs {
		objs = append(objs, masterInput.DeepCopy())
	}
	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	engine := newGooseFSEngineREP(fakeClient, testStatusRuntimeNoWorker, testStatusNamespace)

	ready, err := engine.CheckAndUpdateRuntimeStatus()
	if err == nil {
		t.Errorf("expected error when worker statefulset not found, got ready=%v", ready)
	}
}

func TestCheckAndUpdateRuntimeStatusWithNoMaster(t *testing.T) {
	workerInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeNoMaster + "-worker",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](2),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      2,
				ReadyReplicas: 2,
			},
		},
	}

	runtimeInputs := []*datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testStatusRuntimeNoMaster,
				Namespace: testStatusNamespace,
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Replicas: 2,
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				WorkerPhase:                  testStatusPhaseNotReady,
				FusePhase:                    testStatusPhaseNotReady,
			},
		},
	}

	objs := []runtime.Object{}
	for _, workerInput := range workerInputs {
		objs = append(objs, workerInput.DeepCopy())
	}
	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	engine := newGooseFSEngineREP(fakeClient, testStatusRuntimeNoMaster, testStatusNamespace)

	ready, err := engine.CheckAndUpdateRuntimeStatus()
	if err == nil {
		t.Errorf("expected error when master statefulset not found, got ready=%v", ready)
	}
}

func TestCheckAndUpdateRuntimeStatusWithZeroReplicas(t *testing.T) {
	masterInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "zero-replicas-master",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](1),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	workerInputs := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "zero-replicas-worker",
				Namespace: testStatusNamespace,
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptr.To[int32](0),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:      0,
				ReadyReplicas: 0,
			},
		},
	}

	runtimeInputs := []*datav1alpha1.GooseFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "zero-replicas",
				Namespace: testStatusNamespace,
			},
			Spec: datav1alpha1.GooseFSRuntimeSpec{
				Replicas: 0,
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 0,
				CurrentMasterNumberScheduled: 1,
				CurrentFuseNumberScheduled:   0,
				DesiredMasterNumberScheduled: 1,
				DesiredWorkerNumberScheduled: 0,
				DesiredFuseNumberScheduled:   0,
				WorkerPhase:                  testStatusPhaseNotReady,
				FusePhase:                    testStatusPhaseNotReady,
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

	engine := newGooseFSEngineREP(fakeClient, "zero-replicas", testStatusNamespace)

	patch1 := ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
		func(_ *GooseFSEngine) (string, error) {
			summary := mockGooseFSReportSummary()
			return summary, nil
		})
	defer patch1.Reset()

	patch2 := ApplyFunc(utils.GetDataset,
		func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
			d := &datav1alpha1.Dataset{
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: testStatusUfsTotal,
				},
			}
			return d, nil
		})
	defer patch2.Reset()

	patch3 := ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
		func(_ *GooseFSEngine) cacheHitStates {
			return cacheHitStates{
				bytesReadLocal:  0,
				bytesReadUfsAll: 0,
			}
		})
	defer patch3.Reset()

	ready, err := engine.CheckAndUpdateRuntimeStatus()
	if err != nil {
		t.Errorf("unexpected error for zero replicas case: %v", err)
	}
	if !ready {
		t.Errorf("expected ready=true for zero replicas case when master is ready, got ready=%v", ready)
	}
}
