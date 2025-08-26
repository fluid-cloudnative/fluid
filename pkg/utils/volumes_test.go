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

func TestFilterVolumesByVolumeMounts(t *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
	}
	tests := []struct {
		name string
		args args
		want []corev1.Volume
	}{
		{
			name: "all_volumes_needed",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "test-vol-1",
					},
					{
						Name: "test-vol-2",
					},
					{
						Name: "test-vol-3",
					},
				},
				volumeMounts: []corev1.VolumeMount{
					{
						Name: "test-vol-1",
					},
					{
						Name: "test-vol-2",
					},
					{
						Name: "test-vol-3",
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-vol-1",
				},
				{
					Name: "test-vol-2",
				},
				{
					Name: "test-vol-3",
				},
			},
		},
		{
			name: "volumes_partly_needed",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "test-vol-1",
					},
					{
						Name: "test-vol-2",
					},
					{
						Name: "test-vol-3",
					},
				},
				volumeMounts: []corev1.VolumeMount{
					{
						Name: "test-vol-1",
					},
					{
						Name: "test-vol-2",
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "test-vol-1",
				},
				{
					Name: "test-vol-2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterVolumesByVolumeMounts(tt.args.volumes, tt.args.volumeMounts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterVolumesByVolumeMounts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVolumesDifference(t *testing.T) {
	volume1 := corev1.Volume{Name: "volume1"}
	volume2 := corev1.Volume{Name: "volume2"}
	volume3 := corev1.Volume{Name: "volume3"}
	volume4 := corev1.Volume{Name: "volume4"}

	tests := []struct {
		name     string
		base     []corev1.Volume
		filter   []corev1.Volume
		expected []corev1.Volume
	}{
		{
			name:     "nil_volumes",
			base:     []corev1.Volume{},
			filter:   []corev1.Volume{},
			expected: []corev1.Volume{},
		},
		{
			name:     "nil_base_not_nil_exclude",
			base:     []corev1.Volume{},
			filter:   []corev1.Volume{volume1, volume2},
			expected: []corev1.Volume{},
		},
		{
			name:     "not_nil_base_nil_exclude",
			base:     []corev1.Volume{volume1, volume2},
			filter:   []corev1.Volume{},
			expected: []corev1.Volume{volume1, volume2},
		},
		{
			name:     "same_base_and_exclude",
			base:     []corev1.Volume{volume1, volume2},
			filter:   []corev1.Volume{volume1, volume2},
			expected: []corev1.Volume{},
		},
		{
			name:     "base_include_all_exclude",
			base:     []corev1.Volume{volume1, volume2, volume3, volume4},
			filter:   []corev1.Volume{volume2, volume4},
			expected: []corev1.Volume{volume1, volume3},
		},
		{
			name:     "base_include_no_exclude",
			base:     []corev1.Volume{volume1, volume2},
			filter:   []corev1.Volume{volume3, volume4},
			expected: []corev1.Volume{volume1, volume2},
		},
		{
			name:     "base_include_partial_exclude",
			base:     []corev1.Volume{volume1, volume2, volume3},
			filter:   []corev1.Volume{volume2, volume4},
			expected: []corev1.Volume{volume1, volume3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVolumesDifference(tt.base, tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("expected slice length %d, actual slice length %d", len(tt.expected), len(result))
				return
			}
			expectedMap := make(map[string]bool)
			for _, v := range tt.expected {
				expectedMap[v.Name] = true
			}

			resultMap := make(map[string]bool)
			for _, v := range result {
				resultMap[v.Name] = true
			}
			for name, expectedVolume := range expectedMap {
				resultVolume, exist := resultMap[name]
				if !exist {
					t.Errorf("expected Volume %s, but not exist in return", name)
				}

				if !reflect.DeepEqual(resultVolume, expectedVolume) {
					t.Errorf("expected Volume %v, but got %v", expectedVolume, resultVolume)
				}
			}
		})
	}
}

func TestGetVolumeMountsDifference(t *testing.T) {
	volumeMount1 := corev1.VolumeMount{Name: "volumeMount1"}
	volumeMount2 := corev1.VolumeMount{Name: "volumeMount2"}
	volumeMount3 := corev1.VolumeMount{Name: "volumeMount3"}
	volumeMount4 := corev1.VolumeMount{Name: "volumeMount4"}

	tests := []struct {
		name     string
		base     []corev1.VolumeMount
		filter   []corev1.VolumeMount
		expected []corev1.VolumeMount
	}{
		{
			name:     "nil_volumes",
			base:     []corev1.VolumeMount{},
			filter:   []corev1.VolumeMount{},
			expected: []corev1.VolumeMount{},
		},
		{
			name:     "nil_base_not_nil_exclude",
			base:     []corev1.VolumeMount{},
			filter:   []corev1.VolumeMount{volumeMount1, volumeMount2},
			expected: []corev1.VolumeMount{},
		},
		{
			name:     "not_nil_base_nil_exclude",
			base:     []corev1.VolumeMount{volumeMount1, volumeMount2},
			filter:   []corev1.VolumeMount{},
			expected: []corev1.VolumeMount{volumeMount1, volumeMount2},
		},
		{
			name:     "same_base_and_exclude",
			base:     []corev1.VolumeMount{volumeMount1, volumeMount2},
			filter:   []corev1.VolumeMount{volumeMount1, volumeMount2},
			expected: []corev1.VolumeMount{},
		},
		{
			name:     "base_include_all_exclude",
			base:     []corev1.VolumeMount{volumeMount1, volumeMount2, volumeMount3, volumeMount4},
			filter:   []corev1.VolumeMount{volumeMount2, volumeMount4},
			expected: []corev1.VolumeMount{volumeMount1, volumeMount3},
		},
		{
			name:     "base_include_no_exclude",
			base:     []corev1.VolumeMount{volumeMount1, volumeMount2},
			filter:   []corev1.VolumeMount{volumeMount3, volumeMount4},
			expected: []corev1.VolumeMount{volumeMount1, volumeMount2},
		},
		{
			name:     "base_include_partial_exclude",
			base:     []corev1.VolumeMount{volumeMount1, volumeMount2, volumeMount3},
			filter:   []corev1.VolumeMount{volumeMount2, volumeMount4},
			expected: []corev1.VolumeMount{volumeMount1, volumeMount3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetVolumeMountsDifference(tt.base, tt.filter)
			if len(result) != len(tt.expected) {
				t.Errorf("expected slice length %d, actual slice length %d", len(tt.expected), len(result))
				return
			}
			expectedMap := make(map[string]bool)
			for _, v := range tt.expected {
				expectedMap[v.Name] = true
			}

			resultMap := make(map[string]bool)
			for _, v := range result {
				resultMap[v.Name] = true
			}
			for name, expectedVolumeMount := range expectedMap {
				resultVolumeMount, exist := resultMap[name]
				if !exist {
					t.Errorf("expected VolumeMount %s, but not exist in return", name)
				}

				if !reflect.DeepEqual(resultVolumeMount, expectedVolumeMount) {
					t.Errorf("expected VolumeMount %v, but got %v", expectedVolumeMount, resultVolumeMount)
				}
			}
		})
	}
}
