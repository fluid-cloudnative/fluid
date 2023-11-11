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
	"strings"
	"testing"
)

func loadMountInfoFromString(str string) (map[string]*Mount, error) {
	return readMountInfo(strings.NewReader(str))
}

// Test basic loading of a single mountpoint.
func TestLoadMountInfoBasic(t *testing.T) {
	var mountinfo = `
15 0 259:3 / / rw,relatime shared:1 - ext4 /dev/root rw,data=ordered
`
	mountMap, err := loadMountInfoFromString(mountinfo)
	if err != nil {
		t.Fatal("Failed to load mount")
	}
	mnt, ok := mountMap["/"]
	if !ok {
		t.Fatal("wrong path")
	}

	if mnt.MountPath != "/" {
		t.Error("Wrong path")
	}
	if mnt.FilesystemType != "ext4" {
		t.Error("Wrong filesystem type")
	}
	if len(mnt.PeerGroups) != 1 || mnt.PeerGroups[1] != true {
		t.Error("Wrong peer group")
	}
	if mnt.Subtree != "/" {
		t.Error("Wrong subtree")
	}
	if mnt.ReadOnly {
		t.Error("Wrong readonly flag")
	}
}

func TestLoadMountInfoWithNoPeerGroup(t *testing.T) {
	var mountinfo = `
15 0 259:3 / / rw,relatime - ext4 /dev/root rw,data=ordered
`
	mountMap, err := loadMountInfoFromString(mountinfo)
	if err != nil {
		t.Fatal("Failed to load mount")
	}
	if len(mountMap) != 0 {
		t.Fatal("wrong parse")
	}
}

func TestLoadMountInfoMultiMount(t *testing.T) {
	var mountinfo = `
15 0 259:3 / / rw,relatime shared:1 - ext4 /dev/root rw,data=ordered
15 0 259:3 / / rw,relatime shared:2 - ext4 /dev/root rw,data=ordered
`
	mountMap, err := loadMountInfoFromString(mountinfo)
	if err != nil {
		t.Fatal("Failed to load mount")
	}
	mnt, ok := mountMap["/"]
	if !ok {
		t.Fatal("wrong path")
	}

	if mnt.MountPath != "/" {
		t.Error("Wrong path")
	}
	if mnt.FilesystemType != "ext4" {
		t.Error("Wrong filesystem type")
	}
	expectPeerGroup := 2
	if len(mnt.PeerGroups) != 1 || mnt.PeerGroups[expectPeerGroup] != true {
		t.Error("Wrong peer group")
	}
	if mnt.Subtree != "/" {
		t.Error("Wrong subtree")
	}
	if mnt.ReadOnly {
		t.Error("Wrong readonly flag")
	}
	if mnt.Count != 2 {
		t.Error("Wrong count")
	}
}

func Test_parseMountInfoLine(t *testing.T) {
	peerGroup := map[int]bool{475: true}
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want *Mount
	}{
		{
			name: "test",
			args: args{
				line: "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other",
			},
			want: &Mount{
				Subtree:        "/",
				MountPath:      "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
				FilesystemType: "fuse.juicefs",
				PeerGroups:     peerGroup,
				ReadOnly:       true,
				Count:          1,
			},
		},
		{
			name: "peer group nil",
			args: args{
				line: "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other",
			},
			want: &Mount{
				Subtree:        "/",
				MountPath:      "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
				FilesystemType: "fuse.juicefs",
				PeerGroups:     map[int]bool{},
				ReadOnly:       true,
				Count:          1,
			},
		},
		{
			name: "multiple peer group",
			args: args{
				line: "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 master:478 - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other",
			},
			want: &Mount{
				Subtree:        "/",
				MountPath:      "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
				FilesystemType: "fuse.juicefs",
				PeerGroups:     map[int]bool{475: true, 478: true},
				ReadOnly:       true,
				Count:          1,
			},
		},
		{
			name: "peer group err",
			args: args{
				line: "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:abc - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other",
			},
			want: nil,
		},
		{
			name: "len err1",
			args: args{
				line: "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other",
			},
			want: nil,
		},
		{
			name: "len err2",
			args: args{
				line: "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other -",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMountInfoLine(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMountInfoLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peerGroupFromString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name      string
		args      args
		wantPgTag string
		wantPg    int
		wantErr   bool
	}{
		{
			name: "test-shared",
			args: args{
				str: "shared:475",
			},
			wantPgTag: "shared",
			wantPg:    475,
			wantErr:   false,
		},
		{
			name: "test-master",
			args: args{
				str: "master:475",
			},
			wantPgTag: "master",
			wantPg:    475,
			wantErr:   false,
		},
		{
			name: "test-unbindable",
			args: args{
				str: "unbindable",
			},
			wantPgTag: "",
			wantPg:    -1,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgTag, pg, err := peerGroupFromString(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("peerGroupFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pgTag != tt.wantPgTag {
				t.Errorf("peerGroupFromString() got = %v, want %v", pgTag, tt.wantPgTag)
			}
			if pg != tt.wantPg {
				t.Errorf("peerGroupFromString() got1 = %v, want %v", pg, tt.wantPg)
			}
		})
	}
}

func Test_unescapeString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				str: "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
			},
			want: "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
		},
		{
			name: "test-ascii",
			args: args{
				str: "\\123abc",
			},
			want: "Sabc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unescapeString(tt.args.str); got != tt.want {
				t.Errorf("unescapeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
