/*

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

func TestWorkerSidecarEnabled(t *testing.T) {
	type testCase struct {
		name        string
		annotations map[string]string
		expect      bool
	}

	testcases := []testCase{
		{
			name: "enable_worker",
			annotations: map[string]string{
				common.WorkerSidecar: "true",
			},
			expect: true,
		}, {
			name: "disable_worker",
			annotations: map[string]string{
				common.WorkerSidecar: "false",
			},
			expect: false,
		}, {
			name: "no_worker",
			annotations: map[string]string{
				"test": "false",
			},
			expect: false,
		},
	}

	for _, testcase := range testcases {
		got := WorkerSidecarEnabled(testcase.annotations)
		if got != testcase.expect {
			t.Errorf("The testcase %s's failed due to expect %v but got %v", testcase.name, testcase.expect, got)
		}
	}
}

func TestFuseSidecarEnabled(t *testing.T) {
	type testCase struct {
		name        string
		annotations map[string]string
		expect      bool
	}

	testcases := []testCase{
		{
			name: "enable_fuse",
			annotations: map[string]string{
				common.FuseSidecar: "true",
			},
			expect: true,
		}, {
			name: "disable_fuse",
			annotations: map[string]string{
				common.FuseSidecar: "false",
			},
			expect: false,
		}, {
			name: "no_fuse",
			annotations: map[string]string{
				"test": "false",
			},
			expect: false,
		},
	}

	for _, testcase := range testcases {
		got := FuseSidecarEnabled(testcase.annotations)
		if got != testcase.expect {
			t.Errorf("The testcase %s's failed due to expect %v but got %v", testcase.name, testcase.expect, got)
		}
	}
}
