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
	"fmt"
	"reflect"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AlluxioEngine Metadata Synchronization Tests", Label("pkg.ddc.alluxio.metadata_test.go"), func() {
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

	Describe("Test AlluxioEngine.SyncMetadata", func() {
		When("the metadata has not been synced", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = ""
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
			})
			It("should sync metadata", func() {
				patch := gomonkey.ApplyPrivateMethod(engine, "syncMetadataInternal", func() error {
					return nil
				})
				defer patch.Reset()
				err := engine.SyncMetadata()
				Expect(err).Should(BeNil())
			})
		})

		When("the metadata has been synced", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = "100Gi"
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
			})
			It("should not call AlluxioEngine.syncMetadataInternal()", func() {
				patch := gomonkey.ApplyPrivateMethod(engine, "syncMetadataInternal", func() error {
					return fmt.Errorf("syncMetadataInternal should not be called")
				})
				defer patch.Reset()
				err := engine.SyncMetadata()
				Expect(err).Should(BeNil())
			})
		})

		When("auto metadata sync is disabled", func() {
			BeforeEach(func() {
				alluxioruntime.Spec.RuntimeManagement.MetadataSyncPolicy.AutoSync = ptr.To(false)
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
			})
			It("should not call AlluxioEngine.syncMetadataInternal()", func() {
				patch := gomonkey.ApplyPrivateMethod(engine, "syncMetadataInternal", func() error {
					return fmt.Errorf("syncMetadataInternal should not be called")
				})
				defer patch.Reset()
				err := engine.SyncMetadata()
				Expect(err).Should(BeNil())
			})
		})

		When("the metadata should be restored", func() {
			BeforeEach(func() {
				dataset.Spec.DataRestoreLocation = &datav1alpha1.DataRestoreLocation{
					Path:     "local:///host1/erf",
					NodeName: "test-node",
				}
				mockedObjects.MasterSts.Spec.Replicas = ptr.To[int32](1)
			})
			It("should restore metadata from the specified path", func() {
				patch := gomonkey.ApplyMethodFunc(engine, "RestoreMetadataInternal", func() error {
					return nil
				})
				defer patch.Reset()
				err := engine.SyncMetadata()
				Expect(err).Should(BeNil())
			})
		})

		When("master statefulset is not existed or healthy", func() {
			BeforeEach(func() {
				resources = []runtime.Object{
					dataset,
					alluxioruntime,
					// mockedObjects.MasterSts,
					mockedObjects.WorkerSts,
					mockedObjects.FuseDs,
					mockedObjects.PersistentVolumeClaim,
					mockedObjects.PersistentVolume,
				}
			})
			It("should return error", func() {
				err := engine.SyncMetadata()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the master engine does not exist"))
			})
		})
	})

	Describe("Test AlluxioEngine.ShouldSyncMetadata()", func() {
		When("auto sync metadata is disabled", func() {
			BeforeEach(func() {
				alluxioruntime.Spec.RuntimeManagement.MetadataSyncPolicy.AutoSync = ptr.To(false)
			})

			It("should return false", func() {
				shouldSync, err := engine.shouldSyncMetadata()
				Expect(err).To(BeNil())
				Expect(shouldSync).To(BeFalse())
			})
		})

		When("the metadata has not been synced", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = ""
			})

			It("should return true", func() {
				shouldSync, err := engine.shouldSyncMetadata()
				Expect(err).To(BeNil())
				Expect(shouldSync).To(BeTrue())
			})
		})

		When("the metadata has been synced", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = "100.00GiB"
			})

			It("should return false", func() {
				shouldSync, err := engine.shouldSyncMetadata()
				Expect(err).To(BeNil())
				Expect(shouldSync).To(BeFalse())
			})
		})

		When("the metadata is in syncing status", func() {
			BeforeEach(func() {
				dataset.Status.UfsTotal = metadataSyncNotDoneMsg
			})

			It("should return true", func() {
				shouldSync, err := engine.shouldSyncMetadata()
				Expect(err).To(BeNil())
				Expect(shouldSync).To(BeTrue())
			})
		})
	})

	Describe("Test AlluxioEngine.syncMetadataInternal()", func() {
		When("given AlluxioEngine works as expected", func() {
			It("should successfully sync metadata and update dataset status", func() {
				patch1 := gomonkey.ApplyMethodFunc(reflect.TypeOf(operations.AlluxioFileUtils{}), "LoadMetadataWithoutTimeout", func(string) error {
					return nil
				})
				defer patch1.Reset()

				patch2 := gomonkey.ApplyMethodFunc(engine, "TotalStorageBytes", func() (int64, error) {
					return 1 << 30, nil
				})
				defer patch2.Reset()

				patch3 := gomonkey.ApplyPrivateMethod(engine, "getDataSetFileNum", func() (string, error) {
					return "15", nil
				})
				defer patch3.Reset()

				// first time sync metadata, asynchronously starts a goroutine to sync metadata
				err := engine.syncMetadataInternal()
				Expect(err).To(BeNil())
				Expect(engine.MetadataSyncDoneCh).ToNot(BeNil())

				dataset, err = utils.GetDataset(engine.Client, engine.name, engine.namespace)
				Expect(err).To(BeNil())
				Expect(dataset.Status.FileNum).To(Equal(metadataSyncNotDoneMsg))
				Expect(dataset.Status.UfsTotal).To(Equal(metadataSyncNotDoneMsg))

				// second time sync metadata, get metadata results from channel and update them to dataset status
				err = engine.syncMetadataInternal()
				Expect(err).To(BeNil())
				Expect(engine.MetadataSyncDoneCh).To(BeNil())
				dataset, err = utils.GetDataset(engine.Client, engine.name, engine.namespace)
				Expect(err).To(BeNil())
				Expect(dataset.Status.UfsTotal).To(Equal("1.00GiB"))
				Expect(dataset.Status.FileNum).To(Equal("15"))
			})
		})
	})

	Describe("Test AlluxioEngine.shouldRestoreMetadata()", func() {
		When("dataset has no DataRestoreLocation", func() {
			BeforeEach(func() {
				dataset.Spec.DataRestoreLocation = nil
			})

			It("should return false", func() {
				shouldRestore, err := engine.shouldRestoreMetadata()
				Expect(err).To(BeNil())
				Expect(shouldRestore).To(BeFalse())
			})
		})

		When("dataset has DataRestoreLocation", func() {
			BeforeEach(func() {
				dataset.Spec.DataRestoreLocation = &datav1alpha1.DataRestoreLocation{
					NodeName: "test-node",
					Path:     "local:///tmp/restore",
				}
			})

			It("should return true", func() {
				shouldRestore, err := engine.shouldRestoreMetadata()
				Expect(err).To(BeNil())
				Expect(shouldRestore).To(BeTrue())
			})
		})
	})

	Describe("Test AlluxioEngine.RestoreMetadataInternal()", func() {
		When("metadata restored in a host path", func() {
			BeforeEach(func() {
				dataset.Spec.DataRestoreLocation = &datav1alpha1.DataRestoreLocation{
					Path:     "local:///host1/erf",
					NodeName: "test-node",
				}

			})

			It("should restore metadata successfully", func() {
				patch := gomonkey.ApplyMethodSeq(operations.AlluxioFileUtils{}, "QueryMetaDataInfoIntoFile", []gomonkey.OutputCell{
					{Times: 1, Values: gomonkey.Params{"1024", nil}},
					{Times: 1, Values: gomonkey.Params{"100", nil}},
				})
				defer patch.Reset()
				err := engine.RestoreMetadataInternal()
				Expect(err).To(BeNil())
				dataset, err := utils.GetDataset(engine.Client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(dataset.Status.UfsTotal).To(Equal("1.00KiB"))
				Expect(dataset.Status.FileNum).To(Equal("100"))
			})
		})

		When("metadata stored in a file in pvc", func() {
			BeforeEach(func() {
				dataset.Spec.DataRestoreLocation = &datav1alpha1.DataRestoreLocation{
					Path: "pvc://mypvc/myrestoredir/",
				}
			})

			It("should restore metadata from pvc", func() {
				patch := gomonkey.ApplyMethodSeq(operations.AlluxioFileUtils{}, "QueryMetaDataInfoIntoFile", []gomonkey.OutputCell{
					{Times: 1, Values: gomonkey.Params{"1073741824", nil}},
					{Times: 1, Values: gomonkey.Params{"500", nil}},
				})
				defer patch.Reset()
				err := engine.RestoreMetadataInternal()
				Expect(err).To(BeNil())
				dataset, err := utils.GetDataset(engine.Client, dataset.Name, dataset.Namespace)
				Expect(err).To(BeNil())
				Expect(dataset.Status.UfsTotal).To(Equal("1.00GiB"))
				Expect(dataset.Status.FileNum).To(Equal("500"))
			})
		})
	})
})
