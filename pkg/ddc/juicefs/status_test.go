package juicefs

import (
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/smartystreets/goconvey/convey"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func TestJuiceFSEngine_CheckAndUpdateRuntimeStatus(t *testing.T) {
	Convey("Test CheckAndUpdateRuntimeStatus ", t, func() {
		Convey("CheckAndUpdateRuntimeStatus success", func() {
			var engine *JuiceFSEngine
			patch1 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfDaemonset",
				func(_ *JuiceFSEngine, dsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfDaemonSet()
					return r, nil
				})
			defer patch1.Reset()
			patch2 := ApplyMethod(reflect.TypeOf(engine), "GetPodMetrics",
				func(_ *JuiceFSEngine, podName, containerName string) (string, error) {
					return mockJuiceFSMetric(), nil
				})
			defer patch2.Reset()
			patch3 := ApplyMethod(reflect.TypeOf(engine), "GetRunningPodsOfStatefulSet",
				func(_ *JuiceFSEngine, stsName string, namespace string) ([]corev1.Pod, error) {
					r := mockRunningPodsOfStatefulSet()
					return r, nil
				})
			defer patch3.Reset()

			var workerInputs = []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs1-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      1,
						ReadyReplicas: 1,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs2-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:      2,
						ReadyReplicas: 2,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-fuse-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
				},
			}

			var fuseInputs = []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs1-fuse",
						Namespace: "fluid",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs2-fuse",
						Namespace: "fluid",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-worker-fuse",
						Namespace: "fluid",
					},
				},
			}

			runtimeInputs := []*datav1alpha1.JuiceFSRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs1",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
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
							utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
							utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
						},
						WorkerPhase: "NotReady",
						FusePhase:   "NotReady",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs2",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
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
							utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
							utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
						},
						WorkerPhase: "NotReady",
						FusePhase:   "NotReady",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-worker",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
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
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "no-fuse",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
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
				},
			}

			objs := []runtime.Object{}

			for _, workerInput := range workerInputs {
				objs = append(objs, workerInput.DeepCopy())
			}

			for _, runtimeInput := range runtimeInputs {
				objs = append(objs, runtimeInput.DeepCopy())
			}

			for _, fuseInput := range fuseInputs {
				objs = append(objs, fuseInput.DeepCopy())
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

			testCases := []struct {
				testName   string
				name       string
				namespace  string
				isErr      bool
				deprecated bool
			}{
				{
					testName:  "juicefs1",
					name:      "juicefs1",
					namespace: "fluid",
				},
				{
					testName:  "juicefs2",
					name:      "juicefs2",
					namespace: "fluid",
				},
				{
					testName:  "no-fuse",
					name:      "no-fuse",
					namespace: "fluid",
					isErr:     true,
				},
				{
					testName:  "no-worker",
					name:      "no-worker",
					namespace: "fluid",
					isErr:     true,
				},
			}

			for _, testCase := range testCases {
				engine := newJuiceFSEngineREP(fakeClient, testCase.name, testCase.namespace)

				_, err := engine.CheckAndUpdateRuntimeStatus()
				if err != nil && !testCase.isErr {
					t.Errorf("testcase %s Failed due to %v", testCase.testName, err)
				}
			}
		})
	})

}
