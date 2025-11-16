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

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestNewThinFileUtils(t *testing.T) {
	var expectedResult = ThinFileUtils{
		podName:   "thin",
		namespace: "default",
		container: common.ThinFuseContainer,
		log:       fake.NullLogger(),
	}
	result := NewThinFileUtils("thin", common.ThinFuseContainer, "default", fake.NullLogger())
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the ThinFileUtils, want: %v, got: %v", expectedResult, result)
	}
}

func TestThinFileUtils_LoadMetadataWithoutTimeout(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Load thin metadata", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	patches := ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := ThinFileUtils{log: fake.NullLogger()}
	err := a.LoadMetadataWithoutTimeout("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutCommon)

	err = a.LoadMetadataWithoutTimeout("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

func TestThinFileUtils_GetUsedSpace(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "192.168.100.11:/nfs/mnt   87687856128  87687856128            0 100% /runtime-mnt/thin/kube-system/thindemo/thin-fuse", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	patches := ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := &ThinFileUtils{log: fake.NullLogger()}
	_, err := a.GetUsedSpace("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutCommon)

	usedSpace, err := a.GetUsedSpace("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if usedSpace != 87687856128 {
		t.Errorf("check failure, want 87687856128, got %d", usedSpace)
	}
}

func TestThinFileUtils_GetFileCount(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "6367897", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	patches := ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := &ThinFileUtils{log: fake.NullLogger()}
	_, err := a.GetFileCount("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutCommon)

	fileCount, err := a.GetFileCount("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if fileCount != 6367897 {
		t.Errorf("check failure, want 6367897, got %d", fileCount)
	}
}

func TestThinFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := &ThinFileUtils{log: fake.NullLogger()}
	_, _, err := a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(ThinFileUtils{}), "exec", ExecWithoutTimeoutCommon)

	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}
