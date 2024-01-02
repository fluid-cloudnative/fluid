/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
)

func TestGetRuntime(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.VineyardRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *datav1alpha1.VineyardRuntime
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				name:      "spark",
				namespace: "default",
			},
			want: &datav1alpha1.VineyardRuntime{
				TypeMeta: v1.TypeMeta{
					Kind:       "VineyardRuntime",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      "spark",
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
			e := &VineyardEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			got, err := e.getRuntime()
			if (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.getRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("VineyardEngine.getRuntime() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetMasterPod(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.VineyardRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *corev1.Pod
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark-master",
						Namespace: "default",
					},
				},
				name:      "spark-master",
				namespace: "default",
			},
			want: &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Name:      "spark-master",
					Namespace: "default",
				},
				TypeMeta: v1.TypeMeta{
					Kind:       "Pod",
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
			s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Pod{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.want)
			e := &VineyardEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotMaster, err := e.getMasterPod(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.getMasterPod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMaster, tt.want) {
				t.Errorf("VineyardEngine.getMasterPod() = %#v, want %#v", gotMaster, tt.want)
			}
		})
	}
}

func TestGetMasterStatefulset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.VineyardRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *appsv1.StatefulSet
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark-master",
						Namespace: "default",
					},
				},
				name:      "spark-master",
				namespace: "default",
			},
			want: &appsv1.StatefulSet{
				ObjectMeta: v1.ObjectMeta{
					Name:      "spark-master",
					Namespace: "default",
				},
				TypeMeta: v1.TypeMeta{
					Kind:       "StatefulSet",
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
			s.AddKnownTypes(appsv1.SchemeGroupVersion, &appsv1.StatefulSet{})
			_ = corev1.AddToScheme(s)
			mockClient := fake.NewFakeClientWithScheme(s, tt.fields.runtime, tt.want)
			e := &VineyardEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotMaster, err := e.getMasterStatefulset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.getMasterStatefulset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMaster, tt.want) {
				t.Errorf("VineyardEngine.getMasterStatefulset() = %#v, want %#v", gotMaster, tt.want)
			}
		})
	}
}

func TestGetDaemonset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.VineyardRuntime
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
				runtime: &datav1alpha1.VineyardRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark-master",
						Namespace: "default",
					},
				},
				name:      "spark-master",
				namespace: "default",
			},
			wantDaemonset: &appsv1.DaemonSet{
				ObjectMeta: v1.ObjectMeta{
					Name:      "spark-master",
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
			e := &VineyardEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotDaemonset, err := e.getDaemonset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("VineyardEngine.getDaemonset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDaemonset, tt.wantDaemonset) {
				t.Errorf("VineyardEngine.getDaemonset() = %#v, want %#v", gotDaemonset, tt.wantDaemonset)
			}
		})
	}
}

func TestGetMountPoint(t *testing.T) {
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
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &VineyardEngine{
				Log:       tt.fields.Log,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			wantMountPath := fmt.Sprintf("%s/%s/%s/vineyard-fuse", tt.fields.MountRoot+"/vineyard", tt.fields.namespace, e.name)
			if gotMountPath := e.getMountPoint(); gotMountPath != wantMountPath {
				t.Errorf("VineyardEngine.getMountPoint() = %v, want %v", gotMountPath, wantMountPath)
			}
		})
	}
}

