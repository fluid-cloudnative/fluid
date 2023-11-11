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

package mountinfo

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

var (
	peerGroup1      = map[int]bool{475: true}
	peerGroup2      = map[int]bool{476: true}
	mockGlobalMount = &Mount{
		Subtree:        "/",
		MountPath:      "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
		FilesystemType: "fuse.juicefs",
		PeerGroups:     peerGroup2,
		ReadOnly:       false,
	}
	mockBindMount = &Mount{
		Subtree:        "/",
		MountPath:      "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
		FilesystemType: "fuse.juicefs",
		PeerGroups:     peerGroup1,
		ReadOnly:       false,
	}
	mockBindSubPathMount = &Mount{
		Subtree:        "/",
		MountPath:      "/var/lib/kubelet/pods/6fe8418f-3f78-4adb-9e02-416d8601c1b6/volume-subpaths/default-jfsdemo/demo/0",
		FilesystemType: "fuse.juicefs",
		PeerGroups:     peerGroup1,
		ReadOnly:       false,
	}
	mockMountPoints = map[string]*Mount{
		"/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse":                                                          mockGlobalMount,
		"/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount": mockBindMount,
		"/var/lib/kubelet/pods/6fe8418f-3f78-4adb-9e02-416d8601c1b6/volume-subpaths/default-jfsdemo/demo/0":          mockBindSubPathMount,
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
				"default-jfsdemo": {mockBindSubPathMount, mockBindMount},
			},
		},
		{
			name: "test-bind-nil",
			args: args{
				mountByPath: map[string]*Mount{
					"kubernetes.io~csi/mount": {
						Subtree:        "/",
						MountPath:      "kubernetes.io~csi/mount",
						FilesystemType: "ext4",
						PeerGroups:     nil,
						ReadOnly:       false,
						Count:          0,
					},
				},
			},
			wantBindMountByName: map[string][]*Mount{},
		},
		{
			name: "test-subpath-nil",
			args: args{
				mountByPath: map[string]*Mount{
					"/volume-subpaths/test": {
						Subtree:        "/",
						MountPath:      "/volume-subpaths/test",
						FilesystemType: "ext4",
						PeerGroups:     nil,
						ReadOnly:       false,
						Count:          0,
					},
				},
			},
			wantBindMountByName: map[string][]*Mount{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBindMountByName := getBindMounts(tt.args.mountByPath)
			if len(gotBindMountByName) != len(tt.wantBindMountByName) {
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
					"default-jfsdemo": {mockBindMount, mockBindSubPathMount},
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
				{
					SourcePath:            "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
					MountPath:             "/var/lib/kubelet/pods/6fe8418f-3f78-4adb-9e02-416d8601c1b6/volume-subpaths/default-jfsdemo/demo/0",
					FilesystemType:        "fuse.juicefs",
					ReadOnly:              false,
					Count:                 0,
					NamespacedDatasetName: "default-jfsdemo",
				},
			},
		},
		{
			name: "test-nil",
			args: args{
				globalMountByName: map[string]*Mount{},
				bindMountByName: map[string][]*Mount{
					"default-jfsdemo": {mockBindMount, mockBindSubPathMount},
				},
			},
			wantBrokenMounts: []MountPoint{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBrokenMounts := getBrokenBindMounts(tt.args.globalMountByName, tt.args.bindMountByName); len(gotBrokenMounts) != len(tt.wantBrokenMounts) {
				t.Errorf("getBrokenBindMounts() = %v, want %v", gotBrokenMounts, tt.wantBrokenMounts)
			}
		})
	}
}

func Test_getGlobalMounts(t *testing.T) {
	t.Setenv(utils.MountRoot, "/runtime-mnt")
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
		{
			name: "test-nil",
			args: args{
				mountByPath: map[string]*Mount{
					"/runtime-mnt/test": {
						Subtree:        "/",
						MountPath:      "/runtime-mnt/test",
						FilesystemType: "fuse.juicefs",
						PeerGroups:     peerGroup2,
						ReadOnly:       false,
					},
				},
			},
			wantGlobalMountByName: map[string]*Mount{},
			wantErr:               false,
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
