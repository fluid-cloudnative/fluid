/*

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
	"strings"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	NOT_EXIST      = "not-exist"
	OTHER_ERR      = "other-err"
	FINE           = "fine"
	EXEC_ERR       = "exec-err"
	TOO_MANY_LINES = "too many lines"
	DATA_NUM       = "data nums not match"
	PARSE_ERR      = "parse err"
)

// a empty logger just for testing ...
type NullLogger struct{}

func (_ NullLogger) Info(_ string, _ ...interface{}) {
	// Do nothing.
}

func TestLoadMetaData(t *testing.T) {
	var tests = []struct {
		path string
		sync bool
		err  error
	}{
		{"/", true, nil},
	}
	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	for _, test := range tests {
		tools := NewAlluxioFileUtils("", "", "", ctrl.Log, []string{})
		err := tools.LoadMetaData(test.path, test.sync)
		// fmt.Println(expectedErr)
		if err == nil {
			t.Errorf("expected %v, got %v", test.path, tools)
		}
	}
}

func (_ NullLogger) Enabled() bool {
	return false
}

func (_ NullLogger) Error(_ error, _ string, _ ...interface{}) {
	// Do nothing.
}

func (log NullLogger) V(_ int) logr.InfoLogger {
	return log
}

func (log NullLogger) WithName(_ string) logr.Logger {
	return log
}

func (log NullLogger) WithValues(_ ...interface{}) logr.Logger {
	return log
}

//imeplement nulllogger to bypass go vet check

func TestAlluxioFileUtils_IsExist(t *testing.T) {

	mockExec := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {

		if strings.Contains(p4[3], NOT_EXIST) {
			return "does not exist", "", errors.New("does not exist")

		} else if strings.Contains(p4[3], OTHER_ERR) {
			return "", "", errors.New("other error")
		} else {
			return "", "", nil
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
		out   bool
		noErr bool
	}{
		{NOT_EXIST, false, true},
		{OTHER_ERR, false, false},
		{FINE, true, true},
	}
	for _, test := range tests {
		found, err := AlluxioFileUtils{log: NullLogger{}}.IsExist(test.in)
		if found != test.out {
			t.Errorf("input parameter is %s,expected %t, got %t", test.in, test.out, found)
		}
		var noErr bool = (err == nil)
		if test.noErr != noErr {
			t.Errorf("input parameter is %s,expected noerr is %t", test.in, test.noErr)
		}
	}
}

func TestAlluxioFileUtils_Du(t *testing.T) {
	out1, out2, out3 := 111, 222, "%233"
	mockExec := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {

		if strings.Contains(p4[4], EXEC_ERR) {
			return "does not exist", "", errors.New("exec-error")
		} else if strings.Contains(p4[4], TOO_MANY_LINES) {
			return "1\n2\n3\n4\n", "1\n2\n3\n4\n", nil
		} else if strings.Contains(p4[4], DATA_NUM) {
			return "1\n2\t3", "1\n2\t3", nil
		} else if strings.Contains(p4[4], PARSE_ERR) {
			return "1\n1\tdududu\tbbb\t", "1\n1\t2\tbbb\t", nil
		} else {
			return fmt.Sprintf("first line!\n%d\t%d\t(%s)\t2333", out1, out2, out3), "", nil
		}
	}

	err := gohook.HookByIndirectJmp(kubeclient.ExecCommandInContainer, mockExec, nil)
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
		in         string
		out1, out2 int64
		out3       string
		noErr      bool
	}{
		{EXEC_ERR, 0, 0, "", false},
		{TOO_MANY_LINES, 0, 0, "", false},
		{DATA_NUM, 0, 0, "", false},
		{PARSE_ERR, 0, 0, "", false},
		{FINE, int64(out1), int64(out2), out3, true},
	}
	for _, test := range tests {
		o1, o2, o3, err := AlluxioFileUtils{log: NullLogger{}}.Du(test.in)
		var noErr bool = (err == nil)
		if test.noErr != noErr {
			t.Errorf("input parameter is %s,expected noerr is %t", test.in, test.noErr)
		}
		if test.noErr {
			if o1 != test.out1 || o2 != test.out2 || o3 != test.out3 {
				t.Fatalf("input parameter is %s,output is %d,%d, %s", test.in, o1, o2, o3)
			}
		}
	}
}

func TestAlluxioFileUtils_Count(t *testing.T) {
	out1, out2, out3 := 111, 222, 333
	mockExec := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {

		if strings.Contains(p4[3], EXEC_ERR) {
			return "does not exist", "", errors.New("exec-error")
		} else if strings.Contains(p4[3], TOO_MANY_LINES) {
			return "1\n2\n3\n4\n", "1\n2\n3\n4\n", nil
		} else if strings.Contains(p4[3], DATA_NUM) {
			return "1\n2\t3", "1\n2\t3", nil
		} else if strings.Contains(p4[3], PARSE_ERR) {
			return "1\n1\tdududu\tbbb\t", "1\n1\t2\tbbb\t", nil
		} else {
			return fmt.Sprintf("first line!\n%d\t%d\t%d", out1, out2, out3), "", nil
		}
	}

	err := gohook.HookByIndirectJmp(kubeclient.ExecCommandInContainer, mockExec, nil)
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	defer wrappedUnhook()
	if err != nil {
		t.Fatal(err.Error())
	}
	var tests = []struct {
		in               string
		out1, out2, out3 int64
		noErr            bool
	}{
		{EXEC_ERR, 0, 0, 0, false},
		{TOO_MANY_LINES, 0, 0, 0, false},
		{DATA_NUM, 0, 0, 0, false},
		{PARSE_ERR, 0, 0, 0, false},
		{FINE, int64(out1), int64(out2), int64(out3), true},
	}
	for _, test := range tests {
		o1, o2, o3, err := AlluxioFileUtils{log: NullLogger{}}.Count(test.in)
		var noErr bool = (err == nil)
		if test.noErr != noErr {
			t.Errorf("input parameter is %s,expected noerr is %t", test.in, test.noErr)
		}
		if test.noErr {
			if o1 != test.out1 || o2 != test.out2 || o3 != test.out3 {
				t.Fatalf("input parameter is %s,output is %d,%d, %d", test.in, o1, o2, o3)
			}
		}
	}
}
