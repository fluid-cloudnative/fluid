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
package operations

import (
	"errors"
	"reflect"
	"testing"

	"github.com/brahma-adshonor/gohook"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	NotExist     = "not-exist"
	OtherErr     = "other-err"
	FINE         = "fine"
	CommonStatus = `{
  "Setting": {
    "Name": "zww-juicefs",
    "UUID": "73416457-6f3f-490b-abb6-cbc1f837944e",
    "Storage": "minio",
    "Bucket": "http://10.98.166.242:9000/zww-juicefs",
    "AccessKey": "minioadmin",
    "SecretKey": "removed",
    "BlockSize": 4096,
    "Compression": "none",
    "Shards": 0,
    "HashPrefix": false,
    "Capacity": 0,
    "Inodes": 0,
    "KeyEncrypted": false,
    "TrashDays": 2,
    "MetaVersion": 0,
    "MinClientVersion": "",
    "MaxClientVersion": ""
  },
  "Sessions": [
    {
      "Sid": 14,
      "Expire": "2022-02-09T10:01:50Z",
      "Version": "1.0-dev (2022-02-09 748949ac)",
      "HostName": "juicefs-pvc-33d9bdf3-5fb5-42fe-bf48-d3d6156b424b-createvol2dv4j",
      "MountPoint": "/mnt/jfs",
      "ProcessID": 20
    }
  ]
}`
)

func TestNewJuiceFSFileUtils(t *testing.T) {
	var expectedResult = JuiceFileUtils{
		podName:   "juicefs",
		namespace: "default",
		container: common.JuiceFSFuseContainer,
		log:       fake.NullLogger(),
	}
	result := NewJuiceFileUtils("juicefs", common.JuiceFSFuseContainer, "default", fake.NullLogger())
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the JuiceFSFileUtils, want: %v, got: %v", expectedResult, result)
	}
}

func TestJuiceFileUtils_Mkdir(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs mkdir success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	err = a.Mkdir("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.Mkdir("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: fake.NullLogger()}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_GetMetric(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs metrics success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	_, err = a.GetMetric("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := a.GetMetric("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if m != "juicefs metrics success" {
		t.Errorf("expected juicefs metrics success, got %s", m)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_DeleteCacheDirs(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs rmr success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	err = a.DeleteCacheDirs([]string{"/tmp/raw/chunks"})
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteCacheDirs([]string{"/tmp/raw/chunks"})
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_DeleteCacheDir(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs rmr success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	a := JuiceFileUtils{}
	// no error
	err := gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteCacheDir("/tmp/raw/chunks")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()

	// error
	err = gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteCacheDir("/tmp/raw/chunks")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_GetStatus(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return CommonStatus, "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	err = a.DeleteCacheDir("/tmp/raw/chunks")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	got, err := a.GetStatus("test")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if got != CommonStatus {
		t.Errorf("want %s, got: %v", CommonStatus, got)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_LoadMetadataWithoutTimeout(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Load juicefs metadata", "", nil
	}
	ExecWithoutTimeoutErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExecWithoutTimeout := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{log: fake.NullLogger()}
	err = a.LoadMetadataWithoutTimeout("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExecWithoutTimeout()

	err = gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetadataWithoutTimeout("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExecWithoutTimeout()
}

func TestJuiceFileUtils_Count(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "6367897   /tmp", "", nil
	}
	ExecWithoutTimeoutErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	ExecWithoutTimeoutNegative := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "-9223372036854775808   /tmp", "", nil
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: fake.NullLogger()}
	_, err = a.Count("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutNegative, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.Count("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	fileCount, err := a.Count("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if fileCount != 6367897 {
		t.Errorf("check failure, want 6367897, got %d", fileCount)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_GetFileCount(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "6367897", "", nil
	}
	ExecWithoutTimeoutErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: fake.NullLogger()}
	_, err = a.GetFileCount("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	fileCount, err := a.GetFileCount("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if fileCount != 6367897 {
		t.Errorf("check failure, want 6367897, got %d", fileCount)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_GetUsedSpace(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "JuiceFS:test   87687856128  87687856128            0 100% /runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse", "", nil
	}
	ExecWithoutTimeoutErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: fake.NullLogger()}
	_, err = a.GetUsedSpace("/runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	usedSpace, err := a.GetUsedSpace("/runtime-mnt/juicefs/kube-system/jfsdemo/juicefs-fuse")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if usedSpace != 87687856128 {
		t.Errorf("check failure, want 87687856128, got %d", usedSpace)
	}
	wrappedUnhookExec()
}

func TestJuiceFSFileUtils_QueryMetaDataInfoIntoFile(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "JuiceFS  cluster summary", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{log: fake.NullLogger()}

	keySets := []KeyOfMetaDataFile{DatasetName, Namespace, UfsTotal, FileNum, ""}
	for index, keySet := range keySets {
		_, err = a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err == nil {
			t.Errorf("%d check failure, want err, got nil", index)
			return
		}
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for index, keySet := range keySets {
		_, err = a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}
	}
	wrappedUnhookExec()
}

func TestValidDir(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name      string
		args      args
		wantMatch bool
	}{
		{
			name: "test-normal",
			args: args{
				dir: "/tmp/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test1",
			args: args{
				dir: "/t mp/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test2",
			args: args{
				dir: "/t..mp/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test3",
			args: args{
				dir: "/t__mp/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test4",
			args: args{
				dir: "/t--mp/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test5",
			args: args{
				dir: "/",
			},
			wantMatch: false,
		},
		{
			name: "test6",
			args: args{
				dir: ".",
			},
			wantMatch: false,
		},
		{
			name: "test7",
			args: args{
				dir: "/tttt/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test8",
			args: args{
				dir: "//",
			},
			wantMatch: false,
		},
		{
			name: "test9",
			args: args{
				dir: "/0/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test10",
			args: args{
				dir: "/0/1/raw/chunks",
			},
			wantMatch: true,
		},
		{
			name: "test11",
			args: args{
				dir: "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/0/raw/chunks",
			},
			wantMatch: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotMatch := ValidCacheDir(tt.args.dir); gotMatch != tt.wantMatch {
				t.Errorf("ValidDir() = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}
