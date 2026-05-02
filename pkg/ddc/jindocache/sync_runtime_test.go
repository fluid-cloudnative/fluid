/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("JindoCacheEngine SyncRuntime", func() {
	const (
		runtimeName      = "sync-runtime"
		runtimeNamespace = "default"
	)

	newContext := func(fakeClient ctrlclient.Client) cruntime.ReconcileRequestContext {
		return cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{Name: runtimeName, Namespace: runtimeNamespace},
			Client:         fakeClient,
			Log:            fake.NullLogger(),
			RuntimeType:    common.JindoRuntime,
			EngineImpl:     common.JindoRuntime,
		}
	}

	It("should return when syncing the master spec changes resources", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName, Namespace: runtimeNamespace},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("200m"),
						},
					},
				},
			},
		}
		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-master", Namespace: runtimeNamespace},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name:      "master",
					Resources: corev1.ResourceRequirements{},
				}}}},
			},
		}
		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-worker", Namespace: runtimeNamespace},
			Spec:       appsv1.StatefulSetSpec{UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType}},
		}
		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-fuse", Namespace: runtimeNamespace},
			Spec:       appsv1.DaemonSetSpec{UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType}},
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, jindoRuntime, master, worker, fuse)
		engine := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      runtimeName,
			namespace: runtimeNamespace,
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		changed, err := engine.SyncRuntime(newContext(fakeClient))

		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(BeTrue())

		updatedMaster := &appsv1.StatefulSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKeyFromObject(master), updatedMaster)).To(Succeed())
		Expect(updatedMaster.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).NotTo(BeNil())
		Expect(updatedMaster.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).To(Equal("200m"))
	})

	It("should return an error when the runtime cannot be loaded", func() {
		fakeClient := fake.NewFakeClientWithScheme(testScheme)
		engine := &JindoCacheEngine{
			name:      runtimeName,
			namespace: runtimeNamespace,
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		changed, err := engine.SyncRuntime(newContext(fakeClient))

		Expect(err).To(HaveOccurred())
		Expect(changed).To(BeFalse())
	})

	It("should return when syncing the worker spec changes resources", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName, Namespace: runtimeNamespace},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("300m"),
						},
					},
				},
			},
		}
		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-master", Namespace: runtimeNamespace},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name:      "master",
					Resources: corev1.ResourceRequirements{},
				}}}},
			},
		}
		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-worker", Namespace: runtimeNamespace},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name:      "worker",
					Resources: corev1.ResourceRequirements{},
				}}}},
			},
		}
		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-fuse", Namespace: runtimeNamespace},
			Spec: appsv1.DaemonSetSpec{
				UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType},
				Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name:      "fuse",
					Resources: corev1.ResourceRequirements{},
				}}}},
			},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, jindoRuntime, master, worker, fuse)
		engine := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      runtimeName,
			namespace: runtimeNamespace,
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		changed, err := engine.SyncRuntime(newContext(fakeClient))

		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(BeTrue())

		updatedWorker := &appsv1.StatefulSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKeyFromObject(worker), updatedWorker)).To(Succeed())
		Expect(updatedWorker.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()).NotTo(BeNil())
		Expect(updatedWorker.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()).To(Equal("300m"))
	})

	It("should return when syncing the fuse spec changes resources", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName, Namespace: runtimeNamespace},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{Disabled: true},
				Worker: datav1alpha1.JindoCompTemplateSpec{Disabled: true},
				Fuse: datav1alpha1.JindoFuseSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		}
		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: runtimeName + "-jindofs-fuse", Namespace: runtimeNamespace},
			Spec: appsv1.DaemonSetSpec{
				UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType},
				Template: corev1.PodTemplateSpec{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name:      "fuse",
					Resources: corev1.ResourceRequirements{},
				}}}},
			},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, jindoRuntime, fuse)
		engine := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      runtimeName,
			namespace: runtimeNamespace,
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		changed, err := engine.SyncRuntime(newContext(fakeClient))

		Expect(err).NotTo(HaveOccurred())
		Expect(changed).To(BeTrue())

		updatedFuse := &appsv1.DaemonSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKeyFromObject(fuse), updatedFuse)).To(Succeed())
		Expect(updatedFuse.Spec.Template.Spec.Containers[0].Resources.Requests.Memory()).NotTo(BeNil())
		Expect(updatedFuse.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).To(Equal("1Gi"))
	})
})

