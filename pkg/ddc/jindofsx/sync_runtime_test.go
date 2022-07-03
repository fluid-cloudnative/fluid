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

func TestJindoFSxEngine_syncMasterSpec(t *testing.T) {
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
									corev1.ResourceCPU: resource.MustParse("100m"),
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
												corev1.ResourceCPU: resource.MustParse("100m"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantChanged: true,
			wantErr:     false,
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

			e := &JindoFSxEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncMasterSpec(tt.args.ctx, tt.fields.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoFSxEngine.syncMasterSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("JindoFSxEngine.syncMasterSpec() = %v, want %v. got sts resources %v, want %v",
					gotChanged,
					tt.wantChanged,
					tt.args.master.Spec.Template.Spec.Containers[0].Resources,
					tt.wantResource,
				)
			}
		})
	}
}
