package kubectl

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/brahma-adshonor/gohook"
)

func TestCreateConfigMapFromFile(t *testing.T) {
	StatCommon := func(name string) (os.FileInfo, error) {
		return nil, nil
	}
	StatErr := func(name string) (os.FileInfo, error) {
		return nil, errors.New("fail to run the command")
	}
	KubectlCommon := func(args []string) ([]byte, error) {
		return []byte("output"), nil
	}
	KubectlErr := func(args []string) ([]byte, error) {
		return nil, errors.New("fail to exec the kubectl function")
	}

	wrappedUnhookStat := func() {
		err := gohook.UnHook(os.Stat)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookKubectl := func() {
		err := gohook.UnHook(kubectl)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(os.Stat, StatErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = CreateConfigMapFromFile("hbase", "data", "output/tmp/hbase159761306", "default")
	if err == nil {
		t.Errorf("check failure, want err, got nil")
	}
	wrappedUnhookStat()

	err = gohook.Hook(os.Stat, StatCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(kubectl, KubectlErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = CreateConfigMapFromFile("hbase", "data", "output/tmp/hbase159761306", "default")
	if err == nil {
		t.Errorf("check failure, want err, got nil")
	}
	wrappedUnhookKubectl()

	err = gohook.Hook(kubectl, KubectlCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = CreateConfigMapFromFile("hbase", "data", "output/tmp/hbase159761306", "default")
	if err != nil {
		t.Errorf("check failure, want nil, got err %v", err)
	}
	wrappedUnhookKubectl()
	wrappedUnhookStat()
}

func TestKubectl(t *testing.T) {
	CombinedOutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("test-CombineOutput"), nil
	}
	CombinedOutputErr := func(cmd *exec.Cmd) ([]byte, error) {
		return nil, errors.New("fail to run the command")
	}
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
	}

	wrappedUnhookCombinedOutput := func() {
		err := gohook.UnHook((*exec.Cmd).CombinedOutput)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookLookPath := func() {
		err := gohook.UnHook(exec.LookPath)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(exec.LookPath, LookPathErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = kubectl([]string{"args1", "args2"})
	if err == nil {
		t.Errorf("check failure, want err, get nil")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).CombinedOutput, CombinedOutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = kubectl([]string{"args1", "args2"})
	if err == nil {
		t.Errorf("check failure, want err, get nil")
	}
	wrappedUnhookCombinedOutput()

	err = gohook.Hook((*exec.Cmd).CombinedOutput, CombinedOutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = kubectl([]string{"args1", "args2"})
	if err != nil {
		t.Errorf("check faillure, want nil, get err %v", err)
	}
	wrappedUnhookCombinedOutput()
	wrappedUnhookLookPath()
}
