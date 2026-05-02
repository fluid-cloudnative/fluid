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

const (
	healthCheckTestNamespace = "fluid"
	healthCheckTestHbase     = "hbase"
	healthCheckTestSpark     = "spark"
	healthCheckTestName      = "test"
)

var _ = Describe("ThinEngine Health Check", Label("pkg.ddc.thin.health_check_test.go"), func() {
	Describe("CheckRuntimeHealthy", func() {
		var (
			stsInputs         []appsv1.StatefulSet
			daemonSetInputs   []appsv1.DaemonSet
			thinRuntimeInputs []datav1alpha1.ThinRuntime
			datasetInputs     []*datav1alpha1.Dataset
		)

		BeforeEach(func() {
			stsInputs = []appsv1.StatefulSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-worker",
						Namespace: healthCheckTestNamespace,
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
						Namespace: healthCheckTestNamespace,
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
						Namespace: healthCheckTestNamespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				},
			}

			thinRuntimeInputs = []datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestHbase,
						Namespace: healthCheckTestNamespace,
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
						Name:      healthCheckTestName,
						Namespace: healthCheckTestNamespace,
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
						Name:      healthCheckTestHbase,
						Namespace: healthCheckTestNamespace,
					},
					Spec:   datav1alpha1.DatasetSpec{},
					Status: datav1alpha1.DatasetStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestName,
						Namespace: healthCheckTestNamespace,
					},
					Spec:   datav1alpha1.DatasetSpec{},
					Status: datav1alpha1.DatasetStatus{},
				},
			}
		})

		Context("when workers and fuses are healthy", func() {
			var (
				client client.Client
				engine ThinEngine
			)

			BeforeEach(func() {
				testObjs := []runtime.Object{}
				for _, daemonSet := range daemonSetInputs {
					testObjs = append(testObjs, daemonSet.DeepCopy())
				}
				for _, sts := range stsInputs {
					testObjs = append(testObjs, sts.DeepCopy())
				}
				for _, thinRuntime := range thinRuntimeInputs {
					testObjs = append(testObjs, thinRuntime.DeepCopy())
				}
				for _, dataset := range datasetInputs {
					testObjs = append(testObjs, dataset.DeepCopy())
				}

				client = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine = ThinEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					namespace: healthCheckTestNamespace,
					name:      healthCheckTestHbase,
					runtime:   &thinRuntimeInputs[0],
				}
			})

			JustBeforeEach(func() {
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
			})

			It("should return no error and update runtime status", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(HaveOccurred())

				thinRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(thinRuntime.Status.WorkerNumberReady).To(Equal(int32(1)))
				Expect(thinRuntime.Status.WorkerNumberAvailable).To(Equal(int32(1)))
				Expect(thinRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(thinRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))

				_, cond := utils.GetRuntimeCondition(thinRuntime.Status.Conditions, datav1alpha1.RuntimeWorkersReady)
				Expect(cond).NotTo(BeNil())

				_, cond = utils.GetRuntimeCondition(thinRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())

				var datasets datav1alpha1.DatasetList
				err = client.List(context.TODO(), &datasets)
				Expect(err).NotTo(HaveOccurred())

				// Find the hbase dataset by name instead of relying on list order
				var hbaseDataset *datav1alpha1.Dataset
				for i := range datasets.Items {
					if datasets.Items[i].Name == healthCheckTestHbase {
						hbaseDataset = &datasets.Items[i]
						break
					}
				}
				Expect(hbaseDataset).NotTo(BeNil())
				Expect(hbaseDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
				Expect(hbaseDataset.Status.CacheStates).To(Equal(common.CacheStateList{common.Cached: "true"}))
			})
		})

		Context("when workers are not ready", func() {
			var (
				client client.Client
				engine ThinEngine
			)

			BeforeEach(func() {
				testObjs := []runtime.Object{}
				for _, daemonSet := range daemonSetInputs {
					testObjs = append(testObjs, daemonSet.DeepCopy())
				}
				for _, sts := range stsInputs {
					testObjs = append(testObjs, sts.DeepCopy())
				}
				for _, thinRuntime := range thinRuntimeInputs {
					testObjs = append(testObjs, thinRuntime.DeepCopy())
				}
				for _, dataset := range datasetInputs {
					testObjs = append(testObjs, dataset.DeepCopy())
				}

				client = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine = ThinEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					namespace: healthCheckTestNamespace,
					name:      healthCheckTestName,
					runtime:   &thinRuntimeInputs[1],
				}
			})

			JustBeforeEach(func() {
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
			})

			It("should return an error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when a fuse-only runtime is healthy", func() {
			var (
				client client.Client
				engine ThinEngine
			)

			BeforeEach(func() {
				healthyFuse := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestSpark + "-fuse",
						Namespace: healthCheckTestNamespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				}
				runtimeObj := &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestSpark,
						Namespace: healthCheckTestNamespace,
					},
					Spec: datav1alpha1.ThinRuntimeSpec{},
					Status: datav1alpha1.RuntimeStatus{
						CacheStates: map[common.CacheStateName]string{common.Cached: "true"},
					},
				}
				datasetObj := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestSpark,
						Namespace: healthCheckTestNamespace,
					},
					Status: datav1alpha1.DatasetStatus{},
				}

				client = fake.NewFakeClientWithScheme(testScheme, healthyFuse, runtimeObj, datasetObj)
				engine = ThinEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					namespace: healthCheckTestNamespace,
					name:      healthCheckTestSpark,
					runtime:   runtimeObj,
				}
			})

			JustBeforeEach(func() {
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
			})

			It("skips worker checks and updates the dataset to bound", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).NotTo(HaveOccurred())

				updatedRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(updatedRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))

				updatedDataset, err := utils.GetDataset(client, healthCheckTestSpark, healthCheckTestNamespace)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})

		Context("when checking a fuse-only runtime without a fuse daemonset", func() {
			var (
				client client.Client
				engine ThinEngine
			)

			BeforeEach(func() {
				runtimeObj := &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestSpark,
						Namespace: healthCheckTestNamespace,
					},
					Spec: datav1alpha1.ThinRuntimeSpec{},
				}
				datasetObj := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestSpark,
						Namespace: healthCheckTestNamespace,
					},
				}

				client = fake.NewFakeClientWithScheme(testScheme, runtimeObj, datasetObj)
				engine = ThinEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					namespace: healthCheckTestNamespace,
					name:      healthCheckTestSpark,
					runtime:   runtimeObj,
				}
			})

			JustBeforeEach(func() {
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
			})

			It("returns the fuse check error", func() {
				err := engine.CheckRuntimeHealthy()
				Expect(err).To(HaveOccurred())
			})
		})

	})

	Describe("checkFuseHealthy", func() {
		var (
			daemonSetInputs   []appsv1.DaemonSet
			thinRuntimeInputs []datav1alpha1.ThinRuntime
		)

		BeforeEach(func() {
			daemonSetInputs = []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase-fuse",
						Namespace: healthCheckTestNamespace,
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
						Namespace: healthCheckTestNamespace,
					},
					Status: appsv1.DaemonSetStatus{
						NumberUnavailable: 0,
						NumberReady:       1,
						NumberAvailable:   1,
					},
				},
			}

			thinRuntimeInputs = []datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestHbase,
						Namespace: healthCheckTestNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      healthCheckTestSpark,
						Namespace: healthCheckTestNamespace,
					},
				},
			}
		})

		Context("when fuse has unavailable pods", func() {
			var (
				client client.Client
				engine ThinEngine
			)

			BeforeEach(func() {
				testObjs := []runtime.Object{}
				for _, daemonSet := range daemonSetInputs {
					testObjs = append(testObjs, daemonSet.DeepCopy())
				}
				for _, thinRuntimeInput := range thinRuntimeInputs {
					testObjs = append(testObjs, thinRuntimeInput.DeepCopy())
				}
				client = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine = ThinEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					namespace: healthCheckTestNamespace,
					name:      healthCheckTestHbase,
					runtime: &datav1alpha1.ThinRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      healthCheckTestHbase,
							Namespace: healthCheckTestNamespace,
						},
					},
					Recorder: record.NewFakeRecorder(1),
				}
			})

			JustBeforeEach(func() {
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
			})

			It("should update runtime with NotReady phase", func() {
				_, err := engine.checkFuseHealthy()
				Expect(err).NotTo(HaveOccurred())

				thinRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(thinRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(thinRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(thinRuntime.Status.FuseNumberUnavailable).To(Equal(int32(1)))

				_, cond := utils.GetRuntimeCondition(thinRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())
			})
		})

		Context("when all fuse pods are available", func() {
			var (
				client client.Client
				engine ThinEngine
			)

			BeforeEach(func() {
				testObjs := []runtime.Object{}
				for _, daemonSet := range daemonSetInputs {
					testObjs = append(testObjs, daemonSet.DeepCopy())
				}
				for _, thinRuntimeInput := range thinRuntimeInputs {
					testObjs = append(testObjs, thinRuntimeInput.DeepCopy())
				}
				client = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine = ThinEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					namespace: healthCheckTestNamespace,
					name:      healthCheckTestSpark,
					runtime: &datav1alpha1.ThinRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      healthCheckTestSpark,
							Namespace: healthCheckTestNamespace,
						},
					},
					Recorder: record.NewFakeRecorder(1),
				}
			})

			JustBeforeEach(func() {
				runtimeInfo, err := base.BuildRuntimeInfo(engine.name, engine.namespace, common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())
				engine.Helper = ctrl.BuildHelper(runtimeInfo, client, engine.Log)
			})

			It("should update runtime with Ready phase", func() {
				_, err := engine.checkFuseHealthy()
				Expect(err).NotTo(HaveOccurred())

				thinRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(thinRuntime.Status.FuseNumberReady).To(Equal(int32(1)))
				Expect(thinRuntime.Status.FuseNumberAvailable).To(Equal(int32(1)))
				Expect(thinRuntime.Status.FuseNumberUnavailable).To(Equal(int32(0)))

				_, cond := utils.GetRuntimeCondition(thinRuntime.Status.Conditions, datav1alpha1.RuntimeFusesReady)
				Expect(cond).NotTo(BeNil())
			})
		})
	})

	Describe("CheckAndUpdateRuntimeStatus", func() {
		It("initializes fuse-only runtime status and strips mount options", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      healthCheckTestSpark,
					Namespace: healthCheckTestNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "s3://bucket/data",
						Options:    map[string]string{"endpoint": "test"},
						EncryptOptions: []datav1alpha1.EncryptOption{{
							Name: "secret",
						}},
					}},
				},
			}

			runtimeObj := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:              healthCheckTestSpark,
					Namespace:         healthCheckTestNamespace,
					CreationTimestamp: metav1.Now(),
				},
				Spec: datav1alpha1.ThinRuntimeSpec{
					Fuse: datav1alpha1.ThinFuseSpec{},
				},
				Status: datav1alpha1.RuntimeStatus{
					FusePhase: datav1alpha1.RuntimePhaseNone,
				},
			}

			client := fake.NewFakeClientWithScheme(testScheme, dataset, runtimeObj)
			engine := ThinEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				name:      healthCheckTestSpark,
				namespace: healthCheckTestNamespace,
				runtime:   runtimeObj,
			}

			ready, err := engine.CheckAndUpdateRuntimeStatus()

			Expect(err).NotTo(HaveOccurred())
			Expect(ready).To(BeTrue())

			updatedRuntime, err := engine.getRuntime()
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedRuntime.Status.FusePhase).To(Equal(datav1alpha1.RuntimePhaseReady))
			Expect(updatedRuntime.Status.ValueFileConfigmap).To(Equal(engine.getHelmValuesConfigMapName()))
			Expect(updatedRuntime.Status.SetupDuration).NotTo(BeEmpty())
			Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.Cached, "N/A"))
			Expect(updatedRuntime.Status.Mounts).To(HaveLen(1))
			Expect(updatedRuntime.Status.Mounts[0].Options).To(BeNil())
			Expect(updatedRuntime.Status.Mounts[0].EncryptOptions).To(BeNil())

			_, cond := utils.GetRuntimeCondition(updatedRuntime.Status.Conditions, datav1alpha1.RuntimeFusesInitialized)
			Expect(cond).NotTo(BeNil())
		})
	})

	Describe("ThinEngine.getDataSetFileNum", func() {
		It("returns an empty count when no running fuse pod can be found", func() {
			engine := &ThinEngine{
				Client:    fake.NewFakeClientWithScheme(testScheme),
				Log:       fake.NullLogger(),
				name:      healthCheckTestName,
				namespace: healthCheckTestNamespace,
			}

			count, err := engine.getDataSetFileNum()

			Expect(err).To(HaveOccurred())
			Expect(count).To(BeEmpty())
		})
	})
})
