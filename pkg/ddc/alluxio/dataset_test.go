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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Dataset Status Tests", Label("pkg.ddc.alluxio.dataset_test.go"), func() {
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
		}
	})

	JustBeforeEach(func() {
		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = client
	})

	Describe("Test AlluxioEngine.UpdateCacheOfDataset()", func() {
		When("everything works as expected", func() {
			BeforeEach(func() {
				alluxioruntime.Status.CacheStates = map[common.CacheStateName]string{
					common.Cached:           "42.00GiB",
					common.CacheCapacity:    "100.00GiB",
					common.CacheHitRatio:    "95%",
					common.CachedPercentage: "100.0%",
					common.LocalHitRatio:    "60.0%",
					common.RemoteHitRatio:   "40.0%",
				}
			})

			It("should update cache status of dataset", func() {
				err := engine.UpdateCacheOfDataset()
				Expect(err).To(BeNil())

				datasetToCheck, err := utils.GetDataset(client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(datasetToCheck.Status.CacheStates).To(Equal(alluxioruntime.Status.CacheStates))
			})
		})
	})

	Describe("Test AlluxioEngine.UpdateDatasetStatus()", func() {
		When("Dataset's phase transit from NotBound to Bound", func() {
			BeforeEach(func() {
				dataset.Status.Phase = datav1alpha1.NotBoundDatasetPhase
			})
			It("should update dataset's phase to Bound", func() {
				patch := gomonkey.ApplyMethodFunc(engine, "GetHCFSStatus", func() (*datav1alpha1.HCFSStatus, error) {
					return &datav1alpha1.HCFSStatus{
						Endpoint:                    "dummy-endpoint",
						UnderlayerFileSystemVersion: "dummy-version-v1.0.0",
					}, nil
				})
				defer patch.Reset()

				err := engine.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
				Expect(err).To(BeNil())

				datasetToCheck, err := utils.GetDataset(client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(datasetToCheck.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
				Expect(datasetToCheck.Status.Mounts).To(Equal(dataset.Spec.Mounts))
				Expect(datasetToCheck.Status.Runtimes).To(HaveLen(1))
				Expect(datasetToCheck.Status.Runtimes[0]).To(Equal(datav1alpha1.Runtime{
					Name:           alluxioruntime.Name,
					Namespace:      alluxioruntime.Namespace,
					Category:       common.AccelerateCategory,
					Type:           common.AlluxioRuntime,
					MasterReplicas: alluxioruntime.Spec.Master.Replicas,
				}))

			})
		})

		When("Dataset's phase transit from Bound to Failed", func() {
			BeforeEach(func() {
				dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
				dataset.Status.HCFSStatus = &datav1alpha1.HCFSStatus{
					Endpoint:                    "dummy-endpoint",
					UnderlayerFileSystemVersion: "dummy-version-v1.0.0",
				}
			})
			It("should update dataset's phase to Failed", func() {
				err := engine.UpdateDatasetStatus(datav1alpha1.FailedDatasetPhase)
				Expect(err).To(BeNil())

				datasetToCheck, err := utils.GetDataset(client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(datasetToCheck.Status.Phase).To(Equal(datav1alpha1.FailedDatasetPhase))
				Expect(datasetToCheck.Status.Conditions).To(HaveLen(1))
				Expect(datasetToCheck.Status.Conditions[0].Message).To(Equal("The ddc runtime is not ready."))
			})
		})

		When("Dataset's phase transit from Bound to NotBound phase", func() {
			BeforeEach(func() {
				dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
				dataset.Status.HCFSStatus = &datav1alpha1.HCFSStatus{
					Endpoint:                    "dummy-endpoint",
					UnderlayerFileSystemVersion: "dummy-version-v1.0.0",
				}
			})

			It("should update dataset's phase to NotBound phase", func() {
				err := engine.UpdateDatasetStatus(datav1alpha1.NotBoundDatasetPhase)
				Expect(err).To(BeNil())

				datasetToCheck, err := utils.GetDataset(client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(datasetToCheck.Status.Phase).To(Equal(datav1alpha1.NotBoundDatasetPhase))
				Expect(datasetToCheck.Status.Conditions).To(HaveLen(1))
				Expect(datasetToCheck.Status.Conditions[0].Message).To(Equal("The ddc runtime is unknown."))
			})
		})
	})

	Describe("Test AlluxioEngine.BindToDataset()", func() {
		When("Dataset's phase is NotBound", func() {
			BeforeEach(func() {
				dataset.Status.Phase = datav1alpha1.NotBoundDatasetPhase
			})
			It("should update dataset's phase to Bound", func() {
				patch := gomonkey.ApplyMethodFunc(engine, "GetHCFSStatus", func() (*datav1alpha1.HCFSStatus, error) {
					return &datav1alpha1.HCFSStatus{
						Endpoint:                    "dummy-endpoint",
						UnderlayerFileSystemVersion: "dummy-version-v1.0.0",
					}, nil
				})
				defer patch.Reset()

				err := engine.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
				Expect(err).To(BeNil())

				datasetToCheck, err := utils.GetDataset(client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(datasetToCheck.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
				Expect(datasetToCheck.Status.Mounts).To(Equal(dataset.Spec.Mounts))
				Expect(datasetToCheck.Status.Runtimes).To(HaveLen(1))
				Expect(datasetToCheck.Status.Runtimes[0]).To(Equal(datav1alpha1.Runtime{
					Name:           alluxioruntime.Name,
					Namespace:      alluxioruntime.Namespace,
					Category:       common.AccelerateCategory,
					Type:           common.AlluxioRuntime,
					MasterReplicas: alluxioruntime.Spec.Master.Replicas,
				}))

			})
		})
	})
})
