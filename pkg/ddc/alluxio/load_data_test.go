package alluxio

import (
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"testing"
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
			if got := isTargetPathUnderFluidNativeMounts(tt.args.targetPath, tt.args.dataset); got != tt.want {
				t.Errorf("isTargetPathUnderFluidNativeMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}
