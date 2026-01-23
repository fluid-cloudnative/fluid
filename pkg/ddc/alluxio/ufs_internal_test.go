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
	"errors"
	"os"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	mockDatasetName      = "hbase"
	mockNamespace        = "fluid"
	mockSecretName       = "test-secret"
	mockAlluxioPath      = "/data"
	mockMountPoint       = "s3://mybucket/mypath"
	mockEncryptKeyName   = "aws-access-key"
	mockEncryptKeyValue  = "test-access-key-value"
	mockEncryptSecretKey = "access-key"
	mockMountName        = "test-mount"

	// Test assertion messages
	testErrMsg         = "should return error"
	testUFSNotReadyMsg = "UFS is not ready"

	// Test option keys and values
	testSecretKey       = "secret-key"
	testExistingValue   = "existing-value"
	testMountOptKey     = "mount-opt1"
	testMountOptValue   = "mount-val1"
	testMountSecretKey  = "mount-secret-key"
	testSharedOptKey    = "shared-opt1"
	testSharedOptValue  = "shared-val1"
	testSharedAccessKey = "shared-access-key"
	testCommonKey       = "common-key"
)

var errMock = errors.New("mock error")

var _ = Describe("AlluxioEngine UFS Internal Tests", Label("pkg.ddc.alluxio.ufs_internal_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		mockedObjs     mockedObjects
		fakeClient     client.Client
		resources      []runtime.Object
	)

	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: mockNamespace, Name: mockDatasetName})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
		mockedObjs = mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
		resources = []runtime.Object{
			dataset,
			alluxioruntime,
			mockedObjs.MasterSts,
			mockedObjs.WorkerSts,
			mockedObjs.FuseDs,
			mockedObjs.PersistentVolumeClaim,
			mockedObjs.PersistentVolume,
		}
	})

	JustBeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = fakeClient
	})

	Describe("Test AlluxioEngine.usedStorageBytesInternal()", func() {
		It("should return zero value", func() {
			value, err := engine.usedStorageBytesInternal()
			Expect(err).To(BeNil())
			Expect(value).To(Equal(int64(0)))
		})
	})

	Describe("Test AlluxioEngine.freeStorageBytesInternal()", func() {
		It("should return zero value", func() {
			value, err := engine.freeStorageBytesInternal()
			Expect(err).To(BeNil())
			Expect(value).To(Equal(int64(0)))
		})
	})

	Describe("Test AlluxioEngine.totalStorageBytesInternal()", func() {
		When("AlluxioFileUtils.Count succeeds", func() {
			It("should return the total bytes from Count", func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Count", func(alluxioPath string) (int64, int64, int64, error) {
					return 100, 10, 1024000, nil
				})
				defer patch.Reset()

				total, err := engine.totalStorageBytesInternal()
				Expect(err).To(BeNil())
				Expect(total).To(Equal(int64(1024000)))
			})
		})

		When("AlluxioFileUtils.Count fails", func() {
			It(testErrMsg, func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Count", func(alluxioPath string) (int64, int64, int64, error) {
					return 0, 0, 0, errMock
				})
				defer patch.Reset()

				total, err := engine.totalStorageBytesInternal()
				Expect(err).To(HaveOccurred())
				Expect(total).To(Equal(int64(0)))
			})
		})
	})

	Describe("Test AlluxioEngine.totalFileNumsInternal()", func() {
		When("AlluxioFileUtils.GetFileCount succeeds", func() {
			It("should return the file count", func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "GetFileCount", func() (int64, error) {
					return 5000, nil
				})
				defer patch.Reset()

				count, err := engine.totalFileNumsInternal()
				Expect(err).To(BeNil())
				Expect(count).To(Equal(int64(5000)))
			})
		})

		When("AlluxioFileUtils.GetFileCount fails", func() {
			It(testErrMsg, func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "GetFileCount", func() (int64, error) {
					return 0, errMock
				})
				defer patch.Reset()

				count, err := engine.totalFileNumsInternal()
				Expect(err).To(HaveOccurred())
				Expect(count).To(Equal(int64(0)))
			})
		})
	})

	Describe("Test AlluxioEngine.shouldMountUFS()", func() {
		When("dataset cannot be retrieved", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It(testErrMsg, func() {
				should, err := engine.shouldMountUFS()
				Expect(err).To(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})

		When(testUFSNotReadyMsg, func() {
			It("should return error indicating UFS not ready", func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return false
				})
				defer patch.Reset()

				should, err := engine.shouldMountUFS()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(testUFSNotReadyMsg))
				Expect(should).To(BeFalse())
			})
		})

		When("UFS is ready and mount point is Fluid native scheme", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: "local:///mnt/data",
						Path:       "/local-path",
					},
					{
						MountPoint: "pvc://my-pvc",
						Path:       "/pvc-path",
					},
				}
			})

			It("should skip Fluid native mounts and return false", func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patch.Reset()

				should, err := engine.shouldMountUFS()
				Expect(err).To(BeNil())
				Expect(should).To(BeFalse())
			})
		})

		When("UFS is ready and mount point needs mounting", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
					},
				}
			})

			It("should return true when path is not mounted", func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return false, nil
				})
				defer patchIsMounted.Reset()

				should, err := engine.shouldMountUFS()
				Expect(err).To(BeNil())
				Expect(should).To(BeTrue())
			})

			It("should return false when path is already mounted", func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return true, nil
				})
				defer patchIsMounted.Reset()

				should, err := engine.shouldMountUFS()
				Expect(err).To(BeNil())
				Expect(should).To(BeFalse())
			})
		})

		When("IsMounted check fails", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
					},
				}
			})

			It(testErrMsg, func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return false, errMock
				})
				defer patchIsMounted.Reset()

				should, err := engine.shouldMountUFS()
				Expect(err).To(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})
	})

	Describe("Test AlluxioEngine.mountUFS()", func() {
		When("dataset cannot be retrieved", func() {
			BeforeEach(func() {
				resources = []runtime.Object{}
			})

			It(testErrMsg, func() {
				err := engine.mountUFS()
				Expect(err).To(HaveOccurred())
			})
		})

		When(testUFSNotReadyMsg, func() {
			It(testErrMsg, func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return false
				})
				defer patch.Reset()

				err := engine.mountUFS()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(testUFSNotReadyMsg))
			})
		})

		When("mount point is Fluid native scheme", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: "local:///mnt/data",
						Path:       "/local-path",
					},
				}
			})

			It("should skip mounting and return nil", func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patch.Reset()

				err := engine.mountUFS()
				Expect(err).To(BeNil())
			})
		})

		When("mount point is already mounted", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
					},
				}
			})

			It("should skip mounting", func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return true, nil
				})
				defer patchIsMounted.Reset()

				err := engine.mountUFS()
				Expect(err).To(BeNil())
			})
		})

		When("mount point needs mounting", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
						Name:       mockMountName,
					},
				}
			})

			It("should mount successfully", func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return false, nil
				})
				defer patchIsMounted.Reset()

				patchMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Mount", func(alluxioPath string, ufsPath string, options map[string]string, readOnly bool, shared bool) error {
					return nil
				})
				defer patchMount.Reset()

				err := engine.mountUFS()
				Expect(err).To(BeNil())
			})
		})

		When("IsMounted check fails", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
					},
				}
			})

			It(testErrMsg, func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return false, errMock
				})
				defer patchIsMounted.Reset()

				err := engine.mountUFS()
				Expect(err).To(HaveOccurred())
			})
		})

		When("Mount operation fails", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
						Name:       mockMountName,
					},
				}
			})

			It(testErrMsg, func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchIsMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "IsMounted", func(alluxioPath string) (bool, error) {
					return false, nil
				})
				defer patchIsMounted.Reset()

				patchMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Mount", func(alluxioPath string, ufsPath string, options map[string]string, readOnly bool, shared bool) error {
					return errMock
				})
				defer patchMount.Reset()

				err := engine.mountUFS()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Test AlluxioEngine.genEncryptOptions()", func() {
		var secret *corev1.Secret

		BeforeEach(func() {
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockSecretName,
					Namespace: mockNamespace,
				},
				Data: map[string][]byte{
					mockEncryptSecretKey: []byte(mockEncryptKeyValue),
				},
			}
			resources = append(resources, secret)
		})

		When("encrypt options is empty", func() {
			It("should return original options unchanged", func() {
				mOptions := map[string]string{"key1": "value1"}
				result, err := engine.genEncryptOptions([]datav1alpha1.EncryptOption{}, mOptions, mockMountName)
				Expect(err).To(BeNil())
				Expect(result).To(Equal(mOptions))
			})
		})

		When("encrypt option key already exists in options", func() {
			It(testErrMsg, func() {
				mOptions := map[string]string{mockEncryptKeyName: testExistingValue}
				encryptOpts := []datav1alpha1.EncryptOption{
					{
						Name: mockEncryptKeyName,
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: mockSecretName,
								Key:  mockEncryptSecretKey,
							},
						},
					},
				}

				result, err := engine.genEncryptOptions(encryptOpts, mOptions, mockMountName)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("set more than one times"))
				Expect(result).To(HaveKeyWithValue(mockEncryptKeyName, testExistingValue))
			})
		})

		When("secret does not exist", func() {
			It(testErrMsg, func() {
				mOptions := map[string]string{}
				encryptOpts := []datav1alpha1.EncryptOption{
					{
						Name: mockEncryptKeyName,
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "non-existent-secret",
								Key:  mockEncryptSecretKey,
							},
						},
					},
				}

				result, err := engine.genEncryptOptions(encryptOpts, mOptions, mockMountName)
				Expect(err).To(HaveOccurred())
				Expect(result).NotTo(HaveKey(mockEncryptKeyName))
			})
		})

		When("encrypt options are valid", func() {
			It("should extract secret values into options", func() {
				mOptions := map[string]string{"existing-key": testExistingValue}
				encryptOpts := []datav1alpha1.EncryptOption{
					{
						Name: mockEncryptKeyName,
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: mockSecretName,
								Key:  mockEncryptSecretKey,
							},
						},
					},
				}

				result, err := engine.genEncryptOptions(encryptOpts, mOptions, mockMountName)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKeyWithValue(mockEncryptKeyName, mockEncryptKeyValue))
				Expect(result).To(HaveKeyWithValue("existing-key", testExistingValue))
			})
		})
	})

	Describe("Test AlluxioEngine.processUpdatingUFS()", func() {
		var ufsToUpdate *utils.UFSToUpdate

		BeforeEach(func() {
			ufsToUpdate = utils.NewUFSToUpdate(dataset)
		})

		When("using configmap mount storage", func() {
			BeforeEach(func() {
				os.Setenv(MountConfigStorage, ConfigmapStorageName)
			})

			AfterEach(func() {
				os.Unsetenv(MountConfigStorage)
			})

			It("should call updateUFSWithMountConfigMapScript", func() {
				patchConfigMap := gomonkey.ApplyPrivateMethod(engine, "updateUFSWithMountConfigMapScript", func(dataset *datav1alpha1.Dataset) (bool, error) {
					return true, nil
				})
				defer patchConfigMap.Reset()

				patchSyncMetadata := gomonkey.ApplyMethodFunc(engine, "SyncMetadata", func() error {
					return nil
				})
				defer patchSyncMetadata.Reset()

				ready, err := engine.processUpdatingUFS(ufsToUpdate)
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		When("not using configmap mount storage", func() {
			BeforeEach(func() {
				os.Setenv(MountConfigStorage, "secret")
			})

			AfterEach(func() {
				os.Unsetenv(MountConfigStorage)
			})

			It("should call updatingUFSWithMountCommand", func() {
				patchMountCmd := gomonkey.ApplyPrivateMethod(engine, "updatingUFSWithMountCommand", func(dataset *datav1alpha1.Dataset, ufsToUpdate *utils.UFSToUpdate) (bool, error) {
					return true, nil
				})
				defer patchMountCmd.Reset()

				patchSyncMetadata := gomonkey.ApplyMethodFunc(engine, "SyncMetadata", func() error {
					return nil
				})
				defer patchSyncMetadata.Reset()

				ready, err := engine.processUpdatingUFS(ufsToUpdate)
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		When("updateReady is true and has new mount paths", func() {
			BeforeEach(func() {
				os.Setenv(MountConfigStorage, "secret")
				ufsToUpdate.AddMountPaths([]string{mockAlluxioPath})
			})

			AfterEach(func() {
				os.Unsetenv(MountConfigStorage)
			})

			It("should update mount time", func() {
				patchMountCmd := gomonkey.ApplyPrivateMethod(engine, "updatingUFSWithMountCommand", func(dataset *datav1alpha1.Dataset, ufsToUpdate *utils.UFSToUpdate) (bool, error) {
					return true, nil
				})
				defer patchMountCmd.Reset()

				patchSyncMetadata := gomonkey.ApplyMethodFunc(engine, "SyncMetadata", func() error {
					return nil
				})
				defer patchSyncMetadata.Reset()

				ready, err := engine.processUpdatingUFS(ufsToUpdate)
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})
	})

	Describe("Test AlluxioEngine.updatingUFSWithMountCommand()", func() {
		var ufsToUpdate *utils.UFSToUpdate

		BeforeEach(func() {
			ufsToUpdate = utils.NewUFSToUpdate(dataset)
		})

		When(testUFSNotReadyMsg, func() {
			It(testErrMsg, func() {
				patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return false
				})
				defer patch.Reset()

				ready, err := engine.updatingUFSWithMountCommand(dataset, ufsToUpdate)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(testUFSNotReadyMsg))
				Expect(ready).To(BeFalse())
			})
		})

		When("there are paths to add", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
						Name:       mockMountName,
					},
				}
				ufsToUpdate.AddMountPaths([]string{mockAlluxioPath})
			})

			It("should mount the new paths", func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Mount", func(alluxioPath string, ufsPath string, options map[string]string, readOnly bool, shared bool) error {
					return nil
				})
				defer patchMount.Reset()

				ready, err := engine.updatingUFSWithMountCommand(dataset, ufsToUpdate)
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		When("there are paths to remove", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{}
				dataset.Status.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: "hdfs://old-server/old-path",
						Path:       "/old-path",
						Name:       "old-mount",
					},
				}
				ufsToUpdate = utils.NewUFSToUpdate(dataset)
				ufsToUpdate.AnalyzePathsDelta()
			})

			It("should unmount the removed paths", func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchUnMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "UnMount", func(alluxioPath string) error {
					return nil
				})
				defer patchUnMount.Reset()

				ready, err := engine.updatingUFSWithMountCommand(dataset, ufsToUpdate)
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		When("unmount fails", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{}
				dataset.Status.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: "hdfs://old-server/old-path",
						Path:       "/old-path",
						Name:       "old-mount",
					},
				}
				ufsToUpdate = utils.NewUFSToUpdate(dataset)
				ufsToUpdate.AnalyzePathsDelta()
			})

			It(testErrMsg, func() {
				patchReady := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "Ready", func() bool {
					return true
				})
				defer patchReady.Reset()

				patchUnMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "UnMount", func(alluxioPath string) error {
					return errMock
				})
				defer patchUnMount.Reset()

				ready, err := engine.updatingUFSWithMountCommand(dataset, ufsToUpdate)
				Expect(err).To(HaveOccurred())
				Expect(ready).To(BeFalse())
			})
		})
	})

	Describe("Test AlluxioEngine.updateUFSWithMountConfigMapScript()", func() {
		var mountConfigMap *corev1.ConfigMap

		BeforeEach(func() {
			mountConfigMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engine.getMountConfigmapName(),
					Namespace: mockNamespace,
				},
				Data: map[string]string{
					NON_NATIVE_MOUNT_DATA_NAME: "",
				},
			}
			resources = append(resources, mountConfigMap)
		})

		When("mount configmap is not found", func() {
			BeforeEach(func() {
				resources = []runtime.Object{dataset, alluxioruntime}
			})

			It(testErrMsg, func() {
				patchGetConfigmap := gomonkey.ApplyFunc(kubeclient.GetConfigmapByName, func(client client.Client, name string, namespace string) (*corev1.ConfigMap, error) {
					return nil, nil
				})
				defer patchGetConfigmap.Reset()

				ready, err := engine.updateUFSWithMountConfigMapScript(dataset)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("mount configmap"))
				Expect(ready).To(BeFalse())
			})
		})

		When("mount configmap exists and paths match", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
						Name:       mockMountName,
					},
				}
			})

			It("should execute mount scripts and verify paths", func() {
				patchGetConfigmap := gomonkey.ApplyFunc(kubeclient.GetConfigmapByName, func(c client.Client, name string, namespace string) (*corev1.ConfigMap, error) {
					return mountConfigMap, nil
				})
				defer patchGetConfigmap.Reset()

				patchGenMountsInfo := gomonkey.ApplyPrivateMethod(engine, "generateNonNativeMountsInfo", func(dataset *datav1alpha1.Dataset) ([]string, error) {
					return []string{mockAlluxioPath + " " + mockMountPoint}, nil
				})
				defer patchGenMountsInfo.Reset()

				patchUpdateConfigMap := gomonkey.ApplyFunc(kubeclient.UpdateConfigMap, func(c client.Client, cm *corev1.ConfigMap) error {
					return nil
				})
				defer patchUpdateConfigMap.Reset()

				patchExecMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "ExecMountScripts", func() error {
					return nil
				})
				defer patchExecMount.Reset()

				patchGetMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "GetMountedAlluxioPaths", func() ([]string, error) {
					return []string{mockAlluxioPath, "/"}, nil
				})
				defer patchGetMounted.Reset()

				ready, err := engine.updateUFSWithMountConfigMapScript(dataset)
				Expect(err).To(BeNil())
				Expect(ready).To(BeTrue())
			})
		})

		When("execute mount script fails", func() {
			It(testErrMsg, func() {
				patchGetConfigmap := gomonkey.ApplyFunc(kubeclient.GetConfigmapByName, func(c client.Client, name string, namespace string) (*corev1.ConfigMap, error) {
					return mountConfigMap, nil
				})
				defer patchGetConfigmap.Reset()

				patchGenMountsInfo := gomonkey.ApplyPrivateMethod(engine, "generateNonNativeMountsInfo", func(dataset *datav1alpha1.Dataset) ([]string, error) {
					return []string{}, nil
				})
				defer patchGenMountsInfo.Reset()

				patchExecMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "ExecMountScripts", func() error {
					return errMock
				})
				defer patchExecMount.Reset()

				ready, err := engine.updateUFSWithMountConfigMapScript(dataset)
				Expect(err).To(HaveOccurred())
				Expect(ready).To(BeFalse())
			})
		})

		When("mounted paths do not match dataset mount paths", func() {
			BeforeEach(func() {
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						MountPoint: mockMountPoint,
						Path:       mockAlluxioPath,
						Name:       mockMountName,
					},
					{
						MountPoint: "hdfs://another/path",
						Path:       "/another-path",
						Name:       "another-mount",
					},
				}
			})

			It("should return updateReady as false", func() {
				patchGetConfigmap := gomonkey.ApplyFunc(kubeclient.GetConfigmapByName, func(c client.Client, name string, namespace string) (*corev1.ConfigMap, error) {
					return mountConfigMap, nil
				})
				defer patchGetConfigmap.Reset()

				patchGenMountsInfo := gomonkey.ApplyPrivateMethod(engine, "generateNonNativeMountsInfo", func(dataset *datav1alpha1.Dataset) ([]string, error) {
					return []string{mockAlluxioPath + " " + mockMountPoint, "/another-path hdfs://another/path"}, nil
				})
				defer patchGenMountsInfo.Reset()

				patchUpdateConfigMap := gomonkey.ApplyFunc(kubeclient.UpdateConfigMap, func(c client.Client, cm *corev1.ConfigMap) error {
					return nil
				})
				defer patchUpdateConfigMap.Reset()

				patchExecMount := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "ExecMountScripts", func() error {
					return nil
				})
				defer patchExecMount.Reset()

				patchGetMounted := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "GetMountedAlluxioPaths", func() ([]string, error) {
					return []string{mockAlluxioPath, "/"}, nil
				})
				defer patchGetMounted.Reset()

				ready, err := engine.updateUFSWithMountConfigMapScript(dataset)
				Expect(err).To(BeNil())
				Expect(ready).To(BeFalse())
			})
		})
	})
})

