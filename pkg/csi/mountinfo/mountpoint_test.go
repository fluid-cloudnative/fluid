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

package mountinfo

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"os"
	"reflect"
	"testing"
)

var (
	peerGroup1      = 475
	peerGroup2      = 476
	mockGlobalMount = &Mount{
		Subtree:        "/",
		MountPath:      "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
		FilesystemType: "fuse.juicefs",
		PeerGroup:      &peerGroup2,
		ReadOnly:       false,
	}
	mockBindMount = &Mount{
		Subtree:        "/",
		MountPath:      "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
		FilesystemType: "fuse.juicefs",
		PeerGroup:      &peerGroup1,
		ReadOnly:       false,
	}
	mockMountPoints = map[string]*Mount{
		"/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse":                                                          mockGlobalMount,
		"/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount": mockBindMount,
	}
)

func Test_getBindMounts(t *testing.T) {
	type args struct {
		mountByPath map[string]*Mount
	}
	tests := []struct {
		name                string
		args                args
		wantBindMountByName map[string][]*Mount
	}{
		{
			name: "test",
			args: args{
				mountByPath: mockMountPoints,
			},
			wantBindMountByName: map[string][]*Mount{
				"default-jfsdemo": {mockBindMount},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBindMountByName := getBindMounts(tt.args.mountByPath); !reflect.DeepEqual(gotBindMountByName, tt.wantBindMountByName) {
				t.Errorf("getBindMounts() = %v, want %v", gotBindMountByName, tt.wantBindMountByName)
			}
		})
	}
}

func Test_getBrokenBindMounts(t *testing.T) {
	type args struct {
		globalMountByName map[string]*Mount
		bindMountByName   map[string][]*Mount
	}
	tests := []struct {
		name             string
		args             args
		wantBrokenMounts []MountPoint
	}{
		{
			name: "test",
			args: args{
				globalMountByName: map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				},
				bindMountByName: map[string][]*Mount{
					"default-jfsdemo": {mockBindMount},
				},
			},
			wantBrokenMounts: []MountPoint{
				{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBrokenMounts := getBrokenBindMounts(tt.args.globalMountByName, tt.args.bindMountByName); !reflect.DeepEqual(gotBrokenMounts, tt.wantBrokenMounts) {
				t.Errorf("getBrokenBindMounts() = %v, want %v", gotBrokenMounts, tt.wantBrokenMounts)
			}
		})
	}
}

func Test_getGlobalMounts(t *testing.T) {
	os.Setenv(utils.MountRoot, "/runtime-mnt")
	type args struct {
		mountByPath map[string]*Mount
	}
	tests := []struct {
		name                  string
		args                  args
		wantGlobalMountByName map[string]*Mount
		wantErr               bool
	}{
		{
			name: "test",
			args: args{
				mountByPath: mockMountPoints,
			},
			wantGlobalMountByName: map[string]*Mount{
				"default-jfsdemo": mockGlobalMount,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGlobalMountByName, err := getGlobalMounts(tt.args.mountByPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGlobalMounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotGlobalMountByName, tt.wantGlobalMountByName) {
				t.Errorf("getGlobalMounts() gotGlobalMountByName = %v, want %v", gotGlobalMountByName, tt.wantGlobalMountByName)
			}
		})
	}
}
