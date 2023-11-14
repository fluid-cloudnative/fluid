package helm

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/brahma-adshonor/gohook"
)

func TestGenerateValueFile(t *testing.T) {
	_, err := GenerateValueFile("test-value")
	if err != nil {
		t.Errorf("fail to exec the function")
		return
	}
}

func TestGenerateHelmTemplate(t *testing.T) {
	CombinedOutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("test-output"), nil
	}
	CombinedOutputErr := func(cmd *exec.Cmd) ([]byte, error) {
		return nil, errors.New("fail to run the command")
	}
	StatCommon := func(name string) (os.FileInfo, error) {
		return nil, nil
	}
	StatErr := func(name string) (os.FileInfo, error) {
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
	wrappedUnhookStat := func() {
		err := gohook.UnHook(os.Stat)
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
	_, err = GenerateHelmTemplate("fluid", "default", "testValueFile", "testChartName")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(os.Stat, StatErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = GenerateHelmTemplate("fluid", "default", "testValueFile", "testChartName")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookStat()

	err = gohook.Hook(os.Stat, StatCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).CombinedOutput, CombinedOutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = GenerateHelmTemplate("fluid", "default", "testValueFile", "testChartName")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCombinedOutput()

	err = gohook.Hook((*exec.Cmd).CombinedOutput, CombinedOutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = GenerateHelmTemplate("fluid", "default", "testValueFile", "testChartName")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookCombinedOutput()
	wrappedUnhookStat()
	wrappedUnhookLookPath()
}

func TestGetChartVersion(t *testing.T) {
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
	}
	StatCommon := func(name string) (os.FileInfo, error) {
		return nil, nil
	}
	StatErr := func(name string) (os.FileInfo, error) {
		return nil, errors.New("fail to run the command")
	}
	OutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("fluid:v0.6.0"), nil
	}
	OutputErr := func(cmd *exec.Cmd) ([]byte, error) {
		return nil, errors.New("fail to run the command")
	}

	wrappedUnhookOutput := func() {
		err := gohook.UnHook((*exec.Cmd).Output)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookStat := func() {
		err := gohook.UnHook(os.Stat)
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
	_, err = GetChartVersion("fluid:v0.6.0")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(os.Stat, StatErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = GetChartVersion("fluid:v0.6.0")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookStat()

	err = gohook.Hook(os.Stat, StatCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Output, OutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = GetChartVersion("fluid:v0.6.0")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookOutput()

	err = gohook.Hook((*exec.Cmd).Output, OutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	version, err := GetChartVersion("fluid:v0.6.0")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if version != "v0.6.0" {
		t.Errorf("fail to get the version of the helm")
	}
	wrappedUnhookOutput()
	wrappedUnhookStat()
	wrappedUnhookLookPath()
}

func TestGetChartName(t *testing.T) {
	var testCases = []struct {
		chartName      string
		expectedResult string
	}{
		{
			chartName:      "",
			expectedResult: ".",
		},
		{
			chartName:      "/chart/fluid",
			expectedResult: "fluid",
		},
	}

	for _, test := range testCases {
		if name := GetChartName(test.chartName); name != test.expectedResult {
			t.Errorf("fail to get chart name")
		}
	}
}
