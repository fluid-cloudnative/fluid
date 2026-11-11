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
	"strings"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

var _ = Describe("CacheRuntimeClass snapshot", Label("pkg.ddc.cache.engine.runtime_class_snapshot_test.go"), func() {
	const (
		runtimeName = "snap"
		namespace   = "default"
		className   = "snap-class"
		runtimeUID  = "snap-uid"
	)

	var (
		engine       *CacheEngine
		fakeClient   client.Client
		recorder     *record.FakeRecorder
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		dataset      *datav1alpha1.Dataset
		objects      []client.Object
		ctx          cruntime.ReconcileRequestContext
	)

	quantity := resource.MustParse

	templateResources := func() corev1.ResourceRequirements {
		return corev1.ResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceCPU: quantity("1"), corev1.ResourceMemory: quantity("2Gi")},
			Limits:   corev1.ResourceList{corev1.ResourceCPU: quantity("2"), corev1.ResourceMemory: quantity("4Gi")},
		}
	}

	componentDefinition := func(containerName, image string) *datav1alpha1.RuntimeComponentDefinition {
		return &datav1alpha1.RuntimeComponentDefinition{
			Options: map[string]string{"block-size": "64MB"},
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
				Name:      containerName,
				Image:     image,
				Resources: templateResources(),
			}}}},
		}
	}

	// workload stands for an AdvancedStatefulSet the creation path rendered earlier.
	workload := func(componentType common.ComponentType, image string, resources corev1.ResourceRequirements,
		volumes ...corev1.Volume) *workloadv1alpha1.AdvancedStatefulSet {
		return &workloadv1alpha1.AdvancedStatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: common.GetCacheComponentName(runtimeName, componentType), Namespace: namespace},
			Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
				Replicas: ptr.To(int32(1)),
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: string(componentType), Image: image, Resources: resources}},
					Volumes:    volumes,
				}},
			},
		}
	}

	// tieredStoreVolume is the tmpfs volume TransformRuntimeTieredStore adds for a memory level,
	// whose size the worker's memory request and limit were charged.
	tieredStoreVolume := func(size string) corev1.Volume {
		sizeLimit := quantity(size)
		return corev1.Volume{
			Name: tieredStoreVolumeNamePrefix + "0",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium:    corev1.StorageMediumMemory,
				SizeLimit: &sizeLimit,
			}},
		}
	}

	getWorkload := func(componentType common.ComponentType) *workloadv1alpha1.AdvancedStatefulSet {
		asts := &workloadv1alpha1.AdvancedStatefulSet{}
		Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
			Name: common.GetCacheComponentName(runtimeName, componentType), Namespace: namespace}, asts)).To(Succeed())
		return asts
	}

	getSnapshot := func() *corev1.ConfigMap {
		snapshot := &corev1.ConfigMap{}
		Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
			Name: common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName), Namespace: namespace}, snapshot)).To(Succeed())
		return snapshot
	}

	editLiveClass := func(edit func(*datav1alpha1.CacheRuntimeClass)) {
		live := &datav1alpha1.CacheRuntimeClass{}
		Expect(fakeClient.Get(context.TODO(), types.NamespacedName{Name: className}, live)).To(Succeed())
		edit(live)
		Expect(fakeClient.Update(context.TODO(), live)).To(Succeed())
	}

	events := func() []string {
		var got []string
		for {
			select {
			case event := <-recorder.Events:
				got = append(got, event)
			default:
				return got
			}
		}
	}

	BeforeEach(func() {
		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName, Namespace: namespace, UID: "snap-dataset-uid"},
		}
		runtimeObj = &datav1alpha1.CacheRuntime{
			TypeMeta: metav1.TypeMeta{APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind},
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: namespace,
				UID:       runtimeUID,
			},
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: className,
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
				Worker:           datav1alpha1.CacheRuntimeWorkerSpec{Replicas: 1},
				Client: datav1alpha1.CacheRuntimeClientSpec{
					RuntimeComponentCommonSpec: datav1alpha1.RuntimeComponentCommonSpec{Disabled: true},
				},
			},
		}
		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			ObjectMeta:     metav1.ObjectMeta{Name: className},
			FileSystemType: "curvine",
			Topology: &datav1alpha1.RuntimeTopology{
				Master: componentDefinition("master", "curvine:v1"),
				Worker: componentDefinition("worker", "curvine:v1"),
			},
			DataOperationSpecs: []datav1alpha1.DataOperationSpec{{
				Name:    "DataLoad",
				Image:   "curvine-loader:v1",
				Command: []string{"load"},
			}},
		}
		runtimeClass.Topology.Master.ExecutionEntries = &datav1alpha1.ExecutionEntries{
			ReportSummary: &datav1alpha1.ExecutionCommonEntry{Command: []string{"summary-v1"}, TimeoutSeconds: 10},
		}
		objects = []client.Object{dataset, runtimeObj, runtimeClass}
		recorder = record.NewFakeRecorder(64)
	})

	JustBeforeEach(func() {
		fakeClient = fake.NewClientBuilder().
			WithScheme(CacheEngineTestScheme).
			WithObjects(objects...).
			WithStatusSubresource(dataset, runtimeObj).
			Build()
		engine = &CacheEngine{
			name:      runtimeName,
			namespace: namespace,
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
			Recorder:  recorder,
		}
		ctx = cruntime.ReconcileRequestContext{
			Context:        context.Background(),
			Client:         fakeClient,
			Log:            ctrl.Log.WithName("test"),
			RuntimeType:    "cache",
			NamespacedName: types.NamespacedName{Name: runtimeName, Namespace: namespace},
		}
	})

	Describe("binding a runtime that has no components yet", func() {
		It("should snapshot the class into an immutable configmap the runtime controls", func() {
			bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(bound.Name).To(Equal(className))
			Expect(bound.Topology).To(Equal(runtimeClass.Topology))
			Expect(bound.DataOperationSpecs).To(Equal(runtimeClass.DataOperationSpecs))
			Expect(bound.FileSystemType).To(Equal("curvine"))

			snapshot := getSnapshot()
			Expect(metav1.IsControlledBy(snapshot, runtimeObj)).To(BeTrue())
			Expect(snapshot.Immutable).To(Equal(ptr.To(true)))
			Expect(snapshot.Labels).To(HaveKeyWithValue(common.LabelCacheRuntimeName, runtimeName))
			Expect(snapshot.Annotations).To(HaveKeyWithValue(common.AnnotationCacheRuntimeClassName, className))
			Expect(snapshot.Annotations).NotTo(HaveKey(common.AnnotationCacheRuntimeAdoptionPending))
			Expect(events()).To(BeEmpty())
		})

		It("should keep reading the snapshot after the class is edited or deleted", func() {
			bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())

			editLiveClass(func(live *datav1alpha1.CacheRuntimeClass) {
				live.Topology.Worker.Template.Spec.Containers[0].Image = "curvine:v2"
				live.Topology.Worker.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory] = quantity("16Gi")
				live.Topology.Worker.Options["block-size"] = "128MB"
				live.Topology.Master.ExecutionEntries.ReportSummary.Command = []string{"summary-v2"}
				live.DataOperationSpecs[0].Image = "curvine-loader:v2"
			})
			again, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(again).To(Equal(bound))

			Expect(fakeClient.Delete(context.TODO(), &datav1alpha1.CacheRuntimeClass{ObjectMeta: metav1.ObjectMeta{Name: className}})).To(Succeed())
			afterDelete, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(afterDelete).To(Equal(bound))
		})

		It("should fail when the class does not exist", func() {
			runtimeObj.Spec.RuntimeClassName = "missing"
			_, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).To(HaveOccurred())
			Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
				Name: common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName), Namespace: namespace},
				&corev1.ConfigMap{})).NotTo(Succeed())
		})

		It("should stay bound to its snapshot and warn when runtimeClassName changes", func() {
			_, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())

			runtimeObj.Spec.RuntimeClassName = "another-class"
			bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(bound.Name).To(Equal(className))
			Expect(events()).To(ConsistOf(And(ContainSubstring(common.RuntimeClassChanged), ContainSubstring("another-class"))))
		})
	})

	Describe("reading an existing snapshot", func() {
		Context("when the snapshot cannot be decoded", func() {
			BeforeEach(func() {
				objects = append(objects, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName),
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind,
							Name: runtimeName, UID: runtimeUID, Controller: ptr.To(true),
						}},
					},
					Data: map[string]string{runtimeClassSnapshotKey: "{not json"},
				})
			})

			It("should report it and leave it as it is", func() {
				_, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).To(MatchError(ContainSubstring("is invalid")))
				Expect(events()).To(ConsistOf(ContainSubstring(common.RuntimeClassSnapshotInvalid)))
				Expect(getSnapshot().Data[runtimeClassSnapshotKey]).To(Equal("{not json"))
			})
		})

		Context("when the snapshot exists but the informer cache has not caught up with it", func() {
			BeforeEach(func() {
				existing := runtimeClass.DeepCopy()
				existing.Topology.Worker.Template.Spec.Containers[0].Image = "curvine:v0"
				data, err := json.Marshal(existing)
				Expect(err).NotTo(HaveOccurred())
				objects = append(objects, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:        common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName),
						Namespace:   namespace,
						Annotations: map[string]string{common.AnnotationCacheRuntimeClassName: className},
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind,
							Name: runtimeName, UID: runtimeUID, Controller: ptr.To(true),
						}},
					},
					Data: map[string]string{runtimeClassSnapshotKey: string(data)},
				})
			})

			It("should wait for the snapshot that exists instead of taking another", func() {
				lagging := newLaggingCacheClient(fakeClient)
				lagging.hiddenGets[types.NamespacedName{
					Name: common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName), Namespace: namespace}] = 2
				engine.Client = lagging

				bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				Expect(bound.Topology.Worker.Template.Spec.Containers[0].Image).To(Equal("curvine:v0"))
				Expect(getSnapshot().Data[runtimeClassSnapshotKey]).To(ContainSubstring("curvine:v0"))
			})
		})

		Context("when the snapshot was left by an earlier CacheRuntime of the same name", func() {
			BeforeEach(func() {
				stale := runtimeClass.DeepCopy()
				stale.Topology.Worker.Template.Spec.Containers[0].Image = "curvine:v0"
				data, err := json.Marshal(stale)
				Expect(err).NotTo(HaveOccurred())
				objects = append(objects, &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName),
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{{
							APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind,
							Name: runtimeName, UID: "earlier-uid", Controller: ptr.To(true),
						}},
					},
					Data: map[string]string{runtimeClassSnapshotKey: string(data)},
				})
			})

			It("should replace it with a snapshot of its own", func() {
				bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				Expect(bound.Topology.Worker.Template.Spec.Containers[0].Image).To(Equal("curvine:v1"))
				Expect(metav1.IsControlledBy(getSnapshot(), runtimeObj)).To(BeTrue())
			})
		})
	})

	Describe("binding a runtime whose components already exist", func() {
		// The class has moved on since the components were created: curvine:v1 and the
		// template resources, while the components run curvine:v0 and older defaults.
		createdResources := func() corev1.ResourceRequirements {
			return corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: quantity("500m"), corev1.ResourceMemory: quantity("1Gi")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: quantity("1"), corev1.ResourceMemory: quantity("2Gi")},
			}
		}

		Context("when the components already run what the runtime declares", func() {
			BeforeEach(func() {
				// The worker names limits.memory and stores 1Gi in memory: the workload was
				// charged the quota on top of both memory values.
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: quantity("8Gi")},
				}
				workerResources := createdResources()
				workerResources.Requests[corev1.ResourceMemory] = quantity("2Gi")
				workerResources.Limits[corev1.ResourceMemory] = quantity("9Gi")

				objects = append(objects,
					workload(common.ComponentTypeMaster, "curvine:v0", createdResources()),
					workload(common.ComponentTypeWorker, "curvine:v0", workerResources, tieredStoreVolume("1Gi")))
			})

			It("should adopt their defaults from the workloads instead of the class", func() {
				bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				master := bound.Topology.Master.Template.Spec.Containers[0]
				Expect(master.Image).To(Equal("curvine:v0"))
				Expect(component.ResourcesEqual(master.Resources, createdResources())).To(BeTrue())

				worker := bound.Topology.Worker.Template.Spec.Containers[0]
				Expect(worker.Image).To(Equal("curvine:v0"))
				Expect(component.ResourcesEqual(worker.Resources, corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: quantity("500m"), corev1.ResourceMemory: quantity("1Gi")},
					Limits:   corev1.ResourceList{corev1.ResourceCPU: quantity("1")},
				})).To(BeTrue(), "the quota is taken back out and the key the runtime names is not a default")

				// Everything else follows the class as it is now, which is what the runtime
				// read before it was bound.
				Expect(bound.Topology.Master.ExecutionEntries).To(Equal(runtimeClass.Topology.Master.ExecutionEntries))
				Expect(bound.DataOperationSpecs).To(Equal(runtimeClass.DataOperationSpecs))
				Expect(getSnapshot().Annotations).NotTo(HaveKey(common.AnnotationCacheRuntimeAdoptionPending))
				Expect(events()).To(BeEmpty())
			})

			It("should leave the workloads as they are when syncing on top of the adopted defaults", func() {
				bound, pending, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				Expect(pending).To(BeEmpty())
				masterVersion := getWorkload(common.ComponentTypeMaster).ResourceVersion
				workerVersion := getWorkload(common.ComponentTypeWorker).ResourceVersion

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, bound, pending)).To(Succeed())

				Expect(getWorkload(common.ComponentTypeMaster).ResourceVersion).To(Equal(masterVersion))
				Expect(getWorkload(common.ComponentTypeWorker).ResourceVersion).To(Equal(workerVersion))
			})

			It("should still apply a later edit of the runtime on top of the adopted defaults", func() {
				bound, pending, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				runtimeObj.Spec.Worker.Resources.Limits[corev1.ResourceMemory] = quantity("10Gi")
				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, bound, pending)).To(Succeed())

				resources := getWorkload(common.ComponentTypeWorker).Spec.Template.Spec.Containers[0].Resources
				Expect(resources.Limits.Memory().String()).To(Equal("11Gi"))
				Expect(resources.Requests.Memory().String()).To(Equal("2Gi"))
				Expect(resources.Limits.Cpu().String()).To(Equal("1"))
			})
		})

		Context("when a component would change were the runtime's spec synced", func() {
			BeforeEach(func() {
				// The runtime raised limits.memory, but its workload was never synced.
				runtimeObj.Spec.Worker.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{corev1.ResourceMemory: quantity("8Gi")},
				}
				runtimeObj.Spec.Worker.Replicas = 3
				objects = append(objects,
					workload(common.ComponentTypeMaster, "curvine:v0", createdResources()),
					workload(common.ComponentTypeWorker, "curvine:v0", createdResources()))
			})

			It("should hold its resources and image, warn, and still sync its replicas", func() {
				bound, pending, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				Expect(pending).To(Equal(map[common.ComponentType]bool{common.ComponentTypeWorker: true}))
				Expect(getSnapshot().Annotations).To(HaveKeyWithValue(common.AnnotationCacheRuntimeAdoptionPending, "worker"))
				Expect(events()).To(ConsistOf(And(ContainSubstring(common.RuntimeAdoptionPending), ContainSubstring("worker"))))

				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, bound, pending)).To(Succeed())

				worker := getWorkload(common.ComponentTypeWorker)
				Expect(*worker.Spec.Replicas).To(Equal(int32(3)))
				Expect(worker.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).To(Equal("2Gi"))
				Expect(events()).To(ConsistOf(ContainSubstring(common.RuntimeAdoptionPending)))
			})

			It("should hold them in the reconcile that binds the runtime, before the cache sees the snapshot", func() {
				// A runtime set up by an earlier release is bound in Sync, which syncs its spec right
				// after creating the snapshot, before the informer cache has seen it.
				engine.Client = newLaggingCacheClient(fakeClient)

				Expect(engine.Sync(ctx)).To(Succeed())

				Expect(getSnapshot().Annotations).To(HaveKeyWithValue(common.AnnotationCacheRuntimeAdoptionPending, "worker"))
				worker := getWorkload(common.ComponentTypeWorker)
				Expect(*worker.Spec.Replicas).To(Equal(int32(3)))
				Expect(worker.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).To(Equal("2Gi"))
			})

			It("should apply them once the component is removed from the pending list", func() {
				_, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())

				snapshot := getSnapshot()
				delete(snapshot.Annotations, common.AnnotationCacheRuntimeAdoptionPending)
				Expect(fakeClient.Update(context.TODO(), snapshot)).To(Succeed())

				// The next reconcile reads the snapshot again.
				bound, pending, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				Expect(pending).To(BeEmpty())
				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, bound, pending)).To(Succeed())
				Expect(getWorkload(common.ComponentTypeWorker).Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).
					To(Equal("8Gi"))
			})

			It("should release the component once it already runs what the runtime declares", func() {
				bound, pending, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				workerVersion := getWorkload(common.ComponentTypeWorker).ResourceVersion

				runtimeObj.Spec.Worker.Resources.Limits[corev1.ResourceMemory] = quantity("2Gi")
				runtimeObj.Spec.Worker.Replicas = 1
				Expect(engine.syncRuntimeSpec(ctx, runtimeObj, bound, pending)).To(Succeed())

				Expect(getWorkload(common.ComponentTypeWorker).ResourceVersion).To(Equal(workerVersion))
				Expect(getSnapshot().Annotations).NotTo(HaveKey(common.AnnotationCacheRuntimeAdoptionPending))
			})
		})

		Context("when only some components exist", func() {
			BeforeEach(func() {
				objects = append(objects, workload(common.ComponentTypeMaster, "curvine:v0", createdResources()))
			})

			It("should leave the missing ones to be created from the class template", func() {
				bound, _, err := engine.getBoundRuntimeClass(ctx, runtimeObj)
				Expect(err).NotTo(HaveOccurred())
				Expect(bound.Topology.Master.Template.Spec.Containers[0].Image).To(Equal("curvine:v0"))
				Expect(bound.Topology.Worker).To(Equal(runtimeClass.Topology.Worker))
			})
		})
	})

	Describe("syncing a bound runtime", func() {
		BeforeEach(func() {
			objects = append(objects,
				workload(common.ComponentTypeMaster, "curvine:v1", templateResources()),
				workload(common.ComponentTypeWorker, "curvine:v1", templateResources()))
		})

		It("should keep a CacheRuntimeClass edit away from the workloads and the runtime config", func() {
			Expect(engine.Sync(ctx)).To(Succeed())
			config := &corev1.ConfigMap{}
			configKey := types.NamespacedName{Name: common.GetCacheRuntimeConfigConfigMapName(runtimeName), Namespace: namespace}
			Expect(fakeClient.Get(context.TODO(), configKey, config)).To(Succeed())
			masterVersion := getWorkload(common.ComponentTypeMaster).ResourceVersion
			workerVersion := getWorkload(common.ComponentTypeWorker).ResourceVersion

			editLiveClass(func(live *datav1alpha1.CacheRuntimeClass) {
				for _, definition := range []*datav1alpha1.RuntimeComponentDefinition{live.Topology.Master, live.Topology.Worker} {
					definition.Template.Spec.Containers[0].Image = "curvine:v2"
					definition.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceCPU] = quantity("4")
					definition.Options["block-size"] = "128MB"
				}
			})
			Expect(engine.Sync(ctx)).To(Succeed())

			Expect(getWorkload(common.ComponentTypeMaster).ResourceVersion).To(Equal(masterVersion))
			Expect(getWorkload(common.ComponentTypeWorker).ResourceVersion).To(Equal(workerVersion))
			after := &corev1.ConfigMap{}
			Expect(fakeClient.Get(context.TODO(), configKey, after)).To(Succeed())
			Expect(after.Data).To(Equal(config.Data))
			Expect(strings.Join(configMapValues(after.Data), "")).To(ContainSubstring("64MB"))
		})
	})

	Describe("reading the class for data operations", func() {
		It("should read the class as it is until the runtime is bound, and the snapshot afterwards", func() {
			unbound, err := engine.getBoundRuntimeClassReadOnly(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(unbound.DataOperationSpecs[0].Image).To(Equal("curvine-loader:v1"))
			Expect(fakeClient.Get(context.TODO(), types.NamespacedName{
				Name: common.GetCacheRuntimeClassSnapshotConfigMapName(runtimeName), Namespace: namespace},
				&corev1.ConfigMap{})).NotTo(Succeed(), "data operations do not bind the runtime")

			_, _, err = engine.getBoundRuntimeClass(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			editLiveClass(func(live *datav1alpha1.CacheRuntimeClass) {
				live.DataOperationSpecs[0].Image = "curvine-loader:v2"
			})

			bound, err := engine.getBoundRuntimeClassReadOnly(ctx, runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(bound.DataOperationSpecs[0].Image).To(Equal("curvine-loader:v1"))
		})
	})
})

var _ = Describe("adoptComponentBaseline", Label("pkg.ddc.cache.engine.runtime_class_snapshot_test.go"), func() {
	quantity := resource.MustParse

	podSpecWith := func(image string, resources corev1.ResourceRequirements, volumes ...corev1.Volume) *corev1.PodSpec {
		return &corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main", Image: image, Resources: resources}},
			Volumes:    volumes,
		}
	}
	memoryVolume := func(size string) corev1.Volume {
		sizeLimit := quantity(size)
		return corev1.Volume{
			Name:         tieredStoreVolumeNamePrefix + "0",
			VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory, SizeLimit: &sizeLimit}},
		}
	}

	It("should adopt the workload as it is when the runtime declares nothing", func() {
		resources := corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: quantity("2")}}
		baseline := adoptComponentBaseline(common.ComponentTypeMaster, datav1alpha1.RuntimeComponentCommonSpec{},
			podSpecWith("repo/curvine:v1", resources))

		Expect(baseline.image).To(Equal("repo/curvine:v1"))
		Expect(component.ResourcesEqual(baseline.resources, resources)).To(BeTrue())
		Expect(baseline.applied).To(BeTrue())
	})

	It("should not take the quota back out of a master, which is never charged one", func() {
		resources := corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceMemory: quantity("4Gi")}}
		baseline := adoptComponentBaseline(common.ComponentTypeMaster, datav1alpha1.RuntimeComponentCommonSpec{},
			podSpecWith("curvine:v1", resources, memoryVolume("1Gi")))

		Expect(baseline.resources.Limits.Memory().String()).To(Equal("4Gi"))
		Expect(baseline.applied).To(BeTrue())
	})

	It("should not adopt a worker whose memory is not above the charged quota", func() {
		resources := corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceMemory: quantity("1Gi")}}
		baseline := adoptComponentBaseline(common.ComponentTypeWorker, datav1alpha1.RuntimeComponentCommonSpec{},
			podSpecWith("curvine:v1", resources, memoryVolume("1Gi")))

		Expect(baseline.applied).To(BeFalse())
	})

	It("should drop the claims the runtime names and ignore the order the merge appends them in", func() {
		spec := datav1alpha1.RuntimeComponentCommonSpec{Resources: corev1.ResourceRequirements{
			Claims: []corev1.ResourceClaim{{Name: "gpu"}},
		}}
		// The runtime's claim replaced the template's in place, ahead of a template claim.
		baseline := adoptComponentBaseline(common.ComponentTypeWorker, spec, podSpecWith("curvine:v1",
			corev1.ResourceRequirements{Claims: []corev1.ResourceClaim{{Name: "gpu"}, {Name: "rdma"}}}))

		Expect(baseline.resources.Claims).To(Equal([]corev1.ResourceClaim{{Name: "rdma"}}))
		Expect(baseline.applied).To(BeTrue())
	})

	It("should complete a partial runtimeVersion against the image the workload runs", func() {
		applied := adoptComponentBaseline(common.ComponentTypeWorker,
			datav1alpha1.RuntimeComponentCommonSpec{RuntimeVersion: datav1alpha1.VersionSpec{ImageTag: "v1"}},
			podSpecWith("repo/curvine:v1", corev1.ResourceRequirements{}))
		Expect(applied.applied).To(BeTrue())

		changed := adoptComponentBaseline(common.ComponentTypeWorker,
			datav1alpha1.RuntimeComponentCommonSpec{RuntimeVersion: datav1alpha1.VersionSpec{ImageTag: "v2"}},
			podSpecWith("repo/curvine:v1", corev1.ResourceRequirements{}))
		Expect(changed.image).To(Equal("repo/curvine:v1"))
		Expect(changed.applied).To(BeFalse())
	})

	It("should treat a merge the sync would refuse as leaving the workload unchanged", func() {
		// requests.cpu 2 over a runtime limit of 1: the sync emits a conflict and writes nothing.
		spec := datav1alpha1.RuntimeComponentCommonSpec{Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{corev1.ResourceCPU: quantity("1")},
		}}
		baseline := adoptComponentBaseline(common.ComponentTypeMaster, spec, podSpecWith("curvine:v1",
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: quantity("2")},
				Limits:   corev1.ResourceList{corev1.ResourceCPU: quantity("4")},
			}))
		Expect(baseline.applied).To(BeTrue())
	})
})

