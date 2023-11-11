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

	wrappedUnhookExec := func() {
		err := gohook.UnHook(EFCFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(EFCFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &EFCFileUtils{log: fake.NullLogger()}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(EFCFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestEFCFileUtils_IsExist(t *testing.T) {
	mockExec := func(a EFCFileUtils, p []string, verbose bool) (stdout string, stderr string, e error) {
		if strings.Contains(p[1], NotExist) {
			return "No such file or directory", "", errors.New("No such file or directory")
		} else if strings.Contains(p[1], OtherErr) {
			return "", "", errors.New("other error")
		} else {
			return "", "", nil
		}
	}

	err := gohook.Hook(EFCFileUtils.exec, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(EFCFileUtils.exec)
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
		found, err := EFCFileUtils{log: fake.NullLogger()}.IsExist(test.in)
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

func TestEFCFileUtils_DeleteDir(t *testing.T) {
	ExecCommon := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "efc rmr success", "", nil
	}
	ExecErr := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(EFCFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(EFCFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := EFCFileUtils{}
	err = a.DeleteDir("")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(EFCFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteDir("")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestEFCFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "efc mount grep success", "", nil
	}
	ExecErr := func(a EFCFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(EFCFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(EFCFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := EFCFileUtils{}
	ready := a.Ready()
	if ready == true {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(EFCFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ready = a.Ready()
	if ready == false {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}
