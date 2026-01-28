/*
Copyright 2022 The Fluid Authors.

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

package jindofsx

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("JindoFSxEngine_syncMasterSpec", func() {
	var (
		res resource.Quantity
		ctx cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	It("should not change when no resource for jindoruntime", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("emtpy")
		jindoRuntime.SetNamespace("default")

		master := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "emtpy-jindofs-master",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{},
		}

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, master.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "emtpy",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

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

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, master.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "nomaster",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).To(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should not change when master spec is same", func() {
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

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, master.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "same",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})
})

var _ = Describe("JindoFSxEngine_syncWorkerSpec", func() {
	var (
		res resource.Quantity
		ctx cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	It("should not change when no resource for jindoruntime", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("emtpy")
		jindoRuntime.SetNamespace("default")

		worker := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "emtpy-jindofs-worker",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{},
		}

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, worker.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "emtpy",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

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

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, worker.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "noworker",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).To(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should not change when worker spec is same", func() {
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

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, worker.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "same",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})
})

var _ = Describe("JindoFSxEngine_syncFuseSpec", func() {
	var (
		res resource.Quantity
		ctx cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	It("should not change when no resource for jindoruntime", func() {
		jindoRuntime := &datav1alpha1.JindoRuntime{}
		jindoRuntime.SetName("emtpy")
		jindoRuntime.SetNamespace("default")

		fuse := &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "emtpy-jindofs-fuse",
				Namespace: "default",
			},
			Spec: appsv1.DaemonSetSpec{},
		}

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, fuse.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "emtpy",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

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

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, fuse.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "nofuse",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).To(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})

	It("should not change when fuse spec is same", func() {
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

		runtimeObjs := []runtime.Object{}
		runtimeObjs = append(runtimeObjs, fuse.DeepCopy())

		s := runtime.NewScheme()
		s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
		_ = corev1.AddToScheme(s)
		runtimeObjs = append(runtimeObjs, jindoRuntime)
		client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

		e := &JindoFSxEngine{
			runtime:   jindoRuntime,
			name:      "same",
			namespace: "default",
			Log:       fake.NullLogger(),
			Client:    client,
		}

		gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

		Expect(err).NotTo(HaveOccurred())
		Expect(gotChanged).To(BeFalse())
	})
})
