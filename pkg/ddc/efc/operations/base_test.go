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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

const (
	NotExist = "not-exist"
	OtherErr = "other-err"
	FINE     = "fine"
)

func TestNewEFCFileUtils(t *testing.T) {
	var expectedResult = EFCFileUtils{
		podName:   "efcdemo",
		namespace: "default",
		container: "efc-master",
		log:       fake.NullLogger(),
	}
	result := NewEFCFileUtils("efcdemo", "efc-master", "default", fake.NullLogger())
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the EFCFileUtils, want: %v, got: %v", expectedResult, result)
	}
}

func TestEFCFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(EFCFileUtils{}, "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := &EFCFileUtils{log: fake.NullLogger()}
	_, _, err := a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(EFCFileUtils{}, "exec", ExecWithoutTimeoutCommon)

	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

func TestEFCFileUtils_DeleteDir(t *testing.T) {
	ExecCommon := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "efc rmr success", "", nil
	}
	ExecErr := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	patches := gomonkey.ApplyPrivateMethod(EFCFileUtils{}, "exec", ExecErr)
	defer patches.Reset()

	a := EFCFileUtils{}
	err := a.DeleteDir("")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(EFCFileUtils{}, "exec", ExecCommon)

	err = a.DeleteDir("")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

func TestEFCFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "efc mount grep success", "", nil
	}
	ExecErr := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	patches := gomonkey.ApplyPrivateMethod(EFCFileUtils{}, "exec", ExecErr)
	defer patches.Reset()

	a := EFCFileUtils{}
	ready := a.Ready()
	if ready == true {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(EFCFileUtils{}, "exec", ExecCommon)

	ready = a.Ready()
	if ready == false {
		t.Errorf("check failure, want ready")
	}
}
