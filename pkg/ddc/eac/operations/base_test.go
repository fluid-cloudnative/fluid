/*
  Copyright 2022 The Fluid Authors.

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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	NotExist = "not-exist"
	OtherErr = "other-err"
	FINE     = "fine"
)

func TestNewEACFileUtils(t *testing.T) {
	var expectedResult = EACFileUtils{
		podName:   "efcdemo",
		namespace: "default",
		container: "eac-master",
		log:       fake.NullLogger(),
	}
	result := NewEACFileUtils("efcdemo", "eac-master", "default", fake.NullLogger())
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the EACFileUtils, want: %v, got: %v", expectedResult, result)
	}
}

func TestEACFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a EACFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a EACFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(EACFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(EACFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &EACFileUtils{log: fake.NullLogger()}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(EACFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestEACFileUtils_IsExist(t *testing.T) {
	mockExec := func(a EACFileUtils, p []string, verbose bool) (stdout string, stderr string, e error) {
		if strings.Contains(p[1], NotExist) {
			return "No such file or directory", "", errors.New("No such file or directory")
		} else if strings.Contains(p[1], OtherErr) {
			return "", "", errors.New("other error")
		} else {
			return "", "", nil
		}
	}

	err := gohook.Hook(EACFileUtils.exec, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(EACFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	var tests = []struct {
		in    string
		out   bool
		noErr bool
	}{
		{NotExist, false, true},
		{OtherErr, false, false},
		{FINE, true, true},
	}
	for _, test := range tests {
		found, err := EACFileUtils{log: fake.NullLogger()}.IsExist(test.in)
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

func TestEACFileUtils_DeleteDir(t *testing.T) {
	ExecCommon := func(a EACFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "eac rmr success", "", nil
	}
	ExecErr := func(a EACFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(EACFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(EACFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := EACFileUtils{}
	err = a.DeleteDir("")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(EACFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteDir("")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestEACFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a EACFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "eac mount grep success", "", nil
	}
	ExecErr := func(a EACFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(EACFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(EACFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := EACFileUtils{}
	ready := a.Ready()
	if ready == true {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(EACFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ready = a.Ready()
	if ready == false {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}
