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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	cacheDirName          = "cache-dir"
	dataVolumePrefix      = "datavolume-"
	dataVolumeMountPrefix = "datavolumeMount-"
	memTier               = "mem"
	ssdTier               = "ssd"
	hddTier               = "hdd"
)

var _ = Describe("TrimVolumes", func() {
	DescribeTable("should trim volumes correctly",
		func(volumes []corev1.Volume, names []string, wants []string) {
			got := TrimVolumes(volumes, names)
			gotNames := []string{}
			for _, name := range got {
				gotNames = append(gotNames, name.Name)
			}
			Expect(gotNames).To(Equal(wants))
		},
		Entry("no exclude",
			[]corev1.Volume{
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
			[]string{dataVolumePrefix, cacheDirName, memTier, ssdTier, hddTier},
			[]string{"test-1", "fuse-device", "jindofs-fuse-mount"},
		),
		Entry("exclude",
			[]corev1.Volume{
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
			[]string{dataVolumePrefix, cacheDirName, memTier, ssdTier, hddTier},
			[]string{"fuse-device", "jindofs-fuse-mount"},
		),
	)
})

var _ = Describe("TrimVolumeMounts", func() {
	DescribeTable("should trim volume mounts",
		func(volumeMounts []corev1.VolumeMount, names []string, wants []string) {
			got := TrimVolumeMounts(volumeMounts, names)
			gotNames := []string{}
			for _, name := range got {
				gotNames = append(gotNames, name.Name)
			}
			Expect(gotNames).To(Equal(wants))
		},
		Entry("no exclude",
			[]corev1.VolumeMount{
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
			[]string{dataVolumeMountPrefix, cacheDirName, memTier, ssdTier, hddTier},
			[]string{"test-1", "fuse-device", "jindofs-fuse-mount"},
		),
		Entry("exclude",
			[]corev1.VolumeMount{
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
			[]string{dataVolumeMountPrefix, cacheDirName, memTier, ssdTier, hddTier},
			[]string{"fuse-device", "jindofs-fuse-mount"},
		),
	)
})

var _ = Describe("AppendOrOverrideVolume", func() {
	sizeLimitResource := resource.MustParse("10Gi")

	DescribeTable("should append or override volume",
		func(volumes []corev1.Volume, vol corev1.Volume, want []corev1.Volume) {
			got := AppendOrOverrideVolume(volumes, vol)
			Expect(got).To(Equal(want))
		},
		Entry("volume_not_existed",
			[]corev1.Volume{
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
			corev1.Volume{
				Name: "new-vol",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/path/to/hostdir",
					},
				},
			},
			[]corev1.Volume{
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
		),
		Entry("volume_existed",
			[]corev1.Volume{
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
			corev1.Volume{
				Name: "vol-1",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium:    corev1.StorageMediumMemory,
						SizeLimit: &sizeLimitResource,
					},
				},
			},
			[]corev1.Volume{
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
		),
		Entry("volume_overridden",
			[]corev1.Volume{
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
			corev1.Volume{
				Name: "vol-1",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/path/to/hostdir",
					},
				},
			},
			[]corev1.Volume{
				{
					Name: "vol-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/path/to/hostdir",
						},
					},
				},
			},
		),
	)
})

var _ = Describe("AppendOrOverrideVolumeMounts", func() {
	DescribeTable("should append or override volume mount",
		func(volumeMounts []corev1.VolumeMount, vm corev1.VolumeMount, want []corev1.VolumeMount) {
			got := AppendOrOverrideVolumeMounts(volumeMounts, vm)
			Expect(got).To(Equal(want))
		},
		Entry("volume_mount_not_exists",
			[]corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
			},
			corev1.VolumeMount{
				Name:      "new-vol",
				MountPath: "/path/to/container/new-vol",
			},
			[]corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
				{
					Name:      "new-vol",
					MountPath: "/path/to/container/new-vol",
				},
			},
		),
		Entry("volume_mount_existed",
			[]corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
			},
			corev1.VolumeMount{
				Name:      "vol-1",
				MountPath: "/path/to/container-dir",
			},
			[]corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
			},
		),
		Entry("volume_mount_overridden",
			[]corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container-dir",
				},
			},
			corev1.VolumeMount{
				Name:      "vol-1",
				MountPath: "/path/to/container/vol-1",
				ReadOnly:  true,
			},
			[]corev1.VolumeMount{
				{
					Name:      "vol-1",
					MountPath: "/path/to/container/vol-1",
					ReadOnly:  true,
				},
			},
		),
	)
})

