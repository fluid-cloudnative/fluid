/*
Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// MockExecutions is a mock implementation of CacheFileUtil for testing
type MockExecutions struct {
	MountFunc func(command []string, timeout time.Duration) (stdout string, err error)
}

func (m *MockExecutions) Mount(command []string, timeout time.Duration) (stdout string, err error) {
	if m.MountFunc != nil {
		return m.MountFunc(command, timeout)
	}
	return "", nil
}

var _ = Describe("CacheEngine UpdateOnUFSChange Tests", Label("pkg.ddc.cache.engine.ufs_test.go"), func() {
	var (
		fakeClient client.Client
		engine     *CacheEngine
		testScheme *runtime.Scheme
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(testScheme)).NotTo(HaveOccurred())
		Expect(corev1.AddToScheme(testScheme)).NotTo(HaveOccurred())

		fakeClient = fake.NewClientBuilder().WithScheme(testScheme).WithStatusSubresource(&datav1alpha1.CacheRuntime{}, &datav1alpha1.Dataset{}).Build()

		log := GinkgoLogr
		recorder := record.NewFakeRecorder(100)

		engine = &CacheEngine{
			Client:                 fakeClient,
			Log:                    log,
			Recorder:               recorder,
			name:                   "test-runtime",
			namespace:              "default",
			runtimeType:            "cache",
			engineImpl:             "cache",
			gracefulShutdownLimits: 5,
			retryShutdown:          0,
			syncRetryDuration:      5 * time.Second,
			timeOfLastSync:         time.Now(),
		}
	})

	Describe("PrepareUFS Tests (Public Method)", func() {
		var (
			runtimeClass *datav1alpha1.CacheRuntimeClass
			patches      *gomonkey.Patches
		)

		BeforeEach(func() {
			// Reset patches before each test
			if patches != nil {
				patches.Reset()
			}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when no mount ufs command found", func() {
			It("should return empty string and no error when topology is nil", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
				}
				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})

			It("should return empty string and no error when master is nil", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology:       &datav1alpha1.RuntimeTopology{},
				}
				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})

			It("should return empty string and no error when execution entries is nil", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{},
					},
				}
				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})

			It("should return empty string and no error when MountUFS is nil", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							ExecutionEntries: &datav1alpha1.ExecutionEntries{},
						},
					},
				}
				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})
		})

		Context("when MountUFS exists but getMasterPodInfo fails", func() {
			BeforeEach(func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								MountUFS: &datav1alpha1.ExecutionCommonEntry{
									Command: []string{"/mount.sh"},
								},
							},
						},
					},
				}
			})

			It("should return error from getMasterPodInfo", func() {
				// getMasterPodInfo will fail because Template is not set
				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})
		})

		Context("when MountUFS executes successfully", func() {
			It("should return stdout from Mount command with valid JSON output", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "master",
										},
									},
								},
							},
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								MountUFS: &datav1alpha1.ExecutionCommonEntry{
									Command:        []string{"/mount.sh"},
									TimeoutSeconds: 30,
								},
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				mockExecutions := &MockExecutions{MountFunc: func(command []string, timeout time.Duration) (stdout string, err error) {
					return `{"mounted": ["/mount1", "/mount2"]}`, nil
				}}
				patches = gomonkey.ApplyFunc(NewCacheFileUtil, func(podName, containerName, namespace string, log logr.Logger) CacheFileUtil {
					return mockExecutions
				})

				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(Equal(`{"mounted": ["/mount1", "/mount2"]}`))
			})

			It("should return empty stdout when Mount returns empty output", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "master",
										},
									},
								},
							},
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								MountUFS: &datav1alpha1.ExecutionCommonEntry{
									Command:        []string{"/mount.sh"},
									TimeoutSeconds: 30,
								},
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				mockExecutions := &MockExecutions{MountFunc: func(command []string, timeout time.Duration) (stdout string, err error) {
					return "", nil
				}}
				patches = gomonkey.ApplyFunc(NewCacheFileUtil, func(podName, containerName, namespace string, log logr.Logger) CacheFileUtil {
					return mockExecutions
				})

				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout).To(BeEmpty())
			})

			It("should return error when Mount command fails", func() {
				runtimeClass = &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name: "master",
										},
									},
								},
							},
							ExecutionEntries: &datav1alpha1.ExecutionEntries{
								MountUFS: &datav1alpha1.ExecutionCommonEntry{
									Command:        []string{"/mount.sh"},
									TimeoutSeconds: 30,
								},
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				mockExecutions := &MockExecutions{MountFunc: func(command []string, timeout time.Duration) (stdout string, err error) {
					return "", errors.New("mount command failed")
				}}
				patches = gomonkey.ApplyFunc(NewCacheFileUtil, func(podName, containerName, namespace string, log logr.Logger) CacheFileUtil {
					return mockExecutions
				})

				stdout, err := engine.PrepareUFS(runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("mount command failed"))
				Expect(stdout).To(BeEmpty())
			})
		})
	})

	Describe("UpdateOnUFSChange Tests (Public Method)", func() {
		var patches *gomonkey.Patches

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		// Helper function to update runtime mount time
		updateRuntimeMountTime := func(offset time.Duration) {
			rt := &datav1alpha1.CacheRuntime{}
			Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, rt)).NotTo(HaveOccurred())
			rt.Status.MountTime = &metav1.Time{Time: time.Now().Add(offset)}
			Expect(fakeClient.Status().Update(context.Background(), rt)).NotTo(HaveOccurred())
		}

		// Helper function to create complete test environment for UpdateOnUFSChange
		setupUpdateOnUFSTest := func(mountFunc func([]string, time.Duration) (string, error)) (*datav1alpha1.CacheRuntimeClass, *datav1alpha1.Dataset, *datav1alpha1.CacheRuntime) {
			// Create RuntimeClass with MountUFS
			rc := &datav1alpha1.CacheRuntimeClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-runtime-class",
				},
				FileSystemType: "cache",
				Topology: &datav1alpha1.RuntimeTopology{
					Master: &datav1alpha1.RuntimeComponentDefinition{
						Options: map[string]string{},
						Service: datav1alpha1.RuntimeComponentService{},
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "master",
									},
								},
							},
						},
						ExecutionEntries: &datav1alpha1.ExecutionEntries{
							MountUFS: &datav1alpha1.ExecutionCommonEntry{
								Command:        []string{"/mount.sh"},
								TimeoutSeconds: 30,
							},
						},
					},
					Worker: &datav1alpha1.RuntimeComponentDefinition{
						Options: map[string]string{},
					},
					Client: &datav1alpha1.RuntimeComponentDefinition{
						Options: map[string]string{},
						Service: datav1alpha1.RuntimeComponentService{},
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), rc)).NotTo(HaveOccurred())

			// Create Dataset
			ds := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime",
					Namespace: "default",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "mount1",
							MountPoint: "s3://bucket1/path",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Mounts: []datav1alpha1.Mount{
						{
							Name:       "mount1",
							MountPoint: "s3://bucket1/path",
						},
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), ds)).NotTo(HaveOccurred())

			// Create CacheRuntime
			runtime := &datav1alpha1.CacheRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime",
					Namespace: "default",
					UID:       types.UID("test-uid"),
				},
				Spec: datav1alpha1.CacheRuntimeSpec{
					RuntimeClassName: "test-runtime-class",
					Master: datav1alpha1.CacheRuntimeMasterSpec{
						Replicas: 1,
					},
				},
				Status: datav1alpha1.CacheRuntimeStatus{
					MountTime: &metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				},
			}
			Expect(fakeClient.Create(context.Background(), runtime)).NotTo(HaveOccurred())

			// Create master pod
			masterPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-runtime-master-0",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "master",
							State: corev1.ContainerState{
								Running: &corev1.ContainerStateRunning{
									StartedAt: metav1.Time{Time: time.Now().Add(-2 * time.Hour)},
								},
							},
						},
					},
				},
			}
			Expect(fakeClient.Create(context.Background(), masterPod)).NotTo(HaveOccurred())

			// Trigger UFS update by adding a mount
			datasetToUpdate := &datav1alpha1.Dataset{}
			Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, datasetToUpdate)).NotTo(HaveOccurred())
			datasetToUpdate.Spec.Mounts = append(datasetToUpdate.Spec.Mounts, datav1alpha1.Mount{
				Name:       "mount2",
				MountPoint: "s3://bucket2/path",
			})
			Expect(fakeClient.Update(context.Background(), datasetToUpdate)).NotTo(HaveOccurred())

			// Mock newCacheFileUtils if mountFunc is provided
			if mountFunc != nil {
				mockExecutions := &MockExecutions{MountFunc: mountFunc}
				patches = gomonkey.ApplyFunc(NewCacheFileUtil, func(podName, containerName, namespace string, log logr.Logger) CacheFileUtil {
					return mockExecutions
				})
			}

			return rc, datasetToUpdate, runtime
		}

		Context("when no update is needed", func() {
			It("should return early when no UFS changes and no remount required", func() {
				_, _, rt := setupUpdateOnUFSTest(nil)

				// Remove the added mount to make spec match status (no changes)
				dataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, dataset)).NotTo(HaveOccurred())
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						Name:       "mount1",
						MountPoint: "s3://bucket1/path",
					},
				}
				Expect(fakeClient.Update(context.Background(), dataset)).NotTo(HaveOccurred())

				// Set recent mount time (after pod start) to avoid remount
				updateRuntimeMountTime(-10 * time.Minute)

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).NotTo(HaveOccurred())

				// Verify dataset status remains Bound (not changed to Updating)
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})

		Context("when PrepareUFS returns invalid JSON", func() {
			It("should return JSON parse error", func() {
				_, _, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return "invalid json output", nil
				})

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse mount ufs output"))

				// Verify dataset status was set to Updating before the error
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.UpdatingDatasetPhase))
			})
		})

		Context("when PrepareUFS returns empty output", func() {
			It("should return empty output error", func() {
				_, _, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return "   ", nil // Whitespace only
				})

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty output"))

				// Verify dataset status was set to Updating
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.UpdatingDatasetPhase))
			})
		})

		Context("when syncDatasetMounts fails due to missing mount point", func() {
			It("should return mount point not mounted error", func() {
				_, _, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return `{"mounted": ["/mount1"]}`, nil
				})

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is not yet mounted"))

				// Verify dataset status remains in Updating state
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.UpdatingDatasetPhase))
			})
		})

		Context("when syncDatasetMounts fails due to unexpected mounted paths", func() {
			It("should return unexpected mounted paths error", func() {
				_, _, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return `{"mounted": ["/mount1", "/mount2", "/extra-path"]}`, nil
				})

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unexpected mounted paths remain"))

				// Verify dataset status remains in Updating state
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.UpdatingDatasetPhase))
			})
		})

		Context("when all mount points are correctly mounted", func() {
			It("should successfully update mounts and set status to Bound", func() {
				_, _, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return `{"mounted": ["/mount1", "/mount2"]}`, nil
				})

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).NotTo(HaveOccurred())

				// Verify dataset status was updated to Bound
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))

				// Verify runtime mount time was updated
				updatedRuntime := &datav1alpha1.CacheRuntime{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedRuntime)).NotTo(HaveOccurred())
				Expect(updatedRuntime.Status.MountTime).NotTo(BeNil())
			})
		})

		Context("when dataset has native scheme mounts", func() {
			It("should skip native scheme mounts and succeed", func() {
				_, dataset, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return `{"mounted": ["/mount1", "/mount2"]}`, nil
				})

				// Add a native scheme mount
				dataset.Spec.Mounts = append(dataset.Spec.Mounts, datav1alpha1.Mount{
					Name:       "native-mount",
					MountPoint: "local:///mnt/local",
				})
				Expect(fakeClient.Update(context.Background(), dataset)).NotTo(HaveOccurred())

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).NotTo(HaveOccurred())

				// Verify dataset status was updated to Bound
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})

		Context("when triggered by master pod restart", func() {
			It("should detect restart and perform remount", func() {
				_, _, rt := setupUpdateOnUFSTest(func(command []string, timeout time.Duration) (stdout string, err error) {
					return `{"mounted": ["/mount1", "/mount2"]}`, nil
				})

				// Remove the added mount to make spec match status (no UFS changes)
				dataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, dataset)).NotTo(HaveOccurred())
				dataset.Spec.Mounts = []datav1alpha1.Mount{
					{
						Name:       "mount1",
						MountPoint: "s3://bucket1/path",
					},
				}
				Expect(fakeClient.Update(context.Background(), dataset)).NotTo(HaveOccurred())

				// Set old mount time (before pod start) to trigger remount
				updateRuntimeMountTime(-2 * time.Hour)

				err := engine.UpdateOnUFSChange(rt)
				Expect(err).NotTo(HaveOccurred())

				// Verify dataset status was updated to Bound after remount
				updatedDataset := &datav1alpha1.Dataset{}
				Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime", Namespace: "default"}, updatedDataset)).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
			})
		})
	})
})