func TestGetMountRoot(t *testing.T) {
	tests := []struct {
		name     string
		wantPath string
	}{
		{
			name:     "test",
			wantPath: "/tmp/vineyard",
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

func TestGetWorkerPodExporterPort(t *testing.T) {
	tests := []struct {
		name     string
		engine   *VineyardEngine
		wantPort int32
	}{
		{
			name: "test",
			engine: &VineyardEngine{
				runtime: &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Ports: map[string]int{
								"exporter": 9999,
							},
						},
					},
				},
			},
			wantPort: 9999,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPort := tt.engine.getWorkerPodExporterPort(); gotPort != tt.wantPort {
				t.Errorf("getMountRoot() = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

func TestGetWorkerReplicas(t *testing.T) {
	tests := []struct {
		name         string
		engine       *VineyardEngine
		wantReplicas int32
	}{
		{
			name: "test",
			engine: &VineyardEngine{
				runtime: &datav1alpha1.VineyardRuntime{
					Spec: datav1alpha1.VineyardRuntimeSpec{
						Worker: datav1alpha1.VineyardCompTemplateSpec{
							Replicas: 3,
						},
					},
				},
			},
			wantReplicas: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotReplicas := tt.engine.getWorkerReplicas(); gotReplicas != tt.wantReplicas {
				t.Errorf("getMountRoot() = %v, want %v", gotReplicas, tt.wantReplicas)
			}
		})
	}
}

func TestParseMasterImage(t *testing.T) {
	tests := []struct {
		name            string
		image           string
		tag             string
		imagePullPolicy string
		engine          *VineyardEngine
		wantImage       string
		wantTag         string
		wantPolicy      string
	}{
		{
			name:            "parse master image",
			image:           "test-registry/test-image",
			tag:             "test-tag",
			imagePullPolicy: "IfNotPresent",
			engine:          &VineyardEngine{},
			wantImage:       "test-registry/test-image",
			wantTag:         "test-tag",
			wantPolicy:      "IfNotPresent",
		},
		{
			name:            "parse master image with default image",
			image:           "",
			tag:             "",
			imagePullPolicy: "",
			engine:          &VineyardEngine{},
			wantImage:       "bitnami/etcd",
			wantTag:         "3.5.10",
			wantPolicy:      "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantImage, wantTag, wantPolicy := tt.engine.parseMasterImage(tt.image, tt.tag, tt.imagePullPolicy)
			if wantImage != tt.wantImage || wantTag != tt.wantTag || wantPolicy != tt.wantPolicy {
				t.Errorf("parseMasterImage() = image: %v, tag: %v, policy: %v. want image: %v, tag: %v, policy: %v", wantImage, wantTag, wantPolicy, tt.wantImage, tt.wantTag, tt.wantPolicy)
			}
		})
	}
}

func TestParseWorkerImage(t *testing.T) {
	tests := []struct {
		name            string
		image           string
		tag             string
		imagePullPolicy string
		engine          *VineyardEngine
		wantImage       string
		wantTag         string
		wantPolicy      string
	}{
		{
			name:            "parse worker image",
			image:           "test-registry/test-image",
			tag:             "test-tag",
			imagePullPolicy: "IfNotPresent",
			engine:          &VineyardEngine{},
			wantImage:       "test-registry/test-image",
			wantTag:         "test-tag",
			wantPolicy:      "IfNotPresent",
		},
		{
			name:            "parse worker image with default image",
			image:           "",
			tag:             "",
			imagePullPolicy: "",
			engine:          &VineyardEngine{},
			wantImage:       "vineyardcloudnative/vineyardd",
			wantTag:         "latest",
			wantPolicy:      "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantImage, wantTag, wantPolicy := tt.engine.parseWorkerImage(tt.image, tt.tag, tt.imagePullPolicy)
			if wantImage != tt.wantImage || wantTag != tt.wantTag || wantPolicy != tt.wantPolicy {
				t.Errorf("parseWorkerImage() = image: %v, tag: %v, policy: %v. want image: %v, tag: %v, policy: %v", wantImage, wantTag, wantPolicy, tt.wantImage, tt.wantTag, tt.wantPolicy)
			}
		})
	}
}

func TestParseFuseImage(t *testing.T) {
	tests := []struct {
		name            string
		image           string
		tag             string
		imagePullPolicy string
		engine          *VineyardEngine
		wantImage       string
		wantTag         string
		wantPolicy      string
	}{
		{
			name:            "parse fuse image",
			image:           "test-registry/test-image",
			tag:             "test-tag",
			imagePullPolicy: "IfNotPresent",
			engine:          &VineyardEngine{},
			wantImage:       "test-registry/test-image",
			wantTag:         "test-tag",
			wantPolicy:      "IfNotPresent",
		},
		{
			name:            "parse fuse image with default image",
			image:           "",
			tag:             "",
			imagePullPolicy: "",
			engine:          &VineyardEngine{},
			wantImage:       "vineyardcloudnative/mount-vineyard-socket",
			wantTag:         "latest",
			wantPolicy:      "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantImage, wantTag, wantPolicy := tt.engine.parseFuseImage(tt.image, tt.tag, tt.imagePullPolicy)
			if wantImage != tt.wantImage || wantTag != tt.wantTag || wantPolicy != tt.wantPolicy {
				t.Errorf("parseFuseImage() = image: %v, tag: %v, policy: %v. want image: %v, tag: %v, policy: %v", wantImage, wantTag, wantPolicy, tt.wantImage, tt.wantTag, tt.wantPolicy)
			}
		})
	}
}
