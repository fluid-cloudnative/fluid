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
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
)

var _ = Describe("ThinEngine Health Check Tests", Label("pkg.ddc.thin.health_check_test.go"), func() {
	Describe("CheckRuntimeHealthy", func() {
		var (
			stsInputs         []appsv1.StatefulSet
			daemonSetInputs   []appsv1.DaemonSet
			ThinRuntimeInputs []datav1alpha1.ThinRuntime
			datasetInputs     []*datav1alpha1.Dataset
			testObjs          []runtime.Object
			testClient        client.Client
			engines           []ThinEngine
		)

		BeforeEach(func() {
			stsInputs = []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker",
						Namespace: "fluid",
					},
					Status: appsv1.StatefulSetStatus{
						Replicas:          1,
						ReadyReplicas:     1,
						AvailableReplicas: 1,
					},
				},
				{
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
				},
			}

			daemonSetInputs = []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-fuse",
						Namespace: "fluid",
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				},
			}

			ThinRuntimeInputs = []datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 1,
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
					},
					Status: datav1alpha1.RuntimeStatus{
						CacheStates: map[common.CacheStateName]string{common.Cached: "true"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						Replicas: 1,
						Worker: datav1alpha1.ThinCompTemplateSpec{
							Enabled: true,
						},
					},
					Status: datav1alpha1.RuntimeStatus{
						CacheStates: map[common.CacheStateName]string{common.Cached: "true"},
					},
				},
			}

			datasetInputs = []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec:   datav1alpha1.DatasetSpec{},
					Status: datav1alpha1.DatasetStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "fluid",
					},
					Spec:   datav1alpha1.DatasetSpec{},
					Status: datav1alpha1.DatasetStatus{},
				},
			}

			testObjs = []runtime.Object{}
			for _, daemonSet := range daemonSetInputs {
				testObjs = append(testObjs, daemonSet.DeepCopy())
			}
			for _, sts := range stsInputs {
				testObjs = append(testObjs, sts.DeepCopy())
			}
			for _, ThinRuntime := range ThinRuntimeInputs {
				testObjs = append(testObjs, ThinRuntime.DeepCopy())
			}
			for _, dataset := range datasetInputs {
				testObjs = append(testObjs, dataset.DeepCopy())
			}

			testClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

			engines = []ThinEngine{
				{
					Client:    testClient,
					Log:       fake.NullLogger(),
					namespace: "fluid",
					name:      "hbase",
					runtime:   &ThinRuntimeInputs[0],
				},
				{
					Client:    testClient,
					Log:       fake.NullLogger(),
					namespace: "fluid",
					name:      "test",
					runtime:   &ThinRuntimeInputs[1],
				},
			}
		})

		Context("when runtime is healthy", func() {
			It("should succeed and update runtime status correctly", func() {
				engine := engines[0]
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, testClient, engine.Log)

				err = engine.CheckRuntimeHealthy()
				Expect(err).NotTo(HaveOccurred())

				ThinRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(ThinRuntime.Status.WorkerNumberReady).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.WorkerNumberAvailable).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))

				_, cond := utils.GetRuntimeCondition(ThinRuntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
				Expect(cond).NotTo(BeNil())

				_, cond = utils.GetRuntimeCondition(ThinRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())

				var datasets datav1alpha1.DatasetList
				err = testClient.List(context.TODO(), &datasets)
				Expect(err).NotTo(HaveOccurred())
				Expect(datasets.Items[0].Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
				Expect(datasets.Items[0].Status.CacheStates[common.Cached]).To(Equal("true"))
			})
		})

		Context("when runtime is unhealthy", func() {
			It("should return error and update runtime status correctly", func() {
				engine := engines[1]
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, testClient, engine.Log)

				err = engine.CheckRuntimeHealthy()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("checkFuseHealthy", func() {
		var (
			daemonSetInputs   []appsv1.DaemonSet
			ThinRuntimeInputs []datav1alpha1.ThinRuntime
			testObjs          []runtime.Object
			testClient        client.Client
			engines           []ThinEngine
		)

		BeforeEach(func() {
			daemonSetInputs = []appsv1.DaemonSet{
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

			ThinRuntimeInputs = []datav1alpha1.ThinRuntime{
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

			testObjs = []runtime.Object{}
			for _, daemonSet := range daemonSetInputs {
				testObjs = append(testObjs, daemonSet.DeepCopy())
			}
			for _, ThinRuntimeInput := range ThinRuntimeInputs {
				testObjs = append(testObjs, ThinRuntimeInput.DeepCopy())
			}

			testClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

			engines = []ThinEngine{
				{
					Client:    testClient,
					Log:       fake.NullLogger(),
					namespace: "fluid",
					name:      "hbase",
					runtime: &datav1alpha1.ThinRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase",
							Namespace: "fluid",
						},
					},
					Recorder: record.NewFakeRecorder(1),
				},
				{
					Client:    testClient,
					Log:       fake.NullLogger(),
					namespace: "fluid",
					name:      "spark",
					runtime: &datav1alpha1.ThinRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "spark",
							Namespace: "fluid",
						},
					},
					Recorder: record.NewFakeRecorder(1),
				},
			}
		})

		Context("when fuse has unavailable pods", func() {
			It("should still return ready and update status correctly", func() {
				engine := engines[0]
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, testClient, engine.Log)

				ready, err := engine.checkFuseHealthy()
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())

				ThinRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(ThinRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.FuseNumberUnavailable).To(Equal(int32(1)))

				_, cond := utils.GetRuntimeCondition(ThinRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())
			})
		})

		Context("when fuse is healthy", func() {
			It("should return Ready phase and update status correctly", func() {
				engine := engines[1]
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, testClient, engine.Log)

				ready, err := engine.checkFuseHealthy()
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())

				ThinRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(ThinRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(ThinRuntime.Status.FuseNumberUnavailable).To(Equal(int32(0)))

				_, cond := utils.GetRuntimeCondition(ThinRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())
			})
		})
	})
})
