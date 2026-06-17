/*
Copyright 2020 The Fluid Authors.

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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	NOT_EXIST      = "not-exist"
	OTHER_ERR      = "other-err"
	FINE           = "fine"
	EXEC_ERR       = "exec-err"
	NEGATIVE_RES   = "negative-res"
	TOO_MANY_LINES = "too many lines"
	DATA_NUM       = "data nums not match"
	PARSE_ERR      = "parse err"
)

func TestNewAlluxioFileUtils(t *testing.T) {
	var expectedResult = AlluxioFileUtils{
		podName:   "hbase",
		namespace: "default",
		container: "hbase-container",
		log:       fake.NullLogger(),
	}
	result := NewAlluxioFileUtils("hbase", "hbase-container", "default", fake.NullLogger())
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the AlluxioFileUtils, want: %v, got: %v", expectedResult, result)
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

	mockExec := func(ctx context.Context, p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
		return "", "", nil
	}

	patches := gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, mockExec)
	defer patches.Reset()

	for _, test := range tests {
		tools := NewAlluxioFileUtils("", "", "", ctrl.Log)
		err := tools.LoadMetaData(test.path, test.sync)
		// fmt.Println(expectedErr)
		if err != nil {
			t.Errorf("expected %v, got %v", test.path, tools)
		}
	}
}

// TestAlluxioFileUtils_Du tests the Du method of AlluxioFileUtils.
// It verifies the method's ability to parse the output of the alluxio fs du command
// and handle various error scenarios such as execution errors, too many lines,
// data number mismatches, and parse errors.
//
// Test cases:
// - EXEC_ERR: Tests handling of command execution errors.
// - TOO_MANY_LINES: Tests handling of output with too many lines.
// - DATA_NUM: Tests handling of mismatched data numbers.
// - PARSE_ERR: Tests handling of parse errors.
// - FINE: Tests successful parsing of du output.
func TestAlluxioFileUtils_Du(t *testing.T) {
	out1, out2, out3 := 111, 222, "%233"
	mockExec := func(ctx context.Context, p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {

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

	patches := gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, mockExec)
	defer patches.Reset()

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
		o1, o2, o3, err := AlluxioFileUtils{log: fake.NullLogger()}.Du(test.in)
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

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := AlluxioFileUtils{}
	_, err := a.ReportSummary()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	_, err = a.ReportSummary()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

// TestAlluxioFileUtils_LoadMetadataWithoutTimeout tests LoadMetadataWithoutTimeout
// by mocking the internal exec behavior to cover both failure and success cases.
func TestAlluxioFileUtils_LoadMetadataWithoutTimeout(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecWithoutTimeoutErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := AlluxioFileUtils{log: fake.NullLogger()}
	err := a.LoadMetadataWithoutTimeout("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecWithoutTimeoutCommon)
	err = a.LoadMetadataWithoutTimeout("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

// TestAlluxioFileUtils_LoadMetaData tests the AlluxioFileUtils.LoadMetaData method
// for both failure and success cases by mocking the internal exec command.
// It accepts a *testing.T and has no return value.
func TestAlluxioFileUtils_LoadMetaData(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := AlluxioFileUtils{log: fake.NullLogger()}
	err := a.LoadMetaData("/", true)
	patches.Reset()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches = gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	err = a.LoadMetaData("/", false)
	patches.Reset()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

// TestAlluxioFileUtils_QueryMetaDataInfoIntoFile tests the QueryMetaDataInfoIntoFile method.
// It uses gomonkey to mock the internal exec method, verifying that an error is returned when exec fails
// and that no error occurs when exec succeeds. The test covers all defined KeyOfMetaDataFile types.
func TestAlluxioFileUtils_QueryMetaDataInfoIntoFile(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := AlluxioFileUtils{log: fake.NullLogger()}

	keySets := []KeyOfMetaDataFile{DatasetName, Namespace, UfsTotal, FileNum, ""}
	for index, keySet := range keySets {
		_, err := a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err == nil {
			t.Errorf("%d check failure, want err, got nil", index)
			return
		}
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	for index, keySet := range keySets {
		_, err := a.QueryMetaDataInfoIntoFile(keySet, "/tmp/file")
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}
	}
}

// TestAlluxioFIleUtils_MKdir verifies AlluxioFileUtils.Mkdir by stubbing the private exec method with gomonkey:
// when exec returns an error, Mkdir should return a non-nil error; when exec succeeds, Mkdir should return nil.
// patches.Reset in defer restores the original behavior after the test.
func TestAlluxioFIleUtils_MKdir(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "alluxio mkdir success", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := AlluxioFileUtils{}
	err := a.Mkdir("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)

	err = a.Mkdir("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

// TestAlluxioFIleUtils_Mount verifies that AlluxioFileUtils.Mount returns the expected result
// when mounting an Alluxio path to a UFS path with different read-only and shared flag combinations.
// Parameters:
// - t (*testing.T): The testing object used to report test failures.
// Returns:
// - None: This test reports failures through the testing object.
func TestAlluxioFIleUtils_Mount(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "alluxio mkdir success", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
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

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	for index, test := range testCases {
		err := a.Mount("/", "/", nil, test.readOnly, test.shared)
		if err == nil {
			t.Errorf("%d check failure, want err, got nil", index)
			return
		}
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	for index, test := range testCases {
		err := a.Mount("/", "/", nil, test.readOnly, test.shared)
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}
	}
}

func TestAlluxioFileUtils_IsMounted(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "https://mirrors.bit.edu.cn/apache/hbase/stable  on  /hbase (web, capacity=-1B, used=-1B, read-only, not shared, properties={}) \n /underFSStorage  on  /  (local, capacity=0B, used=0B, not read-only, not shared, properties={})", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, err := a.IsMounted("/hbase")
	if err == nil {
		t.Error("check failure, want err, got nil")
		return
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
	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	for index, test := range testCases {
		mounted, err := a.IsMounted(test.alluxioPath)
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

// TestAlluxioFileUtils_FindUnmountedAlluxioPaths tests the FindUnmountedAlluxioPaths method
// of AlluxioFileUtils. It mocks the internal exec command to simulate the output of
// `alluxio fs mount` and verifies that paths not currently mounted are correctly
// identified from the given alluxioPaths slice.
//
// Test cases cover:
// - all provided paths are already mounted (expect empty result)
// - a mix of mounted and unmounted paths (expect only unmounted ones)
// - an empty input slice (expect empty result)
// - all provided paths are unmounted (expect all paths returned)
//
// Parameters:
// - t (*testing.T): the test context used for reporting failures.
func TestAlluxioFileUtils_FindUnmountedAlluxioPaths(t *testing.T) {
	const returnMessage = `s3://bucket/path/train on /cache (s3, capacity=-1B, used=-1B, not read-only, not shared, properties={alluxio.underfs.s3.inherit.acl=false, alluxio.underfs.s3.endpoint=s3endpoint, aws.secretKey=, aws.accessKeyId=})
/underFSStorage on / (local, capacity=0B, used=0B, not read-only, not shared, properties={})`

	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return returnMessage, "", nil
	}
	a := &AlluxioFileUtils{log: fake.NullLogger()}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	defer patches.Reset()

	var testCases = []struct {
		alluxioPaths           []string
		expectedUnmountedPaths []string
	}{
		{
			alluxioPaths:           []string{"/cache"},
			expectedUnmountedPaths: []string{},
		},
		{
			alluxioPaths:           []string{"/cache", "/cache2"},
			expectedUnmountedPaths: []string{"/cache2"},
		},
		{
			alluxioPaths:           []string{},
			expectedUnmountedPaths: []string{},
		},
		{
			alluxioPaths:           []string{"/cache2"},
			expectedUnmountedPaths: []string{"/cache2"},
		},
	}
	for index, test := range testCases {
		unmountedPaths, err := a.FindUnmountedAlluxioPaths(test.alluxioPaths)
		if err != nil {
			t.Errorf("%d check failure, want nil, got err: %v", index, err)
			return
		}

		if (len(unmountedPaths) != 0 || len(test.expectedUnmountedPaths) != 0) &&
			!reflect.DeepEqual(unmountedPaths, test.expectedUnmountedPaths) {
			t.Errorf("%d check failure, want: %s, got: %s", index, strings.Join(test.expectedUnmountedPaths, ","), strings.Join(unmountedPaths, ","))
			return
		}
	}
}

// TestAlluxioFileUtils_Ready verifies the Ready behavior of AlluxioFileUtils under
// both failure and success conditions by mocking the internal exec command.
// It confirms that Ready returns false when the readiness probe fails and true
// when the probe returns a valid Alluxio cluster summary.
func TestAlluxioFileUtils_Ready(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary: ", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()
	a := &AlluxioFileUtils{log: fake.NullLogger()}
	ready := a.Ready()
	if ready != false {
		t.Errorf("check failure, want false, got %t", ready)
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	ready = a.Ready()
	if ready != true {
		t.Errorf("check failure, want true, got %t", ready)
	}
}

func TestAlluxioFIleUtils_Du(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "File Size     In Alluxio       Path\n577575561     0 (0%)           /hbase", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()
	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, _, _, err := a.Du("/hbase")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
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
}

// TestAlluxioFileUtils_Count tests the Count method of AlluxioFileUtils.
// This function verifies the Count method's ability to parse the output of the alluxio fs count command
// and handle various error scenarios such as execution errors, negative results, too many lines,
// data number mismatches, and parse errors.
//
// Test cases:
// - EXEC_ERR: Tests handling of command execution errors.
// - NEGATIVE_RES: Tests handling of negative result values.
// - TOO_MANY_LINES: Tests handling of output with too many lines.
// - DATA_NUM: Tests handling of mismatched data numbers.
// - PARSE_ERR: Tests handling of parse errors.
// - FINE: Tests successful parsing of count output.
func TestAlluxioFileUtils_Count(t *testing.T) {
	out1, out2, out3 := 111, 222, 333
	mockExec := func(ctx context.Context, p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {

		if strings.Contains(p4[3], EXEC_ERR) {
			return "does not exist", "", errors.New("exec-error")
		} else if strings.Contains(p4[3], NEGATIVE_RES) {
			return "12324\t45463\t-9223372036854775808", "", nil
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

	patches := gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithFullOutput, mockExec)
	defer patches.Reset()

	var tests = []struct {
		in               string
		out1, out2, out3 int64
		noErr            bool
	}{
		{EXEC_ERR, 0, 0, 0, false},
		{NEGATIVE_RES, 0, 0, 0, false},
		{TOO_MANY_LINES, 0, 0, 0, false},
		{DATA_NUM, 0, 0, 0, false},
		{PARSE_ERR, 0, 0, 0, false},
		{FINE, int64(out1), int64(out2), int64(out3), true},
	}
	for _, test := range tests {
		o1, o2, o3, err := AlluxioFileUtils{log: fake.NullLogger()}.Count(test.in)
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

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, err := a.GetFileCount()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecWithoutTimeoutCommon)
	fileCount, err := a.GetFileCount()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if fileCount != 6367897 {
		t.Errorf("check failure, want 6367897, got %d", fileCount)
	}
}

func TestAlluxioFIleUtils_ReportMetrics(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "report [category] [category args]\nReport Alluxio running cluster information.\n", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}

	_, err := a.ReportMetrics()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	_, err = a.ReportMetrics()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

// TestAlluxioFIleUtils_ReportCapacity tests the ReportCapacity method of AlluxioFileUtils.
// This test verifies both the error handling when the underlying command execution fails
// and the successful path when the command returns valid cluster capacity information.
func TestAlluxioFIleUtils_ReportCapacity(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "report [category] [category args]\nReport Alluxio running cluster information.\n", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, err := a.ReportCapacity()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	_, err = a.ReportCapacity()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

// TestAlluxioFileUtils_exec tests the private exec method of AlluxioFileUtils.
// It mocks the exec implementation to verify that an error is returned when
// command execution fails and that no error is returned when execution succeeds.
// The test also covers calls with both disabled and enabled verbose output.
func TestAlluxioFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecWithoutTimeoutErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, _, err := a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecWithoutTimeoutCommon)
	_, _, err = a.exec([]string{"alluxio", "fsadmin", "report", "capacity"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
}

func TestAlluxioFileUtils_MasterPodName(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary: \n    Master Address: 192.168.0.193:20009\n    Web Port: 20010", "", nil
	}
	ExecErr := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	patches := gomonkey.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecErr)
	defer patches.Reset()

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	_, err := a.MasterPodName()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}

	patches.ApplyPrivateMethod(reflect.TypeOf(AlluxioFileUtils{}), "exec", ExecCommon)
	address, err := a.MasterPodName()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if address != "192.168.0.193" {
		t.Errorf("check failure, want: %s, got: %s", "192.168.0.193", address)
	}
}

func TestAlluxioFileUtils_ExecMountScripts(t *testing.T) {
	// Mock exec to avoid invoking the mounted script during the unit test.
	ExecCommon := func(command []string, verbose bool) (stdout string, stderr string, err error) {
		return strings.Join(command, " "), "", nil
	}
	ExecErr := func(command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	a := &AlluxioFileUtils{log: fake.NullLogger()}
	patch1 := gomonkey.ApplyPrivateMethod(*a, "exec", ExecErr)

	// ExecMountScripts should return the error reported by exec.
	err := a.ExecMountScripts()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	patch1.Reset()

	patch2 := gomonkey.ApplyPrivateMethod(*a, "exec", ExecCommon)
	// ExecMountScripts should complete without error when exec succeeds.
	err = a.ExecMountScripts()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	patch2.Reset()
}
