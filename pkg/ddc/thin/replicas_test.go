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

package thin

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	ctrlhelper "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newThinEngineREP(client client.Client, name string, namespace string) *ThinEngine {

	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.ThinRuntime)
	engine := &ThinEngine{
		runtime:     &v1alpha1.ThinRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	return engine
}

func TestSyncReplicas(t *testing.T) {
	nodeInputs := []*corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-spark",
				Labels: map[string]string{
					"fluid.io/dataset-num":            "1",
					"fluid.io/s-thin-fluid-spark":     "true",
					"fluid.io/s-fluid-spark":          "true",
					"fluid.io/s-h-thin-d-fluid-spark": "5B",
					"fluid.io/s-h-thin-m-fluid-spark": "1B",
					"fluid.io/s-h-thin-t-fluid-spark": "6B",
					"fluid_exclusive":                 "fluid_spark",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-share",
				Labels: map[string]string{
					"fluid.io/dataset-num":             "2",
					"fluid.io/s-thin-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":          "true",
					"fluid.io/s-h-thin-d-fluid-hadoop": "5B",
					"fluid.io/s-h-thin-m-fluid-hadoop": "1B",
					"fluid.io/s-h-thin-t-fluid-hadoop": "6B",
					"fluid.io/s-thin-fluid-hbase":      "true",
					"fluid.io/s-fluid-hbase":           "true",
					"fluid.io/s-h-thin-d-fluid-hbase":  "5B",
					"fluid.io/s-h-thin-m-fluid-hbase":  "1B",
					"fluid.io/s-h-thin-t-fluid-hbase":  "6B",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node-hadoop",
				Labels: map[string]string{
					"fluid.io/dataset-num":             "1",
					"fluid.io/s-thin-fluid-hadoop":     "true",
					"fluid.io/s-fluid-hadoop":          "true",
					"fluid.io/s-h-thin-d-fluid-hadoop": "5B",
					"fluid.io/s-h-thin-m-fluid-hadoop": "1B",
					"fluid.io/s-h-thin-t-fluid-hadoop": "6B",
					"node-select":                      "true",
				},
			},
		},
	}
	runtimeInputs := []*v1alpha1.ThinRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: v1alpha1.ThinRuntimeSpec{
				Replicas: 3, // 2
			},
			Status: v1alpha1.RuntimeStatus{
				DesiredWorkerNumberScheduled: 2,
				Conditions: []v1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
					utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop",
				Namespace: "fluid",
			},
			Spec: v1alpha1.ThinRuntimeSpec{
				Replicas: 1,
			},
			Status: v1alpha1.RuntimeStatus{
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				Conditions: []v1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
					utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj",
				Namespace: "fluid",
			},
			Spec: v1alpha1.ThinRuntimeSpec{
				Replicas: 2,
			},
			Status: v1alpha1.RuntimeStatus{
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				Conditions: []v1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
					utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
				},
			},
		},
	}
	statefulSetInputs := []*appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-worker",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-worker",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-worker",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-fuse",
				Namespace: "fluid",
			},
		},
	}
	daemonsetInputs := []*appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-fuse",
				Namespace: "fluid",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-fuse",
				Namespace: "fluid",
			},
		},
	}
	dataSetInputs := []*v1alpha1.Dataset{
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

	objs := []runtime.Object{}
	for _, nodeInput := range nodeInputs {
		objs = append(objs, nodeInput.DeepCopy())
	}
	for _, runtimeInput := range runtimeInputs {
		objs = append(objs, runtimeInput.DeepCopy())
	}
	for _, statefulSetInput := range statefulSetInputs {
		objs = append(objs, statefulSetInput.DeepCopy())
	}
	for _, daemonsetInput := range daemonsetInputs {
		objs = append(objs, daemonsetInput.DeepCopy())
	}
	for _, dataSetInput := range dataSetInputs {
		objs = append(objs, dataSetInput.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	testCases := []struct {
		name      string
		namespace string
		Type      v1alpha1.RuntimeConditionType
		isErr     bool
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			Type:      "FusesScaledOut",
			isErr:     false,
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			Type:      "FusesScaledIn",
			isErr:     false,
		},
		{
			name:      "obj",
			namespace: "fluid",
			Type:      "",
			isErr:     false,
		},
	}
	for _, testCase := range testCases {
		engine := newThinEngineREP(fakeClient, testCase.name, testCase.namespace)
		runtimeInfo, err := base.BuildRuntimeInfo(testCase.name, testCase.namespace, common.ThinRuntime)
		if err != nil {
			t.Errorf("ThinEngine.CheckWorkersReady() error = %v", err)
		}

		engine.Helper = ctrlhelper.BuildHelper(runtimeInfo, fakeClient, engine.Log)
		err = engine.SyncReplicas(cruntime.ReconcileRequestContext{
			Log:      fake.NullLogger(),
			Recorder: record.NewFakeRecorder(300),
		})
		if err != nil {
			t.Errorf("sync replicas failed,err:%s", err.Error())
		}
		rt, _ := engine.getRuntime()
		if len(rt.Status.Conditions) == 4 {
			Type := rt.Status.Conditions[3].Type
			if Type != testCase.Type {
				t.Errorf("runtime condition want %s, got %s", testCase.Type, Type)
			}
		}
	}
}
