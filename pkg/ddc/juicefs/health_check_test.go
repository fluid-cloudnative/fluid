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
package juicefs

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JuiceFSEngine Health Check Tests", Label("pkg.ddc.juicefs.health_check_test.go"), func() {
	var (
		juicefsruntime *datav1alpha1.JuiceFSRuntime
		dataset        *datav1alpha1.Dataset
		engine         *JuiceFSEngine
		client         client.Client
		resources      []runtime.Object
	)

	BeforeEach(func() {
		juicefsruntime = &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.JuiceFSRuntimeSpec{
				Replicas: 1,
			},
			Status: datav1alpha1.RuntimeStatus{
				CacheStates: map[common.CacheStateName]string{
					common.Cached: "true",
				},
			},
		}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{},
			Status: datav1alpha1.DatasetStatus{
				HCFSStatus: &datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			},
		}

		engine = &JuiceFSEngine{
			Log:       fake.NullLogger(),
			namespace: "fluid",
			name:      "hbase",
			runtime:   juicefsruntime,
			Recorder:  record.NewFakeRecorder(1),
		}

		resources = []runtime.Object{dataset, juicefsruntime}
	})

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(testScheme, resources...)
		engine.Client = client
		runtimeInfo, _ := base.BuildRuntimeInfo(engine.name, engine.namespace, common.JuiceFSRuntime)
		engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
	})

	Describe("Test JuiceFSEngine.CheckRuntimeHealthy()", func() {
		When("all components are healthy", func() {
			BeforeEach(func() {
				workerSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker",
						Namespace: "fluid",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:          1,
						ReadyReplicas:     1,
						AvailableReplicas: 1,
					},
				}

				fuseDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-fuse",
						Namespace: "fluid",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				}

				resources = append(resources, workerSts, fuseDaemonSet)
			})

			It("Should update runtime and dataset status to healthy", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).To(BeNil())

				// Check runtime status
				gotRuntime, err := engine.getRuntime()
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.WorkerNumberReady).To(Equal(int32(1)))
				Expect(gotRuntime.Status.WorkerNumberAvailable).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberUnavailable).To(Equal(int32(0)))

				// Check runtime conditions
				_, cond := utils.GetRuntimeCondition(gotRuntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
				Expect(cond).NotTo(BeNil())
				_, cond = utils.GetRuntimeCondition(gotRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())

				// Check dataset status
				var datasets datav1alpha1.DatasetList
				err = client.List(context.TODO(), &datasets)
				Expect(err).To(BeNil())
				Expect(datasets.Items).To(HaveLen(1))
				Expect(datasets.Items[0].Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
				Expect(datasets.Items[0].Status.CacheStates[common.Cached]).To(Equal("true"))
				Expect(datasets.Items[0].Status.HCFSStatus).NotTo(BeNil())
				Expect(datasets.Items[0].Status.HCFSStatus.Endpoint).To(Equal("test Endpoint"))
				Expect(datasets.Items[0].Status.HCFSStatus.UnderlayerFileSystemVersion).To(Equal("Underlayer HCFS Compatible Version"))
			})
		})

		When("worker is not ready", func() {
			BeforeEach(func() {
				workerSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-worker",
						Namespace: "fluid",
					},
					Spec: appsv1.StatefulSetSpec{
						Replicas: ptr.To[int32](1),
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:          1,
						ReadyReplicas:     0,
						AvailableReplicas: 0,
					},
				}

				juicefsruntime.Name = "test"
				dataset.Name = "test"
				engine.name = "test"
				engine.runtime = juicefsruntime

				resources = []runtime.Object{dataset, juicefsruntime, workerSts}
			})

			It("Should return error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(BeNil())
			})
		})
	})

	Describe("Test JuiceFSEngine.checkFuseHealthy()", func() {
		When("fuse has unavailable pods", func() {
			BeforeEach(func() {
				fuseDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-fuse",
						Namespace: "fluid",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 1,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				}

				resources = append(resources, fuseDaemonSet)
			})

			It("Should update runtime status with unavailable count", func() {
				_, err := engine.checkFuseHealthy()
				Expect(err).To(BeNil())

				// Check runtime status
				gotRuntime, err := engine.getRuntime()
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberUnavailable).To(Equal(int32(1)))

				// Check runtime condition
				_, cond := utils.GetRuntimeCondition(gotRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())
			})
		})

		When("fuse has no unavailable pods", func() {
			BeforeEach(func() {
				fuseDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "spark-fuse",
						Namespace: "fluid",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				}

				juicefsruntime.Name = "spark"
				dataset.Name = "spark"
				engine.name = "spark"
				engine.runtime = juicefsruntime

				resources = []runtime.Object{dataset, juicefsruntime, fuseDaemonSet}
			})

			It("Should update runtime status correctly", func() {
				_, err := engine.checkFuseHealthy()
				Expect(err).To(BeNil())

				// Check runtime status
				gotRuntime, err := engine.getRuntime()
				Expect(err).To(BeNil())
				Expect(gotRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(gotRuntime.Status.FuseNumberUnavailable).To(Equal(int32(0)))

				// Check runtime condition
				_, cond := utils.GetRuntimeCondition(gotRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())
			})
		})
	})
})
