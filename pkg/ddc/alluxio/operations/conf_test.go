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
	"strings"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func TestAlluxioFileUtils_GetConf(t *testing.T) {
	mockExec := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		if strings.Contains(cmd[2], OTHER_ERR) {
			return "", "", errors.New("other error")
		} else {
			return "conf", "", nil
		}
	}
	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	defer wrappedUnhook()

	var tests = []struct {
		in    string
		out   string
		noErr bool
	}{
		{in: OTHER_ERR, out: "", noErr: false},
		{in: FINE, out: "conf", noErr: true},
	}

	for _, test := range tests {
		stdout, err := AlluxioFileUtils{log: fake.NullLogger()}.GetConf(test.in)
		if stdout != test.out {
			t.Errorf("input parameter is %s,expected %s, got %s", test.in, test.out, stdout)
		}
		noerror := err == nil
		if noerror != test.noErr {
			t.Errorf("input parameter is %s,expected noerr is %t", test.in, test.noErr)
		}
	}
}
