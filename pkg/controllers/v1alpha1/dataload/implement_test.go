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

package dataload

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func TestIsTargetPathUnderFluidNativeMounts(t *testing.T) {
	type args struct {
		targetPath string
		dataset    v1alpha1.Dataset
	}

	mockDataset := func(name, mountPoint, path string) v1alpha1.Dataset {
		return v1alpha1.Dataset{
			Spec: v1alpha1.DatasetSpec{
				Mounts: []v1alpha1.Mount{
					{
						Name:       name,
						MountPoint: mountPoint,
						Path:       path,
					},
				},
			},
		}
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_no_fluid_native",
			args: args{
				targetPath: "/imagenet",
				dataset:    mockDataset("imagenet", "oss://imagenet-data/", ""),
			},
			want: false,
		},
		{
			name: "test_pvc",
			args: args{
				targetPath: "/imagenet",
				dataset:    mockDataset("imagenet", "pvc://nfs-imagenet", ""),
			},
			want: true,
		},
		{
			name: "test_hostpath",
			args: args{
				targetPath: "/imagenet",
				dataset:    mockDataset("imagenet", "local:///hostpath_imagenet", ""),
			},
			want: true,
		},
		{
			name: "test_target_subpath",
			args: args{
				targetPath: "/imagenet/data/train",
				dataset:    mockDataset("imagenet", "pvc://nfs-imagenet", ""),
			},
			want: true,
		},
		{
			name: "test_mount_path",
			args: args{
				targetPath: "/dataset/data/train",
				dataset:    mockDataset("imagenet", "pvc://nfs-imagenet", "/dataset"),
			},
			want: true,
		},
		{
			name: "test_other",
			args: args{
				targetPath: "/dataset",
				dataset:    mockDataset("imagenet", "pvc://nfs-imagenet", ""),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsTargetPathUnderFluidNativeMounts(tt.args.targetPath, tt.args.dataset); got != tt.want {
				t.Errorf("isTargetPathUnderFluidNativeMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}
