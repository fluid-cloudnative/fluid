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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// laggingCacheClient stands in for a client reading through an informer cache that has not
// seen some configmaps yet: Get reports them as not found.
type laggingCacheClient struct {
	client.Client
	// hideCreated hides every configmap created through this client.
	hideCreated bool
	created     map[string]bool
	// hideGets hides the snapshot configmap from that many Gets.
	hideGets int
	// stale, when set, is returned for the snapshot configmap instead of what is stored.
	stale *corev1.ConfigMap
}

func (c *laggingCacheClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if err := c.Client.Create(ctx, obj, opts...); err != nil {
		return err
	}
	if _, ok := obj.(*corev1.ConfigMap); ok {
		if c.created == nil {
			c.created = map[string]bool{}
		}
		c.created[obj.GetName()] = true
	}
	return nil
}

func (c *laggingCacheClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*corev1.ConfigMap); ok {
		hidden := c.hideCreated && c.created[key.Name]
		if key.Name == common.GetCacheRuntimeClassSnapshotConfigMapName("test-runtime") && c.hideGets > 0 {
			c.hideGets--
			hidden = true
		}
		if hidden {
			return apierrors.NewNotFound(schema.GroupResource{Resource: "configmaps"}, key.Name)
		}
		if c.stale != nil && key.Name == c.stale.Name {
			c.stale.DeepCopyInto(obj.(*corev1.ConfigMap))
			return nil
		}
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

var _ = Describe("CacheRuntimeClass snapshot", Label("pkg.ddc.cache.engine.runtime_class_snapshot_test.go"), func() {
	const (
		masterSts    = "test-runtime-master"
		workerSts    = "test-runtime-worker"
		snapshotName = "fluid-runtime-class-snapshot-test-runtime"
	)

	var (
		engine       *CacheEngine
		recorder     *record.FakeRecorder
		fakeClient   client.Client
		dataset      *datav1alpha1.Dataset
		runtimeObj   *datav1alpha1.CacheRuntime
		runtimeClass *datav1alpha1.CacheRuntimeClass
		ctx          cruntime.ReconcileRequestContext
	)

	limits := func(memory string) corev1.ResourceRequirements {
		return corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse(memory)}}
	}

	componentDefinition := func(name, image string, resources corev1.ResourceRequirements) *datav1alpha1.RuntimeComponentDefinition {
		return &datav1alpha1.RuntimeComponentDefinition{
			Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: name, Image: image, Resources: resources}},
			}},
		}
	}

	newWorkload := func(name, container, image string, replicas int32, resources corev1.ResourceRequirements) *workloadv1alpha1.AdvancedStatefulSet {
		return &workloadv1alpha1.AdvancedStatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
				Replicas: ptr.To(replicas),
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: container, Image: image, Resources: resources}},
				}},
			},
			Status: workloadv1alpha1.AdvancedStatefulSetStatus{
				ReadyReplicas: replicas, CurrentReplicas: replicas, AvailableReplicas: replicas,
			},
		}
	}

	// build creates the client with the fixture objects plus extra, and an engine reading through it.
	build := func(extra ...client.Object) {
		objects := append([]client.Object{dataset, runtimeObj, runtimeClass}, extra...)
		fakeClient = fake.NewClientBuilder().
			WithScheme(CacheEngineTestScheme).
			WithObjects(objects...).
			WithStatusSubresource(dataset, runtimeObj).
			Build()
		recorder = record.NewFakeRecorder(20)
		engine = &CacheEngine{
			name:      "test-runtime",
			namespace: "default",
			Client:    fakeClient,
			Log:       ctrl.Log.WithName("test"),
			Recorder:  recorder,
		}
		ctx = cruntime.ReconcileRequestContext{
			Context:        context.Background(),
			Client:         fakeClient,
			Log:            ctrl.Log.WithName("test"),
			NamespacedName: types.NamespacedName{Name: "test-runtime", Namespace: "default"},
		}
	}

	// existingComponents are the workloads Setup creates from the class as the fixture declares it.
	existingComponents := func() []client.Object {
		return []client.Object{
			newWorkload(masterSts, "master", "test-master:v1", 1, limits("2Gi")),
			newWorkload(workerSts, "worker", "test-worker:v1", 2, limits("4Gi")),
			&appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Name: "test-runtime-client", Namespace: "default"},
				Spec: appsv1.DaemonSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "client", Image: "test-client:v1"}},
				}}},
			},
		}
	}

	getSnapshot := func() *corev1.ConfigMap {
		snapshot := &corev1.ConfigMap{}
		Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: snapshotName, Namespace: "default"}, snapshot)).To(Succeed())
		return snapshot
	}

	getWorkload := func(name string) *workloadv1alpha1.AdvancedStatefulSet {
		sts := &workloadv1alpha1.AdvancedStatefulSet{}
		Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: "default"}, sts)).To(Succeed())
		return sts
	}

	// editClass edits the live CacheRuntimeClass the way an administrator would.
	editClass := func(edit func(*datav1alpha1.CacheRuntimeClass)) {
		live := &datav1alpha1.CacheRuntimeClass{}
		Expect(fakeClient.Get(context.Background(), types.NamespacedName{Name: "test-class"}, live)).To(Succeed())
		edit(live)
		Expect(fakeClient.Update(context.Background(), live)).To(Succeed())
	}

	bind := func() *datav1alpha1.CacheRuntimeClass {
		bound, err := engine.getBoundRuntimeClass(context.Background(), runtimeObj)
		Expect(err).NotTo(HaveOccurred())
		return bound
	}

	BeforeEach(func() {
		dataset = &datav1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default", UID: "dataset-uid"}}
		runtimeObj = &datav1alpha1.CacheRuntime{
			TypeMeta:   metav1.TypeMeta{APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind},
			ObjectMeta: metav1.ObjectMeta{Name: "test-runtime", Namespace: "default", UID: "runtime-uid"},
			Spec: datav1alpha1.CacheRuntimeSpec{
				RuntimeClassName: "test-class",
				Master:           datav1alpha1.CacheRuntimeMasterSpec{Replicas: 1},
				Worker:           datav1alpha1.CacheRuntimeWorkerSpec{Replicas: 2},
			},
		}
		runtimeClass = &datav1alpha1.CacheRuntimeClass{
			ObjectMeta:     metav1.ObjectMeta{Name: "test-class"},
			FileSystemType: "test-fs",
			Topology: &datav1alpha1.RuntimeTopology{
				Master: componentDefinition("master", "test-master:v1", limits("2Gi")),
				Worker: componentDefinition("worker", "test-worker:v1", limits("4Gi")),
				Client: componentDefinition("client", "test-client:v1", corev1.ResourceRequirements{}),
			},
		}
		runtimeClass.Topology.Master.ExecutionEntries = &datav1alpha1.ExecutionEntries{
			ReportSummary: &datav1alpha1.ExecutionCommonEntry{Command: []string{"summary-v1"}, TimeoutSeconds: 10},
		}
	})

	Describe("binding a runtime that has no components yet", func() {
		BeforeEach(func() { build() })

		It("should create an immutable snapshot controlled by the runtime", func() {
			bound := bind()

			Expect(bound.Topology.Worker.Template.Spec.Containers[0].Image).To(Equal("test-worker:v1"))

			snapshot := getSnapshot()
			Expect(snapshot.Immutable).To(Equal(ptr.To(true)))
			Expect(metav1.IsControlledBy(snapshot, runtimeObj)).To(BeTrue())
			Expect(snapshot.Labels).To(HaveKeyWithValue(common.LabelCacheRuntimeName, "test-runtime"))
			Expect(snapshot.Annotations).To(HaveKeyWithValue(common.AnnotationCacheRuntimeClassName, "test-class"))

			stored := &datav1alpha1.CacheRuntimeClass{}
			Expect(json.Unmarshal([]byte(snapshot.Data[runtimeClassSnapshotKey]), stored)).To(Succeed())
			Expect(stored.Name).To(Equal("test-class"))
			Expect(stored.ResourceVersion).To(BeEmpty(), "only the class content is kept")
			Expect(stored.Topology.Master.ExecutionEntries.ReportSummary.Command).To(Equal([]string{"summary-v1"}))
		})

		It("should keep returning the bound class after the class is edited", func() {
			bind()
			editClass(func(live *datav1alpha1.CacheRuntimeClass) {
				live.Topology.Worker.Template.Spec.Containers[0].Image = "test-worker:v2"
				live.Topology.Worker.Template.Spec.Containers[0].Resources = limits("16Gi")
				live.Topology.Master.ExecutionEntries.ReportSummary.Command = []string{"summary-v2"}
			})

			bound := bind()

			worker := bound.Topology.Worker.Template.Spec.Containers[0]
			Expect(worker.Image).To(Equal("test-worker:v1"))
			Expect(worker.Resources.Limits.Memory().String()).To(Equal("4Gi"))
			Expect(bound.Topology.Master.ExecutionEntries.ReportSummary.Command).To(Equal([]string{"summary-v1"}))
		})

		It("should keep returning the bound class after the class is deleted", func() {
			bind()
			Expect(fakeClient.Delete(context.Background(), runtimeClass)).To(Succeed())

			bound := bind()
			Expect(bound.Name).To(Equal("test-class"))
		})

		It("should fail when the class does not exist", func() {
			Expect(fakeClient.Delete(context.Background(), runtimeClass)).To(Succeed())

			_, err := engine.getBoundRuntimeClass(context.Background(), runtimeObj)
			Expect(apierrors.IsNotFound(err)).To(BeTrue(), "got %v", err)
		})

		It("should stay bound and warn when runtimeClassName changes", func() {
			bind()
			runtimeObj.Spec.RuntimeClassName = "other-class"

			bound := bind()

			Expect(bound.Name).To(Equal("test-class"))
			Expect(recorder.Events).To(Receive(And(
				HavePrefix(corev1.EventTypeWarning+" "+common.RuntimeClassChanged+" "),
				ContainSubstring(`"other-class"`))))
		})
	})

	Describe("reading a snapshot", func() {
		It("should report a malformed snapshot and leave it unchanged", func() {
			malformed := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: snapshotName, Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind,
						Name: "test-runtime", UID: "runtime-uid", Controller: ptr.To(true),
					}},
				},
				Data: map[string]string{runtimeClassSnapshotKey: "{"},
			}
			build(malformed)

			_, err := engine.getBoundRuntimeClass(context.Background(), runtimeObj)

			Expect(err).To(MatchError(ContainSubstring("is invalid")))
			Expect(getSnapshot().Data[runtimeClassSnapshotKey]).To(Equal("{"))
			Expect(recorder.Events).To(Receive(HavePrefix(corev1.EventTypeWarning + " " + common.RuntimeClassSnapshotInvalid + " ")))
		})

		It("should report a snapshot without the class key", func() {
			build(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: snapshotName, Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind,
					Name: "test-runtime", UID: "runtime-uid", Controller: ptr.To(true),
				}},
			}})

			_, err := engine.getBoundRuntimeClass(context.Background(), runtimeObj)
			Expect(err).To(MatchError(ContainSubstring(runtimeClassSnapshotKey)))
		})

		It("should replace a snapshot left by an earlier runtime of the same name", func() {
			stale := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: snapshotName, Namespace: "default",
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: datav1alpha1.GroupVersion.String(), Kind: datav1alpha1.CacheRuntimeKind,
						Name: "test-runtime", UID: "earlier-uid", Controller: ptr.To(true),
					}},
				},
				Data: map[string]string{runtimeClassSnapshotKey: "{}"},
			}
			build(stale)

			bound := bind()

			Expect(bound.Name).To(Equal("test-class"))
			Expect(metav1.IsControlledBy(getSnapshot(), runtimeObj)).To(BeTrue())
		})

		It("should not delete the runtime's own snapshot when the cache still shows an earlier one", func() {
			build()
			bind()
			own := getSnapshot()
			// The cache still returns the snapshot of an earlier runtime of the same name,
			// which the runtime's own snapshot has since replaced.
			stale := own.DeepCopy()
			stale.UID = "earlier-snapshot-uid"
			stale.OwnerReferences[0].UID = "earlier-uid"
			// The fake client ignores UID preconditions; enforce them like the API server.
			guarded := interceptor.NewClient(fakeClient.(client.WithWatch), interceptor.Funcs{
				Delete: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
					options := &client.DeleteOptions{}
					options.ApplyOptions(opts)
					current := &corev1.ConfigMap{}
					if err := c.Get(ctx, client.ObjectKeyFromObject(obj), current); err != nil {
						return err
					}
					if options.Preconditions == nil || options.Preconditions.UID == nil {
						Fail("the snapshot was deleted without a UID precondition")
					}
					if *options.Preconditions.UID != current.UID {
						return apierrors.NewConflict(schema.GroupResource{Resource: "configmaps"}, obj.GetName(), nil)
					}
					return c.Delete(ctx, obj, opts...)
				},
			})
			engine.Client = &laggingCacheClient{Client: guarded, stale: stale}

			_, err := engine.getBoundRuntimeClass(context.Background(), runtimeObj)

			Expect(apierrors.IsConflict(errors.Unwrap(err))).To(BeTrue(), "got %v", err)
			Expect(getSnapshot().UID).To(Equal(own.UID))
		})

		It("should wait for a snapshot that exists but the cache does not show yet", func() {
			build()
			bind()
			engine.Client = &laggingCacheClient{Client: fakeClient, hideGets: 3}
			// The class changed since: had a new snapshot been taken, it would show.
			editClass(func(live *datav1alpha1.CacheRuntimeClass) {
				live.Topology.Worker.Template.Spec.Containers[0].Image = "test-worker:v2"
			})

			bound := bind()

			Expect(bound.Topology.Worker.Template.Spec.Containers[0].Image).To(Equal("test-worker:v1"))
		})

		It("should neither bind nor read the live class when only reading an unbound runtime", func() {
			build()

			_, err := engine.getBoundRuntimeClassReadOnly(context.Background(), runtimeObj)

			Expect(err).To(MatchError(ContainSubstring("not bound")))
			snapshot, err := engine.readRuntimeClassSnapshot(context.Background(), runtimeObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(snapshot).To(BeNil(), "a read-only lookup must not bind the runtime")
		})

		It("should leave a snapshot it does not see as the runtime's alone when only reading", func() {
			// The dataset controller's cache can show a runtime of the same name that has since
			// been recreated: the snapshot belongs to the current one and must stay.
			build()
			bind()
			earlier := runtimeObj.DeepCopy()
			earlier.UID = "earlier-uid"

			_, err := engine.getBoundRuntimeClassReadOnly(context.Background(), earlier)

			Expect(err).To(MatchError(ContainSubstring("not controlled")))
			Expect(metav1.IsControlledBy(getSnapshot(), runtimeObj)).To(BeTrue())
		})

		It("should read the snapshot once the runtime is bound when only reading", func() {
			build()
			bind()
			editClass(func(live *datav1alpha1.CacheRuntimeClass) {
				live.Topology.Worker.Template.Spec.Containers[0].Image = "test-worker:v2"
			})

			bound, err := engine.getBoundRuntimeClassReadOnly(context.Background(), runtimeObj)

			Expect(err).NotTo(HaveOccurred())
			Expect(bound.Topology.Worker.Template.Spec.Containers[0].Image).To(Equal("test-worker:v1"))
		})
	})

	Describe("Sync", func() {
		It("should use the class it bound even when the cache does not show the new snapshot", func() {
			// The snapshot created by this Sync is invisible to the client for the rest of it.
			// Re-reading it in a later step would try to bind again and time out waiting.
			build(existingComponents()...)
			engine.Client = &laggingCacheClient{Client: fakeClient, hideCreated: true}
			ctx.Client = engine.Client

			Expect(engine.Sync(ctx)).To(Succeed())

			Expect(getSnapshot().Annotations).To(HaveKeyWithValue(common.AnnotationCacheRuntimeClassName, "test-class"))
		})

		It("should leave the workloads and the runtime config alone when the class is edited", func() {
			build(existingComponents()...)
			Expect(engine.Sync(ctx)).To(Succeed())
			config := &corev1.ConfigMap{}
			configKey := types.NamespacedName{Name: common.GetCacheRuntimeConfigConfigMapName("test-runtime"), Namespace: "default"}
			Expect(fakeClient.Get(context.Background(), configKey, config)).To(Succeed())
			master, worker := getWorkload(masterSts).ResourceVersion, getWorkload(workerSts).ResourceVersion

			editClass(func(live *datav1alpha1.CacheRuntimeClass) {
				live.Topology.Master.Template.Spec.Containers[0].Image = "test-master:v2"
				live.Topology.Worker.Template.Spec.Containers[0].Resources = limits("16Gi")
				live.Topology.Master.Options = map[string]string{"new": "option"}
			})
			Expect(engine.Sync(ctx)).To(Succeed())

			Expect(getWorkload(masterSts).ResourceVersion).To(Equal(master))
			Expect(getWorkload(workerSts).ResourceVersion).To(Equal(worker))
			after := &corev1.ConfigMap{}
			Expect(fakeClient.Get(context.Background(), configKey, after)).To(Succeed())
			Expect(after.ResourceVersion).To(Equal(config.ResourceVersion))
		})
	})

})
