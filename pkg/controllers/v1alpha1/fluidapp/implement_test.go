/*
  Copyright 2026 The Fluid Authors.

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

package fluidapp

import (
	"fmt"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

var _ = Describe("FluidAppReconcilerImplement", func() {
	const (
		fuseMountPath    = "/mnt/fuse"
		juicefsFuseMount = "juicefs-fuse-mount"
		juicefsMountPath = "/mnt/jfs"
	)

	Describe("NewFluidAppReconcilerImplement", func() {
		It("should create a new reconciler", func() {
			reconciler := NewFluidAppReconcilerImplement(nil, fake.NullLogger(), nil)
			Expect(reconciler).NotTo(BeNil())
			Expect(reconciler.Log).NotTo(BeNil())
		})
	})

	Describe("umountFuseSidecars", func() {
		var patches *gomonkey.Patches
		var reconciler *FluidAppReconcilerImplement

		BeforeEach(func() {
			mockExec := func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
				return "", "", nil
			}
			patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainer, mockExec)
			reconciler = &FluidAppReconcilerImplement{
				Log: fake.NullLogger(),
			}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when pod has no fuse containers", func() {
			It("should succeed without errors", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: "test"}},
					},
				}

				err := reconciler.umountFuseSidecars(pod)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when fuse container has no mount path", func() {
			It("should succeed without errors", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{Name: common.FuseContainerName + "-0"}},
					},
				}

				err := reconciler.umountFuseSidecars(pod)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when fuse container has prestop hook", func() {
			It("should succeed without errors", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: common.FuseContainerName + "-0",
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{Command: []string{"umount"}},
								},
							},
						}},
					},
				}

				err := reconciler.umountFuseSidecars(pod)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when fuse container has mount path", func() {
			It("should succeed and umount the path", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name: common.FuseContainerName + "-0",
							VolumeMounts: []corev1.VolumeMount{{
								Name:      juicefsFuseMount,
								MountPath: juicefsMountPath,
							}},
						}},
					},
				}

				err := reconciler.umountFuseSidecars(pod)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when pod has multiple fuse sidecars", func() {
			It("should succeed and umount all paths", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: common.FuseContainerName + "-0",
								VolumeMounts: []corev1.VolumeMount{{
									Name:      juicefsFuseMount,
									MountPath: juicefsMountPath,
								}},
							},
							{
								Name: common.FuseContainerName + "-1",
								VolumeMounts: []corev1.VolumeMount{{
									Name:      juicefsFuseMount,
									MountPath: juicefsMountPath,
								}},
							},
						},
					},
				}

				err := reconciler.umountFuseSidecars(pod)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("umountFuseSidecar", func() {
		var patches *gomonkey.Patches
		var reconciler *FluidAppReconcilerImplement

		BeforeEach(func() {
			reconciler = &FluidAppReconcilerImplement{
				Log: fake.NullLogger(),
			}
		})

		AfterEach(func() {
			if patches != nil {
				patches.Reset()
			}
		})

		Context("when fuse container has empty name", func() {
			It("should return nil", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test"},
				}
				container := corev1.Container{Name: ""}

				patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainer, func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
					return "", "", nil
				})

				err := reconciler.umountFuseSidecar(pod, container)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when exec fails with 'not mounted' error", func() {
			It("should return nil and not retry", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				}
				container := corev1.Container{
					Name: common.FuseContainerName + "-0",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "fuse-mount",
						MountPath: fuseMountPath,
					}},
				}

				patches = gomonkey.ApplyFunc(kubeclient.GetMountPathInContainer, func(c corev1.Container) (string, error) {
					return fuseMountPath, nil
				})
				patches.ApplyFunc(kubeclient.ExecCommandInContainer, func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
					return "", "not mounted", fmt.Errorf("umount failed")
				})

				err := reconciler.umountFuseSidecar(pod, container)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when exec fails with exit code 137", func() {
			It("should return nil and not retry", func() {
				pod := &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
				}
				container := corev1.Container{
					Name: common.FuseContainerName + "-0",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "fuse-mount",
						MountPath: fuseMountPath,
					}},
				}

				patches = gomonkey.ApplyFunc(kubeclient.GetMountPathInContainer, func(c corev1.Container) (string, error) {
					return fuseMountPath, nil
				})
				patches.ApplyFunc(kubeclient.ExecCommandInContainer, func(p1, p2, p3 string, p4 []string) (stdout string, stderr string, e error) {
					return "", "", fmt.Errorf("exit code 137")
				})

				err := reconciler.umountFuseSidecar(pod, container)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
