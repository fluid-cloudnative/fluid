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
