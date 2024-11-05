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

package juicefs

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func TestTransformWorkerVolumes(t *testing.T) {
	type testCase struct {
		name      string
		runtime   *datav1alpha1.JuiceFSRuntime
		expect    *JuiceFS
		expectErr bool
	}

	testCases := []testCase{
		{
			name: "all",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, Worker: datav1alpha1.JuiceFSCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &JuiceFS{
				Worker: Worker{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "test",
							MountPath: "/test",
						},
					},
				},
			},
			expectErr: false,
		}, {
			name: "onlyVolumeMounts",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Worker: datav1alpha1.JuiceFSCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &JuiceFS{
				Worker: Worker{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "test",
							MountPath: "/test",
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		engine := &JuiceFSEngine{}
		got := &JuiceFS{}
		err := engine.transformWorkerVolumes(testCase.runtime, got)
		if err != nil && !testCase.expectErr {
			t.Errorf("Got unexpected error %v", err)
		}

		if testCase.expectErr {
			continue
		}

		if !reflect.DeepEqual(got, testCase.expect) {
			t.Errorf("want %v, got %v for testcase %s", testCase.expect, got, testCase.name)
		}

	}

}

func TestTransformFuseVolumes(t *testing.T) {
	type testCase struct {
		name      string
		runtime   *datav1alpha1.JuiceFSRuntime
		expect    *JuiceFS
		expectErr bool
	}

	testCases := []testCase{
		{
			name: "all",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, Fuse: datav1alpha1.JuiceFSFuseSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &JuiceFS{
				Fuse: Fuse{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "test",
							MountPath: "/test",
						},
					},
				},
			},
			expectErr: false,
		}, {
			name: "onlyVolumeMounts",
			runtime: &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &JuiceFS{
				Fuse: Fuse{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "test",
							MountPath: "/test",
						},
					},
				},
			},
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		engine := &JuiceFSEngine{}
		got := &JuiceFS{}
		err := engine.transformFuseVolumes(testCase.runtime, got)
		if err != nil && !testCase.expectErr {
			t.Errorf("Got unexpected error %v for testcase %s", err, testCase.name)
		}

		if testCase.expectErr {
			continue
		}

		if !reflect.DeepEqual(got, testCase.expect) {
			t.Errorf("want %v, got %v for testcase %s", testCase.expect, got, testCase.name)
		}

	}

}

func TestJuiceFSEngine_transformWorkerCacheVolumes(t *testing.T) {
	dir := corev1.HostPathDirectoryOrCreate
	type args struct {
		runtime *datav1alpha1.JuiceFSRuntime
		value   *JuiceFS
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantVolumes      []corev1.Volume
		wantVolumeMounts []corev1.VolumeMount
	}{
		{
			name: "test-normal",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
					},
				},
			},
			wantErr: false,
			wantVolumes: []corev1.Volume{{
				Name: "cache-dir-1",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/cache",
						Type: &dir,
					},
				},
			}},
			wantVolumeMounts: []corev1.VolumeMount{{
				Name:      "cache-dir-1",
				MountPath: "/cache",
			}},
		},
		{
			name: "test-option-overwrite",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							Options: map[string]string{"cache-dir": "/worker-cache1:/worker-cache2"},
						},
					},
				},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
					},
				},
			},
			wantErr: false,
			wantVolumes: []corev1.Volume{
				{
					Name: "cache-dir-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/cache",
							Type: &dir,
						},
					},
				},
				{
					Name: "cache-dir-2",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/worker-cache1",
							Type: &dir,
						},
					},
				},
				{
					Name: "cache-dir-3",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/worker-cache2",
							Type: &dir,
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "cache-dir-1",
					MountPath: "/cache",
				},
				{
					Name:      "cache-dir-2",
					MountPath: "/worker-cache1",
				},
				{
					Name:      "cache-dir-3",
					MountPath: "/worker-cache2",
				},
			},
		},
		{
			name: "test-volume-overwrite",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							Options: map[string]string{"cache-dir": "/worker-cache1:/worker-cache2"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cache",
									MountPath: "/worker-cache2",
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cache",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
					},
					Worker: Worker{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "cache",
								MountPath: "/worker-cache2",
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cache",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			wantVolumes: []corev1.Volume{
				{
					Name: "cache",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "cache-dir-1",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/cache",
							Type: &dir,
						},
					},
				},
				{
					Name: "cache-dir-2",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/worker-cache1",
							Type: &dir,
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "cache",
					MountPath: "/worker-cache2",
				},
				{
					Name:      "cache-dir-1",
					MountPath: "/cache",
				},
				{
					Name:      "cache-dir-2",
					MountPath: "/worker-cache1",
				},
			},
		},
		{
			name: "test-emptyDir",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							Options: map[string]string{"cache-dir": "/worker-cache1"},
						},
					},
				},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {
							Path: "/cache",
							Type: string(common.VolumeTypeEmptyDir),
							VolumeSource: &datav1alpha1.VolumeSource{VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							}},
						},
					},
					Worker: Worker{},
				},
			},
			wantErr: false,
			wantVolumes: []corev1.Volume{
				{
					Name: "cache-dir-1",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
				{
					Name: "cache-dir-2",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/worker-cache1",
							Type: &dir,
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "cache-dir-1",
					MountPath: "/cache",
				},
				{
					Name:      "cache-dir-2",
					MountPath: "/worker-cache1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			if err := j.transformWorkerCacheVolumes(tt.args.runtime, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("transformWorkerCacheVolumes() error = %v, wantErr %v", err, tt.wantErr)
			}

			// compare volumes
			if len(tt.args.value.Worker.Volumes) != len(tt.wantVolumes) {
				t.Errorf("want volumes %v, got %v for testcase %s", tt.wantVolumes, tt.args.value.Worker.Volumes, tt.name)
			}
			wantVolumeMap := make(map[string]corev1.Volume)
			for _, v := range tt.wantVolumes {
				wantVolumeMap[v.Name] = v
			}
			for _, v := range tt.args.value.Worker.Volumes {
				if wv := wantVolumeMap[v.Name]; !reflect.DeepEqual(wv, v) {
					t.Errorf("want volumes %v, got %v for testcase %s", tt.wantVolumes, tt.args.value.Worker.Volumes, tt.name)
				}
			}

			// compare volumeMounts
			if len(tt.args.value.Worker.VolumeMounts) != len(tt.wantVolumeMounts) {
				t.Errorf("want volumeMounts %v, got %v for testcase %s", tt.wantVolumeMounts, tt.args.value.Worker.VolumeMounts, tt.name)
			}
			wantVolumeMountsMap := make(map[string]corev1.VolumeMount)
			for _, v := range tt.wantVolumeMounts {
				wantVolumeMountsMap[v.Name] = v
			}
			for _, v := range tt.args.value.Worker.VolumeMounts {
				if wv := wantVolumeMountsMap[v.Name]; !reflect.DeepEqual(wv, v) {
					t.Errorf("want volumeMounts %v, got %v for testcase %s", tt.wantVolumeMounts, tt.args.value.Worker.VolumeMounts, tt.name)
				}
			}
		})
	}
}

