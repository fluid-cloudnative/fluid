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
		_ = os.Unsetenv(utils.MountRoot)
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

func TestParseVersionFromImageTag(t *testing.T) {
	tests := []struct {
		name     string
		imageTag string
		want     jindoVersion
		wantErr  bool
	}{
		{
			name:     "valid version without tag",
			imageTag: "6.2.0",
			want: jindoVersion{
				Major: 6,
				Minor: 2,
				Patch: 0,
				Tag:   "",
			},
			wantErr: false,
		},
		{
			name:     "valid version without tag and with a 'v' prefix",
			imageTag: "v6.2.0",
			want: jindoVersion{
				Major: 6,
				Minor: 2,
				Patch: 0,
				Tag:   "",
			},
			wantErr: false,
		},
		{
			name:     "valid version with tag",
			imageTag: "6.9.1-202501020304",
			want: jindoVersion{
				Major: 6,
				Minor: 9,
				Patch: 1,
				Tag:   "202501020304",
			},
			wantErr: false,
		},
		{
			name:     "valid version with tag and a 'v' prefix",
			imageTag: "v6.9.1-202501020304",
			want: jindoVersion{
				Major: 6,
				Minor: 9,
				Patch: 1,
				Tag:   "202501020304",
			},
			wantErr: false,
		},
		{
			name:     "valid version with complex tag",
			imageTag: "4.5.2-community-edition",
			want: jindoVersion{
				Major: 4,
				Minor: 5,
				Patch: 2,
				Tag:   "community-edition",
			},
			wantErr: false,
		},
		{
			name:     "invalid version format",
			imageTag: "invalid.version",
			want:     jindoVersion{},
			wantErr:  true,
		},
		{
			name:     "empty version",
			imageTag: "",
			want:     jindoVersion{},
			wantErr:  true,
		},
		{
			name:     "version with spaces",
			imageTag: " 4.2.1-beta ",
			want: jindoVersion{
				Major: 4,
				Minor: 2,
				Patch: 1,
				Tag:   "beta",
			},
			wantErr: false,
		},
		{
			name:     "non-numeric major version",
			imageTag: "x.1.2",
			want:     jindoVersion{},
			wantErr:  true,
		},
		{
			name:     "non-numeric minor version",
			imageTag: "1.x.2",
			want:     jindoVersion{},
			wantErr:  true,
		},
		{
			name:     "non-numeric patch version",
			imageTag: "1.2.x",
			want:     jindoVersion{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVersionFromImageTag(tt.imageTag)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVersionFromImageTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Major != tt.want.Major || got.Minor != tt.want.Minor || got.Patch != tt.want.Patch || got.Tag != tt.want.Tag) {
				t.Errorf("parseVersionFromImageTag() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
