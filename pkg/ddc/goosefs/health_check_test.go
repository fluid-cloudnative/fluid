package goosefs

import (
	"fmt"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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
			Log:       log.NullLogger{},
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
			Log:       log.NullLogger{},
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
		fmt.Println(goosefsruntime)
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
	var daemonSetInputs = []appsv1.DaemonSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-worker",
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
				Name:      "spark-worker",
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
			Log:       log.NullLogger{},
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
			Log:       log.NullLogger{},
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
			expectedRuntimeWorkerNumberReady: 1,
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
			Log:       log.NullLogger{},
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
			Log:       log.NullLogger{},
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
