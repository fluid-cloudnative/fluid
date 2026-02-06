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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	metadataTestNamespace        = "fluid"
	metadataTestNameHbase        = "hbase"
	metadataTestNameSpark        = "spark"
	metadataTestNameHadoop       = "hadoop"
	metadataTestNameNoSync       = "no-sync"
	metadataTestNameTest1        = "test1"
	metadataTestNameTest2        = "test2"
	metadataTestUfsTotalInitial  = "2Gi"
	metadataTestUfsTotalExpected = "2GB"
	metadataTestFileNumExpected  = "5"
	metadataTestMetaQueryResult  = "1024"
	metadataTestRestorePath      = "local:///host1/erf"
	metadataTestRestoreNode      = "test-node"
)

var _ = Describe("ShouldSyncMetadata", Label("pkg.ddc.juicefs.metadata_test.go"), func() {
	var (
		testClient client.Client
		testObjs   []runtime.Object
	)

	BeforeEach(func() {
		datasetInputs := []datav1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameHbase,
					Namespace: metadataTestNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: metadataTestUfsTotalInitial,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameSpark,
					Namespace: metadataTestNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: "",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameNoSync,
					Namespace: metadataTestNamespace,
				},
			},
		}
		runtimeInputs := []datav1alpha1.JuiceFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameHbase,
					Namespace: metadataTestNamespace,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameSpark,
					Namespace: metadataTestNamespace,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameNoSync,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					RuntimeManagement: datav1alpha1.RuntimeManagement{
						MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
							AutoSync: ptr.To(false),
						},
					},
				},
			},
		}

		testObjs = []runtime.Object{}
		for _, datasetInput := range datasetInputs {
			testObjs = append(testObjs, datasetInput.DeepCopy())
		}
		for _, runtimeInput := range runtimeInputs {
			testObjs = append(testObjs, runtimeInput.DeepCopy())
		}
		testClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when dataset already has UfsTotal set", func() {
		It("should return false", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameHbase,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			should, err := engine.shouldSyncMetadata()
			Expect(err).NotTo(HaveOccurred())
			Expect(should).To(BeFalse())
		})
	})

	Context("when dataset has empty UfsTotal", func() {
		It("should return true", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameSpark,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			should, err := engine.shouldSyncMetadata()
			Expect(err).NotTo(HaveOccurred())
			Expect(should).To(BeTrue())
		})
	})

	Context("when AutoSync is disabled in runtime", func() {
		It("should return false", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameNoSync,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			should, err := engine.shouldSyncMetadata()
			Expect(err).NotTo(HaveOccurred())
			Expect(should).To(BeFalse())
		})
	})
})

var _ = Describe("SyncMetadata", Label("pkg.ddc.juicefs.metadata_test.go"), func() {
	var (
		testClient client.Client
		testObjs   []runtime.Object
	)

	BeforeEach(func() {
		datasetInputs := []datav1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameHbase,
					Namespace: metadataTestNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: metadataTestUfsTotalInitial,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameSpark,
					Namespace: metadataTestNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: "",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameHadoop,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{
					DataRestoreLocation: &datav1alpha1.DataRestoreLocation{
						Path:     metadataTestRestorePath,
						NodeName: metadataTestRestoreNode,
					},
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: "",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameNoSync,
					Namespace: metadataTestNamespace,
				},
				Status: datav1alpha1.DatasetStatus{
					UfsTotal: "",
				},
			},
		}
		runtimeInputs := []datav1alpha1.JuiceFSRuntime{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameHbase,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					RuntimeManagement: datav1alpha1.RuntimeManagement{},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameSpark,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					RuntimeManagement: datav1alpha1.RuntimeManagement{},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameHadoop,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					RuntimeManagement: datav1alpha1.RuntimeManagement{},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameNoSync,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					RuntimeManagement: datav1alpha1.RuntimeManagement{
						MetadataSyncPolicy: datav1alpha1.MetadataSyncPolicy{
							AutoSync: ptr.To(false),
						},
					},
				},
			},
		}

		testObjs = []runtime.Object{}
		for _, datasetInput := range datasetInputs {
			testObjs = append(testObjs, datasetInput.DeepCopy())
		}
		for _, runtimeInput := range runtimeInputs {
			testObjs = append(testObjs, runtimeInput.DeepCopy())
		}
		testClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when syncing metadata for hbase engine", func() {
		It("should succeed without error", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameHbase,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			err := engine.SyncMetadata()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when syncing metadata for spark engine", func() {
		It("should succeed without error", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameSpark,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			err := engine.SyncMetadata()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when syncing metadata for hadoop engine with data restore", func() {
		It("should succeed using mocked QueryMetaDataInfoIntoFile", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameHadoop,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			patches := gomonkey.ApplyMethodFunc(operations.JuiceFileUtils{}, "QueryMetaDataInfoIntoFile",
				func(key operations.KeyOfMetaDataFile, filename string) (string, error) {
					return metadataTestMetaQueryResult, nil
				})
			defer patches.Reset()

			err := engine.SyncMetadata()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when AutoSync is disabled in runtime", func() {
		It("should not sync metadata and return no error", func() {
			engine := JuiceFSEngine{
				name:      metadataTestNameNoSync,
				namespace: metadataTestNamespace,
				Client:    testClient,
				Log:       fake.NullLogger(),
			}

			err := engine.SyncMetadata()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("SyncMetadataInternal", Label("pkg.ddc.juicefs.metadata_test.go"), func() {
	var (
		testClient client.Client
		testObjs   []runtime.Object
	)

	BeforeEach(func() {
		datasetInputs := []datav1alpha1.Dataset{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameTest1,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      metadataTestNameTest2,
					Namespace: metadataTestNamespace,
				},
				Spec: datav1alpha1.DatasetSpec{},
			},
		}

		testObjs = []runtime.Object{}
		for _, datasetInput := range datasetInputs {
			testObjs = append(testObjs, datasetInput.DeepCopy())
		}
		testClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when engine has MetadataSyncDoneCh", func() {
		It("should update dataset status with sync result", func() {
			engine := JuiceFSEngine{
				name:               metadataTestNameTest1,
				namespace:          metadataTestNamespace,
				Client:             testClient,
				Log:                fake.NullLogger(),
				MetadataSyncDoneCh: make(chan base.MetadataSyncResult),
			}

			result := base.MetadataSyncResult{
				StartTime: time.Now(),
				UfsTotal:  metadataTestUfsTotalExpected,
				Done:      true,
				FileNum:   metadataTestFileNumExpected,
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
			err = testClient.Get(context.Background(), key, dataset)
			Expect(err).NotTo(HaveOccurred())
			Expect(dataset.Status.UfsTotal).To(Equal(metadataTestUfsTotalExpected))
			Expect(dataset.Status.FileNum).To(Equal(metadataTestFileNumExpected))
		})
	})

	Context("when engine has nil MetadataSyncDoneCh", func() {
		It("should succeed without error", func() {
			engine := JuiceFSEngine{
				name:               metadataTestNameTest2,
				namespace:          metadataTestNamespace,
				Client:             testClient,
				Log:                fake.NullLogger(),
				MetadataSyncDoneCh: nil,
			}

			err := engine.syncMetadataInternal()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
