package operations

import (
	"errors"
	"github.com/brahma-adshonor/gohook"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	"reflect"
	"strings"
	"testing"
)

const (
	NOT_EXIST = "not-exist"
	OTHER_ERR = "other-err"
	FINE      = "fine"
)

// an empty logger just for testing ...
type NullLogger struct{}

func (log NullLogger) Info(_ string, _ ...interface{}) {
	// Do nothing.
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

func TestNewJuiceFSFileUtils(t *testing.T) {
	var expectedResult = JuiceFileUtils{
		podName:   "juicefs",
		namespace: "default",
		container: common.JuiceFSFuseContainer,
		log:       NullLogger{},
	}
	result := NewJuiceFileUtils("juicefs", common.JuiceFSFuseContainer, "default", NullLogger{})
	if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("fail to create the JuiceFSFileUtils, want: %v, got: %v", expectedResult, result)
	}
}

func TestJuiceFileUtils_IsExist(t *testing.T) {
	mockExec := func(a JuiceFileUtils, p []string, verbose bool) (stdout string, stderr string, e error) {
		if strings.Contains(p[1], NOT_EXIST) {
			return "No such file or directory", "", errors.New("No such file or directory")
		} else if strings.Contains(p[1], OTHER_ERR) {
			return "", "", errors.New("other error")
		} else {
			return "", "", nil
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, mockExec, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

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
		found, err := JuiceFileUtils{log: NullLogger{}}.IsExist(test.in)
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

func TestJuiceFileUtils_Mkdir(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs mkdir success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	err = a.Mkdir("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.Mkdir("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_exec(t *testing.T) {
	ExecWithoutTimeoutCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Type: COUNTER, Value: 6,367,897", "", nil
	}
	ExecWithoutTimeoutErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}

	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.execWithoutTimeout)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JuiceFileUtils{log: NullLogger{}}
	_, _, err = a.exec([]string{"mkdir", "abc"}, false)
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.execWithoutTimeout, ExecWithoutTimeoutCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, _, err = a.exec([]string{"mkdir", "abc"}, true)
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_GetMetric(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs metrics success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	_, err = a.GetMetric()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	m, err := a.GetMetric()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if m != "juicefs metrics success" {
		t.Errorf("expected juicefs metrics success, got %s", m)
	}
	wrappedUnhookExec()
}

func TestJuiceFileUtils_DeleteDir(t *testing.T) {
	ExecCommon := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "juicefs rmr success", "", nil
	}
	ExecErr := func(a JuiceFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JuiceFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JuiceFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := JuiceFileUtils{}
	err = a.DeleteDir("")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JuiceFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.DeleteDir("")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}
