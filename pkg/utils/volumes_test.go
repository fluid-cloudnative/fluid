/*
Copyright 2021 The Fluid Authors.

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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTrimVolumes(t *testing.T) {
	testCases := map[string]struct {
		volumes []corev1.Volume
		names   []string
		wants   []string
	}{
		"no exclude": {
			volumes: []corev1.Volume{
				{
					Name: "test-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/runtime_mnt/dataset1",
						},
					}},
				{
					Name: "fuse-device",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/dev/fuse",
						},
					},
				},
				{
					Name: "jindofs-fuse-mount",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/runtime-mnt/jindo/big-data/dataset1",
						},
					},
				},
			},
			names: []string{"datavolume-", "cache-dir", "mem", "ssd", "hdd"},
			wants: []string{"test-1", "fuse-device", "jindofs-fuse-mount"},
		}, "exclude": {
			volumes: []corev1.Volume{
				{
					Name: "datavolume-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/runtime_mnt/dataset1",
						},
					}},
				{
					Name: "fuse-device",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/dev/fuse",
						},
					},
				},
				{
					Name: "jindofs-fuse-mount",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/runtime-mnt/jindo/big-data/dataset1",
						},
					},
				},
			},
			names: []string{"datavolume-", "cache-dir", "mem", "ssd", "hdd"},
			wants: []string{"fuse-device", "jindofs-fuse-mount"},
		},
	}

	for name, testCase := range testCases {
		got := TrimVolumes(testCase.volumes, testCase.names)
		gotNames := []string{}
		for _, name := range got {
			gotNames = append(gotNames, name.Name)
		}

		if !reflect.DeepEqual(gotNames, testCase.wants) {
			t.Errorf("%s check failure, want:%v, got:%v", name, testCase.names, gotNames)
		}
	}
}

func TestTrimVolumeMounts(t *testing.T) {
	testCases := map[string]struct {
		volumeMounts []corev1.VolumeMount
		names        []string
		wants        []string
	}{
		"no exclude": {
			volumeMounts: []corev1.VolumeMount{
				{
					Name: "test-1",
				},
				{
					Name: "fuse-device",
				},
				{
					Name: "jindofs-fuse-mount",
				},
			},
			names: []string{"datavolumeMount-", "cache-dir", "mem", "ssd", "hdd"},
			wants: []string{"test-1", "fuse-device", "jindofs-fuse-mount"},
		}, "exclude": {
			volumeMounts: []corev1.VolumeMount{
				{
					Name: "datavolumeMount-1",
				},
				{
					Name: "fuse-device",
				},
				{
					Name: "jindofs-fuse-mount",
				},
			},
			names: []string{"datavolumeMount-", "cache-dir", "mem", "ssd", "hdd"},
			wants: []string{"fuse-device", "jindofs-fuse-mount"},
		},
	}

	for name, testCase := range testCases {
		got := TrimVolumeMounts(testCase.volumeMounts, testCase.names)
		gotNames := []string{}
		for _, name := range got {
			gotNames = append(gotNames, name.Name)
		}

		if !reflect.DeepEqual(gotNames, testCase.wants) {
			t.Errorf("%s check failure, want:%v, got:%v", name, testCase.names, gotNames)
		}
	}
}

func TestAppendOrOverrideVolume(t *testing.T) {
	sizeLimitResource := resource.MustParse("10Gi")

	type args struct {
		volumes []corev1.Volume
		vol     corev1.Volume
	}
	tests := []struct {
		name string
		args args
		want []corev1.Volume
	}{
		{
			name: "volume_not_existed",
			args: args{
				volumes: []corev1.Volume{
					corev1.Volume{
						Name: "vol-1",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium:    corev1.StorageMediumMemory,
								SizeLimit: &sizeLimitResource,
							},
						},
					},
				},
				vol: corev1.Volume{
					Name: "new-vol",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/path/to/hostdir",
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "vol-1",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: &sizeLimitResource,
						},
					},
				},
				{
					Name: "new-vol",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/path/to/hostdir",
						},
					},
				},
			},
		},
		{
			name: "volume_existed",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "vol-1",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium:    corev1.StorageMediumMemory,
								SizeLimit: &sizeLimitResource,
							},
						},
					},
				},
				vol: corev1.Volume{
					Name: "vol-1",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: &sizeLimitResource,
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "vol-1",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium:    corev1.StorageMediumMemory,
							SizeLimit: &sizeLimitResource,
						},
					},
				},
			},
		},
		{
			name: "volume_overridden",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "vol-1",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium:    corev1.StorageMediumMemory,
								SizeLimit: &sizeLimitResource,
							},
						},
					},
				},
				vol: corev1.Volume{
					Name: "vol-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/path/to/hostdir",
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "vol-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/path/to/hostdir",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppendOrOverrideVolume(tt.args.volumes, tt.args.vol); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendOrOverrideVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendOrOverrideVolumeMounts(t *testing.T) {
	type args struct {
		volumeMounts []corev1.VolumeMount
		vm           corev1.VolumeMount
	}
	tests := []struct {
		name string
		args args
		want []corev1.VolumeMount
	}{
		{
			name: "volume_mount_not_exists",
			args: args{
				volumeMounts: []corev1.VolumeMount{
					{
						Name:      "vol-1",
						MountPath: "/path/to/container-dir",
					},
				},
				vm: corev1.VolumeMount{
					Name:      "new-vol",
					MountPath: "/path/to/container/new-vol",
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
				{
					Name:      "new-vol",
					MountPath: "/path/to/container/new-vol",
				},
			},
		},
		{
			name: "volume_mount_existed",
			args: args{
				volumeMounts: []corev1.VolumeMount{
					{
						Name:      "vol-1",
						MountPath: "/path/to/container-dir",
					},
				},
				vm: corev1.VolumeMount{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
			},
		},
		{
			name: "volume_mount_overridden",
			args: args{
				volumeMounts: []corev1.VolumeMount{
					{
						Name:      "vol-1",
						MountPath: "/path/to/container-dir",
					},
				},
				vm: corev1.VolumeMount{
					Name:      "vol-1",
					MountPath: "/path/to/container/vol-1",
					ReadOnly:  true,
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container/vol-1",
					ReadOnly:  true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AppendOrOverrideVolumeMounts(tt.args.volumeMounts, tt.args.vm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppendOrOverrideVolumeMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}