var _ = Describe("JindoCacheEngine_syncMasterSpec", func() {
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	It("should handle empty resource for jindoruntime", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("empty")
		jindoRuntime.SetNamespace("default")

		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "empty-jindofs-master",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{},
		}

		runtimeObjs := []runtime.Object{master.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "empty",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should return error when master not found", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("nomaster")
		jindoRuntime.SetNamespace("default")

		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nomaster",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{},
		}

		runtimeObjs := []runtime.Object{master.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "nomaster",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).To(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should not change master when spec is the same", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.Memory,
							Quota:      &res,
						},
					},
				},
				Master: datav1alpha1.JindoCompTemplateSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		}
		jindoRuntime.SetName("same")
		jindoRuntime.SetNamespace("default")

		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "same-jindofs-master",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "master",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("100m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
		}

		runtimeObjs := []runtime.Object{master.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "same",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should update master strategy to OnDelete before syncing resources", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("strategy")
		jindoRuntime.SetNamespace("default")

		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "strategy-jindofs-master",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.RollingUpdateStatefulSetStrategyType},
			},
		}

		runtimeObjs := []runtime.Object{master.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		fakeClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "strategy",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())

		updatedMaster := &appsv1.StatefulSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKey{Name: master.Name, Namespace: master.Namespace}, updatedMaster)).To(Succeed())
		Expect(updatedMaster.Spec.UpdateStrategy.Type).To(Equal(appsv1.OnDeleteStatefulSetStrategyType))
	})

	It("should update master resources when runtime spec changes", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
				},
			},
		}
		jindoRuntime.SetName("change")
		jindoRuntime.SetNamespace("default")

		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "change-jindofs-master",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "master",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						}},
					},
				},
			},
		}

		runtimeObjs := []runtime.Object{master.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		fakeClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "change",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeTrue())

		updatedMaster := &appsv1.StatefulSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKey{Name: master.Name, Namespace: master.Namespace}, updatedMaster)).To(Succeed())
		Expect(updatedMaster.Spec.Template.Spec.Containers[0].Resources).To(Equal(jindoRuntime.Spec.Master.Resources))
	})
})

var _ = Describe("JindoCacheEngine_syncWorkerSpec", func() {
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	It("should handle empty resource for jindoruntime", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("empty")
		jindoRuntime.SetNamespace("default")

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "empty-jindofs-worker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{},
		}

		runtimeObjs := []runtime.Object{worker.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "empty",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should return error when worker not found", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("noworker")
		jindoRuntime.SetNamespace("default")

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "noworker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{},
		}

		runtimeObjs := []runtime.Object{worker.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "noworker",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).To(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should not change worker when spec is the same", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.Memory,
							Quota:      &res,
						},
					},
				},
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: res,
						},
					},
				},
			},
		}
		jindoRuntime.SetName("same")
		jindoRuntime.SetNamespace("default")

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "same-jindofs-worker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "worker",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("100m"),
										corev1.ResourceMemory: resource.MustParse("320Gi"),
									},
								},
							},
						},
					},
				},
			},
		}

		runtimeObjs := []runtime.Object{worker.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "same",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should update worker strategy to OnDelete before syncing resources", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("strategy")
		jindoRuntime.SetNamespace("default")

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "strategy-jindofs-worker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.RollingUpdateStatefulSetStrategyType},
			},
		}

		runtimeObjs := []runtime.Object{worker.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		fakeClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "strategy",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())

		updatedWorker := &appsv1.StatefulSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKey{Name: worker.Name, Namespace: worker.Namespace}, updatedWorker)).To(Succeed())
		Expect(updatedWorker.Spec.UpdateStrategy.Type).To(Equal(appsv1.OnDeleteStatefulSetStrategyType))
	})

	It("should update worker resources when runtime spec changes", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("300m"),
							corev1.ResourceMemory: resource.MustParse("3Gi"),
						},
					},
				},
			},
		}
		jindoRuntime.SetName("change")
		jindoRuntime.SetNamespace("default")

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "change-jindofs-worker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType},
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "worker",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						}},
					},
				},
			},
		}

		runtimeObjs := []runtime.Object{worker.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		fakeClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "change",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeTrue())

		updatedWorker := &appsv1.StatefulSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKey{Name: worker.Name, Namespace: worker.Namespace}, updatedWorker)).To(Succeed())
		Expect(updatedWorker.Spec.Template.Spec.Containers[0].Resources).To(Equal(jindoRuntime.Spec.Worker.Resources))
	})
})

