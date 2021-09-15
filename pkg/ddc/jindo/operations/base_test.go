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
	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestNewJindoFileUtils(t *testing.T) {
	expectedResult := JindoFileUtils{
		podName:   "hadoop",
		namespace: "default",
		container: "hadoop",
		log:       log.NullLogger{},
	}

	result := NewJindoFileUtils("hadoop", "default", "hadoop", log.NullLogger{})
	if reflect.DeepEqual(result, expectedResult) {
		t.Errorf("check failure, expected %v, get %v", expectedResult, result)
	}
}

func TestJindoFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stdout", "", nil
	}
	ExecWithoutTimeoutErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(JindoFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JindoFileUtils{log: log.NullLogger{}}
	_, _, err = a.exec([]string{"/sdk/bin/jindo", "jfs", "-report"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JindoFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"/sdk/bin/jindo", "jfs", "-report"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJindoFileUtils_execWithoutTimeout(t *testing.T) {
	mockExecCommon := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "conf", "", nil
	}
	mockExecErr := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "err", "", errors.New("other error")
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JindoFileUtils{log: log.NullLogger{}}
	_, _, err = a.execWithoutTimeout([]string{"/sdk/bin/jindo", "jfs", "-report"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhook()

	err = gohook.Hook(kubeclient.ExecCommandInContainer, mockExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.execWithoutTimeout([]string{"/sdk/bin/jindo", "jfs", "-report"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhook()
}

func TestJindoFileUtils_ReportSummary(t *testing.T) {
	ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stdout", "", nil
	}
	ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JindoFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JindoFileUtils{}
	_, err = a.ReportSummary()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JindoFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportSummary()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJindoFileUtils_GetUfsTotalSize(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "2      1    108 testUrl", "", nil
	}
	ExecWithoutTimeoutErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(JindoFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JindoFileUtils{log: log.NullLogger{}}
	_, err = a.GetUfsTotalSize("/tmpDictionary", false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JindoFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.GetUfsTotalSize("/tmpDictionary", false)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJindoFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stdout ", "", nil
	}
	ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JindoFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JindoFileUtils{log: log.NullLogger{}}
	ready := a.Ready()
	if ready != false {
		t.Errorf("check failure, want false, got %t", ready)
	}
	wrappedUnhookExec()

	err = gohook.Hook(JindoFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ready = a.Ready()
	if ready != true {
		t.Errorf("check failure, want true, got %t", ready)
	}
	wrappedUnhookExec()
}

func TestJindoFileUtils_IsExist(t *testing.T) {
	ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stdout", "", nil
	}
	ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JindoFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JindoFileUtils{}
	_, err = a.IsExist("/data")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JindoFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.IsExist("/data")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJindoFileUtils_LoadMetadataWithoutTimeout(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stdout", "", nil
	}
	ExecWithoutTimeoutErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExecWithoutTimeout := func() {
		err := gohook.UnHook(JindoFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JindoFileUtils{log: log.NullLogger{}}
	err = a.LoadMetadataWithoutTimeout("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExecWithoutTimeout()

	err = gohook.Hook(JindoFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetadataWithoutTimeout("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExecWithoutTimeout()
}