// laggingCacheClient stands for the informer cache the controller's client reads from, which
// does not see an object the moment it is written: Get does not find the configmaps created
// through it, nor the configmap named key the first hiddenGets[key] times.
type laggingCacheClient struct {
	client.Client
	created    map[types.NamespacedName]bool
	hiddenGets map[types.NamespacedName]int
}

func newLaggingCacheClient(c client.Client) *laggingCacheClient {
	return &laggingCacheClient{
		Client:     c,
		created:    map[types.NamespacedName]bool{},
		hiddenGets: map[types.NamespacedName]int{},
	}
}

func (c *laggingCacheClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := c.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	if _, isConfigMap := obj.(*corev1.ConfigMap); isConfigMap {
		c.created[client.ObjectKeyFromObject(obj)] = true
	}
	return nil
}

func (c *laggingCacheClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, isConfigMap := obj.(*corev1.ConfigMap); isConfigMap {
		if c.created[key] {
			return apierrors.NewNotFound(corev1.Resource("configmaps"), key.Name)
		}
		if c.hiddenGets[key] > 0 {
			c.hiddenGets[key]--
			return apierrors.NewNotFound(corev1.Resource("configmaps"), key.Name)
		}
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func configMapValues(m map[string]string) []string {
	values := make([]string, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}
	return values
}
