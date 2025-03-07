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

// TestTransformWorkerVolumes is a unit test function that tests the transformWorkerVolumes method of the AlluxioEngine.
// It defines a series of test cases to verify the correctness of volume and volume mount transformations for the Alluxio worker.
// Each test case includes an input AlluxioRuntime object and an expected Alluxio object, along with a flag to indicate if an error is expected.
// The function iterates through each test case, applies the transformWorkerVolumes method, and checks if the output matches the expected result.
// If an error occurs and it is not expected, or if the output does not match the expected result, the test fails with an appropriate error message.
func TestTransformWorkerVolumes(t *testing.T) {
	// Define a testCase struct to represent each test case.
	type testCase struct {
		name      string         // Name of the test case.
		runtime   *datav1alpha1.AlluxioRuntime // Input AlluxioRuntime object.
		expect    *Alluxio       // Expected Alluxio object after transformation.
		expectErr bool           // Flag to indicate if an error is expected.
	}

	// Define a list of test cases to be executed.
	testCases := []testCase{
		{
			name: "all", // Test case name.
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Volumes: []corev1.Volume{ // Define volumes in the runtime.
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test", // Use a secret as the volume source.
								},
							},
						},
					},
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{ // Define volume mounts for the worker.
							{
								Name:      "test",
								MountPath: "/test", // Mount path for the volume.
							},
						},
					},
				},
			},
			expect: &Alluxio{
				Worker: Worker{
					Volumes: []corev1.Volume{ // Expected volumes in the Alluxio worker.
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test", // Expected secret volume source.
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{ // Expected volume mounts in the Alluxio worker.
						{
							Name:      "test",
							MountPath: "/test", // Expected mount path.
						},
					},
				},
			},
			expectErr: false, // No error is expected for this test case.
		},
		{
			name: "onlyVolumeMounts", // Test case name.
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					Worker: datav1alpha1.AlluxioCompTemplateSpec{
						VolumeMounts: []corev1.VolumeMount{ // Define only volume mounts without volumes.
							{
								Name:      "test",
								MountPath: "/test", // Mount path for the volume.
							},
						},
					},
				},
			},
			expect: &Alluxio{
				Worker: Worker{
					Volumes: []corev1.Volume{ // Expected volumes in the Alluxio worker.
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test", // Expected secret volume source.
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{ // Expected volume mounts in the Alluxio worker.
						{
							Name:      "test",
							MountPath: "/test", // Expected mount path.
						},
					},
				},
			},
			expectErr: true, // An error is expected for this test case due to missing volumes.
		},
	}

	// Iterate through each test case and execute the test.
	for _, testCase := range testCases {
		engine := &AlluxioEngine{} // Initialize the AlluxioEngine.
		got := &Alluxio{}          // Initialize the output Alluxio object.
		err := engine.transformWorkerVolumes(testCase.runtime, got) // Apply the transformation.

		// Check if an error occurred and handle accordingly.
		if err != nil && !testCase.expectErr {
			t.Errorf("Got unexpected error %v", err) // Fail the test if an unexpected error occurs.
		}

		// Skip further checks if an error is expected.
		if testCase.expectErr {
			continue
		}

		// Compare the output with the expected result.
		if !reflect.DeepEqual(got, testCase.expect) {
			t.Errorf("want %v, got %v for testcase %s", testCase.expect, got, testCase.name) // Fail the test if the output does not match the expected result.
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
