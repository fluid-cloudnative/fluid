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

package ctrl

import (
	"context"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
)

// func newAlluxioEngineREP(client client.Client, name string, namespace string) *alluxio.AlluxioEngine {

// 	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "alluxio", datav1alpha1.TieredStore{})
// 	engine := &alluxio.AlluxioEngine{
// 		runtime:     &datav1alpha1.AlluxioRuntime{},
// 		name:        name,
// 		namespace:   namespace,
// 		Client:      client,
// 		runtimeInfo: runTimeInfo,
// 		Log:         fake.NullLogger(),
// 	}
// 	return engine
// }

func TestSyncReplicas(t *testing.T) {

	runtimeInputs := []*datav1alpha1.JindoRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 3, // 2
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2, // 0
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 3,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   3,
				WorkerPhase:                  "NotReady",
				FusePhase:                    "NotReady",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 2,
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 3,
				CurrentMasterNumberScheduled: 3,
				CurrentFuseNumberScheduled:   3,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 3,
				DesiredFuseNumberScheduled:   2,
				WorkerPhase:                  "NotReady",
				FusePhase:                    "NotReady",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 2,
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,

				WorkerPhase: "NotReady",
				FusePhase:   "NotReady",
			},
		},
	}

	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
		},
	}

	statefulsetInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-worker",
				Namespace: "big-data",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(3),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(3),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	objs := []runtime.Object{}

	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, dataSetInput := range dataSetInputs {
		objs = append(objs, dataSetInput.DeepCopy())
	}
	for _, statefulsetInput := range statefulsetInputs {
		objs = append(objs, statefulsetInput.DeepCopy())
	}

	s := runtime.NewScheme()
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	testCases := []struct {
		caseName         string
		name             string
		namespace        string
		Type             datav1alpha1.RuntimeConditionType
		worker           *appsv1.StatefulSet
		isErr            bool
		expectConditions int
	}{
		{
			caseName:  "scaleOut",
			name:      "hbase",
			namespace: "fluid",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-jindofs-worker",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(2),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 1,
				},
			},
			Type:             datav1alpha1.RuntimeWorkerScaledOut,
			isErr:            false,
			expectConditions: 2,
		},
		{
			caseName:  "scaleIn",
			name:      "hadoop",
			namespace: "fluid",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop-jindofs-worker",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 1,
				},
			},
			Type:             datav1alpha1.RuntimeWorkerScaledIn,
			isErr:            false,
			expectConditions: 2,
		},
		{

			caseName:  "noAction",
			name:      "obj",
			namespace: "fluid",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "obj-jindofs-worker",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas: 1,
				},
			},
			Type:             "",
			isErr:            false,
			expectConditions: 0,
		},
	}
	for _, testCase := range testCases {

		runtimeInfo, err := base.BuildRuntimeInfo(testCase.name, testCase.namespace, "jindo", datav1alpha1.TieredStore{})
		if err != nil {
			t.Errorf("testcase %s failed due to %v", testCase.name, err)
		}

		var runtime *datav1alpha1.JindoRuntime = &datav1alpha1.JindoRuntime{}

		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.namespace,
			Name:      testCase.name,
		}, runtime)

		if err != nil {
			t.Errorf("testCase %s sync replicas failed,err:%v", testCase.caseName, err)
		}

		statefulset := &appsv1.StatefulSet{}
		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.worker.Namespace,
			Name:      testCase.worker.Name,
		}, statefulset)
		if err != nil {
			t.Errorf("sync replicas failed,err:%s", err.Error())
		}

		h := BuildHelper(runtimeInfo, fakeClient, fake.NullLogger())
		err = h.SyncReplicas(cruntime.ReconcileRequestContext{
			Log:      fake.NullLogger(),
			Recorder: record.NewFakeRecorder(300),
		}, runtime, runtime.Status, statefulset)

		if err != nil {
			t.Errorf("sync replicas failed,err:%s", err.Error())
		}

		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.namespace,
			Name:      testCase.name,
		}, runtime)

		if err != nil {
			t.Errorf("sync replicas failed,err:%v", err)
		}

		// rt, _ := engine.getRuntime()
		if len(runtime.Status.Conditions) != testCase.expectConditions {

			t.Errorf("runtime condition want count %v, got  condtions %v", testCase.expectConditions, runtime.Status.Conditions)

		}

		found := false
		for _, cond := range runtime.Status.Conditions {

			if cond.Type == testCase.Type {
				found = true
				break
			}
		}

		if !found && testCase.expectConditions > 0 {
			t.Errorf("runtime condition want conditionType %v, got  conditions %v", testCase.Type, runtime.Status.Conditions)

		}
	}
}
