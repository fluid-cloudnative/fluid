/*
Copyright 2023 The Fluid Author.

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

package utils

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetDataBackup(t *testing.T) {
	mockDataBackupName := "fluid-test-databackup"
	mockDataBackupNamespace := "default"
	initDataBackup := &datav1alpha1.DataBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockDataBackupName,
			Namespace: mockDataBackupNamespace,
		},
	}

	s := runtime.NewScheme()
	s.AddKnownTypes(datav1alpha1.GroupVersion, initDataBackup)

	fakeClient := fake.NewFakeClientWithScheme(s, initDataBackup)

	testCases := map[string]struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		"test get DataBackup case 1": {
			name:      mockDataBackupName,
			namespace: mockDataBackupNamespace,
			wantName:  mockDataBackupName,
			notFound:  false,
		},
		"test get DataBackup case 2": {
			name:      mockDataBackupName + "not-exist",
			namespace: mockDataBackupNamespace,
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range testCases {
		gotDataBackup, err := GetDataBackup(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil && gotDataBackup != nil {
				t.Errorf("%s check failure, want get err, but get nil", k)
			}
		} else {
			if gotDataBackup.Name != item.wantName {
				t.Errorf("%s check failure, want DataLoad name:%s, got DataLoad name:%s", k, item.wantName, gotDataBackup.Name)
			}
		}
	}

}

func TestGetAddressOfMaster(t *testing.T) {
	mockNodeName := "idc1-host2"
	mockIP := "129.23.1.3"
	var mockRpcPort int32 = 34
	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			NodeName: mockNodeName,
			Containers: []corev1.Container{
				{
					Name: "alluxio-master",
					Ports: []corev1.ContainerPort{
						{
							Name:     "rpc",
							HostPort: mockRpcPort,
						},
						{
							Name:     "rpc-test",
							HostPort: 5201,
						},
					},
				},
				{
					Name: "job-master",
					Ports: []corev1.ContainerPort{
						{
							Name:     "rpc",
							HostPort: 5203,
						},
						{
							Name:     "rpc-test",
							HostPort: 5204,
						},
					},
				},
			},
		},
		Status: corev1.PodStatus{
			HostIP: mockIP,
		},
	}

	nodeName, ip, rpcPort := GetAddressOfMaster(pod)

	if nodeName != mockNodeName {
		t.Errorf("nodeName get failure, should be %s, but get %s", mockNodeName, nodeName)
	}
	if ip != mockIP {
		t.Errorf("ip get failure, should be %s, but get %s", mockIP, ip)
	}
	if rpcPort != mockRpcPort {
		t.Errorf("rpcPort get failure, should be %v, but get %v", mockRpcPort, rpcPort)
	}
}

func TestParseBackupRestorePath(t *testing.T) {
	backupRestorePath := "local:///host1/erf"
	pvcName, path, err := ParseBackupRestorePath(backupRestorePath)
	if pvcName != "" {
		t.Errorf("%s parse failure, there is no pvcName, but get %s", backupRestorePath, pvcName)
	}
	if path != "/host1/erf/" {
		t.Errorf("%s parse failure, path should be /host1/erf/, but get %s", backupRestorePath, path)
	}
	if err != nil {
		t.Errorf("%s parse failure, err should be nil , but get %s", backupRestorePath, err)
	}

	backupRestorePath = ""
	_, _, err = ParseBackupRestorePath(backupRestorePath)
	if err == nil {
		t.Errorf("%s parse failure, err should not be nil", backupRestorePath)
	}

	backupRestorePath = "nfs://132.252.183.2/yum"
	_, _, err = ParseBackupRestorePath(backupRestorePath)
	if err == nil {
		t.Errorf("%s parse failure, err should not be nil", backupRestorePath)
	}

	backupRestorePath = "pvc://pvc1/erf"
	pvcName, path, err = ParseBackupRestorePath(backupRestorePath)
	if pvcName != "pvc1" {
		t.Errorf("%s parse failure, pvcName should be pvc1, but get %s", backupRestorePath, pvcName)
	}
	if path != "/erf/" {
		t.Errorf("%s parse failure, path should be /erf/, but get %s", backupRestorePath, path)
	}
	if err != nil {
		t.Errorf("%s parse failure, err should be nil , but get %s", backupRestorePath, err)
	}

}

func TestGetBackupUserDir(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      string
	}{
		{
			name:      "test1",
			namespace: "ns1",
			want:      "/tmp/backupuser/ns1/test1",
		},
		{
			name:      "test2",
			namespace: "ns2",
			want:      "/tmp/backupuser/ns2/test2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBackupUserDir(tt.namespace, tt.name); got != tt.want {
				t.Errorf("GetBackupUserDir = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRpcPortFromMasterContainer(t *testing.T) {
	type args struct {
		container *corev1.Container
	}
	tests := []struct {
		name        string
		args        args
		wantRpcPort int32
	}{
		{
			name: "alluxio-test",
			args: args{
				container: &corev1.Container{
					Name: "alluxio-master",
					Ports: []corev1.ContainerPort{
						{
							Name:     "rpc",
							HostPort: 34,
						},
						{
							Name:     "rpc-test",
							HostPort: 5201,
						},
					},
				},
			},
			wantRpcPort: 34,
		},
		{
			name: "goosefs-test",
			args: args{
				container: &corev1.Container{
					Name: "goosefs-master",
					Ports: []corev1.ContainerPort{
						{
							Name:     "rpc",
							HostPort: 44,
						},
						{
							Name:     "rpc-test",
							HostPort: 5202,
						},
					},
				},
			},
			wantRpcPort: 44,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRpcPort := GetRpcPortFromMasterContainer(tt.args.container); gotRpcPort != tt.wantRpcPort {
				t.Errorf("GetRpcPortFromMasterContainer() = %v, want %v", gotRpcPort, tt.wantRpcPort)
			}
		})
	}
}

func TestGetDataBackupReleaseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				name: "test",
			},
			want: "test-charts",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataBackupReleaseName(tt.args.name); got != tt.want {
				t.Errorf("GetDataBackupReleaseName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDataBackupPodName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				name: "test",
			},
			want: "test-pod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDataBackupPodName(tt.args.name); got != tt.want {
				t.Errorf("GetDataBackupPodName() = %v, want %v", got, tt.want)
			}
		})
	}
}
