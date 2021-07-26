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

func (log NullLogger) Info(_ string, _ ...interface{}) {
	// Do nothing.
}

func TestNewAlluxioFileUtils(t *testing.T) {
	var expectedResult = AlluxioFileUtils{
		podName:   "hbase",
		namespace: "default",
		container: "hbase-container",
		log:       NullLogger{},
	}
	result := NewAlluxioFileUtils("hbase", "hbase-container", "default", NullLogger{})
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the AlluxioFileUtils")
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
		tools := NewAlluxioFileUtils("", "", "", ctrl.Log)
		err := tools.LoadMetaData(test.path, test.sync)
		// fmt.Println(expectedErr)
		if err == nil {
			t.Errorf("expected %v, got %v", test.path, tools)
		}
	}
}

func (log NullLogger) Enabled() bool {
	return false
}

func (log NullLogger) Error(_ error, _ string, _ ...interface{}) {
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

func TestAlluxioFileUtils_ReportSummary(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := AlluxioFileUtils{}
	_, err = a.ReportSummary()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportSummary()
	if err != nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()
}

func TestAlluxioFileUtils_LoadMetadataWithoutTimeout(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecWithoutTimeoutErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExecWithoutTimeout := func() {
		err := gohook.UnHook(AlluxioFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := AlluxioFileUtils{log: NullLogger{}}
	err = a.LoadMetadataWithoutTimeout("/")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExecWithoutTimeout()

	err = gohook.Hook(AlluxioFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetadataWithoutTimeout("/")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExecWithoutTimeout()
}

func TestAlluxioFileUtils_LoadMetaData(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := AlluxioFileUtils{log: NullLogger{}}
	err = a.LoadMetaData("/", true)
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.LoadMetaData("/", false)
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFileUtils_QueryMetaDataInfoIntoFile(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := AlluxioFileUtils{log: NullLogger{}}

	keySets := []KeyOfMetaDataFile{DatasetName, Namespace, UfsTotal, FileNum, ""}
	for _, keySet := range keySets {
		_, err = a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err == nil {
			t.Errorf("fail to catch the error")
			return
		}
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, keySet := range keySets {
		_, err = a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err != nil {
			t.Errorf("fail to exec the function")
			return
		}
	}
	wrappedUnhookExec()
}

func TestAlluxioFIleUtils_MKdir(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "alluxio mkdir success", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := AlluxioFileUtils{}
	err = a.Mkdir("/")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.Mkdir("/")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFIleUtils_Mount(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "alluxio mkdir success", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	a := AlluxioFileUtils{}
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

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, test := range testCases {
		err = a.Mount("/", "/", nil, test.readOnly, test.shared)
		if err == nil {
			t.Errorf("fail to catch the error")
			return
		}
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, test := range testCases {
		err = a.Mount("/", "/", nil, test.readOnly, test.shared)
		if err != nil {
			t.Errorf("fail to exec the function")
			return
		}
	}
	wrappedUnhookExec()
}

func TestAlluxioFileUtils_IsMounted(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "https://mirrors.bit.edu.cn/apache/hbase/stable  on  /hbase (web, capacity=-1B, used=-1B, read-only, not shared, properties={}) \n /underFSStorage  on  /  (local, capacity=0B, used=0B, not read-only, not shared, properties={})", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}
	_, err = a.IsMounted("/hbase")
	if err == nil {
		t.Errorf("fail to catch the error")
		return
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	var testCases = []struct {
		alluxioPath    string
		expectedResult bool
	}{
		{
			alluxioPath:    "/spark",
			expectedResult: false,
		},
		{
			alluxioPath:    "/hbase",
			expectedResult: true,
		},
	}
	for _, test := range testCases {
		mounted, err := a.IsMounted(test.alluxioPath)
		if err != nil || mounted != test.expectedResult {
			t.Errorf("fail to exec the function")
			return
		}
	}
}

func TestAlluxioFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary: ", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}
	ready := a.Ready()
	if ready != false {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ready = a.Ready()
	if ready != true {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFIleUtils_Du(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "File Size     In Alluxio       Path\n577575561     0 (0%)           /hbase", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}
	_, _, _, err = a.Du("/hbase")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	ufs, cached, cachedPercentage, err := a.Du("/hbase")
	if err != nil || ufs != 577575561 || cached != 0 || cachedPercentage != "0%" {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
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

func TestAlluxioFileUtils_GetFileCount(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}
	_, err = a.GetFileCount()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	fileCount, err := a.GetFileCount()
	if err != nil || fileCount != 6367897 {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFIleUtils_ReportMetrics(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "report [category] [category args]\nReport Alluxio running cluster information.\n", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}

	_, err = a.ReportMetrics()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportMetrics()
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFIleUtils_ReportCapacity(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "report [category] [category args]\nReport Alluxio running cluster information.\n", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}
	_, err = a.ReportCapacity()
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = a.ReportCapacity()
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(AlluxioFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(AlluxioFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &AlluxioFileUtils{log: NullLogger{}}
	_, _, err = a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, false)
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, true)
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookExec()
}

func TestAlluxioFileUtils_execWithoutTimeout(t *testing.T) {
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
	a := &AlluxioFileUtils{log: NullLogger{}}
	_, _, err = a.execWithoutTimeout([]string{"alluxio", "fsadmin", "report", "capacity"}, false)
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhook()

	err = gohook.Hook(kubeclient.ExecCommandInContainer, mockExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.execWithoutTimeout([]string{"alluxio", "fsadmin", "report", "capacity"}, true)
	if err != nil {
		t.Errorf("fail to exec the function")
	}
}

func TestAlluxioFileUtils_MasterPodName(t *testing.T) {
	type fields struct {
		podName   string
		namespace string
		container string
		log       logr.Logger
	}
	tests := []struct {
		name              string
		fields            fields
		wantMasterPodName string
		wantErr           bool
	}{
		{
			name: "test0",
			fields: fields{
				podName:   "test-master-0",
				namespace: "default",
				container: "alluxio-master",
				log:       ctrl.Log,
			},
			wantMasterPodName: "test-master-0",
			wantErr:           true,
		},
		{
			name: "test1",
			fields: fields{
				podName:   "test-master-1",
				namespace: "default",
				container: "alluxio-master",
				log:       ctrl.Log,
			},
			wantMasterPodName: "test-master-1",
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AlluxioFileUtils{
				podName:   tt.fields.podName,
				namespace: tt.fields.namespace,
				container: tt.fields.container,
				log:       tt.fields.log,
			}
			gotMasterPodName, err := a.MasterPodName()
			if (err != nil) != tt.wantErr {
				t.Errorf("AlluxioFileUtils.MasterPodName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMasterPodName != tt.wantMasterPodName {
				t.Errorf("AlluxioFileUtils.MasterPodName() = %v, want %v", gotMasterPodName, tt.wantMasterPodName)
			}
		})
	}
}
