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

package kubeclient

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	utilpointer "k8s.io/utils/pointer"
)

func TestCompareOwnerRefMatcheWithExpected(t *testing.T) {
	type fields struct {
		controller *appsv1.StatefulSet
		child      runtime.Object
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "NoController",
			fields: fields{
				controller: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "big-data",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				child: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1-0",
						Namespace: "big-data",
					},
					Spec: v1.PodSpec{},
				},
			},
		}, {name: "the_controller_uid_is_not_matched",
			fields: fields{
				controller: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "big-data",
						UID:       "uid",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				child: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-0",
						Namespace: "big-data",
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "app/v1",
							UID:        "uid1",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Spec: v1.PodSpec{},
				},
			},
			want: false,
		},
		{name: "ControllerEqual",
			fields: fields{
				controller: &appsv1.StatefulSet{
					TypeMeta: metav1.TypeMeta{
						Kind:       "StatefulSet",
						APIVersion: "app/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "big-data",
						UID:       "uid2",
					},
					Spec: appsv1.StatefulSetSpec{},
				},
				child: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2-0",
						Namespace: "big-data",
						OwnerReferences: []metav1.OwnerReference{{
							Kind:       "StatefulSet",
							APIVersion: "app/v1",
							UID:        "uid2",
							Name:       "test2",
							Controller: utilpointer.BoolPtr(true),
						}},
					},
					Spec: v1.PodSpec{},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.fields.controller)
			_ = v1.AddToScheme(s)
			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.fields.controller)
			runtimeObjs = append(runtimeObjs, tt.fields.child)
			mockClient := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			metaObj, err := meta.Accessor(tt.fields.child)
			if err != nil {
				t.Errorf(" meta.Accessor = %v", err)
			}
			controllerRef := metav1.GetControllerOf(metaObj)
			want, err := compareOwnerRefMatcheWithExpected(mockClient, controllerRef, metaObj.GetNamespace(), tt.fields.controller)
			if err != nil {
				t.Errorf("compareOwnerRefMatcheWithExpected = %v", err)
			}

			if want != tt.want {
				t.Errorf("test case %s compareOwnerRefMatcheWithExpected() = %v, want %v",
					tt.name,
					want,
					tt.want)
			}
		})
	}
}
