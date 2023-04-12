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

package goosefs

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		result := common.IsFluidNativeScheme(test.mountPoint)
		if result != test.expect {
			t.Errorf("expect %v for %s, but got %v", test.expect, test.mountPoint, result)
		}
	}
}

func TestGooseFSEngine_getInitUserDir(t *testing.T) {
	type fields struct {
		runtime                *datav1alpha1.GooseFSRuntime
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
			fields: fields{runtime: &datav1alpha1.GooseFSRuntime{
				TypeMeta:   v1.TypeMeta{},
				ObjectMeta: v1.ObjectMeta{},
				Spec:       datav1alpha1.GooseFSRuntimeSpec{},
				Status:     datav1alpha1.RuntimeStatus{},
			}, name: "test", namespace: "default", runtimeType: "goosefs", Log: fake.NullLogger()},
			want: fmt.Sprintf("/tmp/fluid/%s/%s", "default", "test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
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
				t.Errorf("GooseFSEngine.getInitUserDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGooseFSEngine_getInitUsersArgs(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.GooseFSRuntime
		Log     logr.Logger
		Client  client.Client
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
				runtime: &datav1alpha1.GooseFSRuntime{
					TypeMeta:   v1.TypeMeta{},
					ObjectMeta: v1.ObjectMeta{},
					Spec: datav1alpha1.GooseFSRuntimeSpec{RunAs: &datav1alpha1.User{UID: f(int64(1000)), GID: f(int64(1000)),
						UserName: "test", GroupName: "a"}},
					Status: datav1alpha1.RuntimeStatus{},
				},
			},
			want: []string{"1000:test:1000", "1000:a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.GetInitUsersArgs(tt.fields.runtime.Spec.RunAs)
			var ne bool
			for i, src := range got {
				if src != tt.want[i] {
					ne = false
				}
			}
			if ne {
				t.Errorf("GooseFSEngine.getInitUsersArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMountRootWithEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/var/lib/mymount/goosefs"},
	}
	for _, tc := range testCases {
		t.Setenv(utils.MountRoot, tc.input)
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
		{"/var/lib/mymount", "/goosefs"},
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

func mockExecCommandInContainerForGetFileCount() (stdout string, stderr string, err error) {
	r := `Master.FilesCompleted  (Type: COUNTER, Value: 1,000)`
	return r, "", nil
}

func mockExecCommandInContainerForWorkerUsedCapacity() (stdout string, stderr string, err error) {
	r := `Capacity information for all workers:
	Total Capacity: 4096.00MB
		Tier: MEM  Size: 4096.00MB
	Used Capacity: 443.89MB
		Tier: MEM  Size: 443.89MB
	Used Percentage: 10%
	Free Percentage: 90%
 
Worker Name      Last Heartbeat   Storage       MEM
192.168.1.147    0                capacity      2048.00MB
								  used          443.89MB (21%)
192.168.1.146    0                capacity      2048.00MB
								  used          0B (0%)`
	return r, "", nil
}

func TestGetDataSetFileNum(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "test0",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			want:    "1000",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}

			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForGetFileCount()
				return stdout, stderr, err
			})
			defer patch1.Reset()

			got, err := e.getDataSetFileNum()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.getDataSetFileNum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GooseFSEngine.getDataSetFileNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRuntime(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *datav1alpha1.GooseFSRuntime
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				name:      "spark",
				namespace: "default",
			},
			want: &datav1alpha1.GooseFSRuntime{
				TypeMeta: v1.TypeMeta{
					Kind:       "GooseFSRuntime",
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
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			got, err := e.getRuntime()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.getRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GooseFSEngine.getRuntime() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestGetMasterStatefulset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
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
				runtime: &datav1alpha1.GooseFSRuntime{
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
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotMaster, err := e.getMasterStatefulset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.getMasterStatefulset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMaster, tt.want) {
				t.Errorf("GooseFSEngine.getMasterStatefulset() = %#v, want %#v", gotMaster, tt.want)
			}
		})
	}
}

func TestGetDaemonset(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
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
				runtime: &datav1alpha1.GooseFSRuntime{
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
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    mockClient,
			}
			gotDaemonset, err := e.getDaemonset(tt.fields.name, tt.fields.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.getDaemonset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDaemonset, tt.wantDaemonset) {
				t.Errorf("GooseFSEngine.getDaemonset() = %#v, want %#v", gotDaemonset, tt.wantDaemonset)
			}
		})
	}
}

