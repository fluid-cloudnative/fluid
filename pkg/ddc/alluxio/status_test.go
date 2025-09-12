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

package alluxio

import (
	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Runtime Status Tests", Label("pkg.ddc.alluxio.status_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		mockedObjects  mockedObjects
		client         client.Client
		resources      []runtime.Object
	)

	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: "fluid", Name: "hbase"})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
		mockedObjects = mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
		resources = []runtime.Object{
			dataset,
			alluxioruntime,
			mockedObjects.MasterSts,
			mockedObjects.WorkerSts,
			mockedObjects.FuseDs,
			mockedObjects.PersistentVolumeClaim,
			mockedObjects.PersistentVolume,
		}
	})

	// JustBeforeEach is guaranteed to run after every BeforeEach()
	// So it's easy to modify resources' specs with an extra BeforeEach()
	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	Describe("Test AlluxioEngine.CheckAndUpdateRuntimeStatus()", func() {
		When("Alluxio master and worker are all ready", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.Replicas = 1
				mockedObjects.MasterSts.Status.ReadyReplicas = 1

				alluxioruntime.Spec.Replicas = 3
				mockedObjects.WorkerSts.Spec.Replicas = ptr.To[int32](3)
				mockedObjects.WorkerSts.Status.Replicas = 3
				mockedObjects.WorkerSts.Status.ReadyReplicas = 3
				mockedObjects.WorkerSts.Status.CurrentReplicas = 3
			})

			It("should return ready and successfully update runtime status", func() {
				mockedCacheStates := cacheStates{
					cacheCapacity:    "18.00GiB",
					cached:           "4.20GiB",
					cachedPercentage: "100.0%",
					cacheHitStates: cacheHitStates{
						cacheHitRatio:  "100.0%",
						localHitRatio:  "33.3%",
						remoteHitRatio: "66.7%",

						localThroughputRatio:  "30.0%",
						remoteThroughputRatio: "70.0%",
						cacheThroughputRatio:  "100.0%",

						bytesReadLocal:  int64(1 << 30),
						bytesReadRemote: int64(2 << 30),
						bytesReadUfsAll: int64(1 << 30),
					},
				}
				patch := gomonkey.ApplyPrivateMethod(engine, "queryCacheStatus", func() (cacheStates, error) {
					return mockedCacheStates, nil
				})
				defer patch.Reset()

				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				updatedRuntime, err := utils.GetAlluxioRuntime(engine.Client, engine.name, engine.namespace)
				Expect(err).To(BeNil())

				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.Cached, mockedCacheStates.cached))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.CacheCapacity, mockedCacheStates.cacheCapacity))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.CachedPercentage, mockedCacheStates.cachedPercentage))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.CacheHitRatio, mockedCacheStates.cacheHitStates.cacheHitRatio))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.LocalHitRatio, mockedCacheStates.cacheHitStates.localHitRatio))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.RemoteHitRatio, mockedCacheStates.cacheHitStates.remoteHitRatio))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.CacheThroughputRatio, mockedCacheStates.cacheHitStates.cacheThroughputRatio))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.LocalThroughputRatio, mockedCacheStates.cacheHitStates.localThroughputRatio))
				Expect(updatedRuntime.Status.CacheStates).To(HaveKeyWithValue(common.RemoteThroughputRatio, mockedCacheStates.cacheHitStates.remoteThroughputRatio))
			})
		})

		When("alluxio master is ready but alluxio worker is partial ready", func() {
			BeforeEach(func() {
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
				mockedObjects.MasterSts.Status.ReadyReplicas = 1

				alluxioruntime.Spec.Replicas = 3
				mockedObjects.WorkerSts.Spec.Replicas = ptr.To[int32](3)
				mockedObjects.WorkerSts.Status.ReadyReplicas = 2
				mockedObjects.WorkerSts.Status.CurrentReplicas = 2
			})

			It("should return ready and runtime.status.workerPhase should be set to partial ready", func() {
				patch := gomonkey.ApplyPrivateMethod(engine, "queryCacheStatus", func() (cacheStates, error) {
					return cacheStates{}, nil
				})
				defer patch.Reset()
				ready, err := engine.CheckAndUpdateRuntimeStatus()
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())

				_, err = utils.GetAlluxioRuntime(engine.Client, engine.name, engine.namespace)
				Expect(err).To(BeNil())
			})
		})
	})

	When("the runtime has been set non-nil node affinity", func() {
		BeforeEach(func() {
			mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
			mockedObjects.MasterSts.Status.ReadyReplicas = 1
			mockedObjects.WorkerSts.Spec.Replicas = ptr.To[int32](3)
			mockedObjects.WorkerSts.Status.ReadyReplicas = 3
			mockedObjects.WorkerSts.Spec.Template.Spec.NodeSelector = map[string]string{
				"test-node-selector": "value1",
			}

			mockedObjects.WorkerSts.Spec.Template.Spec.Affinity = &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "test-node-affinity",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"value2"},
									},
								},
							},
						},
					},
				},
			}
		})

		It("should update node affinity info to runtime status", func() {
			patch := gomonkey.ApplyPrivateMethod(engine, "queryCacheStatus", func() (cacheStates, error) {
				return cacheStates{}, nil
			})
			defer patch.Reset()
			ready, err := engine.CheckAndUpdateRuntimeStatus()
			Expect(err).To(BeNil())
			Expect(ready).To(BeTrue())

			updatedRuntime, err := utils.GetAlluxioRuntime(engine.Client, engine.name, engine.namespace)
			Expect(err).To(BeNil())
			Expect(updatedRuntime.Status.CacheAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).To(ContainElements(
				corev1.NodeSelectorRequirement{
					Key:      "test-node-affinity",
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"value2"},
				},
				corev1.NodeSelectorRequirement{
					Key:      "test-node-selector",
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"value1"},
				},
			))

		})
	})
})
