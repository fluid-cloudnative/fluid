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

package jindocache

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newJindoCacheEngineREP(client client.Client, name string, namespace string) *JindoCacheEngine {

	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "jindo", datav1alpha1.TieredStore{})
	engine := &JindoCacheEngine{
		runtime:     &datav1alpha1.JindoRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	engine.Helper = ctrl.BuildHelper(runTimeInfo, client, engine.Log)
	return engine
}

func TestSyncReplicas(t *testing.T) {
	nodeInputs := []*v1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":             "1",
					"fluid.io/s-Jindo-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":           "true",
					"fluid.io/s-h-Jindo-d-fluid-spark": "5B",
					"fluid.io/s-h-Jindo-m-fluid-spark": "1B",
					"fluid.io/s-h-Jindo-t-fluid-spark": "6B",
					"fluid_exclusive":                  "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":              "2",
					"fluid.io/s-Jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-Jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-Jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-Jindo-t-fluid-hadoop": "6B",
					"fluid.io/s-Jindo-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":            "true",
					"fluid.io/s-h-Jindo-d-fluid-hbase":  "5B",
					"fluid.io/s-h-Jindo-m-fluid-hbase":  "1B",
					"fluid.io/s-h-Jindo-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":              "1",
					"fluid.io/s-Jindo-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":           "true",
					"fluid.io/s-h-Jindo-d-fluid-hadoop": "5B",
					"fluid.io/s-h-Jindo-m-fluid-hadoop": "1B",
					"fluid.io/s-h-Jindo-t-fluid-hadoop": "6B",
					"node-select":                       "true",
				},
			},
		},
	}
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
				WorkerPhase:                  "NotReady",
				FusePhase:                    "NotReady",
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 3, // 2
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 0,
				CurrentMasterNumberScheduled: 0, // 0
				CurrentFuseNumberScheduled:   0,
				DesiredMasterNumberScheduled: 0,
				DesiredWorkerNumberScheduled: 0,
				DesiredFuseNumberScheduled:   0,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "disabled",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Replicas: 3, // 2
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Disabled: true,
				},
			},
			Status: datav1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 0,
				CurrentMasterNumberScheduled: 0, // 0
				CurrentFuseNumberScheduled:   0,
				DesiredMasterNumberScheduled: 0,
				DesiredWorkerNumberScheduled: 0,
				DesiredFuseNumberScheduled:   0,
			},
		},
	}
	workersInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(2),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(2),
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-jindofs-worker",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32Ptr(2),
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
	fuseInputs := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-fuse",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-fuse",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-jindofs-fuse",
				Namespace: "fluid",
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deprecated-jindofs-worker",
				Namespace: "fluid",
			},
		},
	}

	objs := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		objs = append(objs, nodeInput.DeepCopy())
	}
	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, workerInput := range workersInputs {
		objs = append(objs, workerInput.DeepCopy())
	}
	for _, fuseInput := range fuseInputs {
		objs = append(objs, fuseInput.DeepCopy())
	}
	for _, dataSetInput := range dataSetInputs {
		objs = append(objs, dataSetInput.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	testCases := []struct {
		testName       string
		name           string
		namespace      string
		Type           datav1alpha1.RuntimeConditionType
		isErr          bool
		condtionLength int
		deprecated     bool
	}{
		{
			testName:       "scaleout",
			name:           "hbase",
			namespace:      "fluid",
			Type:           datav1alpha1.RuntimeWorkerScaledOut,
			isErr:          false,
			condtionLength: 3,
		},
		{
			testName:       "scalein",
			name:           "hadoop",
			namespace:      "fluid",
			Type:           datav1alpha1.RuntimeWorkerScaledIn,
			isErr:          false,
			condtionLength: 3,
		},
		{
			testName:       "noscale",
			name:           "obj",
			namespace:      "fluid",
			Type:           "",
			isErr:          false,
			condtionLength: 0,
		}, {
			testName:       "disabled",
			name:           "disabled",
			namespace:      "fluid",
			Type:           "",
			isErr:          false,
			condtionLength: 0,
			deprecated:     false,
		}, {
			testName:       "notFound",
			name:           "notFound",
			namespace:      "fluid",
			Type:           "",
			isErr:          true,
			condtionLength: 0,
			deprecated:     false,
		}, {
			testName:       "deprecated",
			name:           "deprecated",
			namespace:      "fluid",
			Type:           "",
			isErr:          false,
			condtionLength: 0,
			deprecated:     true,
		},
	}
	for _, testCase := range testCases {
		engine := newJindoCacheEngineREP(fakeClient, testCase.name, testCase.namespace)
		err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
			Log:      fake.NullLogger(),
			Recorder: record.NewFakeRecorder(300),
		})
		if err != nil {
			if testCase.isErr {
				continue
			}
			t.Errorf("sync replicas failed,err:%s", err.Error())
		}
		rt, _ := engine.getRuntime()
		found := false
		if testCase.deprecated {
			continue
		}

		for _, cond := range rt.Status.Conditions {

			if cond.Type == testCase.Type {
				found = true
				break
			}
		}

		if !found && testCase.condtionLength > 0 {
			t.Errorf("testCase: %s runtime condition want conditionType %v, got  conditions %v", testCase.testName, testCase.Type, rt.Status.Conditions)
		}
	}
}