var _ = Describe("AlluxioEngine genUFSMountOptions Additional Tests", Label("pkg.ddc.alluxio.ufs_internal_test.go"), func() {
	var (
		dataset        *datav1alpha1.Dataset
		alluxioruntime *datav1alpha1.AlluxioRuntime
		engine         *AlluxioEngine
		mockedObjs     mockedObjects
		fakeClient     client.Client
		resources      []runtime.Object
	)

	BeforeEach(func() {
		dataset, alluxioruntime = mockFluidObjectsForTests(types.NamespacedName{Namespace: mockNamespace, Name: mockDatasetName})
		engine = mockAlluxioEngineForTests(dataset, alluxioruntime)
		mockedObjs = mockAlluxioObjectsForTests(dataset, alluxioruntime, engine)
		resources = []runtime.Object{
			dataset,
			alluxioruntime,
			mockedObjs.MasterSts,
			mockedObjs.WorkerSts,
			mockedObjs.FuseDs,
			mockedObjs.PersistentVolumeClaim,
			mockedObjs.PersistentVolume,
		}
	})

	JustBeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, resources...)
		engine.Client = fakeClient
	})

	Describe("Test AlluxioEngine.genUFSMountOptions with multiple encrypt options", func() {
		var secret *corev1.Secret

		BeforeEach(func() {
			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mockSecretName,
					Namespace: mockNamespace,
				},
				Data: map[string][]byte{
					mockEncryptSecretKey: []byte("ak-value"),
					testSecretKey:        []byte("sk-value"),
				},
			}
			resources = append(resources, secret)
		})

		When("extractEncryptOptions is true with shared and mount encrypt options", func() {
			It("should merge all options correctly", func() {
				mount := datav1alpha1.Mount{
					Name: mockMountName,
					Options: map[string]string{
						testMountOptKey: testMountOptValue,
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: testMountSecretKey,
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: mockSecretName,
									Key:  testSecretKey,
								},
							},
						},
					},
				}

				sharedOptions := map[string]string{
					testSharedOptKey: testSharedOptValue,
				}

				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: testSharedAccessKey,
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: mockSecretName,
								Key:  mockEncryptSecretKey,
							},
						},
					},
				}

				result, err := engine.genUFSMountOptions(mount, sharedOptions, sharedEncryptOptions, true)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKeyWithValue(testMountOptKey, testMountOptValue))
				Expect(result).To(HaveKeyWithValue(testSharedOptKey, testSharedOptValue))
				Expect(result).To(HaveKeyWithValue(testSharedAccessKey, "ak-value"))
				Expect(result).To(HaveKeyWithValue(testMountSecretKey, "sk-value"))
			})
		})

		When("extractEncryptOptions is false", func() {
			It("should use mount file paths instead of secret values", func() {
				mount := datav1alpha1.Mount{
					Name: mockMountName,
					Options: map[string]string{
						testMountOptKey: testMountOptValue,
					},
					EncryptOptions: []datav1alpha1.EncryptOption{
						{
							Name: testMountSecretKey,
							ValueFrom: datav1alpha1.EncryptOptionSource{
								SecretKeyRef: datav1alpha1.SecretKeySelector{
									Name: mockSecretName,
									Key:  testSecretKey,
								},
							},
						},
					},
				}

				sharedOptions := map[string]string{
					testSharedOptKey: testSharedOptValue,
				}

				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: testSharedAccessKey,
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: mockSecretName,
								Key:  mockEncryptSecretKey,
							},
						},
					},
				}

				result, err := engine.genUFSMountOptions(mount, sharedOptions, sharedEncryptOptions, false)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKeyWithValue(testMountOptKey, testMountOptValue))
				Expect(result).To(HaveKeyWithValue(testSharedOptKey, testSharedOptValue))
				Expect(result).To(HaveKeyWithValue(testSharedAccessKey, "/etc/fluid/secrets/"+mockSecretName+"/access-key"))
				Expect(result).To(HaveKeyWithValue(testMountSecretKey, "/etc/fluid/secrets/"+mockSecretName+"/secret-key"))
			})
		})

		When("genEncryptOptions fails for shared encrypt options", func() {
			It(testErrMsg, func() {
				mount := datav1alpha1.Mount{
					Name:    mockMountName,
					Options: map[string]string{},
				}

				sharedOptions := map[string]string{}

				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "non-existent-key",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "non-existent-secret",
								Key:  "key",
							},
						},
					},
				}

				result, err := engine.genUFSMountOptions(mount, sharedOptions, sharedEncryptOptions, true)
				Expect(err).To(HaveOccurred())
				Expect(result).NotTo(HaveKey("non-existent-key"))
			})
		})
	})

	Describe("Test AlluxioEngine.genUFSMountOptions option override behavior", func() {
		When("mount options override shared options", func() {
			It("should use mount option value", func() {
				mount := datav1alpha1.Mount{
					Name: mockMountName,
					Options: map[string]string{
						testCommonKey: "mount-value",
					},
				}

				sharedOptions := map[string]string{
					testCommonKey: "shared-value",
				}

				result, err := engine.genUFSMountOptions(mount, sharedOptions, []datav1alpha1.EncryptOption{}, true)
				Expect(err).To(BeNil())
				Expect(result).To(HaveKeyWithValue(testCommonKey, "mount-value"))
			})
		})
	})
})
