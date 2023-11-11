/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package jindocache

import (
	"testing"

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

func TestJindoCacheEngine_syncMasterSpec(t *testing.T) {
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
		// runtimeType            string

	}
	type args struct {
		ctx cruntime.ReconcileRequestContext
		// runtime *datav1alpha1.JindoRuntime
		master *appsv1.StatefulSet
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantChanged  bool
		wantErr      bool
		wantResource corev1.ResourceRequirements
	}{
		{
			name: "Not resource for jindoruntime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			}, args: args{
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-master",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     false,
		}, {
			name: "Master not found",
			fields: fields{
				name:      "nomaster",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			}, args: args{
				master: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nomaster",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     true,
		}, {
			name: "Master not change",
			fields: fields{
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
			}, args: args{
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
			wantChanged: false,
			wantErr:     false,
			wantResource: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.args.master.DeepCopy())

			s := runtime.NewScheme()
			// tt.fields.runtime = &datav1alpha1.JindoRuntime{
			// 	ObjectMeta: metav1.ObjectMeta{
			// 		Name:      tt.fields.name,
			// 		Namespace: tt.fields.namespace,
			// 	},
			// }
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.master)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			// runtimeObjs = append(runtimeObjs, tt.args.master)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoCacheEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncMasterSpec(tt.args.ctx, tt.fields.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("Testcase %s JindoCacheEngine.syncMasterSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("Testcase %s JindoCacheEngine.syncMasterSpec() = %v, want %v. got sts resources %v after updated, want %v",
					tt.name,
					gotChanged,
					tt.wantChanged,
					tt.args.master.Spec.Template.Spec.Containers[0].Resources,
					tt.wantResource,
				)
			}

		})
	}
}

func TestJindoCacheEngine_syncWorkerSpec(t *testing.T) {
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
		// runtimeType            string

	}
	type args struct {
		ctx cruntime.ReconcileRequestContext
		// runtime *datav1alpha1.JindoRuntime
		worker *appsv1.StatefulSet
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantChanged  bool
		wantErr      bool
		wantResource corev1.ResourceRequirements
	}{
		{
			name: "Not resource for jindoruntime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			}, args: args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-worker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     false,
		}, {
			name: "worker not found",
			fields: fields{
				name:      "noworker",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			}, args: args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noworker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     true,
		}, {
			name: "worker not change",
			fields: fields{
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
			}, args: args{
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
			wantChanged: false,
			wantErr:     false,
			wantResource: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.args.worker.DeepCopy())

			s := runtime.NewScheme()
			// tt.fields.runtime = &datav1alpha1.JindoRuntime{
			// 	ObjectMeta: metav1.ObjectMeta{
			// 		Name:      tt.fields.name,
			// 		Namespace: tt.fields.namespace,
			// 	},
			// }
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.worker)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			// runtimeObjs = append(runtimeObjs, tt.args.worker)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoCacheEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncWorkerSpec(tt.args.ctx, tt.fields.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("Testcase %s JindoCacheEngine.syncWorkerSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("Testcase %s JindoCacheEngine.syncWorkerSpec() = %v, want %v. got sts resources %v after updated, want %v",
					tt.name,
					gotChanged,
					tt.wantChanged,
					tt.args.worker.Spec.Template.Spec.Containers[0].Resources,
					tt.wantResource,
				)
			}

		})
	}
}

func TestJindoCacheEngine_syncFuseSpec(t *testing.T) {
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
		// runtimeType            string

	}
	type args struct {
		ctx cruntime.ReconcileRequestContext
		// runtime *datav1alpha1.JindoRuntime
		fuse *appsv1.DaemonSet
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantChanged  bool
		wantErr      bool
		wantResource corev1.ResourceRequirements
	}{
		{
			name: "Not resource for jindoruntime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			}, args: args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-jindofs-fuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     false,
		}, {
			name: "fuse not found",
			fields: fields{
				name:      "nofuse",
				namespace: "default",
				runtime:   &datav1alpha1.JindoRuntime{},
			}, args: args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nofuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     true,
		}, {
			name: "fuse not change",
			fields: fields{
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
			}, args: args{
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
			wantChanged: false,
			wantErr:     false,
			wantResource: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.args.fuse.DeepCopy())

			s := runtime.NewScheme()
			// tt.fields.runtime = &datav1alpha1.JindoRuntime{
			// 	ObjectMeta: metav1.ObjectMeta{
			// 		Name:      tt.fields.name,
			// 		Namespace: tt.fields.namespace,
			// 	},
			// }
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.fuse)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			// runtimeObjs = append(runtimeObjs, tt.args.fuse)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JindoCacheEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncFuseSpec(tt.args.ctx, tt.fields.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("testcase %s: JindoCacheEngine.syncFuseSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("testcase %s JindoCacheEngine.syncFuseSpec() = %v, want %v. got sts resources %v after updated, want %v",
					tt.name,
					gotChanged,
					tt.wantChanged,
					tt.args.fuse.Spec.Template.Spec.Containers[0].Resources,
					tt.wantResource,
				)
			}

		})
	}
}