func TestGetMasterPodInfo(t *testing.T) {
	type fields struct {
		name string
	}
	tests := []struct {
		name              string
		fields            fields
		wantPodName       string
		wantContainerName string
	}{
		{
			name: "test",
			fields: fields{
				name: "spark",
			},
			wantPodName:       "spark-master-0",
			wantContainerName: "goosefs-master",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name: tt.fields.name,
			}
			gotPodName, gotContainerName := e.getMasterPodInfo()
			if gotPodName != tt.wantPodName {
				t.Errorf("GooseFSEngine.getMasterPodInfo() gotPodName = %v, want %v", gotPodName, tt.wantPodName)
			}
			if gotContainerName != tt.wantContainerName {
				t.Errorf("GooseFSEngine.getMasterPodInfo() gotContainerName = %v, want %v", gotContainerName, tt.wantContainerName)
			}
		})
	}
}

func TestGetMasterStatefulsetName(t *testing.T) {
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
				name: "spark",
			},
			wantDsName: "spark-master",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name: tt.fields.name,
			}
			if gotDsName := e.getMasterName(); gotDsName != tt.wantDsName {
				t.Errorf("GooseFSEngine.getMasterStatefulsetName() = %v, want %v", gotDsName, tt.wantDsName)
			}
		})
	}
}

func TestGetWorkerDaemonsetName(t *testing.T) {
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
				name: "spark",
			},
			wantDsName: "spark-worker",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name: tt.fields.name,
			}
			if gotDsName := e.getWorkerName(); gotDsName != tt.wantDsName {
				t.Errorf("GooseFSEngine.getWorkerDaemonsetName() = %v, want %v", gotDsName, tt.wantDsName)
			}
		})
	}
}

func TestGetFuseDaemonsetName(t *testing.T) {
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
				name: "spark",
			},
			wantDsName: "spark-fuse",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name: tt.fields.name,
			}
			if gotDsName := e.getFuseDaemonsetName(); gotDsName != tt.wantDsName {
				t.Errorf("GooseFSEngine.getFuseDaemonsetName() = %v, want %v", gotDsName, tt.wantDsName)
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
			e := &GooseFSEngine{
				Log:       tt.fields.Log,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			wantMountPath := fmt.Sprintf("%s/%s/%s/goosefs-fuse", tt.fields.MountRoot+"/goosefs", tt.fields.namespace, e.name)
			if gotMountPath := e.getMountPoint(); gotMountPath != wantMountPath {
				t.Errorf("GooseFSEngine.getMountPoint() = %v, want %v", gotMountPath, wantMountPath)
			}
		})
	}
}

