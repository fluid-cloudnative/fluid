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
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

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

var _ = Describe("AlluxioEngine UFS related tests", func() {
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

	Describe("Test AlluxioEngine.ShouldCheckUFS()", func() {
		It("should always return true for AlluxioEngine", func() {
			should, err := engine.ShouldCheckUFS()
			Expect(err).To(BeNil())
			Expect(should).To(BeTrue())
		})
	})

	Describe("Test AlluxioEngine.PrepareUFS()", func() {
		When("mount with configmap", func() {
			BeforeEach(func() {
				os.Setenv(MountConfigStorage, ConfigmapStorageName)
			})

			AfterEach(func() {
				os.Unsetenv(MountConfigStorage)
			})

			When("Alluxioruntime's spec's master replica is 1", func() {
				BeforeEach(func() {
					alluxioruntime.Spec.Master.Replicas = 1
				})

				It("should exec mount script defined in the configmap", func() {
					patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "ExecMountScripts", func() error {
						return nil
					})
					defer patch.Reset()

					patch2 := gomonkey.ApplyMethodFunc(engine, "SyncMetadata", func() error {
						return nil
					})
					defer patch2.Reset()
					err := engine.PrepareUFS()
					Expect(err).To(BeNil())
				})
			})

			When("AlluxioRuntime's spec's master replicas is 0", func() {
				BeforeEach(func() {
					alluxioruntime.Spec.Master.Replicas = 0
				})

				It("should skip executing mount script", func() {
					patch := gomonkey.ApplyMethodFunc(operations.AlluxioFileUtils{}, "ExecMountScripts", func() error {
						return fmt.Errorf("AlluxioFileUtils{}.ExecMountScripts shouldn't be called")
					})
					defer patch.Reset()

					patch2 := gomonkey.ApplyMethodFunc(engine, "SyncMetadata", func() error {
						return nil
					})
					defer patch2.Reset()

					err := engine.PrepareUFS()
					Expect(err).To(BeNil())
				})
			})
		})

		When("mount with remote exec", func() {
			It("should mount ufs onto Alluxio with kube-exec", func() {
				patch1 := gomonkey.ApplyPrivateMethod(engine, "shouldMountUFS", func() (bool, error) {
					return true, nil
				})
				defer patch1.Reset()

				patch2 := gomonkey.ApplyPrivateMethod(engine, "mountUFS", func() error {
					return nil
				})
				defer patch2.Reset()

				patch3 := gomonkey.ApplyMethodFunc(engine, "SyncMetadata", func() error {
					return nil
				})
				defer patch3.Reset()

				err := engine.PrepareUFS()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Test AlluxioEngine.genUFSMountOptions()", func() {
		When("there's no encrypt options", func() {
			Context("and same key exists in both shared options and mount options", func() {
				It("should override the key-value pair in shared options with the one in mount options", func() {
					mount := datav1alpha1.Mount{
						Options: map[string]string{
							"key1":     "opt-value1",
							"key2":     "opt-value2",
							"some-key": "some-value",
						},
					}

					sharedOptions := map[string]string{
						"key1": "shared-opt-value1",
						"key2": "shared-opt-value2",
						"key3": "shared-opt-value3",
						"key4": "shared-opt-value4",
					}

					gotOptions, err := engine.genUFSMountOptions(mount, sharedOptions, []datav1alpha1.EncryptOption{}, true)
					Expect(err).To(BeNil())
					Expect(gotOptions).To(HaveKeyWithValue("key1", "opt-value1"))
					Expect(gotOptions).To(HaveKeyWithValue("key2", "opt-value2"))
					Expect(gotOptions).To(HaveKeyWithValue("some-key", "some-value"))
					Expect(gotOptions).To(HaveKeyWithValue("key3", "shared-opt-value3"))
					Expect(gotOptions).To(HaveKeyWithValue("key4", "shared-opt-value4"))
				})
			})
		})

		When("extractEncryptOptions == true", func() {
			BeforeEach(func() {
				secret := corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "sensitive-options",
						Namespace: dataset.Namespace,
					},
					Data: map[string][]byte{
						"sensitive-key1": []byte("sensitive-value1"),
						"sensitive-key2": []byte("sensitive-value2"),
					},
				}
				resources = append(resources, &secret)
			})
			It("should extract encrypt options directly from secrets", func() {
				mount := datav1alpha1.Mount{
					Options: map[string]string{
						"foo": "bar1",
					},
				}
				sharedOptions := map[string]string{
					"foo": "bar2",
				}
				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "my-sensitive-key1",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "sensitive-options",
								Key:  "sensitive-key1",
							},
						},
					},
					{
						Name: "my-sensitive-key2",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "sensitive-options",
								Key:  "sensitive-key2",
							},
						},
					},
				}

				gotOptions, err := engine.genUFSMountOptions(mount, sharedOptions, sharedEncryptOptions, true)
				Expect(err).To(BeNil())
				Expect(gotOptions).To(HaveKeyWithValue("foo", "bar1"))
				Expect(gotOptions).To(HaveKeyWithValue("my-sensitive-key1", "sensitive-value1"))
				Expect(gotOptions).To(HaveKeyWithValue("my-sensitive-key2", "sensitive-value2"))
			})
		})

		When("extractEncryptOptions == false", func() {
			It("should transform encrypt options to secret mount path", func() {
				mount := datav1alpha1.Mount{
					Options: map[string]string{
						"foo": "bar1",
					},
				}

				sharedOptions := map[string]string{
					"foo": "bar2",
				}

				sharedEncryptOptions := []datav1alpha1.EncryptOption{
					{
						Name: "my-sensitive-key1",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "mysecret",
								Key:  "sensitive-key1",
							},
						},
					},
					{
						Name: "my-sensitive-key2",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{
								Name: "mysecret",
								Key:  "sensitive-key2",
							},
						},
					},
				}

				gotOptions, err := engine.genUFSMountOptions(mount, sharedOptions, sharedEncryptOptions, false)
				Expect(err).To(BeNil())
				Expect(gotOptions).To(HaveKeyWithValue("foo", "bar1"))
				Expect(gotOptions).To(HaveKeyWithValue("my-sensitive-key1", "/etc/fluid/secrets/mysecret/sensitive-key1"))
				Expect(gotOptions).To(HaveKeyWithValue("my-sensitive-key2", "/etc/fluid/secrets/mysecret/sensitive-key2"))
			})
		})
	})

	Describe("Test AlluxioEngine.CheckIfRemountRequired", func() {
		Context("if everything works as expected", func() {
			var timeNow metav1.Time
			BeforeEach(func() {
				timeNow = metav1.Now()
				masterPodName, masterContainerName := engine.getMasterPodInfo()
				masterPod := corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      masterPodName,
						Namespace: engine.namespace,
					},
					Status: corev1.PodStatus{
						ContainerStatuses: []corev1.ContainerStatus{
							{
								Name: masterContainerName,
								State: corev1.ContainerState{
									Running: &corev1.ContainerStateRunning{
										StartedAt: timeNow,
									},
								},
							},
						},
					},
				}

				resources = append(resources, &masterPod)
			})

			When("mount time in runtime's status is after master container's start time", func() {
				var mountTime metav1.Time
				BeforeEach(func() {
					mountTime = metav1.NewTime(timeNow.Add(time.Second))
					alluxioruntime.Status.MountTime = &mountTime
				})

				It("shouldn't try to remount", func() {
					pitfallPatch := gomonkey.ApplyMethodFunc(engine, "FindUnmountedUFS", func() ([]string, error) {
						return []string{}, fmt.Errorf("should not call function FindUnmountedUFS() here")
					})
					defer pitfallPatch.Reset()

					origRuntime, err := utils.GetAlluxioRuntime(engine.Client, alluxioruntime.Name, alluxioruntime.Namespace)
					Expect(err).To(BeNil())
					origUfsToUpdate := utils.NewUFSToUpdate(dataset)
					ufsToUpdate := utils.NewUFSToUpdate(dataset)
					engine.checkIfRemountRequired(ufsToUpdate)
					Expect(ufsToUpdate).To(Equal(origUfsToUpdate))

					// mount time should not be updated
					gotRuntime, err := utils.GetAlluxioRuntime(engine.Client, alluxioruntime.Name, alluxioruntime.Namespace)
					Expect(err).To(BeNil())

					// cannot compare mountTime and gotRuntime.Status.MountTime directly because
					// time.Time has nanosecond precision problem when involving json marshal/unmarshal(called in client.Get(xxx))
					Expect(*gotRuntime.Status.MountTime).To(Equal(*origRuntime.Status.MountTime))
				})
			})

			When("mount time in runtime's status is before master container's start time", func() {
				var mountTime metav1.Time
				BeforeEach(func() {
					mountTime = metav1.NewTime(timeNow.Add(-time.Second))
					alluxioruntime.Status.MountTime = &mountTime
				})

				Context("and there are unmounted ufs", func() {
					It("should try to remount", func() {
						patch := gomonkey.ApplyMethodFunc(engine, "FindUnmountedUFS", func() ([]string, error) {
							return []string{"s3://mybucket", "oss://myossbucket"}, nil
						})
						defer patch.Reset()

						ufsToUpdate := utils.NewUFSToUpdate(dataset)
						engine.checkIfRemountRequired(ufsToUpdate)
						Expect(ufsToUpdate.ToAdd()).To(Equal([]string{"s3://mybucket", "oss://myossbucket"}))
					})
				})

				Context("and there are no unmounted ufs", func() {
					It("should update the mount time in runtime's status", func() {
						patch := gomonkey.ApplyMethodFunc(engine, "FindUnmountedUFS", func() ([]string, error) {
							return []string{}, nil
						})
						defer patch.Reset()

						origRuntime, err := utils.GetAlluxioRuntime(engine.Client, alluxioruntime.Name, alluxioruntime.Namespace)
						Expect(err).To(BeNil())

						origUfsToUpdate := utils.NewUFSToUpdate(dataset)
						ufsToUpdate := utils.NewUFSToUpdate(dataset)
						engine.checkIfRemountRequired(ufsToUpdate)
						Expect(origUfsToUpdate).To(Equal(ufsToUpdate))

						gotRuntime, err := utils.GetAlluxioRuntime(engine.Client, alluxioruntime.Name, alluxioruntime.Namespace)
						Expect(err).To(BeNil())
						Expect(gotRuntime.Status.MountTime.After(origRuntime.Status.MountTime.Time)).To(BeTrue())
					})
				})

			})
		})

	})
})

