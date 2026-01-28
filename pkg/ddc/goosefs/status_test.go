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

package goosefs

import (
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testStatusNamespace           = "fluid"
	testStatusRuntimeHadoop       = "hadoop"
	testStatusRuntimeHbase        = "hbase"
	testStatusRuntimeNoWorker     = "no-worker"
	testStatusRuntimeNoMaster     = "no-master"
	testStatusRuntimeZeroReplicas = "zero-replicas"
	testStatusPhaseNotReady       = "NotReady"
	testStatusUfsTotal            = "19.07MiB"
)

func newGooseFSEngineForStatus(c client.Client, name string, namespace string) *GooseFSEngine {
	runTimeInfo, err := base.BuildRuntimeInfo(name, namespace, common.GooseFSRuntime)
	Expect(err).NotTo(HaveOccurred())
	engine := &GooseFSEngine{
		runtime:     &datav1alpha1.GooseFSRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      c,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	engine.Helper = ctrl.BuildHelper(runTimeInfo, c, engine.Log)
	return engine
}

var _ = Describe("GooseFSEngine Runtime Status Tests", Label("pkg.ddc.goosefs.status_test.go"), func() {
	var patches *gomonkey.Patches

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	Describe("CheckAndUpdateRuntimeStatus", func() {
		Context("when master and worker are all ready", func() {
			var (
				engine     *GooseFSEngine
				fakeClient client.Client
			)

			BeforeEach(func() {
				masterInputs := []*appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeHadoop + "-master",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](1),
						},
						Status: appsv1.StatefulSetStatus{
							ReadyReplicas: 1,
						},
					},
				}

				workerInputs := []appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeHadoop + "-worker",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](3),
						},
						Status: appsv1.StatefulSetStatus{
							Replicas:      3,
							ReadyReplicas: 2,
						},
					},
				}

				runtimeInputs := []*datav1alpha1.GooseFSRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeHadoop,
							Namespace: testStatusNamespace,
						},
						Spec: datav1alpha1.GooseFSRuntimeSpec{
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
							WorkerPhase: testStatusPhaseNotReady,
							FusePhase:   testStatusPhaseNotReady,
						},
					},
				}

				objs := []runtime.Object{}
				for _, masterInput := range masterInputs {
					objs = append(objs, masterInput.DeepCopy())
				}

				for _, workerInput := range workerInputs {
					objs = append(objs, workerInput.DeepCopy())
				}

				for _, runtimeInput := range runtimeInputs {
					objs = append(objs, runtimeInput.DeepCopy())
				}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
				engine = newGooseFSEngineForStatus(fakeClient, testStatusRuntimeHadoop, testStatusNamespace)
			})

			It("should return ready with no error", func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
					func(_ *GooseFSEngine) (string, error) {
						summary := mockGooseFSReportSummary()
						return summary, nil
					})

				patches.ApplyFunc(utils.GetDataset,
					func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
						d := &datav1alpha1.Dataset{
							Status: datav1alpha1.DatasetStatus{
								UfsTotal: testStatusUfsTotal,
							},
						}
						return d, nil
					})

				patches.ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
					func(_ *GooseFSEngine) cacheHitStates {
						return cacheHitStates{
							bytesReadLocal:  20310917,
							bytesReadUfsAll: 32243712,
						}
					})

				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())
			})
		})

		Context("when master is not ready", func() {
			var (
				engine     *GooseFSEngine
				fakeClient client.Client
			)

			BeforeEach(func() {
				masterInputs := []*appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeHbase + "-master",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](1),
						},
						Status: appsv1.StatefulSetStatus{
							ReadyReplicas: 0,
						},
					},
				}

				workerInputs := []appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeHbase + "-worker",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](2),
						},
						Status: appsv1.StatefulSetStatus{
							Replicas:      2,
							ReadyReplicas: 2,
						},
					},
				}

				runtimeInputs := []*datav1alpha1.GooseFSRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeHbase,
							Namespace: testStatusNamespace,
						},
						Spec: datav1alpha1.GooseFSRuntimeSpec{
							Replicas: 2,
						},
						Status: datav1alpha1.RuntimeStatus{
							CurrentWorkerNumberScheduled: 2,
							CurrentMasterNumberScheduled: 2,
							CurrentFuseNumberScheduled:   2,
							DesiredMasterNumberScheduled: 2,
							DesiredWorkerNumberScheduled: 2,
							DesiredFuseNumberScheduled:   2,
							Conditions: []datav1alpha1.RuntimeCondition{
								utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason, "The workers are initialized.", corev1.ConditionTrue),
								utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason, "The fuses are initialized.", corev1.ConditionTrue),
							},
							WorkerPhase: testStatusPhaseNotReady,
							FusePhase:   testStatusPhaseNotReady,
						},
					},
				}

				objs := []runtime.Object{}
				for _, masterInput := range masterInputs {
					objs = append(objs, masterInput.DeepCopy())
				}

				for _, workerInput := range workerInputs {
					objs = append(objs, workerInput.DeepCopy())
				}

				for _, runtimeInput := range runtimeInputs {
					objs = append(objs, runtimeInput.DeepCopy())
				}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
				engine = newGooseFSEngineForStatus(fakeClient, testStatusRuntimeHbase, testStatusNamespace)
			})

			It("should return not ready", func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
					func(_ *GooseFSEngine) (string, error) {
						summary := mockGooseFSReportSummary()
						return summary, nil
					})

				patches.ApplyFunc(utils.GetDataset,
					func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
						d := &datav1alpha1.Dataset{
							Status: datav1alpha1.DatasetStatus{
								UfsTotal: testStatusUfsTotal,
							},
						}
						return d, nil
					})

				patches.ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
					func(_ *GooseFSEngine) cacheHitStates {
						return cacheHitStates{
							bytesReadLocal:  20310917,
							bytesReadUfsAll: 32243712,
						}
					})

				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeFalse())
			})
		})

		Context("when worker statefulset is not found", func() {
			var (
				engine     *GooseFSEngine
				fakeClient client.Client
			)

			BeforeEach(func() {
				masterInputs := []*appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeNoWorker + "-master",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](1),
						},
						Status: appsv1.StatefulSetStatus{
							ReadyReplicas: 1,
						},
					},
				}

				runtimeInputs := []*datav1alpha1.GooseFSRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeNoWorker,
							Namespace: testStatusNamespace,
						},
						Spec: datav1alpha1.GooseFSRuntimeSpec{
							Replicas: 2,
						},
						Status: datav1alpha1.RuntimeStatus{
							CurrentWorkerNumberScheduled: 2,
							CurrentMasterNumberScheduled: 2,
							CurrentFuseNumberScheduled:   2,
							DesiredMasterNumberScheduled: 2,
							DesiredWorkerNumberScheduled: 2,
							DesiredFuseNumberScheduled:   2,
							WorkerPhase:                  testStatusPhaseNotReady,
							FusePhase:                    testStatusPhaseNotReady,
						},
					},
				}

				objs := []runtime.Object{}
				for _, masterInput := range masterInputs {
					objs = append(objs, masterInput.DeepCopy())
				}
				for _, runtimeInput := range runtimeInputs {
					objs = append(objs, runtimeInput.DeepCopy())
				}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
				engine = newGooseFSEngineForStatus(fakeClient, testStatusRuntimeNoWorker, testStatusNamespace)
			})

			It("should return error", func() {
				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).To(HaveOccurred())
				Expect(ready).To(BeFalse())
			})
		})

		Context("when master statefulset is not found", func() {
			var (
				engine     *GooseFSEngine
				fakeClient client.Client
			)

			BeforeEach(func() {
				workerInputs := []appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeNoMaster + "-worker",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](2),
						},
						Status: appsv1.StatefulSetStatus{
							Replicas:      2,
							ReadyReplicas: 2,
						},
					},
				}

				runtimeInputs := []*datav1alpha1.GooseFSRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeNoMaster,
							Namespace: testStatusNamespace,
						},
						Spec: datav1alpha1.GooseFSRuntimeSpec{
							Replicas: 2,
						},
						Status: datav1alpha1.RuntimeStatus{
							CurrentWorkerNumberScheduled: 2,
							CurrentMasterNumberScheduled: 2,
							CurrentFuseNumberScheduled:   2,
							DesiredMasterNumberScheduled: 2,
							DesiredWorkerNumberScheduled: 2,
							DesiredFuseNumberScheduled:   2,
							WorkerPhase:                  testStatusPhaseNotReady,
							FusePhase:                    testStatusPhaseNotReady,
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
				fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
				engine = newGooseFSEngineForStatus(fakeClient, testStatusRuntimeNoMaster, testStatusNamespace)
			})

			It("should return error", func() {
				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).To(HaveOccurred())
				Expect(ready).To(BeFalse())
			})
		})

		Context("when runtime has zero worker replicas", func() {
			var (
				engine     *GooseFSEngine
				fakeClient client.Client
			)

			BeforeEach(func() {
				masterInputs := []*appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeZeroReplicas + "-master",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](1),
						},
						Status: appsv1.StatefulSetStatus{
							ReadyReplicas: 1,
						},
					},
				}

				workerInputs := []appsv1.StatefulSet{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeZeroReplicas + "-worker",
							Namespace: testStatusNamespace,
						},
						Spec: appsv1.StatefulSetSpec{
							Replicas: ptr.To[int32](0),
						},
						Status: appsv1.StatefulSetStatus{
							Replicas:      0,
							ReadyReplicas: 0,
						},
					},
				}

				runtimeInputs := []*datav1alpha1.GooseFSRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      testStatusRuntimeZeroReplicas,
							Namespace: testStatusNamespace,
						},
						Spec: datav1alpha1.GooseFSRuntimeSpec{
							Replicas: 0,
						},
						Status: datav1alpha1.RuntimeStatus{
							CurrentWorkerNumberScheduled: 0,
							CurrentMasterNumberScheduled: 1,
							CurrentFuseNumberScheduled:   0,
							DesiredMasterNumberScheduled: 1,
							DesiredWorkerNumberScheduled: 0,
							DesiredFuseNumberScheduled:   0,
							WorkerPhase:                  testStatusPhaseNotReady,
							FusePhase:                    testStatusPhaseNotReady,
						},
					},
				}

				objs := []runtime.Object{}
				for _, masterInput := range masterInputs {
					objs = append(objs, masterInput.DeepCopy())
				}
				for _, workerInput := range workerInputs {
					objs = append(objs, workerInput.DeepCopy())
				}
				for _, runtimeInput := range runtimeInputs {
					objs = append(objs, runtimeInput.DeepCopy())
				}
				fakeClient = fake.NewFakeClientWithScheme(testScheme, objs...)
				engine = newGooseFSEngineForStatus(fakeClient, testStatusRuntimeZeroReplicas, testStatusNamespace)
			})

			It("should return ready when master is ready and zero workers are expected", func() {
				patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetReportSummary",
					func(_ *GooseFSEngine) (string, error) {
						summary := mockGooseFSReportSummary()
						return summary, nil
					})

				patches.ApplyFunc(utils.GetDataset,
					func(_ client.Reader, _ string, _ string) (*datav1alpha1.Dataset, error) {
						d := &datav1alpha1.Dataset{
							Status: datav1alpha1.DatasetStatus{
								UfsTotal: testStatusUfsTotal,
							},
						}
						return d, nil
					})

				patches.ApplyMethod(reflect.TypeOf(engine), "GetCacheHitStates",
					func(_ *GooseFSEngine) cacheHitStates {
						return cacheHitStates{
							bytesReadLocal:  0,
							bytesReadUfsAll: 0,
						}
					})

				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())
			})
		})
	})
})
