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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

var _ = Describe("Volume Transformation", func() {
	Describe("transformWorkerVolumes", func() {
		type testCase struct {
			name      string
			runtime   *datav1alpha1.JuiceFSRuntime
			expect    *JuiceFS
			expectErr bool
		}

		DescribeTable("should transform worker volumes correctly",
			func(tc testCase) {
				engine := &JuiceFSEngine{}
				got := &JuiceFS{}
				err := engine.transformWorkerVolumes(tc.runtime, got)

				if tc.expectErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(got).To(Equal(tc.expect))
				}
			},
			Entry("all volumes and mounts specified", testCase{
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
						},
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
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
				expectErr: false,
			}),
			Entry("only volume mounts without volumes", testCase{
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
				expect:    nil,
				expectErr: true,
			}),
		)
	})

	Describe("transformFuseVolumes", func() {
		type testCase struct {
			name      string
			runtime   *datav1alpha1.JuiceFSRuntime
			expect    *JuiceFS
			expectErr bool
		}

		DescribeTable("should transform fuse volumes correctly",
			func(tc testCase) {
				engine := &JuiceFSEngine{}
				got := &JuiceFS{}
				err := engine.transformFuseVolumes(tc.runtime, got)

				if tc.expectErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(got).To(Equal(tc.expect))
				}
			},
			Entry("all volumes and mounts specified", testCase{
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
						},
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
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test",
								MountPath: "/test",
							},
						},
					},
				},
				expectErr: false,
			}),
			Entry("only volume mounts without volumes", testCase{
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
				expect:    nil,
				expectErr: true,
			}),
		)
	})

	Describe("transformWorkerCacheVolumes", func() {

		type testArgs struct {
			runtime *datav1alpha1.JuiceFSRuntime
			value   *JuiceFS
			options map[string]string
		}

		type testCase struct {
			name             string
			args             testArgs
			wantErr          bool
			wantVolumes      []corev1.Volume
			wantVolumeMounts []corev1.VolumeMount
		}

		DescribeTable("should transform worker cache volumes correctly",
			func(tc testCase) {
				dir := corev1.HostPathDirectoryOrCreate
				_ = dir

				j := &JuiceFSEngine{}
				err := j.transformWorkerCacheVolumes(tc.args.runtime, tc.args.value, tc.args.options)

				if tc.wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(len(tc.args.value.Worker.Volumes)).To(Equal(len(tc.wantVolumes)))

				wantVolumeMap := make(map[string]corev1.Volume)
				for _, v := range tc.wantVolumes {
					wantVolumeMap[v.Name] = v
				}
				for _, v := range tc.args.value.Worker.Volumes {
					Expect(wantVolumeMap[v.Name]).To(Equal(v))
				}

				Expect(len(tc.args.value.Worker.VolumeMounts)).To(Equal(len(tc.wantVolumeMounts)))

				wantVolumeMountsMap := make(map[string]corev1.VolumeMount)
				for _, v := range tc.wantVolumeMounts {
					wantVolumeMountsMap[v.Name] = v
				}
				for _, v := range tc.args.value.Worker.VolumeMounts {
					Expect(wantVolumeMountsMap[v.Name]).To(Equal(v))
				}
			},
			Entry("test-normal", func() testCase {
				dir := corev1.HostPathDirectoryOrCreate
				return testCase{
					name: "test-normal",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{},
						value: &JuiceFS{
							CacheDirs: map[string]cache{
								"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
							},
						},
						options: map[string]string{"cache-dir": "/cache"},
					},
					wantErr: false,
					wantVolumes: []corev1.Volume{{
						Name: "cache-dir-0",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/cache",
								Type: &dir,
							},
						},
					}},
					wantVolumeMounts: []corev1.VolumeMount{{
						Name:      "cache-dir-0",
						MountPath: "/cache",
					}},
				}
			}()),
			Entry("test-option-overwrite", func() testCase {
				dir := corev1.HostPathDirectoryOrCreate
				return testCase{
					name: "test-option-overwrite",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{
							Spec: datav1alpha1.JuiceFSRuntimeSpec{
								Worker: datav1alpha1.JuiceFSCompTemplateSpec{},
							},
						},
						value: &JuiceFS{
							CacheDirs: map[string]cache{
								"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
							},
						},
						options: map[string]string{"cache-dir": "/cache:/worker-cache1:/worker-cache2"},
					},
					wantErr: false,
					wantVolumes: []corev1.Volume{
						{
							Name: "cache-dir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/cache",
									Type: &dir,
								},
							},
						},
						{
							Name: "cache-dir-1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/worker-cache1",
									Type: &dir,
								},
							},
						},
						{
							Name: "cache-dir-2",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/worker-cache2",
									Type: &dir,
								},
							},
						},
					},
					wantVolumeMounts: []corev1.VolumeMount{
						{Name: "cache-dir-0", MountPath: "/cache"},
						{Name: "cache-dir-1", MountPath: "/worker-cache1"},
						{Name: "cache-dir-2", MountPath: "/worker-cache2"},
					},
				}
			}()),
			Entry("test-volume-overwrite", func() testCase {
				dir := corev1.HostPathDirectoryOrCreate
				return testCase{
					name: "test-volume-overwrite",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{
							Spec: datav1alpha1.JuiceFSRuntimeSpec{
								Worker: datav1alpha1.JuiceFSCompTemplateSpec{
									VolumeMounts: []corev1.VolumeMount{
										{Name: "cache", MountPath: "/worker-cache2"},
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
									{Name: "cache", MountPath: "/worker-cache2"},
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
						options: map[string]string{"cache-dir": "/worker-cache1:/worker-cache2"},
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
							Name: "cache-dir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/worker-cache1",
									Type: &dir,
								},
							},
						},
					},
					wantVolumeMounts: []corev1.VolumeMount{
						{Name: "cache", MountPath: "/worker-cache2"},
						{Name: "cache-dir-0", MountPath: "/worker-cache1"},
					},
				}
			}()),
			Entry("test-emptyDir", func() testCase {
				dir := corev1.HostPathDirectoryOrCreate
				return testCase{
					name: "test-emptyDir",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{
							Spec: datav1alpha1.JuiceFSRuntimeSpec{
								Worker: datav1alpha1.JuiceFSCompTemplateSpec{},
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
						options: map[string]string{"cache-dir": "/worker-cache1"},
					},
					wantErr: false,
					wantVolumes: []corev1.Volume{
						{
							Name: "cache-dir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/worker-cache1",
									Type: &dir,
								},
							},
						},
					},
					wantVolumeMounts: []corev1.VolumeMount{
						{Name: "cache-dir-0", MountPath: "/worker-cache1"},
					},
				}
			}()),
		)
	})

	Describe("transformFuseCacheVolumes", func() {
		type testArgs struct {
			runtime *datav1alpha1.JuiceFSRuntime
			value   *JuiceFS
			options map[string]string
		}

		type testCase struct {
			name             string
			args             testArgs
			wantErr          bool
			wantVolumes      []corev1.Volume
			wantVolumeMounts []corev1.VolumeMount
		}

		DescribeTable("should transform fuse cache volumes correctly",
			func(tc testCase) {
				j := &JuiceFSEngine{}
				err := j.transformFuseCacheVolumes(tc.args.runtime, tc.args.value, tc.args.options)

				if tc.wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(len(tc.args.value.Fuse.Volumes)).To(Equal(len(tc.wantVolumes)))

				wantVolumeMap := make(map[string]corev1.Volume)
				for _, v := range tc.wantVolumes {
					wantVolumeMap[v.Name] = v
				}
				for _, v := range tc.args.value.Fuse.Volumes {
					Expect(wantVolumeMap[v.Name]).To(Equal(v))
				}

				Expect(len(tc.args.value.Fuse.VolumeMounts)).To(Equal(len(tc.wantVolumeMounts)))

				wantVolumeMountsMap := make(map[string]corev1.VolumeMount)
				for _, v := range tc.wantVolumeMounts {
					wantVolumeMountsMap[v.Name] = v
				}
				for _, v := range tc.args.value.Fuse.VolumeMounts {
					Expect(wantVolumeMountsMap[v.Name]).To(Equal(v))
				}
			},
			Entry("test-normal", func() testCase {
				dir := corev1.HostPathDirectoryOrCreate
				return testCase{
					name: "test-normal",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{},
						value: &JuiceFS{
							CacheDirs: map[string]cache{
								"1": {Path: "/cache", Type: string(common.VolumeTypeHostPath)},
							},
						},
						options: map[string]string{"cache-dir": "/cache"},
					},
					wantErr: false,
					wantVolumes: []corev1.Volume{{
						Name: "cache-dir-0",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/cache",
								Type: &dir,
							},
						},
					}},
					wantVolumeMounts: []corev1.VolumeMount{{
						Name:      "cache-dir-0",
						MountPath: "/cache",
					}},
				}
			}()),
			Entry("test-volume-overwrite", func() testCase {
				return testCase{
					name: "test-volume-overwrite",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{
							Spec: datav1alpha1.JuiceFSRuntimeSpec{
								Fuse: datav1alpha1.JuiceFSFuseSpec{
									VolumeMounts: []corev1.VolumeMount{
										{Name: "cache", MountPath: "/cache"},
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
				}
			}()),
			Entry("test-emptyDir", func() testCase {
				dir := corev1.HostPathDirectoryOrCreate
				return testCase{
					name: "test-emptyDir",
					args: testArgs{
						runtime: &datav1alpha1.JuiceFSRuntime{
							Spec: datav1alpha1.JuiceFSRuntimeSpec{
								Fuse: datav1alpha1.JuiceFSFuseSpec{},
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
						options: map[string]string{"cache-dir": "/fuse-cache1"},
					},
					wantErr: false,
					wantVolumes: []corev1.Volume{
						{
							Name: "cache-dir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/fuse-cache1",
									Type: &dir,
								},
							},
						},
					},
					wantVolumeMounts: []corev1.VolumeMount{
						{Name: "cache-dir-0", MountPath: "/fuse-cache1"},
					},
				}
			}()),
		)
	})
})
