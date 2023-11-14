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
	"context"
	"reflect"
	"testing"

	utilpointer "k8s.io/utils/pointer"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestCheckRuntimeHealthy(t *testing.T) {
	var stsInputs = []appsv1.StatefulSet{
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
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(1),
			},
			Status: appsv1.StatefulSetStatus{
				Replicas:        1,
				ReadyReplicas:   0,
				CurrentReplicas: 0,
			},
		},
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
	testObjs := []runtime.Object{}
	for _, daemonSet := range daemonSetInputs {
		testObjs = append(testObjs, daemonSet.DeepCopy())
	}
	for _, sts := range stsInputs {
		testObjs = append(testObjs, sts.DeepCopy())
	}

	var juicefsruntimeInputs = []datav1alpha1.JuiceFSRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Replicas: 1,
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Replicas: 1,
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		},
	}
	for _, juicefsruntime := range juicefsruntimeInputs {
		testObjs = append(testObjs, juicefsruntime.DeepCopy())
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
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
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

	engines := []JuiceFSEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime:   &juicefsruntimeInputs[0],
		},
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "test",
			runtime:   &juicefsruntimeInputs[1],
		},
	}

	var testCase = []struct {
		engine                             JuiceFSEngine
		expectedErrorNil                   bool
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
		{
			engine:                             engines[1],
			expectedErrorNil:                   false,
			expectedWorkerPhase:                "",
			expectedRuntimeWorkerNumberReady:   0,
			expectedRuntimeWorkerAvailable:     0,
			expectedRuntimeFuseNumberReady:     0,
			expectedRuntimeFuseNumberAvailable: 0,
			expectedDataset: datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
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

		juicefsruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}
		if juicefsruntime.Status.WorkerNumberReady != test.expectedRuntimeWorkerNumberReady ||
			juicefsruntime.Status.WorkerNumberAvailable != test.expectedRuntimeWorkerAvailable {
			t.Errorf("fail to update the runtime")
			return
		}
		if juicefsruntime.Status.FuseNumberReady != test.expectedRuntimeFuseNumberReady ||
			juicefsruntime.Status.FuseNumberAvailable != test.expectedRuntimeFuseNumberAvailable {
			t.Errorf("fail to update the runtime")
			return
		}
		_, cond := utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
		_, cond = utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
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

	var juicefsruntimeInputs = []datav1alpha1.JuiceFSRuntime{
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
	for _, juicefsruntimeInput := range juicefsruntimeInputs {
		testObjs = append(testObjs, juicefsruntimeInput.DeepCopy())
	}
	client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

	engines := []JuiceFSEngine{
		{
			Client:    client,
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime: &datav1alpha1.JuiceFSRuntime{
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
			runtime: &datav1alpha1.JuiceFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "spark",
					Namespace: "fluid",
				},
			},
		},
	}

	var testCase = []struct {
		engine                             JuiceFSEngine
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
			expectedWorkerPhase:                datav1alpha1.RuntimePhaseReady,
			expectedErrorNil:                   true,
			expectedRuntimeFuseNumberReady:     1,
			expectedRuntimeFuseNumberAvailable: 1,
		},
	}

	for _, test := range testCase {
		err := test.engine.checkFuseHealthy()
		if err != nil && test.expectedErrorNil == true ||
			err == nil && test.expectedErrorNil == false {
			t.Errorf("fail to exec the CheckFuseHealthy function with err %v", err)
			return
		}

		juicefsruntime, err := test.engine.getRuntime()
		if err != nil {
			t.Errorf("fail to get the runtime with the error %v", err)
			return
		}

		if juicefsruntime.Status.FuseNumberReady != test.expectedRuntimeFuseNumberReady ||
			juicefsruntime.Status.FuseNumberAvailable != test.expectedRuntimeFuseNumberAvailable {
			t.Errorf("fail to update the runtime")
			return
		}

		_, cond := utils.GetRuntimeCondition(juicefsruntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
		if cond == nil {
			t.Errorf("fail to update the condition")
			return
		}
	}
}
