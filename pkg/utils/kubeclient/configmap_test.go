/*

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

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsConfigMapExist(t *testing.T) {
	namespace := "default"
	testConfigMapInputs := []*v1.ConfigMap{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1",
			Namespace: namespace},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "test2"},
	}}

	testConfigMaps := []runtime.Object{}

	for _, ns := range testConfigMapInputs {
		testConfigMaps = append(testConfigMaps, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ConfigMap doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want: false,
		},
		{
			name: "ConfigMap exists",
			args: args{
				name:      "test1",
				namespace: namespace,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if want, _ := IsConfigMapExist(client, tt.args.name, tt.args.namespace); want != tt.want {
				t.Errorf("testcase %v IsConfigMapExist()'s expected is %v, result is %v", tt.name, tt.want, want)
			}
		})
	}
}

func TestGetConfigmapByName(t *testing.T) {
	namespace := "default"
	testConfigMapInputs := []*v1.ConfigMap{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1",
			Namespace: namespace},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "test2"},
	}}

	testConfigMaps := []runtime.Object{}

	for _, ns := range testConfigMapInputs {
		testConfigMaps = append(testConfigMaps, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ConfigMap doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want: false,
		},
		{
			name: "ConfigMap exists",
			args: args{
				name:      "test1",
				namespace: namespace,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := GetConfigmapByName(client, tt.args.name, tt.args.namespace); err != nil {
				t.Errorf("testcase %v GetConfigmapByName()'s err is %v", tt.name, err)
			}
		})
	}
}

func TestDeleteConfigMap(t *testing.T) {
	namespace := "default"
	testConfigMapInputs := []*v1.ConfigMap{{
		ObjectMeta: metav1.ObjectMeta{Name: "test1",
			Namespace: namespace},
	}, {
		ObjectMeta: metav1.ObjectMeta{Name: "test2"},
	}}

	testConfigMaps := []runtime.Object{}

	for _, ns := range testConfigMapInputs {
		testConfigMaps = append(testConfigMaps, ns.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testConfigMaps...)

	type args struct {
		name      string
		namespace string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "ConfigMap doesn't exist",
			args: args{
				name:      "notExist",
				namespace: namespace,
			},
			want: false,
		},
		{
			name: "ConfigMap exists",
			args: args{
				name:      "test1",
				namespace: namespace,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteConfigMap(client, tt.args.name, tt.args.namespace); err != nil {
				t.Errorf("testcase %v DeleteConfigMap()'s err is %v", tt.name, err)
			}
		})
	}
}

func TestCopyConfigMap(t *testing.T) {
	type args struct {
		client    client.Client
		src       types.NamespacedName
		dst       types.NamespacedName
		reference metav1.OwnerReference
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "copy success",
			args: args{
				client: fake.NewFakeClient(&v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "src-config",
						Namespace: "src",
					},
					Data: map[string]string{
						"check.sh": "/bin/sh check",
					},
				}),
				src: types.NamespacedName{
					Name:      "src-config",
					Namespace: "src",
				},
				dst: types.NamespacedName{
					Name:      "src-config",
					Namespace: "dst",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyConfigMap(tt.args.client, tt.args.src, tt.args.dst, tt.args.reference); (err != nil) != tt.wantErr {
				t.Errorf("CopyConfigMap() error = %v, wantErr %v", err, tt.wantErr)
				_, err := GetConfigmapByName(tt.args.client, tt.args.dst.Name, tt.args.dst.Namespace)
				if err != nil {
					t.Errorf("Get copyied configmap error: %v", err)
				}
			}
		})
	}
}
