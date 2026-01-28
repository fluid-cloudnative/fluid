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

package jindofsx

import (
	"context"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("JindoFSxEngine Dataset Operations", func() {
	Describe("UpdateCacheOfDataset", func() {
		It("should update cache of dataset successfully", func() {
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

			engine := &JindoFSxEngine{
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
			Expect(reflect.DeepEqual(datasets.Items[0].Status, expectedDataset.Status)).To(BeTrue())
		})
	})

	Describe("UpdateCacheOfDatasetWithoutMaster", func() {
		It("should update cache of dataset without master successfully", func() {
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

			engine := &JindoFSxEngine{
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
			Expect(reflect.DeepEqual(datasets.Items[0].Status, expectedDataset.Status)).To(BeTrue())
		})
	})

	Describe("UpdateDatasetStatus", func() {
		var (
			fakeClient client.Client
			engine     *JindoFSxEngine
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
			fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

			engine = &JindoFSxEngine{
				Client:    fakeClient,
				Log:       fake.NullLogger(),
				name:      "hbase",
				namespace: "fluid",
				runtime:   testRuntimeInputs[0],
			}
		})

		DescribeTable("should update dataset status with different phases",
			func(phase datav1alpha1.DatasetPhase, expectedPhase datav1alpha1.DatasetPhase) {
				err := engine.UpdateDatasetStatus(phase)
				Expect(err).NotTo(HaveOccurred())

				var datasets datav1alpha1.DatasetList
				err = fakeClient.List(context.TODO(), &datasets)
				Expect(err).NotTo(HaveOccurred())
				Expect(datasets.Items[0].Status.Phase).To(Equal(expectedPhase))
				Expect(datasets.Items[0].Status.CacheStates).To(Equal(common.CacheStateList{
					common.Cached: "true",
				}))
				Expect(datasets.Items[0].Status.HCFSStatus).To(Equal(&datav1alpha1.HCFSStatus{
					Endpoint:                    "test Endpoint",
					UnderlayerFileSystemVersion: "Underlayer HCFS Compatible Version",
				}))
			},
			Entry("BoundDatasetPhase", datav1alpha1.BoundDatasetPhase, datav1alpha1.BoundDatasetPhase),
			Entry("FailedDatasetPhase", datav1alpha1.FailedDatasetPhase, datav1alpha1.FailedDatasetPhase),
			Entry("NoneDatasetPhase", datav1alpha1.NoneDatasetPhase, datav1alpha1.NoneDatasetPhase),
		)
	})

	Describe("BindToDataset", func() {
		It("should bind to dataset successfully", func() {
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

			engine := &JindoFSxEngine{
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
