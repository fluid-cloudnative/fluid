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
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestEnabled(t *testing.T) {
	type testCase struct {
		name        string
		annotations map[string]string
		key         string
		expect      bool
	}

	testcases := []testCase{
		{
			name: "include_key",
			annotations: map[string]string{
				"mytest": "true",
			},
			key:    "mytest",
			expect: true,
		}, {
			name: "exclude_key",
			annotations: map[string]string{
				"mytest": "true",
			},
			key:    "my",
			expect: false,
		},
	}

	for _, testcase := range testcases {
		got := enabled(testcase.annotations, testcase.key)
		if got != testcase.expect {
			t.Errorf("The testcase %s's failed due to expect %v but got %v", testcase.name, testcase.expect, got)
		}
	}
}

func TestServerlessEnabled(t *testing.T) {
	type testCase struct {
		name        string
		annotations map[string]string
		expect      bool
	}

	DeprecatedServerlessPlatformKey = "serverless.fluid.io/platform"

	testcases := []testCase{
		{
			name: "enable_Serverless",
			annotations: map[string]string{
				common.InjectServerless: "true",
			},
			expect: true,
		}, {
			name: "enable_Serverless_2",
			annotations: map[string]string{
				common.InjectFuseSidecar: "true",
			},
			expect: true,
		}, {
			name: "disable_Serverless",
			annotations: map[string]string{
				common.InjectServerless: "false",
			},
			expect: false,
		}, {
			name: "no_Serverless",
			annotations: map[string]string{
				"test": "false",
			},
			expect: false,
		},
		{
			name: "support_ask_platform",
			annotations: map[string]string{
				"serverless.fluid.io/platform": "foo",
			},
			expect: true,
		},
	}

	for _, testcase := range testcases {
		got := ServerlessEnabled(testcase.annotations)
		if got != testcase.expect {
			t.Errorf("The testcase %s's failed due to expect %v but got %v", testcase.name, testcase.expect, got)
		}
	}
}

func TestInjectionEnabled(t *testing.T) {
	type testCase struct {
		name        string
		annotations map[string]string
		expect      bool
	}

	testcases := []testCase{
		{
			name: "enable_Injection_done",
			annotations: map[string]string{
				common.InjectSidecarDone: "true",
			},
			expect: true,
		}, {
			name: "disable_Injection_done",
			annotations: map[string]string{
				common.InjectSidecarDone: "false",
			},
			expect: false,
		}, {
			name: "no_Injection",
			annotations: map[string]string{
				"test": "false",
			},
			expect: false,
		},
	}

	for _, testcase := range testcases {
		got := InjectSidecarDone(testcase.annotations)
		if got != testcase.expect {
			t.Errorf("The testcase %s's failed due to expect %v but got %v", testcase.name, testcase.expect, got)
		}
	}
}

func TestCacheDirInjectionEnabled(t *testing.T) {
	type testCase struct {
		name        string
		annotations map[string]string
		expect      bool
	}

	testcases := []testCase{
		{
			name: "enable_Injection_done",
			annotations: map[string]string{
				common.InjectCacheDir: "true",
			},
			expect: true,
		}, {
			name: "disable_Injection_done",
			annotations: map[string]string{
				common.InjectCacheDir: "false",
			},
			expect: false,
		}, {
			name: "no_Injection",
			annotations: map[string]string{
				"test": "false",
			},
			expect: false,
		},
	}

	for _, testcase := range testcases {
		got := InjectCacheDirEnabled(testcase.annotations)
		if got != testcase.expect {
			t.Errorf("The testcase %s's failed due to expect %v but got %v", testcase.name, testcase.expect, got)
		}
	}
}

func TestServerlessPlatformMatched(t *testing.T) {
	type envPlatform struct {
		ServerlessPlatformKey string
		ServerlessPlatformVal string
	}
	tests := []struct {
		name      string
		infos     map[string]string
		envs      *envPlatform
		wantMatch bool
	}{
		{
			name:  "test_default_platform",
			infos: map[string]string{"serverless.fluid.io/platform": "test"},
			envs: &envPlatform{
				ServerlessPlatformKey: "",
				ServerlessPlatformVal: "",
			},
			wantMatch: false,
		},
		{
			name:  "test_platform_env_set",
			infos: map[string]string{"serverless.fluid.io/platform": "test"},
			envs: &envPlatform{
				ServerlessPlatformKey: "serverless.fluid.io/platform",
				ServerlessPlatformVal: "test",
			},
			wantMatch: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envs != nil {
				DeprecatedServerlessPlatformKey = tt.envs.ServerlessPlatformKey
			}
			if gotMatch := serverlessPlatformMatched(tt.infos); gotMatch != tt.wantMatch {
				t.Errorf("ServerlessPlatformMatched() = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

func TestAppControllerDisabled(t *testing.T) {
	tests := []struct {
		name      string
		infos     map[string]string
		key       string
		wantMatch bool
	}{
		{
			name:      "test_default_platform",
			infos:     map[string]string{"disabled.fluid.io/platform": "test"},
			key:       "",
			wantMatch: false,
		},
		{
			name:      "test_platform_env_set",
			infos:     map[string]string{"serverless.fluid.io/platform": "test"},
			key:       "serverless.fluid.io/platform",
			wantMatch: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disableApplicationController = tt.key
			if gotMatch := AppControllerDisabled(tt.infos); gotMatch != tt.wantMatch {
				t.Errorf("AppControllerDisabled() = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}
