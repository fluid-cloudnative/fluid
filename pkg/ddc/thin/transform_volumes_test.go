/*
  Copyright 2022 The Fluid Authors.

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
