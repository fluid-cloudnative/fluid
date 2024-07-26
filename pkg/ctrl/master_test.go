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
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	utilpointer "k8s.io/utils/pointer"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestCheckMasterHealthy(t *testing.T) {
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

				CurrentMasterNumberScheduled: 1, // 0
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 1,
				MasterNumberReady:            1,
				DesiredFuseNumberScheduled:   3,
				MasterPhase:                  "NotReady",
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
				CurrentMasterNumberScheduled: 0,
				CurrentFuseNumberScheduled:   3,
				DesiredMasterNumberScheduled: 1,
				DesiredFuseNumberScheduled:   2,
				MasterPhase:                  "NotReady",
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
				CurrentMasterNumberScheduled: 2,
				CurrentFuseNumberScheduled:   2,
				DesiredMasterNumberScheduled: 2,
				MasterNumberReady:            2,
				DesiredFuseNumberScheduled:   2,
				MasterPhase:                  "NotReady",
				FusePhase:                    "NotReady",
			},
		},
	}

	podList := &corev1.PodList{
		Items: []corev1.Pod{{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase-jindofs-master-0",
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
				Name:      "hbase-jindofs-master",
				Namespace: "big-data",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
				Replicas:      1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hadoop-jindofs-master",
				Namespace: "fluid",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: utilpointer.Int32(3),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 0,
				Replicas:      1,
			},
		}, {
			ObjectMeta: metav1.ObjectMeta{
				Name:      "obj-jindofs-master",
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
		master    *appsv1.StatefulSet
		TypeValue bool
		isErr     bool
	}{
		{
			caseName:  "Healthy",
			name:      "hbase",
			namespace: "fluid",
			master: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase-jindofs-master",
					Namespace: "big-data",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(2),
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
			master: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hadoop-jindofs-master",
					Namespace: "fluid",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(3),
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
			Namespace: testCase.master.Namespace,
			Name:      testCase.master.Name,
		}, statefulset)
		if err != nil {
			t.Errorf("sync replicas failed,err:%s", err.Error())
		}

		h := BuildHelper(runtimeInfo, fakeClient, fake.NullLogger())

		err = h.CheckMasterHealthy(record.NewFakeRecorder(300),
			runtime, runtime.Status, statefulset)

		if testCase.isErr == (err == nil) {
			t.Errorf("check master's healthy failed,err:%v", err)
		}

		err = fakeClient.Get(context.TODO(), types.NamespacedName{
			Namespace: testCase.namespace,
			Name:      testCase.name,
		}, runtime)

		if err != nil {
			t.Errorf("check master's healthy failed,err:%s", err.Error())
		}

		if runtime.Status.MasterPhase != testCase.Phase {
			t.Errorf("testcase %s is failed, expect phase %v, got %v", testCase.caseName,
				testCase.Phase,
				runtime.Status.MasterPhase)
		}

	}
}

func Test_recheckMasterHealthyByEachContainerStartedTime(t *testing.T) {
	startT1, _ := time.Parse(time.RFC3339, "2023-08-07T20:17:08+08:00")
	startT2, _ := time.Parse(time.RFC3339, "2023-08-07T20:18:08+08:00")

	testCases := []struct {
		name        string
		sts         *appsv1.StatefulSet
		pods        []*corev1.Pod
		expectedSts *appsv1.StatefulSet
		ready       bool
	}{
		{
			name: "check health at first shot",
			sts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ds-test-master-0",
						Namespace: "default",
						Labels:    map[string]string{"foo": "bar"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "main",
								Image: "busybox",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "main",
								State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.NewTime(startT1),
								}},
							},
						},
					},
				},
			},
			expectedSts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT1.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			ready: false,
		},
		{
			name: "check health when master recreated",
			sts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT1.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ds-test-master-0",
						Namespace: "default",
						Labels:    map[string]string{"foo": "bar"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "main",
								Image: "busybox",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name:  "main",
								State: corev1.ContainerState{},
							},
						},
					},
				},
			},
			expectedSts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT1.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			ready: false,
		},
		{
			name: "check health when master just started",
			sts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT1.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ds-test-master-0",
						Namespace: "default",
						Labels:    map[string]string{"foo": "bar"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "main",
								Image: "busybox",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "main",
								State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.NewTime(startT2),
								}},
							},
						},
					},
				},
			},
			expectedSts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT2.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			ready: true,
		},
		{
			name: "recheck health when master has started before",
			sts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT2.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			pods: []*corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ds-test-master-0",
						Namespace: "default",
						Labels:    map[string]string{"foo": "bar"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "main",
								Image: "busybox",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: "main",
								State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.NewTime(startT2),
								}},
							},
						},
					},
				},
			},
			expectedSts: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ds-test-master",
					Namespace: "default",
					Annotations: map[string]string{
						common.AnnotationLatestMasterStartedTime: startT2.Format(time.RFC3339),
					},
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: utilpointer.Int32(1),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{
						"foo": "bar",
					}},
				},
			},
			ready: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fb := fakeclient.NewClientBuilder().WithScheme(scheme.Scheme)
			if testCase.sts != nil {
				fb.WithObjects(testCase.sts)
			}
			if len(testCase.pods) > 0 {
				for pi := range testCase.pods {
					fb.WithObjects(testCase.pods[pi])
				}
			}

			h := Helper{client: fb.Build()}
			ready, _, err := h.recheckMasterHealthyByEachContainerStartedTime(testCase.sts)
			if err != nil {
				t.Error(err)
			}
			if ready != testCase.ready {
				t.Errorf("unexpected ready state, expected: %v, got: %v", testCase.ready, ready)
			}
			latestSts := appsv1.StatefulSet{}
			_ = h.client.Get(context.Background(), types.NamespacedName{Name: testCase.sts.Name, Namespace: testCase.sts.Namespace}, &latestSts)
			if !reflect.DeepEqual(latestSts.Annotations, testCase.expectedSts.Annotations) {
				t.Errorf("unexpected updated sts, expected: %+v, got: %+v", testCase.expectedSts.Annotations, latestSts.Annotations)
			}
		})
	}
}
