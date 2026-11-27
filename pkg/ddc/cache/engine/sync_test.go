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
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	cclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("CacheEngine Sync Tests", Label("pkg.ddc.cache.engine.sync_test.go"), func() {
	var (
		engine       *CacheEngine
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		dataset      *datav1alpha1.Dataset
		ctx          cruntime.ReconcileRequestContext
		fakeClient   cclient.WithWatch
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

	Describe("Setup creation conflict recovery during concurrent updates", func() {
		var writer *conflictOnceStatusWriter
		var concurrent *datav1alpha1.CacheRuntime

		BeforeEach(func() {
			runtimeObj.Spec.Client.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
			}
			Expect(fakeClient.Update(ctx.Context, runtimeObj)).To(Succeed())
			ctx.Dataset, ctx.Runtime = dataset, runtimeObj
			_, err := engine.Setup(ctx)
			Expect(err).To(MatchError(ContainSubstring("cpu request 1 exceeds limit 500m")))

			latest, err := engine.getRuntime()
			Expect(err).NotTo(HaveOccurred())
			latest.Spec.Client.Resources.Limits = nil
			Expect(fakeClient.Update(ctx.Context, latest)).To(Succeed())
			latest.Status.Master.Phase = datav1alpha1.RuntimePhaseReady
			latest.Status.Worker.Phase = datav1alpha1.RuntimePhaseReady
			latest.Status.Client.Phase = datav1alpha1.RuntimePhaseNotReady
			Expect(fakeClient.Status().Update(ctx.Context, latest)).To(Succeed())
			ctx.Runtime = latest
			concurrent = latest.DeepCopy()
			writer = &conflictOnceStatusWriter{StatusWriter: fakeClient.Status()}
			engine.Client = &conflictOnceClient{Client: fakeClient, statusWriter: writer}
		})

		It("preserves unrelated conditions and recovers in the same Setup call", func() {
			other := datav1alpha1.RuntimeCondition{Type: datav1alpha1.RuntimeWorkersInitialized, Status: corev1.ConditionTrue}
			writer.beforeConflict = func(ctx context.Context) error {
				concurrent.Status.Conditions = append(concurrent.Status.Conditions, other)
				return fakeClient.Status().Update(ctx, concurrent)
			}
			_, err := engine.Setup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.conflicted).To(BeTrue())
			latest, err := engine.getRuntime()
			Expect(err).NotTo(HaveOccurred())
			Expect(latest.Status.Conditions).To(ContainElement(other))
			_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
			Expect(cond).To(HaveField("Status", corev1.ConditionUnknown))
		})

		It("preserves a newer sync result when retrying the recovery write", func() {
			synced := datav1alpha1.RuntimeCondition{
				Type: datav1alpha1.RuntimeResourcesSynced, Status: corev1.ConditionFalse,
				Reason: common.RuntimeResourcesConflict, Message: "worker resources are not synced",
			}
			writer.beforeConflict = func(ctx context.Context) error {
				concurrent.Status.Conditions = utils.UpdateRuntimeCondition(concurrent.Status.Conditions, synced)
				return fakeClient.Status().Update(ctx, concurrent)
			}
			_, err := engine.Setup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.conflicted).To(BeTrue())
			latest, err := engine.getRuntime()
			Expect(err).NotTo(HaveOccurred())
			Expect(latest.Status.Conditions).To(ContainElement(synced))
		})

		It("waits for the new spec to be validated when the generation changes", func() {
			writer.beforeConflict = func(ctx context.Context) error {
				concurrent.Spec.Client.Resources.Limits = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100m")}
				concurrent.Generation++ // The fake client does not advance the generation for spec updates.
				return fakeClient.Update(ctx, concurrent)
			}
			ready, err := engine.Setup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(ready).To(BeFalse())
			Expect(writer.conflicted).To(BeTrue())
			latest, err := engine.getRuntime()
			Expect(err).NotTo(HaveOccurred())
			_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
			Expect(cond).To(HaveField("Status", corev1.ConditionFalse))
			key := types.NamespacedName{Namespace: engine.namespace, Name: common.GetCacheRuntimeConfigConfigMapName(engine.name)}
			Expect(apierrors.IsNotFound(fakeClient.Get(ctx.Context, key, &corev1.ConfigMap{}))).To(BeTrue())
		})

		It("requeues without a Setup error if every recovery write conflicts", func() {
			attempts := 0
			engine.Client = interceptor.NewClient(fakeClient, interceptor.Funcs{
				SubResourceUpdate: func(_ context.Context, _ cclient.Client, _ string, obj cclient.Object, _ ...cclient.SubResourceUpdateOption) error {
					attempts++
					return apierrors.NewConflict(datav1alpha1.GroupVersion.WithResource("cacheruntimes").GroupResource(),
						obj.GetName(), errors.New("stale resource version"))
				},
			})
			ready, err := engine.Setup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(ready).To(BeFalse())
			Expect(attempts).To(BeNumerically(">", 1))
		})
	})

	Describe("Sync", func() {
		Context("when runtime exists", func() {
			DescribeTable("clears the creation conflict before readiness and confirms resources on the first clean sync", func(legacy bool) {
				runtimeClass.Topology.Client.Template.Spec.Containers[0].Resources.Requests = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}
				Expect(fakeClient.Update(ctx.Context, runtimeClass)).To(Succeed())
				runtimeObj.Spec.Client.Resources.Limits = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}
				Expect(fakeClient.Update(ctx.Context, runtimeObj)).To(Succeed())
				ctx.Dataset, ctx.Runtime = dataset, runtimeObj
				ready, err := engine.Setup(ctx)
				Expect(ready).To(BeFalse())
				Expect(err).To(MatchError("invalid client resources: cpu request 1 exceeds limit 500m"))

				// The workloads already exist, but the master pod is still starting.
				master := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: "test-runtime-master", Namespace: engine.namespace}
				Expect(fakeClient.Get(ctx.Context, key, master)).To(Succeed())
				master.Status.ReadyReplicas = 0
				Expect(fakeClient.Update(ctx.Context, master)).To(Succeed())
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).NotTo(BeNil())
				if legacy {
					cond.Reason = common.RuntimeResourcesConflict
					cond.Message = "client cannot be created: the CacheRuntime and its CacheRuntimeClass template merge into invalid client resources: cpu request 1 exceeds limit 500m"
				} else {
					cond.Message = "Invalid client CPU request."
				}
				latest.Status.Master.Phase = datav1alpha1.RuntimePhaseNotReady
				latest.Status.Worker.Phase = datav1alpha1.RuntimePhaseReady
				latest.Status.Client.Phase = datav1alpha1.RuntimePhaseNotReady
				Expect(fakeClient.Status().Update(ctx.Context, latest)).To(Succeed())
				latest.Spec.Client.Resources.Limits = nil
				Expect(fakeClient.Update(ctx.Context, latest)).To(Succeed())
				ctx.Runtime = latest
				ready, err = engine.Setup(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeFalse())
				latest, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				_, cond = utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Status).To(Equal(corev1.ConditionUnknown))
				Expect(cond.Reason).To(Equal("WaitingForSync"))
				Expect(cond.Message).NotTo(ContainSubstring("cannot be created"))

				Expect(fakeClient.Get(ctx.Context, key, master)).To(Succeed())
				master.Status.ReadyReplicas = 1
				Expect(fakeClient.Update(ctx.Context, master)).To(Succeed())
				ctx.Runtime = latest
				ready, err = engine.Setup(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())
				Expect(engine.Sync(ctx)).To(Succeed())
				latest, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				_, cond = utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			},
				Entry("with a creation reason and different message wording", false),
				Entry("with a creation conflict recorded by an older controller", true),
			)

			It("writes component status and ResourcesSynced in one update", func() {
				runtimeWrites := 0
				engine.Client = interceptor.NewClient(fakeClient, interceptor.Funcs{
					SubResourceUpdate: func(ctx context.Context, c cclient.Client, name string, obj cclient.Object, opts ...cclient.SubResourceUpdateOption) error {
						if _, ok := obj.(*datav1alpha1.CacheRuntime); ok && name == "status" {
							runtimeWrites++
						}
						return c.SubResource(name).Update(ctx, obj, opts...)
					},
				})

				Expect(engine.Sync(ctx)).To(Succeed())
				Expect(runtimeWrites).To(Equal(1))
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(latest.Status.Worker.Phase).To(Equal(datav1alpha1.RuntimePhaseReady))
				_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))

				Expect(engine.Sync(ctx)).To(Succeed())
				Expect(runtimeWrites).To(Equal(1))
			})

			It("preserves a concurrent condition when retrying the combined status update", func() {
				other := datav1alpha1.RuntimeCondition{Type: datav1alpha1.RuntimeWorkersInitialized, Status: corev1.ConditionTrue}
				writer := &conflictOnceStatusWriter{StatusWriter: fakeClient.Status(), beforeConflict: func(ctx context.Context) error {
					latest, err := engine.getRuntime()
					if err != nil {
						return err
					}
					latest.Status.Conditions = append(latest.Status.Conditions, other)
					return fakeClient.Status().Update(ctx, latest)
				}}
				engine.Client = &conflictOnceClient{Client: fakeClient, statusWriter: writer}
				Expect(engine.Sync(ctx)).To(Succeed())
				Expect(writer.conflicted).To(BeTrue())
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(latest.Status.Conditions).To(ContainElements(other, HaveField("Type", datav1alpha1.RuntimeResourcesSynced)))
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
					WithStatusSubresource(dataset, runtimeObj).
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

		Context("when runtime is ready but dataset was left Failed by a previous outage", func() {
			BeforeEach(func() {
				dataset.Status.Phase = datav1alpha1.FailedDatasetPhase
				dataset.Status.Conditions = []datav1alpha1.DatasetCondition{
					{
						Type:   datav1alpha1.DatasetReady,
						Status: corev1.ConditionFalse,
					},
				}

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
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{Name: "client", Image: "test-client:latest"}},
							},
						},
					},
					Status: appsv1.DaemonSetStatus{NumberReady: 0, DesiredNumberScheduled: 0},
				}

				engine.Client = fake.NewClientBuilder().
					WithScheme(CacheEngineTestScheme).
					WithObjects(dataset, runtimeObj, runtimeClass, masterSts, workerSts, clientDs).
					WithStatusSubresource(dataset, runtimeObj).
					Build()
			})

			It("should restore the dataset phase to Bound", func() {
				err := engine.Sync(ctx)
				Expect(err).NotTo(HaveOccurred())

				updatedDataset := &datav1alpha1.Dataset{}
				err = engine.Client.Get(context.Background(), types.NamespacedName{
					Name:      "test-runtime",
					Namespace: "default",
				}, updatedDataset)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))

				idx, cond := utils.GetDatasetCondition(updatedDataset.Status.Conditions, datav1alpha1.DatasetReady)
				Expect(idx).NotTo(Equal(-1))
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			})

			Context("and the sync limiter is closed", func() {
				var patches *gomonkey.Patches
				var getCacheStatesCalled bool

				BeforeEach(func() {
					engine.syncRetryDuration = defaultSyncRetryDuration
					engine.timeOfLastSync = time.Now()

					getCacheStatesCalled = false
					// Patched at the GetCacheStates level, not NewCacheFileUtil: this Context has no
					// ReportSummary execution entries configured, so a real call would fail before ever
					// reaching the exec layer. The point here is only whether GetCacheStates is invoked at all.
					patches = gomonkey.ApplyMethod(reflect.TypeOf(engine), "GetCacheStates",
						func(_ *CacheEngine, _ *datav1alpha1.CacheRuntime, _ *datav1alpha1.CacheRuntimeClass) (common.CacheStateList, error) {
							getCacheStatesCalled = true
							return common.CacheStateList{}, nil
						})
				})

				AfterEach(func() {
					if patches != nil {
						patches.Reset()
					}
				})

				It("should still restore the dataset phase to Bound without fetching cache states", func() {
					err := engine.Sync(ctx)
					Expect(err).NotTo(HaveOccurred())

					updatedDataset := &datav1alpha1.Dataset{}
					err = engine.Client.Get(context.Background(), types.NamespacedName{
						Name:      "test-runtime",
						Namespace: "default",
					}, updatedDataset)
					Expect(err).NotTo(HaveOccurred())
					Expect(updatedDataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
					Expect(getCacheStatesCalled).To(BeFalse(), "GetCacheStates should be skipped while the sync limiter is closed")
				})
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
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj, runtimeClass)
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
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj, runtimeClass)
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
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj, runtimeClass)
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
				err := engine.syncRuntimeValueConfigMap(ctx, runtimeObj, runtimeClass)
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

				err := engine.syncRuntimeValueConfigMap(ctx, minimalRuntime, runtimeClass)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		It("should preserve creation-time worker hostPath and client tieredStore levels while updating mutable options", func() {
			localClass := runtimeClass.DeepCopy()
			localClass.Topology.Client = &datav1alpha1.RuntimeComponentDefinition{
				Options: map[string]string{"client-base": "1"},
			}
			localRuntime := runtimeObj.DeepCopy()
			localRuntime.Spec.Worker.TieredStore = datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
					HostPath: &datav1alpha1.HostPathMediumSource{
						Paths:  []string{"/mnt/nvme0"},
						Quotas: []resource.Quantity{resource.MustParse("50Gi")},
					},
					High: "0.9",
					Low:  "0.6",
				}},
			}
			localRuntime.Spec.Client.TieredStore = datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
					EmptyDir: &datav1alpha1.EmptyDirMediumSource{
						Quota: resource.MustParse("4Gi"),
					},
					High: "0.85",
					Low:  "0.55",
				}},
			}
			Expect(fakeClient.Update(ctx.Context, localRuntime)).To(Succeed())

			// Initial sync creates ConfigMap from creation-time spec.
			Expect(engine.syncRuntimeValueConfigMap(ctx, localRuntime, localClass)).To(Succeed())

			// Mutate both unsupported tieredStore fields and mutable runtime options.
			edited, err := engine.getRuntime()
			Expect(err).NotTo(HaveOccurred())
			edited.Spec.Options = map[string]string{"dynamic-opt": "new-val"}
			edited.Spec.Worker.TieredStore.Levels[0].HostPath.Quotas[0] = resource.MustParse("200Gi")
			edited.Spec.Worker.TieredStore.Levels[0].High = "0.99"
			edited.Spec.Worker.TieredStore.Levels[0].Low = "0.88"
			edited.Spec.Client.TieredStore.Levels[0].EmptyDir.Quota = resource.MustParse("12Gi")
			Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

			Expect(engine.syncRuntimeValueConfigMap(ctx, edited, localClass)).To(Succeed())

			cm := &corev1.ConfigMap{}
			Expect(engine.Client.Get(ctx.Context, types.NamespacedName{
				Name:      common.GetCacheRuntimeConfigConfigMapName(engine.name),
				Namespace: engine.namespace,
			}, cm)).To(Succeed())

			var persisted common.CacheRuntimeConfig
			Expect(json.Unmarshal([]byte(cm.Data[engine.getRuntimeConfigFileName()]), &persisted)).To(Succeed())
			Expect(persisted.Worker.Options["dynamic-opt"]).To(Equal("new-val"))
			Expect(persisted.Client.Options["dynamic-opt"]).To(Equal("new-val"))
			Expect(persisted.Worker.TieredStoreLevels).To(Equal([]common.TieredStoreLevelConfig{{
				MediumType: common.HDD,
				MountPaths: []string{GetHostPathTieredStoreMountPath(0, 0)},
				Quotas:     []string{"50Gi"},
				High:       "0.9",
				Low:        "0.6",
			}}))
			Expect(persisted.Client.TieredStoreLevels).To(Equal([]common.TieredStoreLevelConfig{{
				MediumType: common.HDD,
				MountPaths: []string{GetEmptyDirTieredStoreMountPath(0)},
				Quotas:     []string{"4Gi"},
				High:       "0.85",
				Low:        "0.55",
			}}))

			sh := cm.Data[engine.getRuntimeShFileName()]
			Expect(sh).To(ContainSubstring("export WORKER_TIEREDSTORE_LEVEL_0_QUOTA_0='50Gi'"))
			Expect(sh).To(ContainSubstring("export WORKER_TIEREDSTORE_LEVEL_0_HIGH='0.9'"))
			Expect(sh).To(ContainSubstring("export WORKER_TIEREDSTORE_LEVEL_0_LOW='0.6'"))
			Expect(sh).To(ContainSubstring("export CLIENT_TIEREDSTORE_LEVEL_0_QUOTA_0='4Gi'"))
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

	Describe("master and worker spec synchronization", func() {
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
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(BeEmpty())
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
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(BeEmpty())

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

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(memLimitOf(masterSts)).To(Equal("2Gi"))
				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})
		})

		Context("when the CacheRuntimeClass has no topology", func() {
			It("should sync nothing instead of panicking", func() {
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}
				runtimeClass.Topology = nil

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(memLimitOf(masterSts)).To(Equal("2Gi"))
				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})
		})

		Context("when the CacheRuntime specifies resources", func() {
			It("should apply them to the master workload only", func() {
				runtimeObj.Spec.Master.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(memLimitOf(masterSts)).To(Equal("4Gi"))
				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})

			It("should apply them to the worker workload only", func() {
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(memLimitOf(workerSts)).To(Equal("4Gi"))
				Expect(memLimitOf(masterSts)).To(Equal("2Gi"))
			})
		})

		// Issue #6173: a CacheRuntime that names one key used to replace the container's
		// whole ResourceRequirements, so raising the memory limit alone left the container
		// with no CPU request and no CPU limit at all.
		Context("when the CacheRuntime only names part of the template's resources", func() {
			fullTemplateResources := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("2Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			}

			workloadOf := func(stsName string) *workloadv1alpha1.AdvancedStatefulSet {
				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: stsName, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				return sts
			}

			// The CacheRuntime raises the memory limit and names nothing else.
			BeforeEach(func() {
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				}
			})

			Context("and the template declares a full set of requirements", func() {
				// Re-seed over the limits-only template the outer BeforeEach installs, so
				// the spec has requirements left to lose.
				BeforeEach(func() {
					runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources = *fullTemplateResources.DeepCopy()

					sts := workloadOf(workerSts)
					sts.Spec.Template.Spec.Containers[0].Resources = *fullTemplateResources.DeepCopy()
					Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())
				})

				It("should move the key it names and keep the template's other requirements", func() {
					Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

					synced := workloadOf(workerSts).Spec.Template.Spec.Containers[0].Resources
					Expect(synced.Limits.Memory().String()).To(Equal("8Gi"))
					Expect(synced.Limits.Cpu().String()).To(Equal("2"))
					Expect(synced.Requests.Cpu().String()).To(Equal("1"))
					Expect(synced.Requests.Memory().String()).To(Equal("2Gi"))
				})
			})

			// The master resolves its resources through the same desiredComponentResources
			// as the worker, without the tiered store quota on top.
			Context("and the component is the master", func() {
				BeforeEach(func() {
					runtimeObj.Spec.Master.Resources = corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
					}
					runtimeClass.Topology.Master.Template.Spec.Containers[0].Resources = *fullTemplateResources.DeepCopy()

					sts := workloadOf(masterSts)
					sts.Spec.Template.Spec.Containers[0].Resources = *fullTemplateResources.DeepCopy()
					Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())
				})

				It("should move the key it names and keep the template's other requirements", func() {
					Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

					synced := workloadOf(masterSts).Spec.Template.Spec.Containers[0].Resources
					Expect(synced.Limits.Memory().String()).To(Equal("8Gi"))
					Expect(synced.Limits.Cpu().String()).To(Equal("2"))
					Expect(synced.Requests.Cpu().String()).To(Equal("1"))
					Expect(synced.Requests.Memory().String()).To(Equal("2Gi"))
				})
			})

			Context("and neither side declares requests", func() {
				// The template the outer BeforeEach installs is limits-only, so the merge has
				// no requests to carry over. The sync must leave them unset and converge:
				// a second pass finds nothing to change and does not patch again.
				It("should leave requests unset and stop patching once converged", func() {
					Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

					synced := workloadOf(workerSts).Spec.Template.Spec.Containers[0].Resources
					Expect(synced.Limits.Memory().String()).To(Equal("8Gi"))
					Expect(synced.Requests).To(BeNil())

					converged := workloadOf(workerSts).ResourceVersion
					Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())
					Expect(workloadOf(workerSts).ResourceVersion).To(Equal(converged))
				})
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

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(memLimitOf(workerSts)).To(Equal("2Gi"))
			})

			It("should still let the CacheRuntime win over the template", func() {
				setWorkloadMemLimit(workerSts, "8Gi")
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(memLimitOf(workerSts)).To(Equal("4Gi"))
			})
		})

		Context("when neither the CacheRuntime nor the template specifies resources", func() {
			It("should clear the workload's resources", func() {
				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{}

				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				sts.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("8Gi")},
				}
				Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				Expect(sts.Spec.Template.Spec.Containers[0].Resources.Limits).To(BeEmpty())
			})
		})

		Context("when the merged resources put a request above its limit", func() {
			It("should leave the workload's resources alone and emit a Warning event", func() {
				recorder := record.NewFakeRecorder(10)
				engine.Recorder = recorder

				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
					Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
				}
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
				}

				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				before := sts.Spec.Template.Spec.Containers[0].Resources.DeepCopy()

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(ConsistOf(ContainSubstring("cpu request 1 exceeds limit 500m")))

				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				Expect(equality.Semantic.DeepEqual(sts.Spec.Template.Spec.Containers[0].Resources, *before)).To(BeTrue())
				Expect(recorder.Events).To(Receive(And(
					ContainSubstring(common.RuntimeResourcesConflict),
					ContainSubstring("cpu request 1 exceeds limit 500m"))))
			})
		})

		Context("when the master reports a resource conflict", func() {
			var recorder *record.FakeRecorder
			BeforeEach(func() {
				recorder = record.NewFakeRecorder(10)
				engine.Recorder = recorder
				runtimeObj.Spec.Master.Resources.Requests = corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")}
				Expect(fakeClient.Update(ctx.Context, runtimeClass)).To(Succeed())
			})

			updateRuntimeSpec := func() {
				GinkgoHelper()
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				latest.Spec = runtimeObj.Spec
				Expect(fakeClient.Update(ctx.Context, latest)).To(Succeed())
			}

			eventMessage := func(component string) string {
				GinkgoHelper()
				var event string
				Expect(recorder.Events).To(Receive(&event))
				prefix := corev1.EventTypeWarning + " " + common.RuntimeResourcesConflict + " "
				Expect(event).To(HavePrefix(prefix + component + " "))
				return strings.TrimPrefix(event, prefix)
			}

			expectCondition := func(message string) datav1alpha1.RuntimeCondition {
				GinkgoHelper()
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(latest.Status.Conditions).To(HaveLen(1))
				cond := latest.Status.Conditions[0]
				Expect(cond.Type).To(Equal(datav1alpha1.RuntimeResourcesSynced))
				Expect(cond.Status).To(Equal(corev1.ConditionFalse))
				Expect(cond.Reason).To(Equal(common.RuntimeResourcesConflict))
				Expect(cond.Message).To(Equal(message))
				return cond
			}

			It("records both event messages in one condition when the worker also has an undeclared claim", func() {
				runtimeObj.Spec.Worker.Resources.Claims = []corev1.ResourceClaim{{Name: "gpu"}}
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(Succeed())
				expectCondition(eventMessage("master") + "\n" + eventMessage("worker"))
			})

			DescribeTable("reports a creation conflict after exhausting status retries only for the current generation", func(specCorrected bool) {
				updateRuntimeSpec()
				observed, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				ctx.Dataset, ctx.Runtime = dataset, observed
				if specCorrected {
					corrected := observed.DeepCopy()
					corrected.Spec.Master.Resources.Requests = nil
					corrected.Generation++
					Expect(fakeClient.Update(ctx.Context, corrected)).To(Succeed())
				}

				attempts := 0
				engine.Client = interceptor.NewClient(fakeClient, interceptor.Funcs{
					Get: func(ctx context.Context, c cclient.WithWatch, key cclient.ObjectKey, obj cclient.Object, opts ...cclient.GetOption) error {
						if rt, ok := obj.(*datav1alpha1.CacheRuntime); ok {
							observed.DeepCopyInto(rt) // The informer still sees the original generation.
							return nil
						}
						return c.Get(ctx, key, obj, opts...)
					},
					SubResourceUpdate: func(_ context.Context, _ cclient.Client, _ string, obj cclient.Object, _ ...cclient.SubResourceUpdateOption) error {
						attempts++
						return apierrors.NewConflict(datav1alpha1.GroupVersion.WithResource("cacheruntimes").GroupResource(),
							obj.GetName(), errors.New("stale resource version"))
					},
					// The fake client does not implement GET on the status subresource.
					SubResourceGet: func(ctx context.Context, c cclient.Client, name string, obj, result cclient.Object, _ ...cclient.SubResourceGetOption) error {
						Expect(name).To(Equal("status"))
						return c.Get(ctx, cclient.ObjectKeyFromObject(obj), result)
					},
				})

				ready, err := engine.Setup(ctx)
				Expect(ready).To(BeFalse())
				Expect(attempts).To(BeNumerically(">", 1))
				message := eventMessage("master")
				if specCorrected {
					Expect(err).NotTo(HaveOccurred())
				} else {
					var conflict *componentResourcesConflictError
					Expect(errors.As(err, &conflict)).To(BeTrue())
					Expect(conflict.message).To(Equal(message))
				}
				latest := &datav1alpha1.CacheRuntime{}
				Expect(fakeClient.Get(ctx.Context, cclient.ObjectKeyFromObject(observed), latest)).To(Succeed())
				_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).To(BeNil())
			},
				Entry("when the configuration still conflicts", false),
				Entry("when the API server has the corrected configuration", true),
			)

			Context("when the resource condition cannot be persisted", func() {
				statusErr := errors.New("runtime status update denied")
				BeforeEach(func() {
					updateRuntimeSpec()
					engine.Client = interceptor.NewClient(fakeClient, interceptor.Funcs{
						SubResourceUpdate: func(ctx context.Context, c cclient.Client, name string, obj cclient.Object, opts ...cclient.SubResourceUpdateOption) error {
							if _, ok := obj.(*datav1alpha1.CacheRuntime); ok && name == "status" {
								return statusErr
							}
							return c.SubResource(name).Update(ctx, obj, opts...)
						},
					})
				})

				It("returns both the creation conflict and the status error from Setup", func() {
					ctx.Dataset, ctx.Runtime = dataset, runtimeObj
					ready, err := engine.Setup(ctx)
					Expect(ready).To(BeFalse())
					Expect(errors.Is(err, statusErr)).To(BeTrue())
					var conflict *componentResourcesConflictError
					Expect(errors.As(err, &conflict)).To(BeTrue())
					Expect(conflict.message).To(Equal(eventMessage("master")))
				})

				It("returns both the missing worker and the status error from Sync", func() {
					worker := &workloadv1alpha1.AdvancedStatefulSet{ObjectMeta: metav1.ObjectMeta{Name: workerSts, Namespace: engine.namespace}}
					Expect(fakeClient.Delete(ctx.Context, worker)).To(Succeed())
					err := engine.Sync(ctx)
					Expect(errors.Is(err, statusErr)).To(BeTrue())
					Expect(apierrors.IsNotFound(err)).To(BeTrue())
					Expect(err.Error()).To(ContainSubstring(workerSts))
				})
			})

			It("preserves the master sync conflict through Setup until the corrected override is synced", func() {
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(Succeed())
				message := eventMessage("master")
				expectCondition(message)

				runtimeObj.Spec.Master.Resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
				updateRuntimeSpec()
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				ctx.Dataset, ctx.Runtime = dataset, latest
				ready, err := engine.Setup(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(ready).To(BeTrue())
				expectCondition(message)

				Expect(engine.Sync(ctx)).To(Succeed())
				Expect(recorder.Events).To(BeEmpty())

				latest, err = engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))
				Expect(cond.Reason).To(Equal(datav1alpha1.RuntimeResourcesSyncedReason))

				Expect(engine.Sync(ctx)).To(Succeed())
				unchanged, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(unchanged.ResourceVersion).To(Equal(latest.ResourceVersion))
			})

			It("records the master event even if reading the worker fails", func() {
				worker := &workloadv1alpha1.AdvancedStatefulSet{ObjectMeta: metav1.ObjectMeta{Name: workerSts, Namespace: engine.namespace}}
				Expect(fakeClient.Delete(ctx.Context, worker)).To(Succeed())
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(MatchError(ContainSubstring(workerSts)))
				message := eventMessage("master")
				expectCondition(message)

				runtimeObj.Spec.Master.Resources.Requests = nil
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(MatchError(ContainSubstring(workerSts)))
				expectCondition(message)
			})

			It("reports only the observed master conflict when the worker cannot be read", func() {
				runtimeObj.Spec.Master.Resources.Requests = nil
				runtimeObj.Spec.Worker.Resources.Claims = []corev1.ResourceClaim{{Name: "gpu"}}
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(Succeed())
				expectCondition(eventMessage("worker"))
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				_, old := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				old.LastTransitionTime = metav1.Unix(1, 0)
				Expect(fakeClient.Status().Update(ctx.Context, latest)).To(Succeed())

				worker := &workloadv1alpha1.AdvancedStatefulSet{ObjectMeta: metav1.ObjectMeta{Name: workerSts, Namespace: engine.namespace}}
				Expect(fakeClient.Delete(ctx.Context, worker)).To(Succeed())
				runtimeObj.Spec.Master.Resources.Requests = corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")}
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(MatchError(ContainSubstring(workerSts)))
				current := expectCondition(eventMessage("master"))
				Expect(current.LastTransitionTime.Equal(&old.LastTransitionTime)).To(BeTrue())
				Expect(recorder.Events).To(BeEmpty())
			})

			It("records conflicts and their recovery when client readiness cannot be read", func() {
				ds := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: engine.namespace}}
				Expect(fakeClient.Delete(ctx.Context, ds)).To(Succeed())
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(MatchError(ContainSubstring(ds.Name)))
				expectCondition(eventMessage("master"))

				runtimeObj.Spec.Master.Resources.Requests = nil
				updateRuntimeSpec()
				Expect(engine.Sync(ctx)).To(MatchError(ContainSubstring(ds.Name)))
				latest, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				_, cond := utils.GetRuntimeCondition(latest.Status.Conditions, datav1alpha1.RuntimeResourcesSynced)
				Expect(cond).NotTo(BeNil())
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			})
		})

		Context("when the merged resources make an extended resource request differ from its limit", func() {
			It("should leave the workload's resources alone and emit a Warning event", func() {
				recorder := record.NewFakeRecorder(10)
				engine.Recorder = recorder

				const gpu corev1.ResourceName = "nvidia.com/gpu"
				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{gpu: resource.MustParse("1")},
					Limits:   corev1.ResourceList{gpu: resource.MustParse("1")},
				}
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{gpu: resource.MustParse("2")},
				}

				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				before := sts.Spec.Template.Spec.Containers[0].Resources.DeepCopy()

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(ConsistOf(ContainSubstring("nvidia.com/gpu request 1 must equal limit 2")))

				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				Expect(equality.Semantic.DeepEqual(sts.Spec.Template.Spec.Containers[0].Resources, *before)).To(BeTrue())
				Expect(recorder.Events).To(Receive(And(
					ContainSubstring(common.RuntimeResourcesConflict),
					ContainSubstring("nvidia.com/gpu request 1 must equal limit 2"))))
			})
		})

		Context("when the template adds a DRA claim after the workload was created", func() {
			// The class edit declares the claim at both levels, but the sync only patches
			// container resources: a container claim the workload's pod spec does not
			// declare would make the API server reject every replacement Pod.
			BeforeEach(func() {
				runtimeClass.Topology.Worker.Template.Spec.ResourceClaims = []corev1.PodResourceClaim{{Name: "gpu"}}
				runtimeClass.Topology.Worker.Template.Spec.Containers[0].Resources.Claims = []corev1.ResourceClaim{{Name: "gpu"}}
			})

			It("should sync the resources without the claim and emit a Warning event", func() {
				recorder := record.NewFakeRecorder(10)
				engine.Recorder = recorder
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
				}

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(ConsistOf(ContainSubstring("resource claims gpu")))

				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				Expect(memLimitOf(workerSts)).To(Equal("4Gi"))
				Expect(sts.Spec.Template.Spec.Containers[0].Resources.Claims).To(BeEmpty())
				Expect(recorder.Events).To(HaveLen(1))
				Expect(<-recorder.Events).To(And(
					HavePrefix(corev1.EventTypeWarning+" "+common.RuntimeResourcesConflict+" "),
					ContainSubstring("worker resource claims gpu are not synced")))
			})

			It("should sync the claim once the workload declares it", func() {
				recorder := record.NewFakeRecorder(10)
				engine.Recorder = recorder

				sts := &workloadv1alpha1.AdvancedStatefulSet{}
				key := types.NamespacedName{Name: workerSts, Namespace: "default"}
				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				sts.Spec.Template.Spec.ResourceClaims = []corev1.PodResourceClaim{{Name: "gpu"}}
				Expect(fakeClient.Update(ctx.Context, sts)).To(Succeed())

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(BeEmpty())

				Expect(fakeClient.Get(ctx.Context, key, sts)).To(Succeed())
				Expect(sts.Spec.Template.Spec.Containers[0].Resources.Claims).To(Equal([]corev1.ResourceClaim{{Name: "gpu"}}))
				Expect(recorder.Events).To(BeEmpty())
			})
		})

		// Without DynamicResourceAllocation the API server ignores claims, so a claim the
		// workload does not declare must not hold back the rest of a resources edit.
		Context("when the CacheRuntime names an undeclared claim along with a memory limit", func() {
			DescribeTable("should still sync the memory limit and drop only the claim",
				func(componentType common.ComponentType, stsName string, setResources func(corev1.ResourceRequirements)) {
					recorder := record.NewFakeRecorder(10)
					engine.Recorder = recorder
					setResources(corev1.ResourceRequirements{
						Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("4Gi")},
						Claims: []corev1.ResourceClaim{{Name: "unused"}},
					})

					Expect(engine.syncRuntimeSpec(ctx, runtimeObj, runtimeClass)).To(ConsistOf(ContainSubstring("resource claims unused")))

					sts := &workloadv1alpha1.AdvancedStatefulSet{}
					Expect(fakeClient.Get(ctx.Context, types.NamespacedName{Name: stsName, Namespace: "default"}, sts)).To(Succeed())
					Expect(memLimitOf(stsName)).To(Equal("4Gi"))
					Expect(sts.Spec.Template.Spec.Containers[0].Resources.Claims).To(BeEmpty())
					Expect(recorder.Events).To(HaveLen(1))
					Expect(<-recorder.Events).To(And(
						HavePrefix(corev1.EventTypeWarning+" "+common.RuntimeResourcesConflict+" "),
						ContainSubstring(string(componentType)+" resource claims unused are not synced")))
				},
				Entry("for the master", common.ComponentTypeMaster, masterSts, func(r corev1.ResourceRequirements) {
					runtimeObj.Spec.Master.Resources = r
				}),
				Entry("for the worker", common.ComponentTypeWorker, workerSts, func(r corev1.ResourceRequirements) {
					runtimeObj.Spec.Worker.Resources = r
				}),
			)
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
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(BeEmpty())

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
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(BeEmpty())

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
				Expect(engine.syncRuntimeSpec(ctx, syncRuntime, runtimeClass)).To(BeEmpty())
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

			It("should keep creation-time tieredStoreLevels in the runtime ConfigMap aligned with worker volumes when spec is edited", func() {
				desired := createWorker()
				Expect(tmpfsSizeOf(workerSts)).To(Equal("8Gi"))

				createRuntime, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				Expect(engine.syncRuntimeValueConfigMap(ctx, createRuntime, runtimeClass)).To(Succeed())

				edited, err := engine.getRuntime()
				Expect(err).NotTo(HaveOccurred())
				edited.Spec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("16Gi")
				edited.Spec.Options = map[string]string{"sync-opt": "kept"}
				Expect(fakeClient.Update(ctx.Context, edited)).To(Succeed())

				Expect(engine.syncRuntimeValueConfigMap(ctx, edited, runtimeClass)).To(Succeed())
				syncOnce()

				Expect(memLimitOf(workerSts)).To(Equal(desired))
				Expect(tmpfsSizeOf(workerSts)).To(Equal("8Gi"))

				cm := &corev1.ConfigMap{}
				Expect(fakeClient.Get(ctx.Context, types.NamespacedName{
					Name:      common.GetCacheRuntimeConfigConfigMapName(engine.name),
					Namespace: engine.namespace,
				}, cm)).To(Succeed())

				var persisted common.CacheRuntimeConfig
				Expect(json.Unmarshal([]byte(cm.Data[engine.getRuntimeConfigFileName()]), &persisted)).To(Succeed())
				Expect(persisted.Worker.Options["sync-opt"]).To(Equal("kept"))
				Expect(persisted.Worker.TieredStoreLevels).To(HaveLen(1))
				Expect(persisted.Worker.TieredStoreLevels[0].Quotas).To(Equal([]string{"8Gi"}))
				Expect(cm.Data[engine.getRuntimeShFileName()]).To(ContainSubstring("export WORKER_TIEREDSTORE_LEVEL_0_QUOTA_0='8Gi'"))
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
	})
})
