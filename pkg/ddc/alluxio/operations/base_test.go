
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
	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	tt "github.com/go-logr/logr/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"
)

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
		tools := NewAlluxioFileUtils("", "", "", ctrl.Log)
		err := tools.LoadMetaData(test.path, test.sync)
		// fmt.Println(expectedErr)
		if err == nil {
			t.Errorf("expected %v, got %v", test.path, tools)
		}
	}
}
func TestAlluxioFileUtils_IsExist(t *testing.T) {
	mockExecTramp := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
		t.Fatal("done")
		if strings.Contains(p4[3], "not-exist") {
			return "does not exist", "", errors.New("does not exist")
		} else if strings.Contains(p4[3], "other-expectedErr") {
			return "other error", "other error", errors.New("other error")
		} else {
			return "", "", nil
		}
	}

	mockExec := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
		if strings.Contains(p4[3], "not-exist") {
			return "does not exist", "", errors.New("does not exist")
		} else if strings.Contains(p4[3], "other-expectedErr") {
			return "other error", "other error", errors.New("other error")
		} else {
			return "ok", "ok", nil
		}
	}

	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExec, mockExecTramp)
	if err != nil {
		t.Fatal(err.Error())
	}
	l := tt.NullLogger{}
	var tests = []struct {
		in          string
		out         bool
		expectedErr error
	}{
		{"not-exist", false, nil},
		{"other-expectedErr", false, errors.New("error")},
		{"fine", true, nil},
	}
	for _, test := range tests {
		found, err := AlluxioFileUtils{log: l}.IsExist(test.in)

		if found != test.out {
			t.Errorf("input parameter is %s,expected %t, got %t", test.in, test.out, found)
		}
		if test.expectedErr == nil && err != nil {
			t.Errorf("input parameter is %s,and expectedErr should be nil", test.in)
		}
		if test.expectedErr != nil && err == nil {
			t.Error("wrong")
		}
	}
}
