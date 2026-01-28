/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Dataset", func() {
	Describe("UpdateCacheOfDataset", func() {
		Context("with master enabled", func() {
			It("should update dataset cache status correctly", func() {
				testDatasetInputs := []*datav1alpha1.Dataset{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase",
							Namespace: "fluid",
						},
						Spec: datav1alpha1.DatasetSpec{},
					},
				}
				testObjs := []runtime.Object{}
				for _, datasetInput := range testDatasetInputs {
					testObjs = append(testObjs, datasetInput.DeepCopy())
				}

				testRuntimeInputs := []*datav1alpha1.JindoRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase",
							Namespace: "fluid",
						},
						Spec: datav1alpha1.JindoRuntimeSpec{
							Master: datav1alpha1.JindoCompTemplateSpec{
								Replicas: 1,
							},
						},
						Status: datav1alpha1.RuntimeStatus{
							CacheStates: map[common.CacheStateName]string{
								common.Cached: "true",
							},
						},
					},
				}
				for _, runtimeInput := range testRuntimeInputs {
					testObjs = append(testObjs, runtimeInput.DeepCopy())
				}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := &JindoCacheEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					name:      "hbase",
					namespace: "fluid",
					runtime:   testRuntimeInputs[0],
				}

				err := engine.UpdateCacheOfDataset()
				Expect(err).NotTo(HaveOccurred())

				expectedDataset := datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Status: datav1alpha1.DatasetStatus{
						CacheStates: map[common.CacheStateName]string{
							common.Cached: "true",
						},
						HCFSStatus: &datav1alpha1.HCFSStatus{
							Endpoint:                    "",
							UnderlayerFileSystemVersion: "",
						},
					},
				}

				var datasets datav1alpha1.DatasetList
				err = client.List(context.TODO(), &datasets)
				Expect(err).NotTo(HaveOccurred())
				Expect(datasets.Items[0].Status).To(Equal(expectedDataset.Status))
			})
		})

		Context("without master", func() {
			It("should update dataset cache status with N/A endpoint", func() {
				testDatasetInputs := []*datav1alpha1.Dataset{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase",
							Namespace: "fluid",
						},
						Spec: datav1alpha1.DatasetSpec{},
					},
				}
				testObjs := []runtime.Object{}
				for _, datasetInput := range testDatasetInputs {
					testObjs = append(testObjs, datasetInput.DeepCopy())
				}

				testRuntimeInputs := []*datav1alpha1.JindoRuntime{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hbase",
							Namespace: "fluid",
						},
						Spec: datav1alpha1.JindoRuntimeSpec{
							Master: datav1alpha1.JindoCompTemplateSpec{
								Disabled: true,
								Replicas: 1,
							},
						},
						Status: datav1alpha1.RuntimeStatus{
							CacheStates: map[common.CacheStateName]string{
								common.Cached: "true",
							},
						},
					},
				}
				for _, runtimeInput := range testRuntimeInputs {
					testObjs = append(testObjs, runtimeInput.DeepCopy())
				}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := &JindoCacheEngine{
					Client:    client,
					Log:       fake.NullLogger(),
					name:      "hbase",
					namespace: "fluid",
					runtime:   testRuntimeInputs[0],
				}

				err := engine.UpdateCacheOfDataset()
				Expect(err).NotTo(HaveOccurred())

				expectedDataset := datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Status: datav1alpha1.DatasetStatus{
						CacheStates: map[common.CacheStateName]string{
							common.Cached: "true",
						},
						HCFSStatus: &datav1alpha1.HCFSStatus{
							Endpoint:                    "N/A",
							UnderlayerFileSystemVersion: "",
						},
					},
				}

				var datasets datav1alpha1.DatasetList
				err = client.List(context.TODO(), &datasets)
				Expect(err).NotTo(HaveOccurred())
				Expect(datasets.Items[0].Status).To(Equal(expectedDataset.Status))
			})
		})
	})

	Describe("UpdateDatasetStatus", func() {
		var (
			client            client.Client
			engine            *JindoCacheEngine
			testRuntimeInputs []*datav1alpha1.JindoRuntime
		)

		BeforeEach(func() {
			testDatasetInputs := []*datav1alpha1.Dataset{
				{
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
				},
			}
			testObjs := []runtime.Object{}
			for _, datasetInput := range testDatasetInputs {
				testObjs = append(testObjs, datasetInput.DeepCopy())
			}

			testRuntimeInputs = []*datav1alpha1.JindoRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Replicas: 1,
						},
					},
					Status: datav1alpha1.RuntimeStatus{
						CacheStates: map[common.CacheStateName]string{
							common.Cached: "true",
						},
					},
				},
			}
			for _, runtimeInput := range testRuntimeInputs {
				testObjs = append(testObjs, runtimeInput.DeepCopy())
			}
			client = fake.NewFakeClientWithScheme(testScheme, testObjs...)

			engine = &JindoCacheEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				name:      "hbase",
				namespace: "fluid",
				runtime:   testRuntimeInputs[0],
			}
		})

		DescribeTable("updating dataset phase",
			func(phase datav1alpha1.DatasetPhase, expectedPhase datav1alpha1.DatasetPhase, expectedCacheStates map[common.CacheStateName]string, expectedHCFSStatus *datav1alpha1.HCFSStatus) {
				err := engine.UpdateDatasetStatus(phase)
				Expect(err).NotTo(HaveOccurred())

				var datasets datav1alpha1.DatasetList
				err = client.List(context.TODO(), &datasets)
				Expect(err).NotTo(HaveOccurred())
				Expect(datasets.Items[0].Status.Phase).To(Equal(expectedPhase))
				Expect(datasets.Items[0].Status.CacheStates).To(Equal(expectedCacheStates))
				Expect(datasets.Items[0].Status.HCFSStatus).To(Equal(expectedHCFSStatus))
			},
			Entry("with Bound phase",
				datav1alpha1.BoundDatasetPhase,
				datav1alpha1.BoundDatasetPhase,
				map[common.CacheStateName]string{common.Cached: "true"},
				&datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			),
			Entry("with Failed phase",
				datav1alpha1.FailedDatasetPhase,
				datav1alpha1.FailedDatasetPhase,
				map[common.CacheStateName]string{common.Cached: "true"},
				&datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			),
			Entry("with None phase",
				datav1alpha1.NoneDatasetPhase,
				datav1alpha1.NoneDatasetPhase,
				map[common.CacheStateName]string{common.Cached: "true"},
				&datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				},
			),
		)
	})

	Describe("BindToDataset", func() {
		It("should bind runtime to dataset correctly", func() {
			testDatasetInputs := []*datav1alpha1.Dataset{
				{
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
				},
			}
			testObjs := []runtime.Object{}
			for _, datasetInput := range testDatasetInputs {
				testObjs = append(testObjs, datasetInput.DeepCopy())
			}

			testRuntimeInputs := []*datav1alpha1.JindoRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "hbase",
						Namespace: "fluid",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
					Status: datav1alpha1.RuntimeStatus{
						CacheStates: map[common.CacheStateName]string{
							common.Cached: "true",
						},
					},
				},
			}
			for _, runtimeInput := range testRuntimeInputs {
				testObjs = append(testObjs, runtimeInput.DeepCopy())
			}
			client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

			engine := &JindoCacheEngine{
				Client:    client,
				Log:       fake.NullLogger(),
				name:      "hbase",
				namespace: "fluid",
				runtime:   testRuntimeInputs[0],
			}

			expectedResult := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					CacheStates: map[common.CacheStateName]string{
						common.Cached: "true",
					},
					HCFSStatus: &datav1alpha1.HCFSStatus{
						Endpoint:                    "test Endpoint",
						UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
					},
				},
			}
			err := engine.BindToDataset()
			Expect(err).NotTo(HaveOccurred())

			var datasets datav1alpha1.DatasetList
			err = client.List(context.TODO(), &datasets)
			Expect(err).NotTo(HaveOccurred())
			Expect(datasets.Items[0].Status.Phase).To(Equal(expectedResult.Status.Phase))
			Expect(datasets.Items[0].Status.CacheStates).To(Equal(expectedResult.Status.CacheStates))
			Expect(datasets.Items[0].Status.HCFSStatus).To(Equal(expectedResult.Status.HCFSStatus))
		})
	})
})
