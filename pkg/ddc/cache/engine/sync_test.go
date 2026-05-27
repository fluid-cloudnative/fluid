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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CacheEngine Sync Tests", Label("pkg.ddc.cache.engine.sync_test.go"), func() {
	var (
		engine       *CacheEngine
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		dataset      *datav1alpha1.Dataset
		ctx          cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		// Add apps/v1 for StatefulSet and DaemonSet
		_ = appsv1.AddToScheme(scheme)

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
					WorkloadType: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
						},
					},
				},
				Worker: &datav1alpha1.RuntimeComponentDefinition{
					WorkloadType: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
						},
					},
				},
				Client: &datav1alpha1.RuntimeComponentDefinition{
					WorkloadType: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: "apps/v1"},
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
						},
					},
				},
			},
		}

		// Create master StatefulSet
		masterSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime-master",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: func() *int32 { i := int32(1); return &i }(),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "master", Image: "test-master:latest"}},
					},
				},
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 1,
			},
		}

		// Create worker StatefulSet
		workerSts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-runtime-worker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: func() *int32 { i := int32(2); return &i }(),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "worker", Image: "test-worker:latest"}},
					},
				},
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 2,
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

		fakeClient := fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(dataset, runtimeObj, runtimeClass, masterSts, workerSts, clientDs).
			WithStatusSubresource(runtimeObj).
			Build()

		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
		}

		ctx = cruntime.ReconcileRequestContext{
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
				fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
				engine.Client = fakeClient
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
				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(runtimeObj).
					Build()
				engine.Client = fakeClient
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
				// Create StatefulSets and DaemonSet for status update
				masterSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(1); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 1},
				}
				workerSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(2); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 2},
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
				// Create StatefulSets and DaemonSet for status update
				masterSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(1); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 1},
				}
				workerSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(2); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 2},
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
				// Create StatefulSets and DaemonSet for status update
				masterSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(1); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 1},
				}
				workerSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(2); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 2},
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
				// Create necessary resources
				masterSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-master", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(1); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 1},
				}
				workerSts := &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-worker", Namespace: "default"},
					Spec:       appsv1.StatefulSetSpec{Replicas: func() *int32 { i := int32(2); return &i }()},
					Status:     appsv1.StatefulSetStatus{ReadyReplicas: 2},
				}
				clientDs := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: "default"},
					Status:     appsv1.DaemonSetStatus{NumberReady: 0, DesiredNumberScheduled: 0},
				}
				fakeClient := fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(dataset, minimalRuntime, runtimeClass, masterSts, workerSts, clientDs).
					WithStatusSubresource(minimalRuntime).
					Build()
				engine.Client = fakeClient

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
})
