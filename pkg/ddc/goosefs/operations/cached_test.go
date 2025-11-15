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
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestGooseFSFileUtils_CachedState(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary: \n    Master Address: 192.168.0.193:20009  \n Used Capacity: 0B\n", "", nil
	}
	ExecErr := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	patches := gomonkey.ApplyPrivateMethod(GooseFSFileUtils{}, "exec", ExecErr)
	defer patches.Reset()

	a := &GooseFSFileUtils{log: fake.NullLogger()}
	_, err := a.CachedState()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(GooseFSFileUtils{}, "exec", ExecCommon)

	cached, err := a.CachedState()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if cached != 0 {
		t.Errorf("check failure, want 0, got: %d", cached)
	}
}

func TestGooseFSFIlUtils_CleanCache(t *testing.T) {
	ExecCommonUbuntu := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Ubuntu", "", nil
	}
	ExecCommonAlpine := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alpine", "", nil
	}
	ExecCommonCentos := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", fmt.Errorf("unknow release version for linux")
	}
	ExecErr := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(GooseFSFileUtils{}, "exec", ExecErr)
	defer patches.Reset()

	a := &GooseFSFileUtils{log: fake.NullLogger()}
	err := a.CleanCache("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(GooseFSFileUtils{}, "exec", ExecCommonUbuntu)
	err = a.CleanCache("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}

	patches.ApplyPrivateMethod(GooseFSFileUtils{}, "exec", ExecCommonAlpine)
	err = a.CleanCache("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}

	patches.ApplyPrivateMethod(GooseFSFileUtils{}, "exec", ExecCommonCentos)
	err = a.CleanCache("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
}
