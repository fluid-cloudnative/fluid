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
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
)

func TestGetWorkersAsStatefulset(t *testing.T) {

	statefulsetInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sts-jindofs-worker",
				Namespace: "big-data",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		},
	}

	daemonsetInputs := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ds-jindofs-worker",
				Namespace: "big-data",
			},
		},
	}

	objs := []runtime.Object{}

	for _, runtimeInput := range daemonsetInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, statefulsetInput := range statefulsetInputs {
		objs = append(objs, statefulsetInput.DeepCopy())
	}

	s := runtime.NewScheme()
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	testCases := []struct {
		name            string
		key             types.NamespacedName
		success         bool
		deprecatedError bool
	}{
		{
			name: "noError",
			key: types.NamespacedName{
				Name:      "sts-jindofs-worker",
				Namespace: "big-data",
			},
			success:         true,
			deprecatedError: false,
		}, {
			name: "deprecatedError",
			key: types.NamespacedName{
				Name:      "ds-jindofs-worker",
				Namespace: "big-data",
			},
			success:         false,
			deprecatedError: true,
		}, {
			name: "otherError",
			key: types.NamespacedName{
				Name:      "test-jindofs-worker",
				Namespace: "big-data",
			},
			success:         false,
			deprecatedError: false,
		},
	}

	for _, testCase := range testCases {
		_, err := GetWorkersAsStatefulset(fakeClient, testCase.key)

		if testCase.success != (err == nil) {
			t.Errorf("testcase %s failed due to expect succcess %v, got error %v", testCase.name, testCase.success, err)
		}

		if !testCase.success {
			if testCase.deprecatedError != fluiderrs.IsDeprecated(err) {
				t.Errorf("testcase %s failed due to expect isdeprecated  %v, got  %v", testCase.name, testCase.deprecatedError, fluiderrs.IsDeprecated(err))
			}
		}
	}

}

func TestCheckWorkersHealthy(t *testing.T) {
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

	podList := &corev1.PodList{
		Items: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-worker-0",
				Namespace: "big-data",
				Labels:    map[string]string{"a": "b"},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodFailed,
				Conditions: []corev1.PodCondition{{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				}},
			},
		}},
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
				Replicas: utilpointer.Int32Ptr(2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
				Replicas:      1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
				Replicas:      1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(3),
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

	for _, pod := range podList.Items {
		objs = append(objs, &pod)
	}

	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	testCases := []struct {
		caseName  string
		name      string
		namespace string
		Phase     datav1alpha1.RuntimePhase
		worker    *appsv1.StatefulSet
		TypeValue bool
		isErr     bool
	}{
		{
			caseName:  "Healthy",
			name:      "hbase",
			namespace: "fluid",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-jindofs-worker",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32Ptr(2),
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:      1,
					ReadyReplicas: 1,
				},
			},
			Phase: datav1alpha1.RuntimePhaseReady,

			isErr: false,
		},
		{
			caseName:  "Unhealthy",
			name:      "hadoop",
			namespace: "fluid",
			worker: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop-jindofs-worker",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32Ptr(3),
				},
				Status: appsv1.StatefulSetStatus{
					Replicas:      1,
					ReadyReplicas: 0,
				},
			},
			Phase: datav1alpha1.RuntimePhaseNotReady,
			isErr: true,
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

		h := BuildHelper(runtimeInfo, fakeClient, log.NullLogger{})

		err = h.CheckWorkersHealthy(
			record.NewFakeRecorder(300),
			runtime, runtime.Status, statefulset)

		if testCase.isErr == (err == nil) {
			t.Errorf("check workers‘ healthy failed,err:%s", err.Error())
		}

		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.namespace,
			Name:      testCase.name,
		}, runtime)

		if err != nil {
			t.Errorf("check workers‘ healthy failed,err:%s", err.Error())
		}

		if runtime.Status.WorkerPhase != testCase.Phase {
			t.Errorf("testcase %s is failed, expect phase %v, got %v", testCase.caseName,
				testCase.Phase,
				runtime.Status.WorkerPhase)
		}

	}
}
