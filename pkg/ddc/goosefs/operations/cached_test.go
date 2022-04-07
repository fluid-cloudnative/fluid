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

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestGooseFSFileUtils_CachedState(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary: \n    Master Address: 192.168.0.193:20009  \n Used Capacity: 0B\n", "", nil
	}
	ExecErr := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(GooseFSFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(GooseFSFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &GooseFSFileUtils{log: fake.NullLogger()}
	_, err = a.CachedState()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	cached, err := a.CachedState()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if cached != 0 {
		t.Errorf("check failure, want 0, got: %d", cached)
	}
	wrappedUnhookExec()
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
	wrappedUnhookExec := func() {
		err := gohook.UnHook(GooseFSFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(GooseFSFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &GooseFSFileUtils{log: fake.NullLogger()}
	err = a.CleanCache("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommonUbuntu, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.CleanCache("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommonAlpine, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.CleanCache("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommonCentos, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.CleanCache("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()
}
