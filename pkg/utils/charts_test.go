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

package utils

import (
	"os"
	"testing"
)

func TestPathExists(t *testing.T) {
	path := os.TempDir()
	if !PathExists(path) {
		t.Errorf("result of checking if the path exists is wrong")
	}
	if PathExists(path + "test/") {
		t.Errorf("result of checking if the path exists is wrong")
	}
}

func TestGetChartsDirectory(t *testing.T) {
	f, err := os.CreateTemp("", "test")
	if err != nil {
		t.Errorf("MkdirTemp failed due to %v", err)
	}
	testDir := f.Name()

	t.Setenv("HOME", testDir)
	if GetChartsDirectory() != "/charts" {
		t.Errorf("ChartsDirectory should be /charts if ~/charts not exist")
	}
	homeChartsFolder := os.Getenv("HOME") + "/charts"
	// Make Directory if it doesn't exist.
	_ = os.Mkdir(homeChartsFolder, 0600)
	if GetChartsDirectory() != "/charts" {
		t.Errorf("ChartsDirectory should be ~/charts if ~/charts exist")
	}
}
