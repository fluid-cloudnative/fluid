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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

const (
	metadataTestNamespace  = "fluid"
	metadataTestHbase      = "hbase"
	metadataTestSpark      = "spark"
	metadataTestNoAutoSync = "noautosync"
	metadataTestAutoSync   = "autosync"
	metadataTestHadoop     = "hadoop"
	metadataTestTest1      = "test1"
	metadataTestTest2      = "test2"
	metadataTestUfsTotal   = "2Gi"
)

var _ = Describe("ThinEngine Metadata", Label("pkg.ddc.thin.metadata_test.go"), func() {
	Describe("shouldSyncMetadata", func() {
		var (
			datasetInputs []datav1alpha1.Dataset
			runtimeInputs []datav1alpha1.ThinRuntime
			testObjs      []runtime.Object
			client        client.Client
		)

		BeforeEach(func() {
			datasetInputs = []datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestHbase,
						Namespace: metadataTestNamespace,
					},
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: metadataTestUfsTotal,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestSpark,
						Namespace: metadataTestNamespace,
					},
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: "",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestNoAutoSync,
						Namespace: metadataTestNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestAutoSync,
						Namespace: metadataTestNamespace,
					},
				},
			}

			runtimeInputs = []datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestHbase,
						Namespace: metadataTestNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestSpark,
						Namespace: metadataTestNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestNoAutoSync,
						Namespace: metadataTestNamespace,
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
								AutoSync: ptr.To(false),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestAutoSync,
						Namespace: metadataTestNamespace,
					},
					Spec: datav1alpha1.ThinRuntimeSpec{
						RuntimeManagement: datav1alpha1.RuntimeManagement{
							MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
								AutoSync: ptr.To(true),
							},
						},
					},
				},
			}

			testObjs = []runtime.Object{}
			for _, input := range datasetInputs {
				testObjs = append(testObjs, input.DeepCopy())
			}
			for _, input := range runtimeInputs {
				testObjs = append(testObjs, input.DeepCopy())
			}
		})

		JustBeforeEach(func() {
			client = fake.NewFakeClientWithScheme(testScheme, testObjs...)
		})

		Context("when dataset has UfsTotal set", func() {
			It("should return false for hbase", func() {
				engine := ThinEngine{
					name:      metadataTestHbase,
					namespace: metadataTestNamespace,
					Client:    client,
					Log:       fake.NullLogger(),
				}

				should, err := engine.shouldSyncMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when dataset has empty UfsTotal and no autosync policy", func() {
			It("should return false for spark", func() {
				engine := ThinEngine{
					name:      metadataTestSpark,
					namespace: metadataTestNamespace,
					Client:    client,
					Log:       fake.NullLogger(),
				}

				should, err := engine.shouldSyncMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when autosync is explicitly disabled", func() {
			It("should return false for noautosync", func() {
				engine := ThinEngine{
					name:      metadataTestNoAutoSync,
					namespace: metadataTestNamespace,
					Client:    client,
					Log:       fake.NullLogger(),
				}

				should, err := engine.shouldSyncMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		Context("when autosync is explicitly enabled", func() {
			It("should return true for autosync", func() {
				engine := ThinEngine{
					name:      metadataTestAutoSync,
					namespace: metadataTestNamespace,
					Client:    client,
					Log:       fake.NullLogger(),
				}

				should, err := engine.shouldSyncMetadata()
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeTrue())
			})
		})
	})

	Describe("SyncMetadata", func() {
		var (
			datasetInputs []datav1alpha1.Dataset
			runtimeInputs []datav1alpha1.ThinRuntime
			testObjs      []runtime.Object
			client        client.Client
		)

		BeforeEach(func() {
			datasetInputs = []datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestHbase,
						Namespace: metadataTestNamespace,
					},
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: metadataTestUfsTotal,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestSpark,
						Namespace: metadataTestNamespace,
					},
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: "",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestHadoop,
						Namespace: metadataTestNamespace,
					},
					Spec: datav1alpha1.DatasetSpec{
						DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
							Path:     "local:///host1/erf",
							NodeName: "test-node",
						},
					},
					Status: datav1alpha1.DatasetStatus{
						UfsTotal: "",
					},
				},
			}

			runtimeInputs = []datav1alpha1.ThinRuntime{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestHbase,
						Namespace: metadataTestNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestSpark,
						Namespace: metadataTestNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      metadataTestHadoop,
						Namespace: metadataTestNamespace,
					},
				},
			}

			testObjs = []runtime.Object{}
			for _, input := range datasetInputs {
				testObjs = append(testObjs, input.DeepCopy())
			}
			for _, input := range runtimeInputs {
				testObjs = append(testObjs, input.DeepCopy())
			}
		})

		JustBeforeEach(func() {
			client = fake.NewFakeClientWithScheme(testScheme, testObjs...)
		})

		createEngine := func(name string) ThinEngine {
			return ThinEngine{
				name:      name,
				namespace: metadataTestNamespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}
		}

		Context("when dataset has UfsTotal already set", func() {
			It("should return no error for hbase", func() {
				engine := createEngine(metadataTestHbase)
				err := engine.SyncMetadata()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when dataset has empty UfsTotal", func() {
			It("should return no error for spark", func() {
				engine := createEngine(metadataTestSpark)
				err := engine.SyncMetadata()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when dataset has DataRestoreLocation", func() {
			It("should return no error for hadoop", func() {
				engine := createEngine(metadataTestHadoop)
				err := engine.SyncMetadata()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("syncMetadataInternal", func() {
		Context("when MetadataSyncDoneCh receives result", func() {
			It("should update dataset status with UfsTotal and FileNum", func() {
				datasetInputs := []datav1alpha1.Dataset{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      metadataTestTest1,
							Namespace: metadataTestNamespace,
						},
						Spec: datav1alpha1.DatasetSpec{},
					},
				}
				testObjs := []runtime.Object{}
				for _, datasetInput := range datasetInputs {
					testObjs = append(testObjs, datasetInput.DeepCopy())
				}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:               metadataTestTest1,
					namespace:          metadataTestNamespace,
					Client:             client,
					Log:                fake.NullLogger(),
					MetadataSyncDoneCh: make(chan base.MetadataSyncResult, 1),
				}

				result := base.MetadataSyncResult{
					StartTime: time.Now(),
					UfsTotal:  "2GB",
					Done:      true,
					FileNum:   "5",
				}

				go func() {
					engine.MetadataSyncDoneCh <- result
				}()

				err := engine.syncMetadataInternal()
				Expect(err).NotTo(HaveOccurred())

				key := types.NamespacedName{
					Namespace: engine.namespace,
					Name:      engine.name,
				}

				dataset := &datav1alpha1.Dataset{}
				err = client.Get(context.Background(), key, dataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(dataset.Status.UfsTotal).To(Equal("2GB"))
				Expect(dataset.Status.FileNum).To(Equal("5"))
			})
		})

		Context("when MetadataSyncDoneCh is nil", func() {
			It("should return no error without updating", func() {
				datasetInputs := []datav1alpha1.Dataset{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      metadataTestTest2,
							Namespace: metadataTestNamespace,
						},
						Spec: datav1alpha1.DatasetSpec{},
					},
				}
				testObjs := []runtime.Object{}
				for _, datasetInput := range datasetInputs {
					testObjs = append(testObjs, datasetInput.DeepCopy())
				}
				client := fake.NewFakeClientWithScheme(testScheme, testObjs...)

				engine := ThinEngine{
					name:               metadataTestTest2,
					namespace:          metadataTestNamespace,
					Client:             client,
					Log:                fake.NullLogger(),
					MetadataSyncDoneCh: nil,
				}

				err := engine.syncMetadataInternal()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
