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
