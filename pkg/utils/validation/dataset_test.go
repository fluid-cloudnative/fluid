/*
Copyright 2024 The Fluid Author.

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

package validation

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const mountPoint1 string = "https://mirrors.bit.edu.cn/apache/spark/"
const validMountName1 string = "spark"

const mountPoint2 string = "https://mirrors.bit.edu.cn/apache/flink/"
const validMountName2 string = "flink"

const validMountPath1 string = "/test"
const validMountPath2 string = "mnt/test"

func TestIsValidDatasetWithValidDataset(t *testing.T) {
	type testCase struct {
		name                  string
		input                 v1alpha1.Dataset
		enableMountValidation bool
	}

	testCases := []testCase{
		{
			name:                  "validDatasetWithSingleMount",
			enableMountValidation: true,
			input: v1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name: "demo",
				},
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Name:       validMountName1,
							Path:       validMountPath1,
						},
					},
				},
			},
		},
		{
			name:                  "validDatasetWithMultiMount",
			enableMountValidation: true,
			input: v1alpha1.Dataset{
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Name:       validMountName1,
						},
						{
							MountPoint: mountPoint2,
							Name:       validMountName2,
							Path:       validMountPath2,
						},
					},
				},
			},
		},
		{
			name:                  "validDatasetWithDisableMountValidation",
			enableMountValidation: false,
			input: v1alpha1.Dataset{
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Path:       "/${TEST}",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		got := IsValidDataset(test.input, test.enableMountValidation)
		if got != nil {
			t.Errorf("testcase %s failed, expect no error happened, but got an error: %s", test.name, got.Error())
		}
	}
}

func TestIsValidDatasetWithInvalidDataset(t *testing.T) {
	type testCase struct {
		name  string
		input v1alpha1.Dataset
	}

	testCases := []testCase{
		{
			name: "invalidDatasetMountName",
			input: v1alpha1.Dataset{
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Name:       "$(cat /etc/passwd > /test.txt)",
							Path:       validMountPath1,
						},
					},
				},
			},
		},
		{
			name: "invalidDatasetName",
			input: v1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name: "20-hbase",
				},
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Name:       validMountName1,
							Path:       validMountPath1,
						},
					},
				},
			},
		},
		{
			name: "invalidDatasetMountPath",
			input: v1alpha1.Dataset{
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Name:       validMountName1,
							Path:       "/$(cat /etc/passwd > /test.txt)",
						},
					},
				},
			},
		},
		{
			name: "invalidDatasetMountPathInSecondMount",
			input: v1alpha1.Dataset{
				Spec: v1alpha1.DatasetSpec{
					Mounts: []v1alpha1.Mount{
						{
							MountPoint: mountPoint1,
							Name:       validMountName1,
						},
						{
							MountPoint: mountPoint2,
							Name:       validMountName2,
							Path:       "/test/$(cat /etc/passwd > /test.txt)",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		got := IsValidDataset(test.input, true)
		if got == nil {
			t.Errorf("testcase %s failed, expect an error happened, but got no error", test.name)
		}
	}
}
