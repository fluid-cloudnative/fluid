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
	"os"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	cclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CacheEngine Sync Tests", Label("pkg.ddc.cache.engine.sync_test.go"), func() {
	var (
		engine       *CacheEngine
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		dataset      *datav1alpha1.Dataset
		ctx          cruntime.ReconcileRequestContext
		fakeClient   cclient.Client
	)

	BeforeEach(func() {
		scheme := CacheEngineTestScheme

		// Create dataset (name must match runtime name for cache runtime)
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
				UID:       "test-dataset-uid",
			},
			Spec: datav1alpha1.DatasetSpec{},
		}

		// Create runtime
		runtimeObj = &datav1alpha1.CacheRuntime{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "data.fluid.io/v1alpha1",
				Kind:       "CacheRuntime",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime",
				Namespace: "default",
				UID:       "test-runtime-uid",
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-class",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
				Worker:           datav1alpha1.CacheRuntimeWorkerSpec{Replicas: 2},
				Client:           datav1alpha1.CacheRuntimeClientSpec{},
			},
		}
		// Initialize status fields separately due to embedded struct
		runtimeObj.Status.Master.Phase = datav1alpha1.RuntimePhaseNone
		runtimeObj.Status.Worker.Phase = datav1alpha1.RuntimePhaseNone
		runtimeObj.Status.Client.Phase = datav1alpha1.RuntimePhaseNone

		// Create runtime class
		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			ObjectMeta:     metav1.ObjectMeta{Name: "test-class"},
			FileSystemType: "test-fs",
			Topology: &datav1alpha1.RuntimeTopology{
				Master: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
						},
					},
					ExecutionEntries: &datav1alpha1.ExecutionEntries{
						ReportSummary: &datav1alpha1.ExecutionCommonEntry{
							Command:        []string{"summary"},
							TimeoutSeconds: 10,
						},
					},
				},
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
						},
					},
				},
				Client: &datav1alpha1.RuntimeComponentDefinition{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
						},
					},
				},
			},
		}

		// Create master AdvancedStatefulSet
		masterReplicas := int32(1)
		masterSts := &workloadv1alpha1.AdvancedStatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime-master",
				Namespace: "default",
			},
			Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
				Replicas: &masterReplicas,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
					},
				},
			},
			Status: workloadv1alpha1.AdvancedStatefulSetStatus{
				ReadyReplicas:     1,
				CurrentReplicas:   1,
				AvailableReplicas: 1,
			},
		}

		// Create worker AdvancedStatefulSet
		workerReplicas := int32(2)
		workerSts := &workloadv1alpha1.AdvancedStatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime-worker",
				Namespace: "default",
			},
			Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
				Replicas: &workerReplicas,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
					},
				},
			},
			Status: workloadv1alpha1.AdvancedStatefulSetStatus{
				ReadyReplicas:     2,
				CurrentReplicas:   2,
				AvailableReplicas: 2,
			},
		}

		// Create client DaemonSet
		clientDs := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime-client",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
					},
				},
			},
			Status: appsv1.DaemonSetStatus{
				NumberReady:            0,
				DesiredNumberScheduled: 0,
			},
		}

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(dataset, runtimeObj, runtimeClass, masterSts, workerSts, clientDs).
			WithStatusSubresource(dataset, runtimeObj).
			Build()

		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
		}

		ctx = cruntime.ReconcileRequestContext{
			Client:         fakeClient,
			Context:        context.Background(),
			Log:            ctrl.Log.WithName("test"),
			RuntimeType:    "cache",
			NamespacedName: types.NamespacedName{Name: "test-runtime", Namespace: "default"},
		}
	})

	Describe("Sync", func() {
		Context("when runtime exists", func() {
			It("should sync successfully", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when runtime does not exist", func() {
			BeforeEach(func() {
				scheme := runtime.NewScheme()
				_ = datav1alpha1.AddToScheme(scheme)
				_ = appsv1.AddToScheme(scheme)
				engine.Client = fake.NewClientBuilder().WithScheme(scheme).Build()
			})

			It("should return error", func() {
				err := engine.Sync(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when runtimeClass does not exist", func() {
			BeforeEach(func() {
				runtimeObj.Spec.RuntimeClassName = "non-existent-class"
				scheme := runtime.NewScheme()
				_ = datav1alpha1.AddToScheme(scheme)
				_ = appsv1.AddToScheme(scheme)
				engine.Client = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(runtimeObj).
					Build()
			})

			It("should return error", func() {
				err := engine.Sync(ctx)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with existing configmap", func() {
			BeforeEach(func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fluid-runtime-config-test-runtime",
						Namespace: "default",
					},
					Data: map[string]string{"old-key": "old-value"},
				}
				scheme := runtime.NewScheme()
				_ = datav1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				_ = appsv1.AddToScheme(scheme)
				_ = workloadv1alpha1.AddToScheme(scheme)
				// Create AdvancedStatefulSets and DaemonSet for status update
				masterReplicas := int32(1)
				masterSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &masterReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 1, CurrentReplicas: 1, AvailableReplicas: 1},
				}
				workerReplicas := int32(2)
				workerSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &workerReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 2, CurrentReplicas: 2, AvailableReplicas: 2},
				}
				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: "default"},
					Status:     appsv1.DaemonSetStatus{NumberReady: 0, DesiredNumberScheduled: 0},
				}
				engine.Client = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset, runtimeObj, runtimeClass, configMap, masterSts, workerSts, clientDs).
					WithStatusSubresource(runtimeObj).
					Build()
			})

			It("should update configmap", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify configmap was updated
				cm := &corev1.ConfigMap{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).NotTo(BeNil())
			})
		})

		Context("without existing configmap", func() {
			It("should create configmap", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				// Verify configmap was created
				cm := &corev1.ConfigMap{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).NotTo(BeNil())
				Expect(cm.OwnerReferences).NotTo(BeEmpty())
			})
		})

		Context("when runtime is ready and sync is permitted", func() {
			It("should attempt to sync dataset cache states", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when runtime is ready with ReportSummary configured", func() {
			var patches *gomonkey.Patches

			AfterEach(func() {
				if patches != nil {
					patches.Reset()
				}
			})

			It("should get cache states successfully with mocked Execute", func() {

				mockExecutions := &MockExecutions{
					MockExecute: func(command []string, timeout time.Duration) (stdout string, err error) {
						return `{"cached":"1073741824","cachedPercentage":"50","cacheCapacity":"2147483648","cacheHitRatio":"90","fileNum":"100","ufsTotal":"2147483648"}`, nil
					},
				}

				patches = gomonkey.ApplyFunc(NewCacheFileUtil, func(podName, containerName, namespace string, log logr.Logger) CacheFileUtil {
					return mockExecutions
				})

				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				updateDataset := datav1alpha1.Dataset{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "test-runtime",
					Namespace: "default",
				}, &updateDataset)
				Expect(err).NotTo(HaveOccurred())

				cacheStates := updateDataset.Status.CacheStates
				Expect(cacheStates).NotTo(BeNil())
				Expect(updateDataset.Status.Phase, datav1alpha1.BoundDatasetPhase)
				Expect(cacheStates[common.Cached]).To(Equal("1073741824"))
				Expect(cacheStates[common.CachedPercentage]).To(Equal("50"))
				Expect(cacheStates[common.CacheCapacity]).To(Equal("2147483648"))
				Expect(cacheStates[common.CacheHitRatio]).To(Equal("90"))
				Expect(cacheStates[common.FileNum]).To(Equal("100"))
				Expect(cacheStates[common.UfsTotal]).To(Equal("2147483648"))
			})
		})

		Context("when runtime is not ready (master not ready)", func() {
			BeforeEach(func() {
				masterReplicas := int32(1)
				masterSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime-master",
						Namespace: "default",
					},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &masterReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{
						ReadyReplicas:     0,
						CurrentReplicas:   1,
						AvailableReplicas: 0,
					},
				}

				workerReplicas := int32(2)
				workerSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime-worker",
						Namespace: "default",
					},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &workerReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{
						ReadyReplicas:     2,
						CurrentReplicas:   2,
						AvailableReplicas: 2,
					},
				}

				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime-client",
						Namespace: "default",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
							},
						},
					},
					Status: appsv1.DaemonSetStatus{
						NumberReady:            0,
						DesiredNumberScheduled: 0,
					},
				}

				engine = &CacheEngine{
					name:      "test-runtime",
					namespace: "default",
					Client: fake.NewClientBuilder().
						WithScheme(CacheEngineTestScheme).
						WithObjects(dataset, runtimeObj, runtimeClass, masterSts, workerSts, clientDs).
						WithStatusSubresource(dataset, runtimeObj).
						Build(),
					Log: ctrl.Log.WithName("test"),
				}
			})

			It("should not sync dataset cache states when master is not ready", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				updatedDataset := &datav1alpha1.Dataset{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "test-runtime",
					Namespace: "default",
				}, updatedDataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase, datav1alpha1.FailedDatasetPhase)
				Expect(updatedDataset.Status.CacheStates).To(BeNil(), "expected CacheStates to remain nil when runtime is not ready")
			})
		})

		Context("when runtime is not ready (worker not ready)", func() {
			BeforeEach(func() {
				masterReplicas := int32(1)
				masterSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime-master",
						Namespace: "default",
					},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &masterReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{
						ReadyReplicas:     1,
						CurrentReplicas:   1,
						AvailableReplicas: 1,
					},
				}

				workerReplicas := int32(2)
				workerSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime-worker",
						Namespace: "default",
					},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &workerReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{
						ReadyReplicas:     0,
						CurrentReplicas:   2,
						AvailableReplicas: 0,
					},
				}

				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime-client",
						Namespace: "default",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
							},
						},
					},
					Status: appsv1.DaemonSetStatus{
						NumberReady:            0,
						DesiredNumberScheduled: 0,
					},
				}

				fakeClient := fake.NewClientBuilder().
					WithScheme(CacheEngineTestScheme).
					WithObjects(dataset, runtimeObj, runtimeClass, masterSts, workerSts, clientDs).
					WithStatusSubresource(dataset, runtimeObj).
					Build()

				engine = &CacheEngine{
					name:      "test-runtime",
					namespace: "default",
					Client:    fakeClient,
					Log:       ctrl.Log.WithName("test"),
				}
			})

			It("should not sync dataset cache states when worker is not ready", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				updatedDataset := &datav1alpha1.Dataset{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "test-runtime",
					Namespace: "default",
				}, updatedDataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase, datav1alpha1.FailedDatasetPhase)
				Expect(updatedDataset.Status.CacheStates).To(BeNil(), "expected CacheStates to remain nil when runtime is not ready")
			})
		})
	})

	Describe("syncRuntimeValueConfigMap", func() {
		Context("when configmap does not exist", func() {
			It("should create new configmap", func() {
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				// Verify configmap was created
				cm := &corev1.ConfigMap{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Name).To(Equal("fluid-runtime-config-test-runtime"))
				Expect(cm.Namespace).To(Equal("default"))
				Expect(cm.Data).NotTo(BeNil())
				Expect(len(cm.Data)).To(BeNumerically(">", 0))

				// Verify owner reference
				Expect(cm.OwnerReferences).NotTo(BeEmpty())
				Expect(cm.OwnerReferences[0].Name).To(Equal("test-runtime"))
				Expect(cm.OwnerReferences[0].Kind).To(Equal("CacheRuntime"))
				Expect(*cm.OwnerReferences[0].Controller).To(BeTrue())
				Expect(*cm.OwnerReferences[0].BlockOwnerDeletion).To(BeTrue())
			})
		})

		Context("when configmap already exists", func() {
			BeforeEach(func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fluid-runtime-config-test-runtime",
						Namespace: "default",
					},
					Data: map[string]string{"old-key": "old-value"},
				}
				scheme := runtime.NewScheme()
				_ = datav1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				_ = appsv1.AddToScheme(scheme)
				_ = workloadv1alpha1.AddToScheme(scheme)
				// Create AdvancedStatefulSets and DaemonSet for status update
				masterReplicas := int32(1)
				masterSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       workloadv1alpha1.AdvancedStatefulSetSpec{Replicas: &masterReplicas},
					Status:     workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 1, CurrentReplicas: 1, AvailableReplicas: 1},
				}
				workerReplicas := int32(2)
				workerSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &workerReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 2, CurrentReplicas: 2, AvailableReplicas: 2},
				}
				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: "default"},
					Status:     appsv1.DaemonSetStatus{NumberReady: 0, DesiredNumberScheduled: 0},
				}
				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset, runtimeObj, runtimeClass, configMap, masterSts, workerSts, clientDs).
					WithStatusSubresource(runtimeObj).
					Build()
				engine.Client = fakeClient
			})

			It("should update existing configmap", func() {
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				// Verify configmap was updated
				cm := &corev1.ConfigMap{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).NotTo(HaveKey("old-key")) // Old data should be replaced
			})
		})

		Context("when configmap data is unchanged", func() {
			BeforeEach(func() {
				// First sync to create configmap
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				// Get the created configmap
				cm := &corev1.ConfigMap{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, cm)
				Expect(err).NotTo(HaveOccurred())

				// Store original data
				originalData := make(map[string]string)
				for k, v := range cm.Data {
					originalData[k] = v
				}

				// Update engine's client to use same objects
				scheme := runtime.NewScheme()
				_ = datav1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				_ = appsv1.AddToScheme(scheme)
				_ = workloadv1alpha1.AddToScheme(scheme)
				// Create AdvancedStatefulSets and DaemonSet for status update
				masterReplicas := int32(1)
				masterSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       workloadv1alpha1.AdvancedStatefulSetSpec{Replicas: &masterReplicas},
					Status:     workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 1, CurrentReplicas: 1, AvailableReplicas: 1},
				}
				workerReplicas := int32(2)
				workerSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &workerReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 2, CurrentReplicas: 2, AvailableReplicas: 2},
				}
				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: "default"},
					Status:     appsv1.DaemonSetStatus{NumberReady: 0, DesiredNumberScheduled: 0},
				}
				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset, runtimeObj, runtimeClass, cm, masterSts, workerSts, clientDs).
					WithStatusSubresource(runtimeObj).
					Build()
				engine.Client = fakeClient
			})

			It("should not update configmap when data is same", func() {
				// Second sync with same data
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				// Verify configmap still exists
				cm := &corev1.ConfigMap{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "fluid-runtime-config-test-runtime",
					Namespace: "default",
				}, cm)
				Expect(err).NotTo(HaveOccurred())
				Expect(cm.Data).NotTo(BeNil())
			})
		})

		Context("error handling", func() {
			It("should handle empty runtime gracefully", func() {
				// Create a minimal runtime object
				minimalRuntime := &datav1alpha1.CacheRuntime{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "CacheRuntime",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-runtime",
						Namespace: "default",
						UID:       "test-runtime-uid",
					},
					Spec: datav1alpha1.CacheRuntimeSpec{
						RuntimeClassName: "test-class",
					},
				}

				scheme := runtime.NewScheme()
				_ = datav1alpha1.AddToScheme(scheme)
				_ = corev1.AddToScheme(scheme)
				_ = appsv1.AddToScheme(scheme)
				_ = workloadv1alpha1.AddToScheme(scheme)
				// Create necessary resources
				masterReplicas := int32(1)
				masterSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       workloadv1alpha1.AdvancedStatefulSetSpec{Replicas: &masterReplicas},
					Status:     workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 1, CurrentReplicas: 1, AvailableReplicas: 1},
				}
				workerReplicas := int32(2)
				workerSts := &workloadv1alpha1.AdvancedStatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
						Replicas: &workerReplicas,
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
							},
						},
					},
					Status: workloadv1alpha1.AdvancedStatefulSetStatus{ReadyReplicas: 2, CurrentReplicas: 2, AvailableReplicas: 2},
				}
				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: "default"},
					Status:     appsv1.DaemonSetStatus{NumberReady: 0, DesiredNumberScheduled: 0},
				}
				engine.Client = fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset, minimalRuntime, runtimeClass, masterSts, workerSts, clientDs).
					WithStatusSubresource(minimalRuntime).
					Build()

				err := engine.syncRuntimeValueConfigMap(ctx, minimalRuntime)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("getSyncRetryDuration", func() {
		Context("when environment variable is not set", func() {
			It("should return nil duration", func() {
				// Ensure env var is not set
				os.Unsetenv(syncRetryDurationEnv)

				duration, err := getSyncRetryDuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).To(BeNil())
			})
		})

		Context("when environment variable is set to valid duration", func() {
			AfterEach(func() {
				os.Unsetenv(syncRetryDurationEnv)
			})

			It("should parse and return duration", func() {
				os.Setenv(syncRetryDurationEnv, "5s")

				duration, err := getSyncRetryDuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).NotTo(BeNil())
				Expect(*duration).To(Equal(5 * time.Second))
			})
		})

		Context("when environment variable is set to invalid duration", func() {
			AfterEach(func() {
				os.Unsetenv(syncRetryDurationEnv)
			})

			It("should return error", func() {
				os.Setenv(syncRetryDurationEnv, "invalid")

				duration, err := getSyncRetryDuration()
				Expect(err).To(HaveOccurred())
				Expect(duration).To(BeNil())
			})
		})

		Context("with different duration formats", func() {
			AfterEach(func() {
				os.Unsetenv(syncRetryDurationEnv)
			})

			It("should parse minutes correctly", func() {
				os.Setenv(syncRetryDurationEnv, "2m")

				duration, err := getSyncRetryDuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).NotTo(BeNil())
				Expect(*duration).To(Equal(2 * time.Minute))
			})

			It("should parse milliseconds correctly", func() {
				os.Setenv(syncRetryDurationEnv, "500ms")

				duration, err := getSyncRetryDuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).NotTo(BeNil())
				Expect(*duration).To(Equal(500 * time.Millisecond))
			})

			It("should parse complex duration correctly", func() {
				os.Setenv(syncRetryDurationEnv, "1h30m45s")

				duration, err := getSyncRetryDuration()
				Expect(err).NotTo(HaveOccurred())
				Expect(duration).NotTo(BeNil())
				expected := 1*time.Hour + 30*time.Minute + 45*time.Second
				Expect(*duration).To(Equal(expected))
			})
		})
	})

	Describe("syncRuntimeSpec", func() {
		const masterSts, workerSts = "test-runtime-master", "test-runtime-worker"

		// templateResources mirrors the value the creation path derives from the
		// CacheRuntimeClass template, i.e. what a sync must leave untouched.
		templateResources := corev1.ResourceRequirements{
			Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
		}

		// seedTemplateResources reproduces the post-creation state: the template declares
		// resources and the already-created workload carries them.
		seedTemplateResources := func(stsName string, template *corev1.PodTemplateSpec) {
			template.Spec.Containers[0].Resources = *templateResources.DeepCopy()

			sts := &workloadv1alpha1.AdvancedStatefulSet{}
			key := types.NamespacedName{Name: stsName, Namespace: "default"}
			Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
			sts.Spec.Template.Spec.Containers[0].Resources = *templateResources.DeepCopy()
			Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())
		}

		memLimitOf := func(stsName string) string {
			sts := &workloadv1alpha1.AdvancedStatefulSet{}
			key := types.NamespacedName{Name: stsName, Namespace: "default"}
			Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
			limit := sts.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]
			return limit.String()
		}

		// tmpfsSizeOf reports the size limit of the workload's memory-backed tiered
		// store volume, i.e. the quota the container's memory has to cover.
		tmpfsSizeOf := func(stsName string) string {
			sts := &workloadv1alpha1.AdvancedStatefulSet{}
			key := types.NamespacedName{Name: stsName, Namespace: "default"}
			Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
			for _, volume := range sts.Spec.Template.Spec.Volumes {
				emptyDir := volume.EmptyDir
				if emptyDir != nil && emptyDir.Medium == corev1.StorageMediumMemory && emptyDir.SizeLimit != nil {
					return emptyDir.SizeLimit.String()
				}
			}
			return "<none>"
		}

		BeforeEach(func() {
			seedTemplateResources(masterSts, &runtimeClass.Topology.Master.Template)
			seedTemplateResources(workerSts, &runtimeClass.Topology.Worker.Template)
		})

		Context("when the CacheRuntime drops a runtimeVersion it had set", func() {
			imageOfWorker := func() string {
				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				return sts.Spec.Template.Spec.Containers[0].Image
			}

			syncWorkerVersion := func(version datav1alpha1.VersionSpec) {
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.RuntimeVersion = version
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
			}

			It("should roll the worker back to the template image", func() {
				syncWorkerVersion(datav1alpha1.VersionSpec{ImageTag: "v2"})
				Expect(imageOfWorker()).To(Equal("test-worker:v2"))

				syncWorkerVersion(datav1alpha1.VersionSpec{})
				Expect(imageOfWorker()).To(Equal("test-worker:latest"))
			})

			It("should keep the image and emit a Warning event when the template cannot complete the version", func() {
				recorder := record.NewFakeRecorder(10)
				engine.Recorder = recorder
				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Image =
					"test-worker@sha256:067614b70d25b496e3edc3480747d558ee8a364ef47a67f669f5d96ca5098552"
				before := imageOfWorker()

				syncWorkerVersion(datav1alpha1.VersionSpec{ImageTag: "v2"})

				Expect(imageOfWorker()).To(Equal(before))
				Expect(recorder.Events).To(Receive(And(
					HavePrefix(corev1.EventTypeWarning+" "+common.RuntimeVersionIncomplete+" "),
					ContainSubstring(`imageTag "v2"`))))
				Expect(recorder.Events).To(BeEmpty())
			})

			It("should keep the template tag when only the image is set", func() {
				syncWorkerVersion(datav1alpha1.VersionSpec{Image: "other-worker"})
				Expect(imageOfWorker()).To(Equal("other-worker:latest"))

				syncWorkerVersion(datav1alpha1.VersionSpec{})
				Expect(imageOfWorker()).To(Equal("test-worker:latest"))
			})
		})

		// The client component runs on a DaemonSet. Neither reconcile path carries a
		// client spec edit to that workload: syncRuntimeSpec skips the client outright,
		// and reconcileDaemonSet returns early once the DaemonSet exists. The edit is
		// dropped with no error, no event and no condition.
		Context("when the CacheRuntime edits the client component", func() {
			const clientDs = "test-runtime-client"

			imageOfSts := func(stsName string) string {
				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: stsName, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				return sts.Spec.Template.Spec.Containers[0].Image
			}

			clientContainer := func() corev1.Container {
				ds := &appsv1.DaemonSet{}
				key := types.NamespacedName{Name: clientDs, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, ds)).To(Succeed())
				return ds.Spec.Template.Spec.Containers[0]
			}

			BeforeEach(func() {
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				// The same edit on the master and on the client. The master is the
				// control: its outcome proves the edit and the harness are well formed.
				edited.Spec.Master.RuntimeVersion = datav1alpha1.VersionSpec{Image: "test-master", ImageTag: "v2"}
				edited.Spec.Client.RuntimeVersion = datav1alpha1.VersionSpec{Image: "test-client", ImageTag: "v2"}
				edited.Spec.Client.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())
			})

			It("should apply the edit to the master and silently drop it for the client", func() {
				Expect(imageOfSts(masterSts)).To(Equal("test-master:latest"))
				Expect(clientContainer().Image).To(Equal("test-client:latest"))

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				// Control: the master's workload picks the new image up.
				Expect(imageOfSts(masterSts)).To(Equal("test-master:v2"))

				// The client's DaemonSet keeps the image and the resources it was created
				// with, and syncRuntimeSpec reported success all the same.
				Expect(clientContainer().Image).To(Equal("test-client:latest"))
				Expect(clientContainer().Resources.Limits).To(BeEmpty())
			})

			It("should leave the DaemonSet untouched even on the creation path", func() {
				// The transform derives the right desired state, so the edit is lost
				// downstream of it, inside the workload manager.
				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				value, err := engine.transform(dataset, syncRuntime, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				Expect(value.Client.PodTemplateSpec.Spec.Containers[0].Image).To(Equal("test-client:v2"))

				// SetupClientInternal is called directly, bypassing the ShouldSetupClient
				// gate that skips it once the client has left RuntimePhaseNone, so this
				// isolates reconcileDaemonSet itself.
				Expect(engine.SetupClientInternal(value.Client)).To(Succeed())
				Expect(clientContainer().Image).To(Equal("test-client:latest"))
			})

			It("should gate the creation path off once the client has been set up", func() {
				setup, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				setup.Status.Client.Phase = datav1alpha1.RuntimePhaseReady
				Expect(fakeClient.Status().Update(ctx.Context, setup)).To(Succeed())

				should, err := engine.ShouldSetupClient()
				Expect(err).NotTo(HaveOccurred())
				Expect(should).To(BeFalse())
			})
		})
		Context("when the CacheRuntime does not specify resources", func() {
			It("should leave the template's resources untouched", func() {
				Expect(runtimeObj.Spec.Master.Resources.Limits).To(BeNil())
				Expect(runtimeObj.Spec.Master.Resources.Requests).To(BeNil())
				Expect(runtimeObj.Spec.Worker.Resources.Limits).To(BeNil())
				Expect(runtimeObj.Spec.Worker.Resources.Requests).To(BeNil())

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(Succeed())

				Expect(memLimitOf(masterSts)).To(Equal("2Gi"))
				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})
		})

		Context("when the CacheRuntime specifies resources", func() {
			It("should apply them to the master workload only", func() {
				runtimeObj.Spec.Master.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(Succeed())

				Expect(memLimitOf(masterSts)).To(Equal("4Gi"))
				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})

			It("should apply them to the worker workload only", func() {
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(Succeed())

				Expect(memLimitOf(workerSts)).To(Equal("4Gi"))
				Expect(memLimitOf(masterSts)).To(Equal("2Gi"))
			})
		})

		Context("when the workload no longer matches the template", func() {
			// setWorkloadMemLimit edits the workload behind the runtime's back, standing in
			// for a workload that drifted from the template for any reason.
			setWorkloadMemLimit := func(stsName, limit string) {
				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: stsName, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				sts.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(limit)},
				}
				Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())
			}

			It("should restore the template value when the CacheRuntime specifies none", func() {
				setWorkloadMemLimit(workerSts, "8Gi")
				Expect(runtimeObj.Spec.Worker.Resources.Limits).To(BeNil())
				Expect(runtimeObj.Spec.Worker.Resources.Requests).To(BeNil())

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(Succeed())

				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})

			It("should still let the CacheRuntime win over the template", func() {
				setWorkloadMemLimit(workerSts, "8Gi")
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(Succeed())

				Expect(memLimitOf(workerSts)).To(Equal("4Gi"))
			})
		})

		Context("when neither the CacheRuntime nor the template specifies resources", func() {
			It("should leave the workload's resources untouched", func() {
				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{}

				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				sts.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				}
				Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(Succeed())

				Expect(memLimitOf(workerSts)).To(Equal("8Gi"))
			})
		})

		Context("when the tiered store baseline comes from the template", func() {
			// With the CacheRuntime declaring no resources, the baseline the quota is
			// charged on top of is the CacheRuntimeClass template value. The quota must
			// still be added, otherwise the worker is short by the quota it was created
			// with the moment a sync runs.
			BeforeEach(func() {
				runtimeObj.Spec.Worker.TieredStore = datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				}
				Expect(fakeClient.Update(ctx.Context, runtimeObj)).To(Succeed())
			})

			It("should charge the quota on top of the template value", func() {
				Expect(runtimeObj.Spec.Worker.Resources.Limits).To(BeNil())
				Expect(runtimeObj.Spec.Worker.Resources.Requests).To(BeNil())

				// The workload carries the template value plus the quota, as creation left it.
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				sts.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("10Gi")},
				}
				chargedQuota := resource.MustParse("8Gi")
				sts.Spec.Template.Spec.Volumes = []corev1.Volume{{
					Name: "tiered-store-level-0-memory",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: &chargedQuota,
						},
					},
				}}
				Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				// template 2Gi + quota 8Gi, not the bare template value.
				Expect(memLimitOf(workerSts)).To(Equal("10Gi"))
			})
		})

		Context("when the CacheRuntime declares a memory tiered store", func() {
			// The creation path derives the worker's memory as
			// <user baseline> + <tiered store memory quota>. A later sync recomputes
			// the desired state from the same spec and must land on the same value,
			// otherwise the quota silently disappears on the second reconcile.
			BeforeEach(func() {
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}
				runtimeObj.Spec.Worker.TieredStore = datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				}
				Expect(fakeClient.Update(ctx.Context, runtimeObj)).To(Succeed())
			})

			It("should keep the tiered store quota when syncing after creation", func() {
				// Creation reconcile: derive the desired state and create the workload.
				createRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())

				// Read the baseline before transforming: the transform must not be
				// trusted to leave the runtime spec alone.
				baseline := createRuntime.Spec.Worker.Resources.Limits[corev1.ResourceMemory].DeepCopy()

				value, err := engine.transform(dataset, createRuntime, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				desired := value.Worker.PodTemplateSpec.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]

				// Guard against a vacuous comparison: creation must actually charge the quota.
				Expect(desired.Cmp(baseline)).To(Equal(1), "creation path must charge the tiered store quota")

				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				seeded := &workloadv1alpha1.AdvancedStatefulSet{}
				Expect(fakeClient.Get(ctx.Context, key, seeded)).To(Succeed())
				Expect(fakeClient.Delete(ctx.Context, seeded)).To(Succeed())

				_, err = engine.SetupWorkerComponent(value.Worker)
				Expect(err).NotTo(HaveOccurred())
				Expect(memLimitOf(workerSts)).To(Equal(desired.String()))

				// Sync reconcile: re-read the runtime the way a fresh reconcile would,
				// so the sync can only rely on what is persisted in the spec.
				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				Expect(memLimitOf(workerSts)).To(Equal(desired.String()))
			})

			// createWorker runs the creation reconcile and returns the memory the
			// creation path derives, i.e. baseline + quota.
			createWorker := func() string {
				createRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				value, err := engine.transform(dataset, createRuntime, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
				desired := value.Worker.PodTemplateSpec.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]

				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				seeded := &workloadv1alpha1.AdvancedStatefulSet{}
				Expect(fakeClient.Get(ctx.Context, key, seeded)).To(Succeed())
				Expect(fakeClient.Delete(ctx.Context, seeded)).To(Succeed())
				_, err = engine.SetupWorkerComponent(value.Worker)
				Expect(err).NotTo(HaveOccurred())
				return desired.String()
			}

			syncOnce := func() {
				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
			}

			It("should not let an edited tiered store quota half-apply", func() {
				// tieredStore is not a supported update field, and SyncComponentSpec only
				// patches resources. A spec-derived quota would move the container's memory
				// to baseline+16Gi while the tmpfs volume stayed at the 8Gi it was created
				// with.
				desired := createWorker()
				Expect(tmpfsSizeOf(workerSts)).To(Equal("8Gi"))

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("16Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncOnce()

				Expect(memLimitOf(workerSts)).To(Equal(desired))
				Expect(tmpfsSizeOf(workerSts)).To(Equal("8Gi"))
			})

			It("should recharge a quota that an earlier release stripped", func() {
				// Workloads created before this fix have the bare baseline on the container
				// and the full quota on the tmpfs volume. Recovering the quota from the
				// volume rather than the container is what lets them be repaired in place.
				desired := createWorker()

				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				stripped := &workloadv1alpha1.AdvancedStatefulSet{}
				Expect(fakeClient.Get(ctx.Context, key, stripped)).To(Succeed())
				stripped.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}
				Expect(fakeClient.Update(ctx.Context, stripped)).To(Succeed())
				Expect(memLimitOf(workerSts)).To(Equal("4Gi"))

				syncOnce()

				Expect(memLimitOf(workerSts)).To(Equal(desired))
			})
		})

		Context("tiered store drift detection and warning", func() {
			var fakeRecorder *record.FakeRecorder

			createWorkerWithTieredStore := func(tieredStore datav1alpha1.RuntimeTieredStore) {
				runtimeObj.Spec.Worker.TieredStore = tieredStore
				Expect(fakeClient.Update(ctx.Context, runtimeObj)).To(Succeed())

				rt, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				value, err := engine.transform(dataset, rt, runtimeClass)
				Expect(err).NotTo(HaveOccurred())

				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				seeded := &workloadv1alpha1.AdvancedStatefulSet{}
				Expect(fakeClient.Get(ctx.Context, key, seeded)).To(Succeed())
				Expect(fakeClient.Delete(ctx.Context, seeded)).To(Succeed())
				_, err = engine.SetupWorkerComponent(value.Worker)
				Expect(err).NotTo(HaveOccurred())
			}

			BeforeEach(func() {
				fakeRecorder = record.NewFakeRecorder(100)
				engine.Recorder = fakeRecorder
				engine.timeOfLastTieredStoreWarning = time.Time{}
			})

			It("should not emit a warning event when tiered store matches worker pod volumes", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
						{EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("50Gi"), Medium: corev1.StorageMediumMemory}},
					},
				})

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				Expect(fakeRecorder.Events).To(BeEmpty())
			})

			It("should emit a Warning event when processMemory quota in spec drifts from worker volume sizeLimit", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				})

				// Edit spec quota to 16Gi
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("16Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring("level 0 (processMemory): spec quota 16Gi does not match worker volume tiered-store-level-0-memory sizeLimit 8Gi"))
				Expect(event).To(ContainSubstring("Worker tieredStore cannot be updated in-place after creation"))
			})

			It("should emit a Warning event when emptyDir quota in spec drifts from worker volume sizeLimit", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("50Gi")}},
					},
				})

				// Edit spec quota to 100Gi
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].EmptyDir.Quota = resource.MustParse("100Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring("level 0 (emptyDir): spec quota 100Gi does not match worker volume tiered-store-level-0-index-0 sizeLimit 50Gi"))
			})

			It("should emit a Warning event when hostPath paths are edited", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{HostPath: &datav1alpha1.HostPathMediumSource{
							Paths:  []string{"/mnt/disk1"},
							Quotas: []resource.Quantity{resource.MustParse("100Gi")},
						}},
					},
				})

				// Edit host path to /mnt/disk2
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].HostPath.Paths[0] = "/mnt/disk2"
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring("level 0 path 0 (hostPath): spec path /mnt/disk2 does not match worker volume tiered-store-level-0-index-0 path /mnt/disk1"))
			})

			It("should emit a Warning event when a new level is added to spec that worker pod lacks", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				})

				// Add a second level to spec
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels = append(edited.Spec.Worker.TieredStore.Levels,
					datav1alpha1.RuntimeTieredStoreLevel{
						EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("50Gi")},
					})
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring("level 1 (emptyDir): volume tiered-store-level-1-index-0 not found or missing sizeLimit in worker pod"))
			})

			It("should emit a Warning event when a tieredStore volume exists on pod but was removed from spec", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				})

				// Remove tieredStore from spec
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels = nil
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring("worker pod volume tiered-store-level-0-memory is not declared in spec tieredStore"))
			})

			It("should throttle warning events and re-emit only after interval expires", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				})

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("16Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				// First sync: event emitted
				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
				Expect(fakeRecorder.Events).To(Receive())

				// Immediate second sync: throttled, no event
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
				Expect(fakeRecorder.Events).To(BeEmpty())

				// Advance timestamp past interval: event re-emitted
				engine.timeOfLastTieredStoreWarning = time.Now().Add(-11 * time.Minute)
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
				Expect(fakeRecorder.Events).To(Receive())
			})

			It("should reset throttling cooldown when drift is resolved", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				})

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("16Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				// First sync emits event
				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
				Expect(fakeRecorder.Events).To(Receive())
				Expect(engine.timeOfLastTieredStoreWarning).NotTo(BeZero())

				// User reverts spec back to 8Gi to resolve drift
				edited, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("8Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
				Expect(fakeRecorder.Events).To(BeEmpty())
				Expect(engine.timeOfLastTieredStoreWarning).To(BeZero())

				// Re-introducing drift now fires immediately without waiting for cooldown
				edited, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("32Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())
				Expect(fakeRecorder.Events).To(Receive())
			})

			It("should detect drift even when neither CacheRuntime nor template declares worker resources", func() {
				// Clear worker resources from both runtimeClass template and runtimeObj
				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{}
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{}

				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				})

				// Edit spec quota to 16Gi
				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("16Gi")
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring("level 0 (processMemory): spec quota 16Gi does not match worker volume tiered-store-level-0-memory sizeLimit 8Gi"))
			})

			It("should return gracefully when podSpec is nil", func() {
				engine.checkWorkerTieredStoreDrift(runtimeObj, nil)
				Expect(fakeRecorder.Events).To(BeEmpty())
			})

			It("should detect missing volume for processMemory level", func() {
				runtimeWithPM := runtimeObj.DeepCopy()
				runtimeWithPM.Spec.Worker.TieredStore = datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("8Gi")}},
					},
				}
				emptyPodSpec := &corev1.PodSpec{}
				engine.checkWorkerTieredStoreDrift(runtimeWithPM, emptyPodSpec)

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring("level 0 (processMemory): volume tiered-store-level-0-memory not found or missing sizeLimit in worker pod"))
			})

			It("should detect missing volume for hostPath level", func() {
				runtimeWithHP := runtimeObj.DeepCopy()
				runtimeWithHP.Spec.Worker.TieredStore = datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{HostPath: &datav1alpha1.HostPathMediumSource{
							Paths:  []string{"/mnt/disk1"},
							Quotas: []resource.Quantity{resource.MustParse("100Gi")},
						}},
					},
				}
				emptyPodSpec := &corev1.PodSpec{}
				engine.checkWorkerTieredStoreDrift(runtimeWithHP, emptyPodSpec)

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring("level 0 path 0 (hostPath): volume tiered-store-level-0-index-0 not found in worker pod"))
			})

			It("should emit a Warning event when emptyDir medium is edited from disk to Memory", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{EmptyDir: &datav1alpha1.EmptyDirMediumSource{
							Quota:  resource.MustParse("50Gi"),
							Medium: "",
						}},
					},
				})

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].EmptyDir.Medium = corev1.StorageMediumMemory
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring(`level 0 (emptyDir): spec medium "Memory" does not match worker volume tiered-store-level-0-index-0 medium ""`))
				Expect(event).To(ContainSubstring("Delete and recreate the CacheRuntime to apply changes."))
			})

			It("should emit a Warning event when emptyDir medium is edited from Memory to disk", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{EmptyDir: &datav1alpha1.EmptyDirMediumSource{
							Quota:  resource.MustParse("50Gi"),
							Medium: corev1.StorageMediumMemory,
						}},
					},
				})

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].EmptyDir.Medium = ""
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring(`level 0 (emptyDir): spec medium "" does not match worker volume tiered-store-level-0-index-0 medium "Memory"`))
				Expect(event).To(ContainSubstring("Delete and recreate the CacheRuntime to apply changes."))
			})

			It("should give precedence to hostPath over emptyDir when both are configured on a level matching ConfigMap extraction", func() {
				createWorkerWithTieredStore(datav1alpha1.RuntimeTieredStore{
					Levels: []datav1alpha1.RuntimeTieredStoreLevel{
						{EmptyDir: &datav1alpha1.EmptyDirMediumSource{
							Quota:  resource.MustParse("2Gi"),
							Medium: corev1.StorageMediumMemory,
						}},
					},
				})

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].HostPath = &datav1alpha1.HostPathMediumSource{
					Paths:  []string{"/mnt/cache"},
					Quotas: []resource.Quantity{resource.MustParse("2Gi")},
				}
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				syncRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(Succeed())

				var event string
				Expect(fakeRecorder.Events).To(Receive(&event))
				Expect(event).To(ContainSubstring(corev1.EventTypeWarning))
				Expect(event).To(ContainSubstring(common.RuntimeTieredStoreConflict))
				Expect(event).To(ContainSubstring("hostPath"))
				Expect(event).To(ContainSubstring("volume tiered-store-level-0-index-0 not found in worker pod"))
			})

			It("should return error when reading worker pod spec fails", func() {
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				worker := &workloadv1alpha1.AdvancedStatefulSet{}
				Expect(fakeClient.Get(ctx.Context, key, worker)).To(Succeed())
				Expect(fakeClient.Delete(ctx.Context, worker)).To(Succeed())

				err := engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})
	})
})