var _ = Describe("FilterVolumesByVolumeMounts", func() {
	DescribeTable("should filter volumes by volume mounts",
		func(volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, want []corev1.Volume) {
			got := FilterVolumesByVolumeMounts(volumes, volumeMounts)
			Expect(got).To(Equal(want))
		},
		Entry("all_volumes_needed",
			[]corev1.Volume{
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
			[]corev1.VolumeMount{
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
			[]corev1.Volume{
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
		),
		Entry("volumes_partly_needed",
			[]corev1.Volume{
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
			[]corev1.VolumeMount{
				{
					Name: "test-vol-1",
				},
				{
					Name: "test-vol-2",
				},
			},
			[]corev1.Volume{
				{
					Name: "test-vol-1",
				},
				{
					Name: "test-vol-2",
				},
			},
		),
	)
})

var _ = Describe("GetVolumesDifference", func() {
	volume1 := corev1.Volume{Name: "volume1"}
	volume2 := corev1.Volume{Name: "volume2"}
	volume3 := corev1.Volume{Name: "volume3"}
	volume4 := corev1.Volume{Name: "volume4"}

	DescribeTable("should get volumes difference",
		func(base []corev1.Volume, filter []corev1.Volume, expected []corev1.Volume) {
			result := GetVolumesDifference(base, filter)
			Expect(result).To(ConsistOf(expected))
		},
		Entry("nil_volumes",
			[]corev1.Volume{},
			[]corev1.Volume{},
			[]corev1.Volume{},
		),
		Entry("nil_base_not_nil_exclude",
			[]corev1.Volume{},
			[]corev1.Volume{volume1, volume2},
			[]corev1.Volume{},
		),
		Entry("not_nil_base_nil_exclude",
			[]corev1.Volume{volume1, volume2},
			[]corev1.Volume{},
			[]corev1.Volume{volume1, volume2},
		),
		Entry("same_base_and_exclude",
			[]corev1.Volume{volume1, volume2},
			[]corev1.Volume{volume1, volume2},
			[]corev1.Volume{},
		),
		Entry("base_include_all_exclude",
			[]corev1.Volume{volume1, volume2, volume3, volume4},
			[]corev1.Volume{volume2, volume4},
			[]corev1.Volume{volume1, volume3},
		),
		Entry("base_include_no_exclude",
			[]corev1.Volume{volume1, volume2},
			[]corev1.Volume{volume3, volume4},
			[]corev1.Volume{volume1, volume2},
		),
		Entry("base_include_partial_exclude",
			[]corev1.Volume{volume1, volume2, volume3},
			[]corev1.Volume{volume2, volume4},
			[]corev1.Volume{volume1, volume3},
		),
	)
})

var _ = Describe("GetVolumeMountsDifference", func() {
	volumeMount1 := corev1.VolumeMount{Name: "volumeMount1"}
	volumeMount2 := corev1.VolumeMount{Name: "volumeMount2"}
	volumeMount3 := corev1.VolumeMount{Name: "volumeMount3"}
	volumeMount4 := corev1.VolumeMount{Name: "volumeMount4"}

	DescribeTable("should get volume mounts difference",
		func(base []corev1.VolumeMount, filter []corev1.VolumeMount, expected []corev1.VolumeMount) {
			result := GetVolumeMountsDifference(base, filter)
			Expect(result).To(ConsistOf(expected))
		},
		Entry("nil_volumes",
			[]corev1.VolumeMount{},
			[]corev1.VolumeMount{},
			[]corev1.VolumeMount{},
		),
		Entry("nil_base_not_nil_exclude",
			[]corev1.VolumeMount{},
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
			[]corev1.VolumeMount{},
		),
		Entry("not_nil_base_nil_exclude",
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
			[]corev1.VolumeMount{},
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
		),
		Entry("same_base_and_exclude",
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
			[]corev1.VolumeMount{},
		),
		Entry("base_include_all_exclude",
			[]corev1.VolumeMount{volumeMount1, volumeMount2, volumeMount3, volumeMount4},
			[]corev1.VolumeMount{volumeMount2, volumeMount4},
			[]corev1.VolumeMount{volumeMount1, volumeMount3},
		),
		Entry("base_include_no_exclude",
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
			[]corev1.VolumeMount{volumeMount3, volumeMount4},
			[]corev1.VolumeMount{volumeMount1, volumeMount2},
		),
		Entry("base_include_partial_exclude",
			[]corev1.VolumeMount{volumeMount1, volumeMount2, volumeMount3},
			[]corev1.VolumeMount{volumeMount2, volumeMount4},
			[]corev1.VolumeMount{volumeMount1, volumeMount3},
		),
	)
})
