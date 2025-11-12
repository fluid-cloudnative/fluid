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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestAlluxioFileUtils_CachedState(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary: \n    Master Address: 192.168.0.193:20009  \n Used Capacity: 0B\n", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec, ExecErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, err := a.CachedState()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyFunc(AlluxioFileUtils.exec, ExecCommon)
	cached, err := a.CachedState()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if cached != 0 {
		t.Errorf("check failure, want 0, got: %d", cached)
	}
}

func TestAlluxioFIlUtils_CleanCache(t *testing.T) {
	ExecCommonUbuntu := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Ubuntu", "", nil
	}
	ExecCommonAlpine := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alpine", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyFunc(AlluxioFileUtils.exec, ExecErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	err := a.CleanCache("/", 30)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyFunc(AlluxioFileUtils.exec, ExecCommonUbuntu)
	err = a.CleanCache("/", 30)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}

	patches.ApplyFunc(AlluxioFileUtils.exec, ExecCommonAlpine)
	err = a.CleanCache("/", 30)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}
