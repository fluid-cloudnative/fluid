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

package alluxio

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestTransformMasterVolumes(t *testing.T) {
	type testCase struct {
		name      string
		runtime   *datav1alpha1.AlluxioRuntime
		expect    *Alluxio
		expectErr bool
	}

	testCases := []testCase{
		{
			name: "all",
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, Master: datav1alpha1.AlluxioCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &Alluxio{
				Master: Master{
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
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Master: datav1alpha1.AlluxioCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &Alluxio{
				Master: Master{
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
		engine := &AlluxioEngine{}
		got := &Alluxio{}
		err := engine.transformMasterVolumes(testCase.runtime, got)
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

func TestTransformWorkerVolumes(t *testing.T) {
	type testCase struct {
		name      string
		runtime   *datav1alpha1.AlluxioRuntime
		expect    *Alluxio
		expectErr bool
	}

	testCases := []testCase{
		{
			name: "all",
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, Worker: datav1alpha1.AlluxioCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &Alluxio{
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
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &Alluxio{
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
		engine := &AlluxioEngine{}
		got := &Alluxio{}
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
		runtime   *datav1alpha1.AlluxioRuntime
		expect    *Alluxio
		expectErr bool
	}

	testCases := []testCase{
		{
			name: "all",
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test",
								},
							},
						},
					}, Fuse: datav1alpha1.AlluxioFuseSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &Alluxio{
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
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Fuse: datav1alpha1.AlluxioFuseSpec{
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
			},
			expect: &Alluxio{
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
		engine := &AlluxioEngine{}
		got := &Alluxio{}
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
