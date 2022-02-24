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
)

func TestTrimVolumes(t *testing.T) {
	testCases := map[string]struct {
		volumes []corev1.Volume
		names   []string
		wants   []string
	}{
		"no exlude": {
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
		}, "exlude": {
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
			wants: []string{"datavolume-1", "fuse-device", "jindofs-fuse-mount"},
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

}
