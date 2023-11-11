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

package efc

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
	value := &EFC{}
	engine := &EFCEngine{
		name:      "test",
		namespace: "fluid",
	}

	{
		err := engine.transformMasterOptions(runtime, value, &MountInfo{MountPointPrefix: NasMountPointPrefix})
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if value.Master.Options != "client_owner=fluid-test-master,assign_uuid=fluid-test-master,a=b" {
			t.Errorf("unexpected option %v", value.Master.Options)
		}
	}

	{
		err := engine.transformMasterOptions(runtime, value, &MountInfo{MountPointPrefix: CpfsMountPointPrefix})
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if value.Master.Options != "protocol=nfs3,client_owner=fluid-test-master,assign_uuid=fluid-test-master,a=b" {
			t.Errorf("unexpected option %v", value.Master.Options)
		}
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
	value := &EFC{}
	engine := &EFCEngine{
		name:      "test",
		namespace: "fluid",
	}

	{
		err := engine.transformFuseOptions(runtime, value, &MountInfo{MountPointPrefix: NasMountPointPrefix})
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if value.Fuse.Options != "g_tier_EnableClusterCache=true,g_tier_EnableClusterCachePrefetch=true,assign_uuid=fluid-test-fuse,a=b" {
			t.Errorf("unexpected option %v", value.Fuse.Options)
		}
	}

	{
		err := engine.transformFuseOptions(runtime, value, &MountInfo{MountPointPrefix: CpfsMountPointPrefix})
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if value.Fuse.Options != "protocol=nfs3,g_tier_EnableClusterCache=true,g_tier_EnableClusterCachePrefetch=true,assign_uuid=fluid-test-fuse,a=b" {
			t.Errorf("unexpected option %v", value.Fuse.Options)
		}
	}

	{
		runtime.Spec.Worker.Disabled = true
		err := engine.transformFuseOptions(runtime, value, &MountInfo{MountPointPrefix: NasMountPointPrefix})
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if value.Fuse.Options != "assign_uuid=fluid-test-fuse,a=b" {
			t.Errorf("unexpected option %v", value.Fuse.Options)
		}
	}

	{
		runtime.Spec.Worker.Disabled = true
		err := engine.transformFuseOptions(runtime, value, &MountInfo{MountPointPrefix: CpfsMountPointPrefix})
		if err != nil {
			t.Errorf("unexpected err %v", err)
		}
		if value.Fuse.Options != "protocol=nfs3,assign_uuid=fluid-test-fuse,a=b" {
			t.Errorf("unexpected option %v", value.Fuse.Options)
		}
	}
}

func TestTransformWorkerOptions(t *testing.T) {
	engine := EFCEngine{
		name:      "test",
		namespace: "fluid",
	}
	var tests = []struct {
		runtime     *datav1alpha1.EFCRuntime
		efcValue    *EFC
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
			efcValue: &EFC{
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
			efcValue: &EFC{
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
			efcValue: &EFC{
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
		err := engine.transformWorkerOptions(test.runtime, test.efcValue)
		if (err == nil) != !test.wantError {
			t.Errorf("unexpected err %v", err)
		}
		if test.efcValue.Worker.Options != test.wantOptions {
			t.Errorf("want worker options: %s, got %s", test.wantOptions, test.efcValue.Worker.Options)
		}
	}
}
