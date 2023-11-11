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

package kubeclient

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func TestGetFuseMountInContainer(t *testing.T) {

	tests := []struct {
		name      string
		mountType string
		container corev1.Container
		want      string
	}{
		{
			mountType: common.JindoMountType,
			container: corev1.Container{
				Env: []corev1.EnvVar{
					{
						Name:  "FLUID_FUSE_MOUNTPOINT",
						Value: "/jfs/jindofs-fuse",
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "jindofs-fuse-mount",
						MountPath: "/jfs",
					},
				},
			},
			want: "/jfs",
		}, {
			mountType: common.AlluxioMountType,
			container: corev1.Container{
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "alluxio-fuse-mount",
						MountPath: "/runtime_mnt/alluxio",
					},
				},
			},
			want: "/runtime_mnt/alluxio",
		}, {
			mountType: common.JuiceFSMountType,
			container: corev1.Container{
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "juicefs-fuse-mount",
						MountPath: "/runtime_mnt/jfs",
					},
				},
			},
			want: "/runtime_mnt/jfs",
		}, {
			mountType: "nfs",
			container: corev1.Container{
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "thin-fuse-mount",
						MountPath: "/runtime_mnt/thin",
					},
				},
			},
			want: "/runtime_mnt/thin",
		},
	}

	for _, test := range tests {
		v, err := GetFuseMountInContainer(test.mountType, test.container)
		if err != nil {
			t.Errorf("testcase %v GetFuseMountInContainer() got error %v", test.name, err)
		}

		if v.MountPath != test.want {
			t.Errorf("testcase %v GetFuseMountInContainer() got mountPath %v, but got %v", test.name, test.want, v.MountPath)
		}
	}
}

func TestPVCNames(t *testing.T) {
	tests := []struct {
		name         string
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		want         []string
	}{
		{
			name: "nopvc",
			volumes: []corev1.Volume{
				{
					Name: "duplicate",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
						},
					},
				},
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
							Path: "/runtime-mnt/jindo/big-data/duplicate",
						},
					},
				}, {
					Name: "fluid-ate",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/mnt/disk1",
						},
					},
				}, {
					Name: "check-mount",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "duplicate-jindo-check-mount",
							},
						},
					},
				},
			}, volumeMounts: []corev1.VolumeMount{{
				Name:      "fluid-ate",
				MountPath: "/mnt/disk1",
			}, {
				Name:      "fuse-device",
				MountPath: "/dev/fuse",
			}, {
				Name:      "jindofs-fuse-mount",
				MountPath: "/jfs",
			}, {
				Name:      "check-mount",
				ReadOnly:  true,
				MountPath: "/check-mount.sh",
				SubPath:   "check-mount.sh",
			}},
			want: []string{},
		}, {
			name: "pvc",
			volumes: []corev1.Volume{
				{
					Name: "duplicate",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
						},
					},
				},
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
							Path: "/runtime-mnt/jindo/big-data/duplicate",
						},
					},
				}, {
					Name: "pvc",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "dataset1",
							ReadOnly:  true,
						},
					},
				}, {
					Name: "check-mount",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "duplicate-jindo-check-mount",
							},
						},
					},
				},
			}, volumeMounts: []corev1.VolumeMount{{
				Name:      "pvc",
				MountPath: "/mnt/disk1",
			}, {
				Name:      "fuse-device",
				MountPath: "/dev/fuse",
			}, {
				Name:      "jindofs-fuse-mount",
				MountPath: "/jfs",
			}, {
				Name:      "check-mount",
				ReadOnly:  true,
				MountPath: "/check-mount.sh",
				SubPath:   "check-mount.sh",
			}},
			want: []string{"dataset1"},
		},
	}

	for _, test := range tests {
		got := PVCNames(test.volumeMounts, test.volumes)

		if !checkEqual(got, test.want) {
			t.Errorf("test %s failed, want %v, got %v", test.name, test.want, got)
		}
	}
}

func checkEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetMountPathInContainer(t *testing.T) {
	type args struct {
		container corev1.Container
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test-juicefs",
			args: args{
				container: corev1.Container{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "juicefs-fuse-mount",
						MountPath: "/runtime-mnt/juicefs/default/test",
					}},
				},
			},
			want:    "/runtime-mnt/juicefs/default/test/juicefs-fuse",
			wantErr: false,
		},
		{
			name: "test-jindofs",
			args: args{
				container: corev1.Container{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "jindo-fuse-mount",
						MountPath: "/test",
					}},
					Env: []corev1.EnvVar{{
						Name:  common.FuseMountEnv,
						Value: "/test/jfs",
					}},
				},
			},
			want:    "/test/jfs",
			wantErr: false,
		},
		{
			name: "test-goosefs",
			args: args{
				container: corev1.Container{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "goosefs-fuse-mount",
						MountPath: "/runtime-mnt/goosefs/default/test",
					}},
				},
			},
			want:    "/runtime-mnt/goosefs/default/test/goosefs-fuse",
			wantErr: false,
		},
		{
			name: "test-alluxio",
			args: args{
				container: corev1.Container{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "alluxio-fuse-mount",
						MountPath: "/runtime-mnt/alluxio/default/test",
					}},
				},
			},
			want:    "/runtime-mnt/alluxio/default/test/alluxio-fuse",
			wantErr: false,
		},
		{
			name: "test-wrong",
			args: args{
				container: corev1.Container{
					Name: "test",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "test-fuse-mount",
						MountPath: "/runtime-mnt/juicefs/default/test",
					}},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "test-no-mount",
			args: args{
				container: corev1.Container{
					Name: "test",
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMountPathInContainer(tt.args.container)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMountPathInContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMountPathInContainer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
