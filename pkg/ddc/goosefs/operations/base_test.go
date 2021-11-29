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
	"reflect"
	"strings"
	"testing"

	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

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

func TestNewGooseFSFileUtils(t *testing.T) {
	var expectedResult = GooseFSFileUtils{
		podName:   "hbase",
		namespace: "default",
		container: "hbase-container",
		log:       logf.NullLogger{},
	}
	result := NewGooseFSFileUtils("hbase", "hbase-container", "default", logf.NullLogger{})
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the GooseFSFileUtils, want: %v, got: %v", expectedResult, result)
	}
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
		tools := NewGooseFSFileUtils("", "", "", ctrl.Log)
		err := tools.LoadMetaData(test.path, test.sync)
		// fmt.Println(expectedErr)
		if err == nil {
			t.Errorf("expected %v, got %v", test.path, tools)
		}
	}
}

func TestGooseFSFileUtils_IsExist(t *testing.T) {

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
		found, err := GooseFSFileUtils{log: logf.NullLogger{}}.IsExist(test.in)
		if found != test.out {
			t.Errorf("input parameter is %s,expected %t, got %t", test.in, test.out, found)
		}
		var noErr bool = (err == nil)
		if test.noErr != noErr {
			t.Errorf("input parameter is %s,expected noerr is %t", test.in, test.noErr)
		}
	}
}

