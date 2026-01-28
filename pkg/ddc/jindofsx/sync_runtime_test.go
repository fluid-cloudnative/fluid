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
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
	}
	type args struct {
		ctx    cruntime.ReconcileRequestContext
		master *appsv1.StatefulSet
	}

	DescribeTable("sync master spec",
		func(fields fields, args args, wantChanged bool, wantErr bool, wantResource corev1.ResourceRequirements) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, args.master.DeepCopy())

			s := runtime.NewScheme()
			fields.runtime.SetName(fields.name)
			fields.runtime.SetNamespace(fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, args.master)
			s.AddKnownTypes(datav1alpha1.GroupVersion, fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, fields.runtime)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoFSxEngine{
				runtime:   fields.runtime,
				name:      fields.name,
				namespace: fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncMasterSpec(args.ctx, fields.runtime)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(gotChanged).To(Equal(wantChanged))
		},

		Entry("Not resource for jindoruntime",
			fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			},
			args{
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-master",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			false,
			false,
			corev1.ResourceRequirements{},
		),

		Entry("Master not found",
			fields{
				name:      "nomaster",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			},
			args{
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nomaster",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			false,
			true,
			corev1.ResourceRequirements{},
		),

		Entry("Master not change",
			fields{
				name:      "same",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
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
			},
			args{
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "same-jindofs-master",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{
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
			},
			false,
			false,
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		),
	)
})

var _ = Describe("JindoFSxEngine_syncWorkerSpec", func() {
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
	}
	type args struct {
		ctx    cruntime.ReconcileRequestContext
		worker *appsv1.StatefulSet
	}

	DescribeTable("sync worker spec",
		func(fields fields, args args, wantChanged bool, wantErr bool, wantResource corev1.ResourceRequirements) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, args.worker.DeepCopy())

			s := runtime.NewScheme()
			fields.runtime.SetName(fields.name)
			fields.runtime.SetNamespace(fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, args.worker)
			s.AddKnownTypes(datav1alpha1.GroupVersion, fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, fields.runtime)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoFSxEngine{
				runtime:   fields.runtime,
				name:      fields.name,
				namespace: fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncWorkerSpec(args.ctx, fields.runtime)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(gotChanged).To(Equal(wantChanged))
		},

		Entry("Not resource for jindoruntime",
			fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			},
			args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-worker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			false,
			false,
			corev1.ResourceRequirements{},
		),

		Entry("worker not found",
			fields{
				name:      "noworker",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			},
			args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noworker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			false,
			true,
			corev1.ResourceRequirements{},
		),

		Entry("worker not change",
			fields{
				name:      "same",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
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
			},
			args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "same-jindofs-worker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{
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
			},
			false,
			false,
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		),
	)
})

var _ = Describe("JindoFSxEngine_syncFuseSpec", func() {
	var res resource.Quantity

	BeforeEach(func() {
		res = resource.MustParse("320Gi")
	})

	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
	}
	type args struct {
		ctx  cruntime.ReconcileRequestContext
		fuse *appsv1.DaemonSet
	}

	DescribeTable("sync fuse spec",
		func(fields fields, args args, wantChanged bool, wantErr bool, wantResource corev1.ResourceRequirements) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, args.fuse.DeepCopy())

			s := runtime.NewScheme()
			fields.runtime.SetName(fields.name)
			fields.runtime.SetNamespace(fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, args.fuse)
			s.AddKnownTypes(datav1alpha1.GroupVersion, fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, fields.runtime)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoFSxEngine{
				runtime:   fields.runtime,
				name:      fields.name,
				namespace: fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncFuseSpec(args.ctx, fields.runtime)
			if wantErr {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(gotChanged).To(Equal(wantChanged))
		},

		Entry("Not resource for jindoruntime",
			fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			},
			args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-fuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{},
				},
			},
			false,
			false,
			corev1.ResourceRequirements{},
		),

		Entry("fuse not found",
			fields{
				name:      "nofuse",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			},
			args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nofuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{},
				},
			},
			false,
			true,
			corev1.ResourceRequirements{},
		),

		Entry("fuse not change",
			fields{
				name:      "same",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
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
			},
			args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "same-jindofs-fuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{
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
			},
			false,
			false,
			corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		),
	)
})
