/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestJuiceFSEngine_getDaemonset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
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
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "runtime1",
						Namespace: "default",
					},
				},
				name:      "runtime1",
				namespace: "default",
			},
			wantDaemonset: &appsv1.DaemonSet{
				ObjectMeta: v1.ObjectMeta{
					Name:      "runtime1",
					Namespace: "default",
				},
				TypeMeta: v1.TypeMeta{
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
			e := &JuiceFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotDaemonset, err := e.getDaemonset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("JuicefsEngine.getDaemonset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDaemonset, tt.wantDaemonset) {
				t.Errorf("JuiceFSEngine.getDaemonset() = %#v, want %#v", gotDaemonset, tt.wantDaemonset)
			}
		})
	}
}

func TestJuiceFSEngine_getFuseDaemonsetName(t *testing.T) {
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
				name: "juicefs",
			},
			wantDsName: "juicefs-fuse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JuiceFSEngine{
				name: tt.fields.name,
			}
			if gotDsName := e.getFuseDaemonsetName(); gotDsName != tt.wantDsName {
				t.Errorf("JuiceFSEngine.getFuseDaemonsetName() = %v, want %v", gotDsName, tt.wantDsName)
			}
		})
	}
}

func TestJuiceFSEngine_getMountPoint(t *testing.T) {
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
				name:      "juicefs",
				namespace: "default",
				Log:       log.NullLogger{},
				MountRoot: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JuiceFSEngine{
				Log:       tt.fields.Log,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			os.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			wantMountPath := fmt.Sprintf("%s/%s/%s/juicefs-fuse", tt.fields.MountRoot+"/juicefs", tt.fields.namespace, e.name)
			if gotMountPath := e.getMountPoint(); gotMountPath != wantMountPath {
				t.Errorf("JuiceFSEngine.getMountPoint() = %v, want %v", gotMountPath, wantMountPath)
			}
		})
	}
}

func TestJuiceFSEngine_getRuntime(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *datav1alpha1.JuiceFSRuntime
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "juicefs",
						Namespace: "default",
					},
				},
				name:      "juicefs",
				namespace: "default",
			},
			want: &datav1alpha1.JuiceFSRuntime{
				TypeMeta: v1.TypeMeta{
					Kind:       "JuiceFSRuntime",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "juicefs",
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
			e := &JuiceFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			got, err := e.getRuntime()
			if (err != nil) != tt.wantErr {
				t.Errorf("JuiceFSEngine.getRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JuiceFSEngine.getRuntime() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestJuiceFSEngine_getSecret(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
		name      string
		namespace string
		Client    client.Client
	}
	tests := []struct {
		name       string
		fields     fields
		wantSecret *corev1.Secret
		wantErr    bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.JuiceFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "runtime",
						Namespace: "default",
					},
				},
				name:      "test",
				namespace: "default",
			},
			wantSecret: &corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				TypeMeta: v1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Secret{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.wantSecret)
			e := &JuiceFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotSecret, err := e.getSecret(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("JuicefsEngine.getSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotSecret, tt.wantSecret) {
				t.Errorf("JuiceFSEngine.getSecret() = %#v, want %#v", gotSecret, tt.wantSecret)
			}
		})
	}
}

func TestJuiceFSEngine_parseFuseImage(t *testing.T) {
	type args struct {
		image           string
		tag             string
		imagePullPolicy string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			name: "test0",
			args: args{
				image:           "juicedata/juicefs-csi-driver",
				tag:             "v0.10.4",
				imagePullPolicy: "IfNotPresent",
			},
			want:  "juicedata/juicefs-csi-driver",
			want1: "v0.10.4",
			want2: "IfNotPresent",
		},
		{
			name: "test1",
			args: args{
				image:           "",
				tag:             "",
				imagePullPolicy: "IfNotPresent",
			},
			want:  "juicedata/juicefs-csi-driver",
			want1: "v0.10.5",
			want2: "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JuiceFSEngine{}
			os.Setenv(common.JuiceFSFuseImageEnv, "juicedata/juicefs-csi-driver:v0.10.5")
			got, got1, got2 := e.parseFuseImage(tt.args.image, tt.args.tag, tt.args.imagePullPolicy)
			if got != tt.want {
				t.Errorf("JuiceFSEngine.parseFuseImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("JuiceFSEngine.parseFuseImage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("JuiceFSEngine.parseFuseImage() got2 = %v, want %v", got2, tt.want2)
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
			wantPath: "/tmp/juicefs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("MOUNT_ROOT", "/tmp")
			if gotPath := getMountRoot(); gotPath != tt.wantPath {
				t.Errorf("getMountRoot() = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}

func TestJuiceFSEngine_parseRuntimeImage(t *testing.T) {
	type args struct {
		image           string
		tag             string
		imagePullPolicy string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			name: "test1",
			args: args{
				image:           "juicedata/juicefs-csi-driver",
				tag:             "v0.10.6",
				imagePullPolicy: "Never",
			},
			want:  "juicedata/juicefs-csi-driver",
			want1: "v0.10.6",
			want2: "Never",
		},
		{
			name: "test2",
			args: args{
				image:           "",
				tag:             "",
				imagePullPolicy: "",
			},
			want:  "juicedata/juicefs-csi-driver",
			want1: "v0.10.5",
			want2: "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			got, got1, got2 := j.parseRuntimeImage(tt.args.image, tt.args.tag, tt.args.imagePullPolicy)
			if got != tt.want {
				t.Errorf("parseRuntimeImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseRuntimeImage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("parseRuntimeImage() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestJuiceFSEngine_parseRuntimeImage1(t *testing.T) {
	type args struct {
		image           string
		tag             string
		imagePullPolicy string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 string
		want2 string
	}{
		{
			name: "test1",
			args: args{
				image:           "juicedata/juicefs-csi-driver",
				tag:             "v0.10.6",
				imagePullPolicy: "Never",
			},
			want:  "juicedata/juicefs-csi-driver",
			want1: "v0.10.6",
			want2: "Never",
		},
		{
			name: "test2",
			args: args{
				image:           "",
				tag:             "",
				imagePullPolicy: "",
			},
			want:  "juicedata/juicefs-csi-driver",
			want1: "v0.10.5",
			want2: "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			got, got1, got2 := j.parseRuntimeImage(tt.args.image, tt.args.tag, tt.args.imagePullPolicy)
			if got != tt.want {
				t.Errorf("parseRuntimeImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("parseRuntimeImage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("parseRuntimeImage() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_parseInt64Size(t *testing.T) {
	type args struct {
		sizeStr string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				sizeStr: "10",
			},
			want:    10,
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				sizeStr: "v",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt64Size(tt.args.sizeStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt64Size() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseInt64Size() got = %v, want %v", got, tt.want)
			}
		})
	}
}
