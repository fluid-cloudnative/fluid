package jindo

import (
	"testing"

	v1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func newJindoEngineREP(client client.Client, name string, namespace string) *JindoEngine {

	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, "jindo", v1alpha1.TieredStore{})
	engine := &JindoEngine{
		runtime:     &v1alpha1.JindoRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         log.NullLogger{},
	}
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
	runtimeInputs := []*v1alpha1.JindoRuntime{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: v1alpha1.JindoRuntimeSpec{
				Replicas: 3, // 2
			},
			Status: v1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2, // 0
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 3,
				DesiredWorkerNumberScheduled: 3,
				DesiredFuseNumberScheduled:   3,
				Conditions: []v1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
					utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
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
			Spec: v1alpha1.JindoRuntimeSpec{
				Replicas: 2,
			},
			Status: v1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 3,
				CurrentMasterNumberScheduled: 3,
				CurrentFuseNumberScheduled:   3,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				Conditions: []v1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
					utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
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
			Spec: v1alpha1.JindoRuntimeSpec{
				Replicas: 2,
			},
			Status: v1alpha1.RuntimeStatus{
				CurrentWorkerNumberScheduled: 2,
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				DesiredWorkerNumberScheduled: 2,
				DesiredFuseNumberScheduled:   2,
				Conditions: []v1alpha1.RuntimeCondition{
					utils.NewRuntimeCondition(v1alpha1.RuntimeWorkersInitialized, v1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
					utils.NewRuntimeCondition(v1alpha1.RuntimeFusesInitialized, v1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
				},
				WorkerPhase: "NotReady",
				FusePhase:   "NotReady",
			},
		},
	}
	statefulSetInputs := []*appsv1.StatefulSet{
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
	for _, fuseInput := range fuseInputs {
		objs = append(objs, fuseInput.DeepCopy())
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
		engine := newJindoEngineREP(fakeClient, testCase.name, testCase.namespace)
		err := engine.SyncReplicas(cruntime.ReconcileRequestContext{
			Log:      log.NullLogger{},
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
