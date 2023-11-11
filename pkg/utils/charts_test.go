/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