func mockExecCommandInContainerForTotalStorageBytes() (stdout string, stderr string, err error) {
	r := `File Count               Folder Count             Folder Size
	50000                    1000                     6706560319`
	return r, "", nil
}

func mockExecCommandInContainerForTotalFileNums() (stdout string, stderr string, err error) {
	r := `Master.FilesCompleted  (Type: COUNTER, Value: 1,331,167)`
	return r, "", nil
}

// TestUsedStorageBytes tests the UsedStorageBytes method of the AlluxioEngine.
// It verifies that the method returns the expected used storage value and error status.
// Currently, it checks a basic case where the expected used storage is 0 and no error is expected.
func TestUsedStorageBytes(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name:      "test",
			fields:    fields{},
			wantValue: 0,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{}
			gotValue, err := e.UsedStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.UsedStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.UsedStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestFreeStorageBytes is a unit test for the AlluxioEngine.FreeStorageBytes method.
// This test function defines a set of test cases, each including the expected return value and a flag indicating whether an error is expected.
// The test invokes the FreeStorageBytes method and checks whether the returned value matches the expected result and whether error handling is performed correctly.
func TestFreeStorageBytes(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name:      "test",
			fields:    fields{},
			wantValue: 0,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{}
			gotValue, err := e.FreeStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.FreeStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.FreeStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestTotalStorageBytes verifies the functionality of AlluxioEngine's TotalStorageBytes method.
// It validates whether the method correctly calculates total storage capacity by:
// - Mocking AlluxioRuntime configuration and container command execution
// - Testing both normal scenarios (expected values) and error conditions
// - Using patched container command output to ensure predictable test results
// Each test case checks if returned values match expectations and errors are properly handled.
func TestTotalStorageBytes(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.AlluxioRuntime
		name    string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name: "spark",
					},
				},
			},
			wantValue: 6706560319,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime: tt.fields.runtime,
				name:    tt.fields.name,
			}
			patch1 := gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForTotalStorageBytes()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := e.TotalStorageBytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.TotalStorageBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.TotalStorageBytes() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestTotalFileNums validates the AlluxioEngine's ability to correctly retrieve total file numbers from the Alluxio runtime.
