/*
Copyright 2023 The Fluid Authors.

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

package helm

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/brahma-adshonor/gohook"
)

func TestInstallRelease(t *testing.T) {
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
	CombinedOutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("test-output"), nil
	}
	CombinedOutputErr := func(cmd *exec.Cmd) ([]byte, error) {
		return nil, errors.New("fail to run the command")
	}

	wrappedUnhookLookPath := func() {
		err := gohook.UnHook(exec.LookPath)
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
	wrappedUnhookCombinedOutput := func() {
		err := gohook.UnHook((*exec.Cmd).CombinedOutput)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(exec.LookPath, LookPathErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = InstallRelease("fluid", "default", "testValueFile", "testChartName")
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
	err = InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
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
	err = InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCombinedOutput()

	err = gohook.Hook((*exec.Cmd).CombinedOutput, CombinedOutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookCombinedOutput()
	wrappedUnhookStat()
	wrappedUnhookLookPath()
}

func TestCheckRelease(t *testing.T) {
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
	}
	StartErr := func(cmd *exec.Cmd) error {
		return errors.New("fail to run the command")
	}
	StartCommon := func(cmd *exec.Cmd) error {
		return nil
	}
	WaitErr := func(cmd *exec.Cmd) error {
		return errors.New("fail to run the command")
	}

	wrappedUnhookLookPath := func() {
		err := gohook.UnHook(exec.LookPath)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookStart := func() {
		err := gohook.UnHook((*exec.Cmd).Start)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookWait := func() {
		err := gohook.UnHook((*exec.Cmd).Wait)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	err := gohook.Hook(exec.LookPath, LookPathErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = CheckRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Start, StartErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = CheckRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookStart()

	err = gohook.Hook((*exec.Cmd).Start, StartCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Wait, WaitErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = CheckRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookWait()
	wrappedUnhookStart()
	wrappedUnhookLookPath()
}

func TestDeleteRelease(t *testing.T) {
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
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
	err = DeleteRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Output, OutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = DeleteRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookOutput()

	err = gohook.Hook((*exec.Cmd).Output, OutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = DeleteRelease("fluid", "default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookOutput()
	wrappedUnhookLookPath()
}

func TestListReleases(t *testing.T) {
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
	}
	OutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("fluid:v0.6.0\nfluid:v0.5.0"), nil
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
	_, err = ListReleases("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Output, OutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = ListReleases("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookOutput()

	err = gohook.Hook((*exec.Cmd).Output, OutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	release, err := ListReleases("default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if len(release) != 2 {
		t.Errorf("fail to exec the function ListRelease")
	}
	wrappedUnhookOutput()
	wrappedUnhookLookPath()
}

func TestListReleaseMap(t *testing.T) {
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
	}
	OutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("fluid v0.6.0\nspark v0.5.0"), nil
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
	_, err = ListReleaseMap("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Output, OutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = ListReleaseMap("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookOutput()

	err = gohook.Hook((*exec.Cmd).Output, OutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	release, err := ListReleaseMap("default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if len(release) != 2 {
		t.Errorf("fail to split the strout")
	}
	wrappedUnhookOutput()
	wrappedUnhookLookPath()
}

func TestListAllReleasesWithDetail(t *testing.T) {
	LookPathCommon := func(file string) (string, error) {
		return "test-path", nil
	}
	LookPathErr := func(file string) (string, error) {
		return "", errors.New("fail to run the command")
	}
	OutputCommon := func(cmd *exec.Cmd) ([]byte, error) {
		return []byte("fluid default 1 2021-07-19 16:20:16.166658248 +0800 CST deployed fluid-0.6.0 0.6.0-3c06c0e\nspark default 2 2021-07-19 16:20:16.166658248 +0800 CST deployed spark-0.3.0 0.3.0-3c06c0e"), nil
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
	_, err = ListAllReleasesWithDetail("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookLookPath()

	err = gohook.Hook(exec.LookPath, LookPathCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook((*exec.Cmd).Output, OutputErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = ListAllReleasesWithDetail("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookOutput()

	err = gohook.Hook((*exec.Cmd).Output, OutputCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	release, err := ListAllReleasesWithDetail("default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if len(release) != 2 {
		t.Errorf("fail to split the strout")
	}
	wrappedUnhookOutput()
	wrappedUnhookLookPath()
}

func TestDeleteReleaseIfExists(t *testing.T) {
	CheckReleaseCommonTrue := func(name, namespace string) (exist bool, err error) {
		return true, nil
	}
	CheckReleaseCommonFalse := func(name, namespace string) (exist bool, err error) {
		return false, nil
	}
	CheckReleaseErr := func(name, namespace string) (exist bool, err error) {
		return false, errors.New("fail to run the command")
	}
	DeleteReleaseCommon := func(name, namespace string) (err error) {
		return nil
	}
	DeleteReleaseErr := func(name, namespace string) (err error) {
		return errors.New("fail to run the command")
	}
	wrappedUnhookCheckRelease := func() {
		err := gohook.UnHook(CheckRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	wrappedUnhookDeleteRelease := func() {
		err := gohook.UnHook(DeleteRelease)
		if err != nil {
			t.Fatal(err.Error())
		}
	}
	err := gohook.Hook(CheckRelease, CheckReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = DeleteReleaseIfExists("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookCheckRelease()

	err = gohook.Hook(CheckRelease, CheckReleaseCommonFalse, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = DeleteReleaseIfExists("fluid", "default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	wrappedUnhookCheckRelease()

	err = gohook.Hook(CheckRelease, CheckReleaseCommonTrue, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = gohook.Hook(DeleteRelease, DeleteReleaseErr, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = DeleteReleaseIfExists("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookDeleteRelease()

	err = gohook.Hook(DeleteRelease, DeleteReleaseCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = DeleteReleaseIfExists("fluid", "default")
	if err != nil {
		t.Errorf("fail to catch the error")
	}
	wrappedUnhookDeleteRelease()
	wrappedUnhookCheckRelease()
}
