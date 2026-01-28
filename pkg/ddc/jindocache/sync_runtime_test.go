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
)

var _ = Describe("JindoCacheEngine", func() {
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	Describe("syncMasterSpec", func() {
		DescribeTable("should sync master spec correctly",
			func(name, namespace string, jindoRuntime *datav1alpha1.JindoRuntime, master *appsv1.StatefulSet, wantChanged, wantErr bool) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, master.DeepCopy())

				s := runtime.NewScheme()
				jindoRuntime.SetName(name)
				jindoRuntime.SetNamespace(namespace)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, master)
				s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)

				_ = corev1.AddToScheme(s)
				runtimeObjs = append(runtimeObjs, jindoRuntime)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

				e := &JindoCacheEngine{
					runtime:   jindoRuntime,
					name:      name,
					namespace: namespace,
					Log:       fake.NullLogger(),
					Client:    client,
				}

				ctx := cruntime.ReconcileRequestContext{}
				gotChanged, err := e.syncMasterSpec(ctx, jindoRuntime)

				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(gotChanged).To(Equal(wantChanged))
			},
			Entry("Not resource for jindoruntime",
				"emtpy",
				"default",
				&datav1alpha1.JindoRuntime{},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-master",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				false,
				false,
			),
			Entry("Master not found",
				"nomaster",
				"default",
				&datav1alpha1.JindoRuntime{},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nomaster",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				false,
				true,
			),
			Entry("Master not change",
				"same",
				"default",
				&datav1alpha1.JindoRuntime{
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
				},
				&appsv1.StatefulSet{
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
				},
				false,
				false,
			),
		)
	})

	Describe("syncWorkerSpec", func() {
		DescribeTable("should sync worker spec correctly",
			func(name, namespace string, jindoRuntime *datav1alpha1.JindoRuntime, worker *appsv1.StatefulSet, wantChanged, wantErr bool) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, worker.DeepCopy())

				s := runtime.NewScheme()
				jindoRuntime.SetName(name)
				jindoRuntime.SetNamespace(namespace)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, worker)
				s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)

				_ = corev1.AddToScheme(s)
				runtimeObjs = append(runtimeObjs, jindoRuntime)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

				e := &JindoCacheEngine{
					runtime:   jindoRuntime,
					name:      name,
					namespace: namespace,
					Log:       fake.NullLogger(),
					Client:    client,
				}

				ctx := cruntime.ReconcileRequestContext{}
				gotChanged, err := e.syncWorkerSpec(ctx, jindoRuntime)

				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(gotChanged).To(Equal(wantChanged))
			},
			Entry("Not resource for jindoruntime",
				"emtpy",
				"default",
				&datav1alpha1.JindoRuntime{},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-worker",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				false,
				false,
			),
			Entry("worker not found",
				"noworker",
				"default",
				&datav1alpha1.JindoRuntime{},
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noworker",
						Namespace: "default",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				false,
				true,
			),
			Entry("worker not change",
				"same",
				"default",
				&datav1alpha1.JindoRuntime{
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
				},
				&appsv1.StatefulSet{
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
				},
				false,
				false,
			),
		)
	})

	Describe("syncFuseSpec", func() {
		DescribeTable("should sync fuse spec correctly",
			func(name, namespace string, jindoRuntime *datav1alpha1.JindoRuntime, fuse *appsv1.DaemonSet, wantChanged, wantErr bool) {
				runtimeObjs := []runtime.Object{}
				runtimeObjs = append(runtimeObjs, fuse.DeepCopy())

				s := runtime.NewScheme()
				jindoRuntime.SetName(name)
				jindoRuntime.SetNamespace(namespace)
				s.AddKnownTypes(appsv1.SchemeGroupVersion, fuse)
				s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)

				_ = corev1.AddToScheme(s)
				runtimeObjs = append(runtimeObjs, jindoRuntime)
				client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

				e := &JindoCacheEngine{
					runtime:   jindoRuntime,
					name:      name,
					namespace: namespace,
					Log:       fake.NullLogger(),
					Client:    client,
				}

				ctx := cruntime.ReconcileRequestContext{}
				gotChanged, err := e.syncFuseSpec(ctx, jindoRuntime)

				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(gotChanged).To(Equal(wantChanged))
			},
			Entry("Not resource for jindoruntime",
				"emtpy",
				"default",
				&datav1alpha1.JindoRuntime{},
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-fuse",
						Namespace: "default",
					},
					Spec: appsv1.DaemonSetSpec{},
				},
				false,
				false,
			),
			Entry("fuse not found",
				"nofuse",
				"default",
				&datav1alpha1.JindoRuntime{},
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nofuse",
						Namespace: "default",
					},
					Spec: appsv1.DaemonSetSpec{},
				},
				false,
				true,
			),
			Entry("fuse not change",
				"same",
				"default",
				&datav1alpha1.JindoRuntime{
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
				},
				&appsv1.DaemonSet{
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
				},
				false,
				false,
			),
		)
	})
})
