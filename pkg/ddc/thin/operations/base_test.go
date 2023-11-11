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
	"reflect"
	"testing"

	"github.com/brahma-adshonor/gohook"
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
	wrappedUnhookExecWithoutTimeout := func() {
		err := gohook.UnHook(ThinFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := ThinFileUtils{log: fake.NullLogger()}
	err = a.LoadMetadataWithoutTimeout("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExecWithoutTimeout()

	err = gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetadataWithoutTimeout("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExecWithoutTimeout()
}

func TestThinFileUtils_GetUsedSpace(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "192.168.100.11:/nfs/mnt   87687856128  87687856128            0 100% /runtime-mnt/thin/kube-system/thindemo/thin-fuse", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(ThinFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &ThinFileUtils{log: fake.NullLogger()}
	_, err = a.GetUsedSpace("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	usedSpace, err := a.GetUsedSpace("/tmp")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if usedSpace != 87687856128 {
		t.Errorf("check failure, want 87687856128, got %d", usedSpace)
	}
	wrappedUnhookExec()
}

func TestThinFileUtils_GetFileCount(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "6367897", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(ThinFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &ThinFileUtils{log: fake.NullLogger()}
	_, err = a.GetFileCount("/tmp")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
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

func TestThinFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a ThinFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(ThinFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &ThinFileUtils{log: fake.NullLogger()}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(ThinFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}
