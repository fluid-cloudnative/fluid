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

func TestThinEngine_CheckAndUpdateRuntimeStatus(t *testing.T) {
	Convey("Test CheckAndUpdateRuntimeStatus ", t, func() {
		Convey("CheckAndUpdateRuntimeStatus success", func() {
			var workerInputs = []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1-worker",
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
						Name:      "thin2-worker",
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
						Name:      "thin1-fuse",
						Namespace: "fluid",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin2-fuse",
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

			runtimeInputs := []*datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
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
						Name:      "thin2",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
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
					Spec: datav1alpha1.ThinRuntimeSpec{
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
					Spec: datav1alpha1.ThinRuntimeSpec{
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

			var datasetInputs = []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin1",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "ceph://myceph",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin2",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://my-pvc",
							},
						},
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

			for _, datasetInput := range datasetInputs {
				objs = append(objs, datasetInput)
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
					testName:  "thin1",
					name:      "thin1",
					namespace: "fluid",
				},
				{
					testName:  "thin2",
					name:      "thin2",
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
				engine := newThinEngineREP(fakeClient, testCase.name, testCase.namespace)

				_, err := engine.CheckAndUpdateRuntimeStatus()
				if err != nil && !testCase.isErr {
					t.Errorf("testcase %s Failed due to %v", testCase.testName, err)
				}
			}
		})
	})

}