// The test performs the following operations:
// - Creates mock AlluxioRuntime configurations
// - Overrides Kubernetes exec command interactions
// - Verifies both value accuracy and error handling
//
// Test Components:
// - fields: Contains the Alluxio runtime configuration and engine identity
// - tests: Table-driven test cases with expected values and error conditions
// !
// Flow:
// 1. Initialize AlluxioEngine with test parameters
// 2. Mock Kubernetes command execution using function patch
// 3. Execute TotalFileNums() method
// 4. Validate against expected values and error states
//
// Note:
// - Uses monkey patching for Kubernetes client isolation
// - Requires proper setup of mockExecCommandInContainerForTotalFileNums
func TestTotalFileNums(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.AlluxioRuntime
		name    string
	}
	tests := []struct {
		name      string
		fields    fields
		wantValue int64
		wantErr   bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name: "spark",
					},
				},
			},
			wantValue: 1331167,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime: tt.fields.runtime,
				name:    tt.fields.name,
			}
			patch1 := gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForTotalFileNums()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			gotValue, err := e.TotalFileNums()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.TotalFileNums() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotValue != tt.wantValue {
				t.Errorf("AlluxioEngine.TotalFileNums() = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

// TestShouldCheckUFS validates the AlluxioEngine's logic for determining whether
// it should perform a check on the Under File System (UFS).
//
// The test performs the following operations:
// - Initializes a minimal AlluxioEngine instance
// - Invokes the ShouldCheckUFS() method
// - Verifies the boolean result and error status
//
// Test Components:
// - tests: A table-driven slice defining expected outcomes for each case
//
// Flow:
// 1. Construct a new AlluxioEngine instance (with default or minimal config)
// 2. Call the ShouldCheckUFS() method
// 3. Check if the returned value matches expected 'wantShould'
// 4. Validate that the presence or absence of an error matches 'wantErr'
//
// Note:
// - This test assumes default internal state is sufficient for logic evaluation
// - It can be extended to include more cases or mocked dependencies if needed
func TestShouldCheckUFS(t *testing.T) {
	tests := []struct {
		name       string
		wantShould bool
		wantErr    bool
	}{
		{
			name:       "test",
			wantShould: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{}
			gotShould, err := e.ShouldCheckUFS()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioEngine.ShouldCheckUFS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotShould != tt.wantShould {
				t.Errorf("AlluxioEngine.ShouldCheckUFS() = %v, want %v", gotShould, tt.wantShould)
			}
		})
	}
}

// TestPrepareUFS tests the PrepareUFS method of AlluxioEngine.
// This method prepares the underlying file system (UFS) by checking
// the Alluxio master state, mounting UFS, and performing necessary
// metadata synchronization.
func TestFindUnmountedUFS(t *testing.T) {

	type fields struct {
		mountPoints          []datav1alpha1.Mount
		wantedUnmountedPaths []string
	}

	tests := []fields{
		{
			mountPoints: []datav1alpha1.Mount{
				{
					MountPoint: "s3://bucket/path/train",
					Path:       "/path1",
				},
			},
			wantedUnmountedPaths: []string{"/path1"},
		},
		{
			mountPoints: []datav1alpha1.Mount{
				{
					MountPoint: "local://mnt/test",
					Path:       "/path2",
				},
			},
			wantedUnmountedPaths: []string{},
		},
		{
			mountPoints: []datav1alpha1.Mount{
				{
					MountPoint: "s3://bucket/path/train",
					Path:       "/path1",
				},
				{
					MountPoint: "local://mnt/test",
					Path:       "/path2",
				},
				{
					MountPoint: "hdfs://endpoint/path/train",
					Path:       "/path3",
				},
			},
			wantedUnmountedPaths: []string{"/path1", "/path3"},
		},
	}

	for index, test := range tests {
		t.Run("test", func(t *testing.T) {
			s := runtime.NewScheme()
			runtime := datav1alpha1.AlluxioRuntime{}
			dataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: test.mountPoints,
				},
			}

			s.AddKnownTypes(datav1alpha1.GroupVersion, &runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, &dataset)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, &runtime, &dataset)

			var afsUtils operations.AlluxioFileUtils
			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(afsUtils), "Ready", func(_ operations.AlluxioFileUtils) bool {
				return true
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(afsUtils), "FindUnmountedAlluxioPaths", func(_ operations.AlluxioFileUtils, alluxioPaths []string) ([]string, error) {
				return alluxioPaths, nil
			})
			defer patch2.Reset()

			e := &AlluxioEngine{
				runtime:            &runtime,
				name:               "test",
				namespace:          "default",
				Log:                fake.NullLogger(),
				Client:             mockClient,
				MetadataSyncDoneCh: nil,
			}

			unmountedPaths, err := e.FindUnmountedUFS()
			if err != nil {
				t.Errorf("AlluxioEngine.FindUnmountedUFS() error = %v", err)
				return
			}
			if (len(unmountedPaths) != 0 || len(test.wantedUnmountedPaths) != 0) &&
				!reflect.DeepEqual(unmountedPaths, test.wantedUnmountedPaths) {
				t.Errorf("%d check failure, want: %s, got: %s", index, strings.Join(test.wantedUnmountedPaths, ","), strings.Join(unmountedPaths, ","))
				return
			}
		})
	}
}

