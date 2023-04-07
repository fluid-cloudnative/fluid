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

package cachefs

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestCacheFSxEngine_syncWorkerSpec(t *testing.T) {
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.CacheFSRuntime
		name      string
		namespace string
	}
	type args struct {
		ctx    cruntime.ReconcileRequestContext
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
			name: "Not resource for juicefs runtime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.CacheFSRuntime{},
			}, args: args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-worker",
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
				runtime:   &datav1alpha1.CacheFSRuntime{},
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
				runtime: &datav1alpha1.CacheFSRuntime{
					Spec: datav1alpha1.CacheFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: common.Memory,
									Quota:      &res,
								},
							},
						},
						Worker: datav1alpha1.CacheFSCompTemplateSpec{
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
						Name:      "same-worker",
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
			//runtimeObjs = append(runtimeObjs, tt.args.worker.DeepCopy())

			s := runtime.NewScheme()
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.worker)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			runtimeObjs = append(runtimeObjs, tt.args.worker)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &CacheFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncWorkerSpec(tt.args.ctx, tt.fields.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("Testcase %s JuiceFSEngine.syncWorkerSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("Testcase %s JuiceFSEngine.syncWorkerSpec() = %v, want %v. got sts resources %v after updated, want %v",
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

func TestCacheFSxEngine_syncFuseSpec(t *testing.T) {
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.CacheFSRuntime
		name      string
		namespace string
	}
	type args struct {
		ctx  cruntime.ReconcileRequestContext
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
			name: "Not resource for cachefs runtime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.CacheFSRuntime{},
			}, args: args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-fuse",
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
				runtime:   &datav1alpha1.CacheFSRuntime{},
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
				runtime: &datav1alpha1.CacheFSRuntime{
					Spec: datav1alpha1.CacheFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: common.Memory,
									Quota:      &res,
								},
							},
						},
						Fuse: datav1alpha1.CacheFSFuseSpec{
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
						Name:      "same-fuse",
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

			s := runtime.NewScheme()
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.fuse)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			runtimeObjs = append(runtimeObjs, tt.args.fuse)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &CacheFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncFuseSpec(tt.args.ctx, tt.fields.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("testcase %s: JuiceFSEngine.syncFuseSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("testcase %s JuiceFSEngine.syncFuseSpec() = %v, want %v. got sts resources %v after updated, want %v",
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
