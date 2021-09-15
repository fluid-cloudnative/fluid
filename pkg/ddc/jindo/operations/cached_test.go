package operations

import (
	"errors"
	"github.com/brahma-adshonor/gohook"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
)

func TestJindoFIlUtils_CleanCache(t *testing.T) {
	ExecCommon := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "Test stout", "", nil
	}
	ExecErr := func(a JindoFileUtils, command []string, verbose bool) (stdout string, stderr string, err error) {
		return "", "", errors.New("fail to run the command")
	}
	wrappedUnhookExec := func() {
		err := gohook.UnHook(JindoFileUtils.exec)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(JindoFileUtils.exec, ExecErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	a := &JindoFileUtils{log: log.NullLogger{}}
	err = a.CleanCache()
	if err == nil {
		t.Error("check failure, want err, got nil")
	}
	wrappedUnhookExec()

	err = gohook.Hook(JindoFileUtils.exec, ExecCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = a.CleanCache()
	if err != nil {
		t.Errorf("check failure, want nil, got err: %v", err)
	}
	wrappedUnhookExec()
}

