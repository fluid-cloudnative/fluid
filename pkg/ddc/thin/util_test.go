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

package thin

import (
	"fmt"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestThinEngine_getRuntime(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *datav1alpha1.ThinRuntime
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "thin",
						Namespace: "default",
					},
				},
				name:      "thin",
				namespace: "default",
			},
			want: &datav1alpha1.ThinRuntime{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ThinRuntime",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "thin",
					Namespace: "default",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.want)
			e := &ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			got, err := e.getRuntime()
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.getRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ThinEngine.getRuntime() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestThinEngine_getFuseDaemonsetName(t *testing.T) {
	type fields struct {
		name string
	}
	tests := []struct {
		name       string
		fields     fields
		wantDsName string
	}{
		{
			name: "test",
			fields: fields{
				name: "Thin",
			},
			wantDsName: "Thin-fuse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ThinEngine{
				name: tt.fields.name,
			}
			if gotDsName := e.getFuseDaemonsetName(); gotDsName != tt.wantDsName {
				t.Errorf("ThinEngine.getFuseDaemonsetName() = %v, want %v", gotDsName, tt.wantDsName)
			}
		})
	}
}

func TestThinEngine_getDaemonset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.ThinRuntime
		name      string
		namespace string
		Client    client.Client
	}
	tests := []struct {
		name          string
		fields        fields
		wantDaemonset *appsv1.DaemonSet
		wantErr       bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "runtime1",
						Namespace: "default",
					},
				},
				name:      "runtime1",
				namespace: "default",
			},
			wantDaemonset: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runtime1",
					Namespace: "default",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "DaemonSet",
					APIVersion: "apps/v1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.DaemonSet{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.wantDaemonset)
			e := &ThinEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotDaemonset, err := e.getDaemonset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ThinEngine.getDaemonset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDaemonset, tt.wantDaemonset) {
				t.Errorf("ThinEngine.getDaemonset() = %#v, want %#v", gotDaemonset, tt.wantDaemonset)
			}
		})
	}
}

func TestThinEngine_getMountPoint(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
		MountRoot string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test",
			fields: fields{
				name:      "Thin",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ThinEngine{
				Log:       tt.fields.Log,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			wantMountPath := fmt.Sprintf("%s/%s/%s/thin-fuse", tt.fields.MountRoot+"/thin", tt.fields.namespace, e.name)
			if gotMountPath := e.getTargetPath(); gotMountPath != wantMountPath {
				t.Errorf("ThinEngine.getTargetPath() = %v, want %v", gotMountPath, wantMountPath)
			}
		})
	}
}

func Test_getMountRoot(t *testing.T) {
	tests := []struct {
		name     string
		wantPath string
	}{
		{
			name:     "test",
			wantPath: "/tmp/thin",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MOUNT_ROOT", "/tmp")
			if gotPath := getMountRoot(); gotPath != tt.wantPath {
				t.Errorf("getMountRoot() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}
