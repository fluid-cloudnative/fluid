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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"os"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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

func TestAlluxioEngine_getInitUserDir(t *testing.T) {
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
				Status:     datav1alpha1.RuntimeStatus{},
			}, name: "test", namespace: "default", runtimeType: "alluxio", Log: log.NullLogger{}},
			want: fmt.Sprintf("/tmp/fluid/%s/%s", "default", "test"),
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
			if got := e.getInitUserDir(); got != tt.want {
				t.Errorf("AlluxioEngine.getInitUserDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlluxioEngine_getInitUsersArgs(t *testing.T) {
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
			fields: fields{
				runtime: &datav1alpha1.AlluxioRuntime{
					TypeMeta:   v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{},
					Spec: datav1alpha1.AlluxioRuntimeSpec{RunAs: &datav1alpha1.User{UID: f(int64(1000)), GID: f(int64(1000)),
						UserName: "test", GroupName: "a"}},
					Status: datav1alpha1.RuntimeStatus{},
				},
			},
			want: []string{"1000:test:1000", "1000:a"}},
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
			got := e.getInitUsersArgs(tt.fields.runtime)
			var ne bool
			for i, src := range got {
				if src != tt.want[i] {
					ne = false
				}
			}
			if ne {
				t.Errorf("AlluxioEngine.getInitUsersArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountRootWithEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/var/lib/mymount/alluxio"},
	}
	for _, tc := range testCases {
		os.Setenv(utils.MountRoot, tc.input)
		if tc.expected != getMountRoot() {
			t.Errorf("expected %#v, got %#v",
				tc.expected, getMountRoot())
		}
	}
}

func TestMountRootWithoutEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/alluxio"},
	}

	for _, tc := range testCases {
		os.Unsetenv(utils.MountRoot)
		if tc.expected != getMountRoot() {
			t.Errorf("expected %#v, got %#v",
				tc.expected, getMountRoot())
		}
	}
}
func Test_isPortInUsed(t *testing.T) {
	type args struct {
		port      int
		usedPorts []int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "test",
			args: args{
				port:      20000,
				usedPorts: []int{20000, 20001, 20002, 20003, 20004, 20005, 20006, 20007, 20008},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPortInUsed(tt.args.port, tt.args.usedPorts); got != tt.want {
				t.Errorf("isPortInUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lookUpUsedCapacity(t *testing.T) {
	type args struct {
		node            corev1.Node
		usedCapacityMap map[string]int64
	}

	internalIP := "192.168.1.147"
	var usageForInternalIP int64 = 1024

	internalHost := "slave001"
	var usageForInternalHost int64 = 4096

	usedCapacityMap := map[string]int64{}
	usedCapacityMap[internalIP] = usageForInternalIP
	usedCapacityMap[internalHost] = usageForInternalHost

	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "test_lookUpUsedCapacity_ip",
			args: args{
				node: corev1.Node{
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalIP,
								Address: internalIP,
							},
						},
					},
				},
				usedCapacityMap: usedCapacityMap,
			},
			want: usageForInternalIP,
		},
		{
			name: "test_lookUpUsedCapacity_hostname",
			args: args{
				node: corev1.Node{
					Status: corev1.NodeStatus{
						Addresses: []corev1.NodeAddress{
							{
								Type:    corev1.NodeInternalDNS,
								Address: internalHost,
							},
						},
					},
				},
				usedCapacityMap: usedCapacityMap,
			},
			want: usageForInternalHost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lookUpUsedCapacity(tt.args.node, tt.args.usedCapacityMap); got != tt.want {
				t.Errorf("lookUpUsedCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}