// TestUpdateMountTime verifies if AlluxioEngine's updateMountTime method correctly updates runtime's MountTime status.
// It creates a runtime with outdated MountTime, executes the update method, then checks if MountTime gets refreshed timestamp.
//
// param: t *testing.T - The testing context used for running the test and reporting failures.
//
// returns: None (This is a test function and does not return any value.)
func TestUpdateMountTime(t *testing.T) {
	yesterday := time.Now().AddDate(0, 0, -1)

	type fields struct {
		runtime *datav1alpha1.AlluxioRuntime
	}

	tests := []fields{
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Status: datav1alpha1.RuntimeStatus{
					MountTime: &metav1.Time{
						Time: yesterday,
					},
				},
			},
		},
	}

	for index, test := range tests {
		t.Run("test", func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, test.runtime)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, test.runtime)

			e := &AlluxioEngine{
				runtime:            test.runtime,
				name:               "test",
				namespace:          "default",
				Log:                fake.NullLogger(),
				Client:             mockClient,
				MetadataSyncDoneCh: nil,
			}

			e.updateMountTime()
			runtime, _ := e.getRuntime()
			if runtime.Status.MountTime.Time.Equal(yesterday) {
				t.Errorf("%d check failure, got: %v, unexpected: %v", index, runtime.Status.MountTime.Time, yesterday)
				return
			}
		})
	}
}