func TestGooseFSFileUtils_Du(t *testing.T) {
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
		o1, o2, o3, err := GooseFSFileUtils{log: logf.NullLogger{}}.Du(test.in)
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

func TestGooseFSFileUtils_ReportSummary(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary", "", nil
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
	a := GooseFSFileUtils{}
	_, err = a.ReportSummary()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportSummary()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestGooseFSFileUtils_LoadMetadataWithoutTimeout(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary", "", nil
	}
	ExecWithoutTimeoutErr := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExecWithoutTimeout := func() {
		err := gohook.UnHook(GooseFSFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(GooseFSFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := GooseFSFileUtils{log: logf.NullLogger{}}
	err = a.LoadMetadataWithoutTimeout("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExecWithoutTimeout()

	err = gohook.Hook(GooseFSFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetadataWithoutTimeout("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExecWithoutTimeout()
}

func TestGooseFSFileUtils_LoadMetaData(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary", "", nil
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
	a := GooseFSFileUtils{log: logf.NullLogger{}}
	err = a.LoadMetaData("/", true)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetaData("/", false)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestGooseFSFileUtils_QueryMetaDataInfoIntoFile(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary", "", nil
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
	a := GooseFSFileUtils{log: logf.NullLogger{}}

	keySets := []KeyOfMetaDataFile{DatasetName, Namespace, UfsTotal, FileNum, ""}
	for index, keySet := range keySets {
		_, err = a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err == nil {
			t.Errorf("%d check failure, want err, got nil", index)
			return
		}
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for index, keySet := range keySets {
		_, err = a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}
	}
	wrappedUnhookExec()
}

func TestGooseFSFIleUtils_MKdir(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS mkdir success", "", nil
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
	a := GooseFSFileUtils{}
	err = a.Mkdir("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.Mkdir("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestGooseFSFIleUtils_Mount(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS mkdir success", "", nil
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

	a := GooseFSFileUtils{}
	var testCases = []struct {
		readOnly bool
		shared   bool
		options  map[string]string
	}{
		{
			readOnly: true,
			shared:   true,
			options: map[string]string{
				"testKey": "testValue",
			},
		},
		{
			readOnly: true,
			shared:   false,
		},
		{
			readOnly: false,
			shared:   true,
		},
		{
			readOnly: false,
			shared:   false,
		},
	}

	err := gohook.Hook(GooseFSFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for index, test := range testCases {
		err = a.Mount("/", "/", nil, test.readOnly, test.shared)
		if err == nil {
			t.Errorf("%d check failure, want err, got nil", index)
			return
		}
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for index, test := range testCases {
		err = a.Mount("/", "/", nil, test.readOnly, test.shared)
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}
	}
	wrappedUnhookExec()
}

func TestGooseFSFileUtils_IsMounted(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "https://mirrors.bit.edu.cn/apache/hbase/stable  on  /hbase (web, capacity=-1B, used=-1B, read-only, not shared, properties={}) \n /underFSStorage  on  /  (local, capacity=0B, used=0B, not read-only, not shared, properties={})", "", nil
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
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, err = a.IsMounted("/hbase")
	if err == nil {
		t.Error("check failure, want err, got nil")
		return
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	var testCases = []struct {
		goosefsPath    string
		expectedResult bool
	}{
		{
			goosefsPath:    "/spark",
			expectedResult: false,
		},
		{
			goosefsPath:    "/hbase",
			expectedResult: true,
		},
	}
	for index, test := range testCases {
		mounted, err := a.IsMounted(test.goosefsPath)
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}

		if mounted != test.expectedResult {
			t.Errorf("%d check failure, want: %t, got: %t ", index, mounted, test.expectedResult)
			return
		}
	}
}

func TestGooseFSFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary: ", "", nil
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
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	ready := a.Ready()
	if ready != false {
		t.Errorf("check failure, want false, got %t", ready)
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ready = a.Ready()
	if ready != true {
		t.Errorf("check failure, want true, got %t", ready)
	}
	wrappedUnhookExec()
}

func TestGooseFSFIleUtils_Du(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "File Size     In GooseFS       Path\n577575561     0 (0%)           /hbase", "", nil
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
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, _, _, err = a.Du("/hbase")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ufs, cached, cachedPercentage, err := a.Du("/hbase")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if ufs != 577575561 {
		t.Errorf("check failure, want 577575561, got %d", ufs)
	}
	if cached != 0 {
		t.Errorf("check failure, want 0, got %d", cached)
	}
	if cachedPercentage != "0%" {
		t.Errorf("check failure, want 0, got %s", cachedPercentage)
	}
	wrappedUnhookExec()
}

func TestGooseFSFileUtils_Count(t *testing.T) {
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
		o1, o2, o3, err := GooseFSFileUtils{log: logf.NullLogger{}}.Count(test.in)
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

func TestGooseFSFileUtils_GetFileCount(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(GooseFSFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(GooseFSFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, err = a.GetFileCount()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	fileCount, err := a.GetFileCount()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if fileCount != 6367897 {
		t.Errorf("check failure, want 6367897, got %d", fileCount)
	}
	wrappedUnhookExec()
}

func TestGooseFSFIleUtils_ReportMetrics(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "report [category] [category args]\nReport GooseFS running cluster information.\n", "", nil
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
	a := &GooseFSFileUtils{log: logf.NullLogger{}}

	_, err = a.ReportMetrics()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportMetrics()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestGooseFSFIleUtils_ReportCapacity(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "report [category] [category args]\nReport GooseFS running cluster information.\n", "", nil
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
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, err = a.ReportCapacity()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportCapacity()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestGooseFSFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(GooseFSFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(GooseFSFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, _, err = a.exec([]string{"goosefs", "fsadmin", "report", "capacity"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"goosefs", "fsadmin", "report", "capacity"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestGooseFSFileUtils_execWithoutTimeout(t *testing.T) {
	mockExecCommon := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "conf", "", nil
	}
	mockExecErr := func(podName string, containerName string, namespace string, cmd []string) (stdout string, stderr string, e error) {
		return "err", "", errors.New("other error")
	}
	wrappedUnhook := func() {
		err := gohook.UnHook(kubeclient.ExecCommandInContainer)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(kubeclient.ExecCommandInContainer, mockExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, _, err = a.execWithoutTimeout([]string{"goosefs", "fsadmin", "report", "capacity"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhook()

	err = gohook.Hook(kubeclient.ExecCommandInContainer, mockExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.execWithoutTimeout([]string{"goosefs", "fsadmin", "report", "capacity"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

func TestGooseFSFileUtils_MasterPodName(t *testing.T) {
	ExecCommon := func(a GooseFSFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "GooseFS cluster summary: \n    Master Address: 192.168.0.193:20009\n    Web Port: 20010", "", nil
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
	a := &GooseFSFileUtils{log: logf.NullLogger{}}
	_, err = a.MasterPodName()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(GooseFSFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	address, err := a.MasterPodName()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if address != "192.168.0.193" {
		t.Errorf("check failure, want: %s, got: %s", "192.168.0.193", address)
	}
	wrappedUnhookExec()
}
