/*
Copyright 2022 The Fluid Authors.

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

package jindocache

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
			j := &JindoCacheEngine{
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
