/*
Copyright 2020 The Fluid Authors.

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
