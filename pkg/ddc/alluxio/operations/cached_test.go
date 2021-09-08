package operations

import (
	"errors"
	"github.com/brahma-adshonor/gohook"
	"testing"
)

func TestAlluxioFileUtils_CachedState(t *testing.T) {
	ExecCommon := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alluxio cluster summary: \n    Master Address: 192.168.0.193:20009  \n Used Capacity: 0B\n", "", nil
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
	_, err = a.CachedState()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	cached, err := a.CachedState()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	if cached != 0 {
		t.Errorf("check failure, want 0, got: %d", cached)
	}
	wrappedUnhookExec()
}

func TestAlluxioFIlUtils_CleanCache(t *testing.T) {
	ExecCommonUbuntu := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Ubuntu", "", nil
	}
	ExecCommonAlpine := func(a AlluxioFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Alpine", "", nil
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
	err = a.CleanCache("/")
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommonUbuntu, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.CleanCache("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()

	err = gohook.Hook(AlluxioFileUtils.exec, ExecCommonAlpine, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.CleanCache("/")
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}
