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
	"strings"
	"testing"

	"github.com/brahma-adshonor/gohook"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

const (
	NOT_EXIST = "not-exist"
	OTHER_ERR = "other-err"
	FINE      = "fine"
)

func TestNewJuiceFSFileUtils(t *testing.T) {
	var expectedResult = JuiceFileUtils{
		podName:   "juicefs",
		namespace: "default",
		container: common.JuiceFSFuseContainer,
		log:       logf.NullLogger{},
	}
	result := NewJuiceFileUtils("juicefs", common.JuiceFSFuseContainer, "default", logf.NullLogger{})
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the JuiceFSFileUtils, want: %v, got: %v", expectedResult, result)
	}
}

func TestJuiceFileUtils_IsExist(t *testing.T) {
	mockExec := func(a JuiceFileUtils, p []string, verbose bool) (stdout string, stderr string, e error) {
		if strings.Contains(p[1], NOT_EXIST) {
			return "No such file or directory", "", errors.New("No such file or directory")
		} else if strings.Contains(p[1], OTHER_ERR) {
			return "", "", errors.New("other error")
		} else {
			return "", "", nil
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	var tests = []struct {
		in    string
		out   bool
		noErr bool
	}{
		{NOT_EXIST, false, true},
		{OTHER_ERR, false, false},
		{FINE, true, true},
	}
	for _, test := range tests {
		found, err := JuiceFileUtils{log: logf.NullLogger{}}.IsExist(test.in)
		if found != test.out {
			t.Errorf("input parameter is %s,expected %t, got %t", test.in, test.out, found)
		}
		var noErr bool = (err == nil)
		if test.noErr != noErr {
			t.Errorf("input parameter is %s, expected noerr is %t, got %t", test.in, test.noErr, err)
		}
	}
	wrappedUnhookExec()
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
		err := gohook.UnHook(JuiceFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: logf.NullLogger{}}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
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
	_, err = a.GetMetric()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := a.GetMetric()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if m != "juicefs metrics success" {
		t.Errorf("expected juicefs metrics success, got %s", m)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_DeleteDir(t *testing.T) {
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
	err = a.DeleteDir("")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteDir("")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
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
		err := gohook.UnHook(JuiceFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{log: logf.NullLogger{}}
	err = a.LoadMetadataWithoutTimeout("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExecWithoutTimeout()

	err = gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
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
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: logf.NullLogger{}}
	_, err = a.Count("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
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
		err := gohook.UnHook(JuiceFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: logf.NullLogger{}}
	_, err = a.GetFileCount("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
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

func TestAlluxioFileUtils_QueryMetaDataInfoIntoFile(t *testing.T) {
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
	a := JuiceFileUtils{log: logf.NullLogger{}}

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
