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

	"github.com/agiledragon/gomonkey/v2"
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

	lookPathPatch := gomonkey.ApplyFunc(exec.LookPath, LookPathErr)
	err := InstallRelease("fluid", "default", "testValueFile", "testChartName")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()

	lookPathPatch.ApplyFunc(exec.LookPath, LookPathCommon)
	statPatch := gomonkey.ApplyFunc(os.Stat, StatErr)
	err = InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	statPatch.Reset()

	statPatch.ApplyFunc(os.Stat, StatCommon)
	combineOutputPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "CombinedOutput", CombinedOutputErr)
	err = InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	combineOutputPatch.Reset()

	badValue := "test$bad"
	err = InstallRelease("fluid", badValue, "testValueFile", "/chart/fluid")
	if err == nil {
		t.Errorf("fail to catch the error of %s", badValue)
	}

	combineOutputPatch.ApplyMethod((*exec.Cmd)(nil), "CombinedOutput", CombinedOutputCommon)
	err = InstallRelease("fluid", "default", "testValueFile", "/chart/fluid")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	combineOutputPatch.Reset()
	lookPathPatch.Reset()
	statPatch.Reset()
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

	lookupPatch := gomonkey.ApplyFunc(exec.LookPath, LookPathErr)
	_, err := CheckRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookupPatch.Reset()

	lookupPatch.ApplyFunc(exec.LookPath, LookPathCommon)
	startPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "Start", StartErr)
	_, err = CheckRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	startPatch.Reset()

	badValue := "test$bad"
	_, err = CheckRelease("fluid", badValue)
	if err == nil {
		t.Errorf("fail to catch the error of %s", badValue)
	}

	startPatch.ApplyMethod((*exec.Cmd)(nil), "Start", StartCommon)
	waitPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "Wait", WaitErr)
	_, err = CheckRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	waitPatch.Reset()
	startPatch.Reset()
	lookupPatch.Reset()
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

	lookPathPatch := gomonkey.ApplyFunc(exec.LookPath, LookPathErr)
	err := DeleteRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()

	lookPathPatch.ApplyFunc(exec.LookPath, LookPathCommon)
	outputPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", OutputErr)
	err = DeleteRelease("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	outputPatch.Reset()
	// test check illegal arguements
	badValue := "test$bad"
	err = DeleteRelease("fluid", badValue)
	if err == nil {
		t.Errorf("fail to catch the error of %s", badValue)
	}

	outputPatch.ApplyMethod((*exec.Cmd)(nil), "Output", OutputCommon)
	err = DeleteRelease("fluid", "default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	outputPatch.Reset()
	lookPathPatch.Reset()
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

	lookPathPatch := gomonkey.ApplyFunc(exec.LookPath, LookPathErr)
	_, err := ListReleases("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()

	lookPathPatch.ApplyFunc(exec.LookPath, LookPathCommon)
	outputPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", OutputErr)
	_, err = ListReleases("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	outputPatch.Reset()

	outputPatch.ApplyMethod((*exec.Cmd)(nil), "Output", OutputCommon)
	release, err := ListReleases("default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if len(release) != 2 {
		t.Errorf("fail to exec the function ListRelease")
	}
	outputPatch.Reset()

	_, err = ListReleases("def$ault")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()
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

	lookPathPatch := gomonkey.ApplyFunc(exec.LookPath, LookPathErr)
	_, err := ListReleaseMap("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()

	lookPathPatch.ApplyFunc(exec.LookPath, LookPathCommon)
	outputPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", OutputErr)
	_, err = ListReleaseMap("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	outputPatch.Reset()

	outputPatch.ApplyMethod((*exec.Cmd)(nil), "Output", OutputCommon)
	release, err := ListReleaseMap("default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if len(release) != 2 {
		t.Errorf("fail to split the strout")
	}
	outputPatch.Reset()

	_, err = ListReleaseMap("def$ault")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()
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

	lookPathPatch := gomonkey.ApplyFunc(exec.LookPath, LookPathErr)
	_, err := ListAllReleasesWithDetail("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()

	lookPathPatch.ApplyFunc(exec.LookPath, LookPathCommon)
	outputPatch := gomonkey.ApplyMethod((*exec.Cmd)(nil), "Output", OutputErr)
	_, err = ListAllReleasesWithDetail("default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	outputPatch.Reset()

	outputPatch.ApplyMethod((*exec.Cmd)(nil), "Output", OutputCommon)
	release, err := ListAllReleasesWithDetail("default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	if len(release) != 2 {
		t.Errorf("fail to split the strout")
	}
	outputPatch.Reset()

	_, err = ListAllReleasesWithDetail("def$ault")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	lookPathPatch.Reset()
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

	patches := gomonkey.ApplyFunc(CheckRelease, CheckReleaseErr)
	err := DeleteReleaseIfExists("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}
	patches.Reset()

	patches.ApplyFunc(CheckRelease, CheckReleaseCommonFalse)
	err = DeleteReleaseIfExists("fluid", "default")
	if err != nil {
		t.Errorf("fail to exec the function")
	}
	patches.Reset()

	patches.ApplyFunc(CheckRelease, CheckReleaseCommonTrue)
	patches.ApplyFunc(DeleteRelease, DeleteReleaseErr)
	err = DeleteReleaseIfExists("fluid", "default")
	if err == nil {
		t.Errorf("fail to catch the error")
	}

	patches.ApplyFunc(DeleteRelease, DeleteReleaseCommon)
	err = DeleteReleaseIfExists("fluid", "default")
	if err != nil {
		t.Errorf("fail to catch the error")
	}
	patches.Reset()
}
