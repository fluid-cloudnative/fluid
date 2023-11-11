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

package utils

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"testing"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/mount"
)

func TestMountRootWithEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/var/lib/mymount"},
	}
	for _, tc := range testCases {
		t.Setenv(MountRoot, tc.input)
		mountRoot, err := GetMountRoot()
		if err != nil {
			t.Errorf("Get error %v", err)
		}
		if tc.expected != mountRoot {
			t.Errorf("expected %#v, got %#v",
				tc.expected, mountRoot)
		}
	}
}

func TestMountRootWithoutEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", ""},
	}
	for _, tc := range testCases {
		os.Unsetenv(MountRoot)
		mountRoot, err := GetMountRoot()
		if err == nil {
			t.Errorf("Expected error happened, but no error")
		}

		if err.Error() != "the the value of the env variable named MOUNT_ROOT is illegal" {
			t.Errorf("Get unexpected error %v", err)
		}

		if tc.expected != mountRoot {
			t.Errorf("Unexpected result %s", tc.expected)
		}

	}
}

func TestCheckMountReady(t *testing.T) {
	Convey("TestCheckMountReady", t, func() {
		Convey("CheckMountReady success", func() {
			cmd := &exec.Cmd{}
			patch1 := ApplyMethod(reflect.TypeOf(cmd), "CombinedOutput", func(_ *exec.Cmd) ([]byte, error) {
				return nil, nil
			})
			defer patch1.Reset()

			err := CheckMountReadyAndSubPathExist("/test", "test", "")
			So(err, ShouldBeNil)
		})
		Convey("CheckMountReady false", func() {
			cmd := &exec.Cmd{}
			patch1 := ApplyMethod(reflect.TypeOf(cmd), "CombinedOutput", func(_ *exec.Cmd) ([]byte, error) {
				return nil, errors.New("test")
			})
			defer patch1.Reset()

			err := CheckMountReadyAndSubPathExist("/test", "test", "")
			So(err, ShouldNotBeNil)
		})
		Convey("fluidPath nil", func() {
			err := CheckMountReadyAndSubPathExist("", "test", "")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestIsMounted(t *testing.T) {
	Convey("TestIsMounted", t, func() {
		Convey("IsMounted success", func() {
			patch2 := ApplyFunc(os.Stat, func(filename string) (os.FileInfo, error) {
				return nil, nil
			})
			defer patch2.Reset()
			patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio /var/lib/kubelet/pods/4781fc5b-72f9-4175-9321-2e1f169880ce/volumes/kubernetes.io~csi/default-jfsdemo/mount fuse.juicefs rw,relatime,user_id=0,group_id=0,default_permissions,allow_other 0 0"), nil
			})
			defer patch1.Reset()
			absPath := "/var/lib/kubelet/pods/4781fc5b-72f9-4175-9321-2e1f169880ce/volumes/kubernetes.io~csi/default-jfsdemo/mount"

			mounted, err := IsMounted(absPath)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeTrue)
		})
		Convey("IsMounted false", func() {
			patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio /var/lib/kubelet/pods/4781fc5b-72f9-4175-9321-2e1f169880ce/volumes/kubernetes.io~csi/default-jfsdemo/mount fuse.juicefs rw,relatime,user_id=0,group_id=0,default_permissions,allow_other 0 0"), nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(os.Stat, func(filename string) (os.FileInfo, error) {
				return nil, nil
			})
			defer patch2.Reset()
			absPath := "/test"

			mounted, err := IsMounted(absPath)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeFalse)
		})
		Convey("token len is 1", func() {
			patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio "), nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(os.Stat, func(filename string) (os.FileInfo, error) {
				return nil, nil
			})
			defer patch2.Reset()
			absPath := "/test"

			mounted, err := IsMounted(absPath)
			So(err, ShouldBeNil)
			So(mounted, ShouldBeFalse)
		})
		Convey("IsMounted error", func() {
			patch1 := ApplyFunc(os.ReadFile, func(filename string) ([]byte, error) {
				return []byte("JuiceFS:minio"), errors.New("test")
			})
			defer patch1.Reset()
			absPath := "/test"

			mounted, err := IsMounted(absPath)
			So(err, ShouldNotBeNil)
			So(mounted, ShouldBeFalse)
		})
	})
}

func TestCheckMountPointBroken(t *testing.T) {
	Convey("TestCheckMountPointBroken", t, func() {
		Convey("CheckMountPointBroken success", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return true, errors.New("test")
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mount.IsCorruptedMnt, func(err error) bool {
				return true
			})
			defer patch2.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldBeNil)
			So(broken, ShouldBeTrue)
		})
		Convey("CheckMountPointBroken not broken", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return true, nil
			})
			defer patch1.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldBeNil)
			So(broken, ShouldBeFalse)
		})
		Convey("CheckMountPointBroken not exist", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return false, nil
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mount.IsCorruptedMnt, func(err error) bool {
				return false
			})
			defer patch2.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldNotBeNil)
			So(broken, ShouldBeFalse)
		})
		Convey("CheckMountPointBroken error", func() {
			patch1 := ApplyFunc(mount.PathExists, func(path string) (bool, error) {
				return false, errors.New("test")
			})
			defer patch1.Reset()
			patch2 := ApplyFunc(mount.IsCorruptedMnt, func(err error) bool {
				return false
			})
			defer patch2.Reset()
			broken, err := CheckMountPointBroken("/test")
			So(err, ShouldNotBeNil)
			So(broken, ShouldBeFalse)
		})
		Convey("CheckMountPointBroken nil", func() {
			broken, err := CheckMountPointBroken("")
			So(err, ShouldNotBeNil)
			So(broken, ShouldBeFalse)
		})
	})
}

func TestGetRuntimeNameFromFusePod(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name            string
		args            args
		wantRuntimeName string
		wantErr         bool
	}{
		{
			name: "test-right",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test-fuse-123"},
				},
			},
			wantRuntimeName: "test",
			wantErr:         false,
		},
		{
			name: "test-error",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
				},
			},
			wantRuntimeName: "",
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRuntimeName, err := GetRuntimeNameFromFusePod(tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRuntimeNameFromFusePod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRuntimeName != tt.wantRuntimeName {
				t.Errorf("GetRuntimeNameFromFusePod() gotRuntimeName = %v, want %v", gotRuntimeName, tt.wantRuntimeName)
			}
		})
	}
}

func TestIsFusePod(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test-true",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"role": "juicefs-fuse"}},
				},
			},
			want: true,
		},
		{
			name: "test-false",
			args: args{
				pod: v1.Pod{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFusePod(tt.args.pod); got != tt.want {
				t.Errorf("IsFusePod() = %v, want %v", got, tt.want)
			}
		})
	}
}
