package operations

import (
	"errors"
	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"strings"
	"testing"
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
		stdout, err := AlluxioFileUtils{log: NullLogger{}}.GetConf(test.in)
		if stdout != test.out {
			t.Errorf("input parameter is %s,expected %s, got %s", test.in, test.out, stdout)
		}
		noerror := err == nil
		if noerror != test.noErr {
			t.Errorf("input parameter is %s,expected noerr is %t", test.in, test.noErr)
		}
	}
}
