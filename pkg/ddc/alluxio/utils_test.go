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

package alluxio

import (
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var timestamp_test = time.Now().Format("20060102150405")

func TestIsFluidNativeScheme(t *testing.T) {

	var tests = []struct {
		mountPoint string
		expect     bool
	}{
		{"local:///test",
			true},
		{
			"pvc://test",
			true,
		}, {
			"oss://test",
			false,
		},
	}
	for _, test := range tests {
		engine := &AlluxioEngine{}
		result := engine.isFluidNativeScheme(test.mountPoint)
		if result != test.expect {
			t.Errorf("expect %v for %s, but got %v", test.expect, test.mountPoint, result)
		}
	}
}

func TestAlluxioEngine_getPasswdPath(t *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.AlluxioRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		retryShutdown          int32
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "test",
			fields: fields{runtime: &datav1alpha1.AlluxioRuntime{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       datav1alpha1.AlluxioRuntimeSpec{},
				Status:     datav1alpha1.AlluxioRuntimeStatus{},
			}, name: "test", namespace: "default", runtimeType: "alluxio", Log: log.NullLogger{}},
			want: "/tmp/" + timestamp_test + "_passwd",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				retryShutdown:          tt.fields.retryShutdown,
			}
			if got := e.getPasswdPath(); got != tt.want {
				t.Errorf("AlluxioEngine.getPasswdPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlluxioEngine_getGroupsPath(t *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.AlluxioRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		retryShutdown          int32
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "test",
			fields: fields{runtime: &datav1alpha1.AlluxioRuntime{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       datav1alpha1.AlluxioRuntimeSpec{},
				Status:     datav1alpha1.AlluxioRuntimeStatus{},
			}, name: "test", namespace: "default", runtimeType: "alluxio", Log: log.NullLogger{}},
			want: "/tmp/" + timestamp_test + "_group",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				retryShutdown:          tt.fields.retryShutdown,
			}
			if got := e.getGroupsPath(); got != tt.want {
				t.Errorf("AlluxioEngine.getGroupsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlluxioEngine_getCreateArgs(t *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.AlluxioRuntime
		name                   string
		namespace              string
		runtimeType            string
		Log                    logr.Logger
		Client                 client.Client
		gracefulShutdownLimits int32
		retryShutdown          int32
	}
	f := func(s int64) *int64 {
		return &s
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "test",
			fields: fields{runtime: &datav1alpha1.AlluxioRuntime{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec: datav1alpha1.AlluxioRuntimeSpec{RunAs: &datav1alpha1.User{UID: f(int64(1000)), GID: f(int64(1000)),
					UserName: "test", Groups: []datav1alpha1.Group{
						{ID: int64(1000),
							Name: "a"},
						{ID: int64(2000),
							Name: "b"},
					}}},
				Status: datav1alpha1.AlluxioRuntimeStatus{},
			},
			},
			want: []string{"1000:test:1000", "1000:a", "2000:b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AlluxioEngine{
				runtime:                tt.fields.runtime,
				name:                   tt.fields.name,
				namespace:              tt.fields.namespace,
				runtimeType:            tt.fields.runtimeType,
				Log:                    tt.fields.Log,
				Client:                 tt.fields.Client,
				gracefulShutdownLimits: tt.fields.gracefulShutdownLimits,
				retryShutdown:          tt.fields.retryShutdown,
			}
			got := e.getCreateArgs(tt.fields.runtime)
			var ne bool
			for i, src := range got {
				if src != tt.want[i] {
					ne = false
				}
			}
			if ne {
				t.Errorf("AlluxioEngine.getCreateArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