var _ = Describe("JindoCacheEngine_syncFuseSpec", func() {
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	It("should handle empty resource for jindoruntime", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("empty")
		jindoRuntime.SetNamespace("default")

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "empty-jindofs-fuse",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{},
		}

		runtimeObjs := []runtime.Object{fuse.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "empty",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should return error when fuse not found", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("nofuse")
		jindoRuntime.SetNamespace("default")

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nofuse",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{},
		}

		runtimeObjs := []runtime.Object{fuse.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "nofuse",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).To(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should not change fuse when spec is the same", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{
						{
							MediumType: common.Memory,
							Quota:      &res,
						},
					},
				},
				Fuse: datav1alpha1.JindoFuseSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		}
		jindoRuntime.SetName("same")
		jindoRuntime.SetNamespace("default")

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "same-jindofs-fuse",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "fuse",
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										corev1.ResourceCPU:    resource.MustParse("100m"),
										corev1.ResourceMemory: resource.MustParse("1Gi"),
									},
								},
							},
						},
					},
				},
			},
		}

		runtimeObjs := []runtime.Object{fuse.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "same",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should update fuse strategy to OnDelete before syncing resources", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("strategy")
		jindoRuntime.SetNamespace("default")

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "strategy-jindofs-fuse",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{
				UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.RollingUpdateDaemonSetStrategyType},
			},
		}

		runtimeObjs := []runtime.Object{fuse.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		fakeClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "strategy",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())

		updatedFuse := &appsv1.DaemonSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKey{Name: fuse.Name, Namespace: fuse.Namespace}, updatedFuse)).To(Succeed())
		Expect(updatedFuse.Spec.UpdateStrategy.Type).To(Equal(appsv1.OnDeleteDaemonSetStrategyType))
	})

	It("should update fuse resources and metrics annotation when runtime spec changes", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Fuse: datav1alpha1.JindoFuseSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
					Metrics: datav1alpha1.ClientMetrics{ScrapeTarget: "MountPod"},
				},
			},
		}
		jindoRuntime.SetName("change")
		jindoRuntime.SetNamespace("default")

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "change-jindofs-fuse",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{
				UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"existing": "true"}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: "fuse",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						}},
					},
				},
			},
		}

		runtimeObjs := []runtime.Object{fuse.DeepCopy(), jindoRuntime}
		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)

		fakeClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoCacheEngine{
			runtime:   jindoRuntime,
			name:      "change",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    fakeClient,
		}

		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeTrue())

		updatedFuse := &appsv1.DaemonSet{}
		Expect(fakeClient.Get(context.TODO(), ctrlclient.ObjectKey{Name: fuse.Name, Namespace: fuse.Namespace}, updatedFuse)).To(Succeed())
		Expect(updatedFuse.Spec.Template.Spec.Containers[0].Resources).To(Equal(jindoRuntime.Spec.Fuse.Resources))
		Expect(updatedFuse.Spec.Template.ObjectMeta.Annotations).To(HaveKeyWithValue(common.AnnotationPrometheusFuseMetricsScrapeKey, common.True))
	})
})
