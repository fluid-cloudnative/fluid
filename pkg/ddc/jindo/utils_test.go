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

package jindo

import (
	"os"
	"testing"

	"github.com/go-logr/logr"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestIsFluidNativeScheme(t *testing.T) {

	var tests = []struct {
		mountPoint string
		expect     bool
	}{
		{"local:///test",
			true},
		{
			"pvc://test",
			true,
		}, {
			"oss://test",
			false,
		},
	}
	for _, test := range tests {
		result := common.IsFluidNativeScheme(test.mountPoint)
		if result != test.expect {
			t.Errorf("expect %v for %s, but got %v", test.expect, test.mountPoint, result)
		}
	}
}

func TestMountRootWithEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/var/lib/mymount/jindo"},
	}
	for _, tc := range testCases {
		t.Setenv(utils.MountRoot, tc.input)
		if tc.expected != getMountRoot() {
			t.Errorf("expected %#v, got %#v",
				tc.expected, getMountRoot())
		}
	}
}

func TestMountRootWithoutEnvSet(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"/var/lib/mymount", "/jindo"},
	}

	for _, tc := range testCases {
		os.Unsetenv(utils.MountRoot)
		if tc.expected != getMountRoot() {
			t.Errorf("expected %#v, got %#v",
				tc.expected, getMountRoot())
		}
	}
}

func TestJindoFSEngine_getHostMountPoint(t *testing.T) {
	type fields struct {
		name      string
		namespace string
		Log       logr.Logger
		MountRoot string
	}
	var tests = []struct {
		name          string
		fields        fields
		wantMountPath string
	}{
		{
			name: "test",
			fields: fields{
				name:      "jindofs",
				namespace: "default",
				Log:       fake.NullLogger(),
				MountRoot: "/tmp",
			},
			wantMountPath: "/tmp/jindo/default/jindofs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JindoEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       tt.fields.Log,
			}
			t.Setenv("MOUNT_ROOT", tt.fields.MountRoot)
			if gotMountPath := j.getHostMountPoint(); gotMountPath != tt.wantMountPath {
				t.Errorf("getHostMountPoint() = %v, want %v", gotMountPath, tt.wantMountPath)
			}
		})
	}
}
