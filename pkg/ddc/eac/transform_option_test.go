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

package eac

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformMasterOptions(t *testing.T) {
	runtime := &datav1alpha1.EFCRuntime{
		Spec: datav1alpha1.EFCRuntimeSpec{
			Master: datav1alpha1.EFCCompTemplateSpec{
				Properties: map[string]string{
					"a": "b",
				},
			},
		},
	}
	value := &EAC{}
	engine := &EACEngine{
		name:      "test",
		namespace: "fluid",
	}
	err := engine.transformMasterOptions(runtime, value)
	if err != nil {
		t.Errorf("unexpected err %v", err)
	}
	if value.Master.Options != "client_owner=fluid-test-master,assign_uuid=fluid-test-master,a=b" {
		t.Errorf("unexpected option %v", value.Master.Options)
	}
}

func TestTransformFuseOptions(t *testing.T) {
	runtime := &datav1alpha1.EFCRuntime{
		Spec: datav1alpha1.EFCRuntimeSpec{
			Worker: datav1alpha1.EFCCompTemplateSpec{
				Disabled: false,
			},
			Fuse: datav1alpha1.EFCFuseSpec{
				Properties: map[string]string{
					"a": "b",
				},
			},
		},
	}
	value := &EAC{}
	engine := &EACEngine{
		name:      "test",
		namespace: "fluid",
	}
	err := engine.transformFuseOptions(runtime, value)
	if err != nil {
		t.Errorf("unexpected err %v", err)
	}
	if value.Fuse.Options != "g_tier_EnableClusterCache=true,g_tier_EnableClusterCachePrefetch=true,assign_uuid=fluid-test-fuse,a=b" {
		t.Errorf("unexpected option %v", value.Fuse.Options)
	}

	runtime.Spec.Worker.Disabled = true
	err = engine.transformFuseOptions(runtime, value)
	if err != nil {
		t.Errorf("unexpected err %v", err)
	}
	if value.Fuse.Options != "assign_uuid=fluid-test-fuse,a=b" {
		t.Errorf("unexpected option %v", value.Fuse.Options)
	}
}

func TestTransformWorkerOptions(t *testing.T) {
	engine := EACEngine{
		name:      "test",
		namespace: "fluid",
	}
	var tests = []struct {
		runtime     *datav1alpha1.EFCRuntime
		eacValue    *EAC
		wantError   bool
		wantOptions string
	}{
		{
			runtime: &datav1alpha1.EFCRuntime{
				Spec: datav1alpha1.EFCRuntimeSpec{
					Worker: datav1alpha1.EFCCompTemplateSpec{
						Properties: map[string]string{
							"a": "b",
						},
					},
				},
			},
			eacValue: &EAC{
				Worker: Worker{
					TieredStore: TieredStore{
						Levels: []Level{
							engine.getDefaultTiredStoreLevel0(),
						},
					},
				},
			},
			wantError:   false,
			wantOptions: "cache_media=/cache_dir/fluid/test,server_port=0,cache_capacity_gb=1,tmpfs=true,a=b",
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				Spec: datav1alpha1.EFCRuntimeSpec{
					Worker: datav1alpha1.EFCCompTemplateSpec{
						Properties: map[string]string{
							"a": "b",
						},
					},
				},
			},
			eacValue: &EAC{
				Worker: Worker{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								Level:      0,
								MediumType: "SSD",
								Type:       "emptyDir",
								Quota:      "2Gi",
								Path:       "/test",
							},
						},
					},
				},
			},
			wantError:   false,
			wantOptions: "cache_media=/test,server_port=0,cache_capacity_gb=2,a=b",
		},
		{
			runtime: &datav1alpha1.EFCRuntime{
				Spec: datav1alpha1.EFCRuntimeSpec{
					Worker: datav1alpha1.EFCCompTemplateSpec{
						Properties: map[string]string{
							"a": "b",
						},
					},
				},
			},
			eacValue: &EAC{
				Worker: Worker{
					TieredStore: TieredStore{
						Levels: []Level{
							{
								Level:      0,
								MediumType: "SSD",
								Type:       "emptyDir",
								Quota:      "2k",
								Path:       "/test",
							},
						},
					},
				},
			},
			wantError: true,
		},
	}
	for _, test := range tests {
		err := engine.transformWorkerOptions(test.runtime, test.eacValue)
		if (err == nil) != !test.wantError {
			t.Errorf("unexpected err %v", err)
		}
		if test.eacValue.Worker.Options != test.wantOptions {
			t.Errorf("want worker options: %s, got %s", test.wantOptions, test.eacValue.Worker.Options)
		}
	}
}