func TestGetInitTierPathsEnv(t *testing.T) {
	type fields struct {
		runtime *datav1alpha1.GooseFSRuntime
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				&datav1alpha1.GooseFSRuntime{
					Spec: datav1alpha1.GooseFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									Path: "/mnt/goosefs0",
								},
								{
									Path: "/mnt/goosefs1",
								},
							},
						},
					},
				},
			},
			want: "/mnt/goosefs0:/mnt/goosefs1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime: tt.fields.runtime,
			}
			if got := e.getInitTierPathsEnv(tt.fields.runtime); got != tt.want {
				t.Errorf("GooseFSEngine.getInitTierPathsEnv() = %v, want %v", got, tt.want)
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
			wantPath: "/tmp/goosefs",
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

func TestParseRuntimeImage(t *testing.T) {
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
				image:           "ccr.ccs.tencentyun.com/qcloud/goosefs",
				tag:             "v1.2.0",
				imagePullPolicy: "IfNotPresent",
			},
			want:  "ccr.ccs.tencentyun.com/qcloud/goosefs",
			want1: "v1.2.0",
			want2: "IfNotPresent",
		},
		{
			name: "test0",
			args: args{
				image:           "",
				tag:             "",
				imagePullPolicy: "IfNotPresent",
			},
			want:  "ccr.ccs.tencentyun.com/qcloud/goosefs",
			want1: "v1.2.0",
			want2: "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{}
			t.Setenv(common.GooseFSRuntimeImageEnv, "ccr.ccs.tencentyun.com/qcloud/goosefs:v1.2.0")
			got, got1, got2 := e.parseRuntimeImage(tt.args.image, tt.args.tag, tt.args.imagePullPolicy)
			if got != tt.want {
				t.Errorf("GooseFSEngine.parseRuntimeImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GooseFSEngine.parseRuntimeImage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GooseFSEngine.parseRuntimeImage() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestParseFuseImage(t *testing.T) {
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
				image:           "ccr.ccs.tencentyun.com/qcloud/goosefs-fuse",
				tag:             "v1.2.0",
				imagePullPolicy: "IfNotPresent",
			},
			want:  "ccr.ccs.tencentyun.com/qcloud/goosefs-fuse",
			want1: "v1.2.0",
			want2: "IfNotPresent",
		},
		{
			name: "test0",
			args: args{
				image:           "",
				tag:             "",
				imagePullPolicy: "IfNotPresent",
			},
			want:  "ccr.ccs.tencentyun.com/qcloud/goosefs-fuse",
			want1: "v1.2.0",
			want2: "IfNotPresent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{}
			t.Setenv(common.GooseFSFuseImageEnv, "ccr.ccs.tencentyun.com/qcloud/goosefs-fuse:v1.2.0")
			got, got1, got2 := e.parseFuseImage(tt.args.image, tt.args.tag, tt.args.imagePullPolicy)
			if got != tt.want {
				t.Errorf("GooseFSEngine.parseFuseImage() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GooseFSEngine.parseFuseImage() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("GooseFSEngine.parseFuseImage() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestGetMetadataInfoFile(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				name:      "spark",
				namespace: "default",
			},
			want: fmt.Sprintf("/goosefs_backups/%s-%s.yaml", "spark", "default"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			if got := e.GetMetadataInfoFile(); got != tt.want {
				t.Errorf("GooseFSEngine.GetMetadataInfoFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMetadataFileName(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				name:      "spark",
				namespace: "default",
			},
			want: fmt.Sprintf("metadata-backup-%s-%s.gz", "spark", "default"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			if got := e.GetMetadataFileName(); got != tt.want {
				t.Errorf("GooseFSEngine.GetMetadataFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMetadataInfoFileName(t *testing.T) {
	type fields struct {
		name      string
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				name:      "spark",
				namespace: "default",
			},
			want: fmt.Sprintf("%s-%s.yaml", "spark", "default"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
			}
			if got := e.GetMetadataInfoFileName(); got != tt.want {
				t.Errorf("GooseFSEngine.GetMetadataInfoFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWorkerUsedCapacity(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.GooseFSRuntime
		name      string
		namespace string
		Log       logr.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]int64
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				runtime: &datav1alpha1.GooseFSRuntime{
					ObjectMeta: v1.ObjectMeta{
						Name:      "spark",
						Namespace: "default",
					},
				},
				name:      "spark",
				namespace: "default",
				Log:       fake.NullLogger(),
			},
			want:    map[string]int64{"192.168.1.146": 0, "192.168.1.147": 465452400},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GooseFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}

			patch1 := ApplyFunc(kubeclient.ExecCommandInContainer, func(podName string, containerName string, namespace string, cmd []string) (string, string, error) {
				stdout, stderr, err := mockExecCommandInContainerForWorkerUsedCapacity()
				return stdout, stderr, err
			})
			defer patch1.Reset()
			got, err := e.GetWorkerUsedCapacity()
			if (err != nil) != tt.wantErr {
				t.Errorf("GooseFSEngine.GetWorkerUsedCapacity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GooseFSEngine.GetWorkerUsedCapacity() = %v, want %v", got, tt.want)
			}
		})
	}
}
