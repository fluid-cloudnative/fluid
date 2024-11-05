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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func TestAlluxioFileUtils_GetConf(t *testing.T) {
	mockExec := func(ctx context.Context, podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		if strings.Contains(cmd[2], OTHER_ERR) {
			return "", "", errors.New("other error")
		} else {
			return "conf", "", nil
		}
	}
	err := gohook.Hook(kubeclient.ExecCommandInContainerWithFullOutput, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainerWithFullOutput)
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