func TestJuiceFSEngine_transformFuseCacheVolumes(t *testing.T) {
	dir := corev1.HostPathDirectoryOrCreate
	type args struct {
		runtime *datav1alpha1.JuiceFSRuntime
		value   *JuiceFS
	}
	tests := []struct {
		name             string
		args             args
		wantErr          bool
		wantVolumes      []corev1.Volume
		wantVolumeMounts []corev1.VolumeMount
	}{
		{
			name: "test-normal",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
					},
				},
			},
			wantErr: false,
			wantVolumes: []corev1.Volume{{
				Name: "cache-dir-1",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/cache",
						Type: &dir,
					},
				},
			}},
			wantVolumeMounts: []corev1.VolumeMount{{
				Name:      "cache-dir-1",
				MountPath: "/cache",
			}},
		},
		{
			name: "test-volume-overwrite",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cache",
									MountPath: "/cache",
								},
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cache",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
					},
				},
			},
			wantErr:          false,
			wantVolumes:      nil,
			wantVolumeMounts: nil,
		},
		{
			name: "test-emptyDir",
			args: args{
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						Fuse: datav1alpha1.JuiceFSFuseSpec{
							Options: map[string]string{"cache-dir": "/fuse-cache1"},
						},
					},
				},
				value: &JuiceFS{
					CacheDirs: map[string]cache{
						"1": {
							Path: "/cache",
							Type: string(common.VolumeTypeEmptyDir),
							VolumeSource: &datav1alpha1.VolumeSource{VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							}},
						},
					},
				},
			},
			wantErr: false,
			wantVolumes: []corev1.Volume{
				{
					Name: "cache-dir-1",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
				{
					Name: "cache-dir-2",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/fuse-cache1",
							Type: &dir,
						},
					},
				},
			},
			wantVolumeMounts: []corev1.VolumeMount{
				{
					Name:      "cache-dir-1",
					MountPath: "/cache",
				},
				{
					Name:      "cache-dir-2",
					MountPath: "/fuse-cache1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{}
			if err := j.transformFuseCacheVolumes(tt.args.runtime, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("transformFuseCacheVolumes() error = %v, wantErr %v", err, tt.wantErr)
			}

			// compare volumes
			if len(tt.args.value.Fuse.Volumes) != len(tt.wantVolumes) {
				t.Errorf("want volumes %v, got %v for testcase %s", tt.wantVolumes, tt.args.value.Fuse.Volumes, tt.name)
			}
			wantVolumeMap := make(map[string]corev1.Volume)
			for _, v := range tt.wantVolumes {
				wantVolumeMap[v.Name] = v
			}
			for _, v := range tt.args.value.Fuse.Volumes {
				if wv := wantVolumeMap[v.Name]; !reflect.DeepEqual(wv, v) {
					t.Errorf("want volumes %v, got %v for testcase %s", tt.wantVolumes, tt.args.value.Fuse.Volumes, tt.name)
				}
			}

			// compare volumeMounts
			if len(tt.args.value.Fuse.VolumeMounts) != len(tt.wantVolumeMounts) {
				t.Errorf("want volumeMounts %v, got %v for testcase %s", tt.wantVolumeMounts, tt.args.value.Fuse.VolumeMounts, tt.name)
			}
			wantVolumeMountsMap := make(map[string]corev1.VolumeMount)
			for _, v := range tt.wantVolumeMounts {
				wantVolumeMountsMap[v.Name] = v
			}
			for _, v := range tt.args.value.Fuse.VolumeMounts {
				if wv := wantVolumeMountsMap[v.Name]; !reflect.DeepEqual(wv, v) {
					t.Errorf("want volumeMounts %v, got %v for testcase %s", tt.wantVolumeMounts, tt.args.value.Fuse.VolumeMounts, tt.name)
				}
			}
		})
	}
}
