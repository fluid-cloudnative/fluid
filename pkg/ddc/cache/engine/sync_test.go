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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CacheEngine Sync Tests", Label("pkg.ddc.cache.engine.sync_test.go"), func() {
	var (
		fakeClient client.Client
		engine     *CacheEngine
		ctx        cruntime.ReconcileRequestContext
		testScheme *runtime.Scheme
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(testScheme)).NotTo(HaveOccurred())
		Expect(corev1.AddToScheme(testScheme)).NotTo(HaveOccurred())

		fakeClient = fake.NewClientBuilder().WithScheme(testScheme).Build()

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

		ctx = cruntime.ReconcileRequestContext{
			Context: context.Background(),
			NamespacedName: types.NamespacedName{
				Name:      "test-runtime",
				Namespace: "default",
			},
			Runtime:     &datav1alpha1.CacheRuntime{},
			RuntimeType: "cache",
			EngineImpl:  "cache",
			Client:      fakeClient,
			Log:         log,
			Recorder:    recorder,
		}
	})

	Describe("Sync - Main Entry Point Tests", func() {
		Context("when CacheRuntime does not exist", func() {
			It("should return error from getRuntime", func() {
				err := engine.Sync(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when CacheRuntime exists but CacheRuntimeClass is missing", func() {
			BeforeEach(func() {
				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-123"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "non-existent-class",
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
			})

			It("should fail when generating configmap data due to missing runtime class", func() {
				err := engine.Sync(ctx)
				Expect(err).To(HaveOccurred())
				// The error occurs during generateRuntimeConfigData when trying to get runtime class
			})
		})

		Context("when CacheRuntime and CacheRuntimeClass exist but Dataset is missing", func() {
			BeforeEach(func() {
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{},
							Service: datav1alpha1.RuntimeComponentService{},
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
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-456"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class",
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
			})

			It("should fail when generating configmap data due to missing dataset", func() {
				err := engine.Sync(ctx)
				Expect(err).To(HaveOccurred())
				// Error occurs in generateRuntimeConfigData when GetDataset fails
			})
		})

		Context("when all dependencies exist and ConfigMap needs to be created", func() {
			BeforeEach(func() {
				// Create Dataset
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test-mount",
								MountPoint: "local:///mnt/test",
								Path:       "/data",
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), dataset)).NotTo(HaveOccurred())

				// Create RuntimeClass
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{"master-key": "master-value"},
							Service: datav1alpha1.RuntimeComponentService{},
						},
						Worker: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{"worker-key": "worker-value"},
						},
						Client: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{"client-key": "client-value"},
							Service: datav1alpha1.RuntimeComponentService{},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				// Create Runtime
				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-789"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class",
						Master: datav1alpha1.CacheRuntimeMasterSpec{
							Replicas: 1,
						},
						Worker: datav1alpha1.CacheRuntimeWorkerSpec{
							Replicas: 2,
						},
						Client: datav1alpha1.CacheRuntimeClientSpec{
							RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
								Disabled: false,
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
			})

			It("should create ConfigMap with correct owner reference and data", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify ConfigMap was created
				configMap := &corev1.ConfigMap{}
				err = fakeClient.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, configMap)
				Expect(err).NotTo(HaveOccurred())

				// Verify owner reference
				Expect(configMap.OwnerReferences).To(HaveLen(1))
				Expect(configMap.OwnerReferences[0].Name).To(Equal("test-runtime"))
				Expect(configMap.OwnerReferences[0].UID).To(Equal(types.UID("test-uid-789")))
				Expect(*configMap.OwnerReferences[0].Controller).To(BeTrue())

				// Verify ConfigMap data contains runtime.json key
				Expect(configMap.Data).To(HaveKey("runtime.json"))
				Expect(configMap.Data["runtime.json"]).NotTo(BeEmpty())
			})
		})

		Context("when ConfigMap already exists with same data", func() {
			BeforeEach(func() {
				// Create Dataset
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test-mount",
								MountPoint: "local:///mnt/test",
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), dataset)).NotTo(HaveOccurred())

				// Create RuntimeClass
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{},
							Service: datav1alpha1.RuntimeComponentService{},
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
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				// Create Runtime
				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-sync"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class",
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
			})

			It("should not update ConfigMap when data is unchanged", func() {
				// First call to Sync will create the ConfigMap
				err := engine.Sync(ctx)
				// May fail at UFS step, but ConfigMap should be created
				Expect(err).NotTo(HaveOccurred())

				// Get the created ConfigMap
				originalCM := &corev1.ConfigMap{}
				err = fakeClient.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, originalCM)
				Expect(err).NotTo(HaveOccurred())
				originalData := originalCM.Data["runtime.json"]
				Expect(originalData).NotTo(BeEmpty())

				// Second call to Sync should not change the ConfigMap data
				err = engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify ConfigMap data was not changed
				updatedCM := &corev1.ConfigMap{}
				err = fakeClient.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, updatedCM)
				Expect(err).NotTo(HaveOccurred())
				// The data content should remain exactly the same
				Expect(updatedCM.Data["runtime.json"]).To(Equal(originalData))
			})
		})

		Context("when ConfigMap exists with different data", func() {
			BeforeEach(func() {
				// Create Dataset
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "new-mount",
								MountPoint: "s3://new-bucket",
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), dataset)).NotTo(HaveOccurred())

				// Create RuntimeClass
				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{},
							Service: datav1alpha1.RuntimeComponentService{},
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
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				// Create Runtime
				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-update"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class",
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())

				// Pre-create ConfigMap with old data
				oldCM := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fluid-runtime-config-test-runtime",
						Namespace: "default",
					},
					Data: map[string]string{
						"runtime.json": `{"old":"data"}`,
					},
				}
				Expect(fakeClient.Create(context.Background(), oldCM)).NotTo(HaveOccurred())
			})

			It("should update ConfigMap with new data", func() {
				err := engine.Sync(ctx)
				// May fail at UFS step, but ConfigMap should be updated
				Expect(err).NotTo(HaveOccurred())

				// Verify ConfigMap was updated
				updatedCM := &corev1.ConfigMap{}
				err = fakeClient.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, updatedCM)
				Expect(err).NotTo(HaveOccurred())

				// Data should no longer contain old data
				Expect(updatedCM.Data["runtime.json"]).NotTo(ContainSubstring(`"old":"data"`))
				// Should contain new mount information
				Expect(updatedCM.Data["runtime.json"]).To(ContainSubstring("new-mount"))
				Expect(updatedCM.Data["runtime.json"]).To(ContainSubstring("s3://new-bucket"))
			})
		})

		Context("when UpdateOnUFSChange encounters various scenarios", func() {
			BeforeEach(func() {
				// Setup basic dependencies for all UFS tests
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test-mount",
								MountPoint: "local:///mnt/test",
							},
						},
					},
					Status: datav1alpha1.DatasetStatus{
						Phase: datav1alpha1.BoundDatasetPhase,
					},
				}
				Expect(fakeClient.Create(context.Background(), dataset)).NotTo(HaveOccurred())

				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{},
							Service: datav1alpha1.RuntimeComponentService{},
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
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())
			})

			Context("when master is disabled", func() {
				BeforeEach(func() {
					rt := &datav1alpha1.CacheRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-runtime",
							Namespace: "default",
							UID:       types.UID("test-uid-disabled"),
						},
						Spec: datav1alpha1.CacheRuntimeSpec{
							RuntimeClassName: "test-runtime-class",
							Master: datav1alpha1.CacheRuntimeMasterSpec{
								RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
									Disabled: true,
								},
							},
						},
					}
					Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
				})

				It("should skip UFS update when master is disabled", func() {
					err := engine.Sync(ctx)
					// Should succeed because UFS update is skipped when master is disabled
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when runtime class has no MountUFS execution entry", func() {
				BeforeEach(func() {
					// Update runtime class without MountUFS
					runtimeClass := &datav1alpha1.CacheRuntimeClass{}
					Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-runtime-class"}, runtimeClass)).NotTo(HaveOccurred())
					runtimeClass.Topology.Master.ExecutionEntries = nil
					Expect(fakeClient.Update(context.Background(), runtimeClass)).NotTo(HaveOccurred())

					rt := &datav1alpha1.CacheRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-runtime",
							Namespace: "default",
							UID:       types.UID("test-uid-no-mount"),
						},
						Spec: datav1alpha1.CacheRuntimeSpec{
							RuntimeClassName: "test-runtime-class",
						},
					}
					Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
				})

				It("should skip UFS update when no MountUFS command is defined", func() {
					err := engine.Sync(ctx)
					// Should succeed because UFS update is skipped when no MountUFS is defined
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when no UFS changes are detected", func() {
				BeforeEach(func() {
					now := time.Now()
					rt := &datav1alpha1.CacheRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-runtime",
							Namespace: "default",
							UID:       types.UID("test-uid-stable"),
						},
						Spec: datav1alpha1.CacheRuntimeSpec{
							RuntimeClassName: "test-runtime-class",
						},
						Status: datav1alpha1.CacheRuntimeStatus{
							MountTime: &metav1.Time{Time: now.Add(-5 * time.Minute)},
						},
					}
					Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())

					// Create master pod with old start time
					masterPod := &corev1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-runtime-master-0",
							Namespace: "default",
						},
						Status: corev1.PodStatus{
							ContainerStatuses: []corev1.ContainerStatus{
								{
									Name: "master",
									State: corev1.ContainerState{
										Running: &corev1.ContainerStateRunning{
											StartedAt: metav1.Time{Time: now.Add(-1 * time.Hour)},
										},
									},
								},
							},
						},
					}
					Expect(fakeClient.Create(context.Background(), masterPod)).NotTo(HaveOccurred())
				})

				It("should skip UFS update when no changes detected", func() {
					err := engine.Sync(ctx)
					// Should succeed because no UFS changes are detected
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when runtime spec has disabled components", func() {
			BeforeEach(func() {
				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "test-mount",
								MountPoint: "local:///mnt/test",
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), dataset)).NotTo(HaveOccurred())

				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{},
							Service: datav1alpha1.RuntimeComponentService{},
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
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-disabled-comp"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class",
						Master: datav1alpha1.CacheRuntimeMasterSpec{
							RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
								Disabled: true,
							},
						},
						Worker: datav1alpha1.CacheRuntimeWorkerSpec{
							RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
								Disabled: true,
							},
						},
						Client: datav1alpha1.CacheRuntimeClientSpec{
							RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{
								Disabled: true,
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
			})

			It("should generate configmap without disabled components", func() {
				err := engine.Sync(ctx)
				// May fail at UFS step since master is disabled, but ConfigMap should be created
				Expect(err).NotTo(HaveOccurred())

				configMap := &corev1.ConfigMap{}
				err = fakeClient.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, configMap)
				Expect(err).NotTo(HaveOccurred())

				// Verify configmap data does not contain master/worker/client sections
				jsonData := configMap.Data["runtime.json"]
				Expect(jsonData).NotTo(BeEmpty())
				// When all components are disabled, they should not appear in the config
				Expect(jsonData).NotTo(ContainSubstring(`"master":`))
				Expect(jsonData).NotTo(ContainSubstring(`"worker":`))
				Expect(jsonData).NotTo(ContainSubstring(`"client":`))
			})
		})

		Context("when dataset has shared options and encrypt options", func() {
			BeforeEach(func() {
				// Create secret for encrypt options
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"access-key": []byte("my-access-key"),
					},
				}
				Expect(fakeClient.Create(context.Background(), secret)).NotTo(HaveOccurred())

				dataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						SharedOptions: map[string]string{
							"shared-opt": "shared-value",
						},
						SharedEncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: "SHARED_SECRET",
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "test-secret",
										Key:  "access-key",
									},
								},
							},
						},
						Mounts: []datav1alpha1.Mount{
							{
								Name:       "s3-mount",
								MountPoint: "s3://my-bucket",
								Options: map[string]string{
									"mount-opt": "mount-value",
								},
								EncryptOptions: []datav1alpha1.EncryptOption{
									{
										Name: "MOUNT_SECRET",
										ValueFrom: datav1alpha1.EncryptOptionSource{
											SecretKeyRef: datav1alpha1.SecretKeySelector{
												Name: "test-secret",
												Key:  "access-key",
											},
										},
									},
								},
							},
						},
					},
				}
				Expect(fakeClient.Create(context.Background(), dataset)).NotTo(HaveOccurred())

				runtimeClass := &datav1alpha1.CacheRuntimeClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-runtime-class",
					},
					FileSystemType: "cache",
					Topology: &datav1alpha1.RuntimeTopology{
						Master: &datav1alpha1.RuntimeComponentDefinition{
							Options: map[string]string{},
							Service: datav1alpha1.RuntimeComponentService{},
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
				Expect(fakeClient.Create(context.Background(), runtimeClass)).NotTo(HaveOccurred())

				rt := &datav1alpha1.CacheRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       types.UID("test-uid-encrypt"),
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-runtime-class",
					},
				}
				Expect(fakeClient.Create(context.Background(), rt)).NotTo(HaveOccurred())
			})

			It("should include shared and encrypt options in configmap", func() {
				err := engine.Sync(ctx)
				// May fail at UFS step, but ConfigMap should be created with correct options
				Expect(err).NotTo(HaveOccurred())

				configMap := &corev1.ConfigMap{}
				err = fakeClient.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, configMap)
				Expect(err).NotTo(HaveOccurred())

				jsonData := configMap.Data["runtime.json"]
				Expect(jsonData).To(ContainSubstring("shared-opt"))
				Expect(jsonData).To(ContainSubstring("shared-value"))
				Expect(jsonData).To(ContainSubstring("mount-opt"))
				Expect(jsonData).To(ContainSubstring("mount-value"))
				// Encrypt options should be converted to file paths
				Expect(jsonData).To(ContainSubstring("/etc/fluid/secrets/test-secret/access-key"))
			})
		})
	})

})
