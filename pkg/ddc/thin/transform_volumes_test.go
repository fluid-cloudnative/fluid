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

package thin

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestThinEngine_transformFuseVolumes(t1 *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		value        *ThinValue
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantVolumes      int
		wantVolumeMounts int
	}{
		{
			name: "test-normal",
			args: args{
				volumes: []corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}},
				volumeMounts: []corev1.VolumeMount{{
					Name:      "vol1",
					MountPath: "/data",
				}},
				value: &ThinValue{},
			},
			wantErr:          false,
			wantVolumes:      1,
			wantVolumeMounts: 1,
		},
		{
			name: "test-err",
			args: args{
				volumes: []corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}},
				volumeMounts: []corev1.VolumeMount{{
					Name:      "vol2",
					MountPath: "/data",
				}},
				value: &ThinValue{},
			},
			wantErr:          true,
			wantVolumes:      0,
			wantVolumeMounts: 0,
		},
		{
			name: "test-2",
			args: args{
				volumes: []corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}, {
					Name: "vol2",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}},
				volumeMounts: []corev1.VolumeMount{{
					Name:      "vol2",
					MountPath: "/data",
				}},
				value: &ThinValue{},
			},
			wantErr:          false,
			wantVolumes:      1,
			wantVolumeMounts: 1,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{}
			if err := t.transformFuseVolumes(tt.args.volumes, tt.args.volumeMounts, tt.args.value); (err != nil) != tt.wantErr {
				t1.Errorf("transformFuseVolumes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tt.args.value.Fuse.VolumeMounts) != tt.wantVolumeMounts ||
				len(tt.args.value.Fuse.Volumes) != tt.wantVolumes {
				t1.Errorf("values %v", tt.args.value)
			}
		})
	}
}

func TestThinEngine_transformWorkerVolumes(t1 *testing.T) {
	type args struct {
		volumes      []corev1.Volume
		volumeMounts []corev1.VolumeMount
		value        *ThinValue
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantVolumes      int
		wantVolumeMounts int
	}{
		{
			name: "test-normal",
			args: args{
				volumes: []corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}},
				volumeMounts: []corev1.VolumeMount{{
					Name:      "vol1",
					MountPath: "/data",
				}},
				value: &ThinValue{},
			},
			wantErr:          false,
			wantVolumes:      1,
			wantVolumeMounts: 1,
		},
		{
			name: "test-err",
			args: args{
				volumes: []corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}},
				volumeMounts: []corev1.VolumeMount{{
					Name:      "vol2",
					MountPath: "/data",
				}},
				value: &ThinValue{},
			},
			wantErr:          true,
			wantVolumes:      0,
			wantVolumeMounts: 0,
		},
		{
			name: "test-2",
			args: args{
				volumes: []corev1.Volume{{
					Name: "vol1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}, {
					Name: "vol2",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/test",
						},
					},
				}},
				volumeMounts: []corev1.VolumeMount{{
					Name:      "vol2",
					MountPath: "/data",
				}},
				value: &ThinValue{},
			},
			wantErr:          false,
			wantVolumes:      1,
			wantVolumeMounts: 1,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ThinEngine{}
			if err := t.transformWorkerVolumes(tt.args.volumes, tt.args.volumeMounts, tt.args.value); (err != nil) != tt.wantErr {
				t1.Errorf("transformWorkerVolumes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(tt.args.value.Worker.VolumeMounts) != tt.wantVolumeMounts ||
				len(tt.args.value.Worker.Volumes) != tt.wantVolumes {
				t1.Errorf("values %v", tt.args.value)
			}
		})
	}
}
