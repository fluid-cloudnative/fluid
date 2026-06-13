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
	"context"

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
	const expectedJuiceFSMountCmd = "/mnt/jfs/juicefs-fuse"

	var patches *gomonkey.Patches

	AfterEach(func() {
		if patches != nil {
			patches.Reset()
		}
	})

	Describe("umountFuseSidecars", func() {
		It("returns nil when there is no fuse sidecar container", func() {
			i := &FluidAppReconcilerImplement{Log: fake.NullLogger()}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "test"}}},
			}

			Expect(i.umountFuseSidecars(pod)).To(Succeed())
		})

		It("returns nil when the fuse sidecar mount path lookup is empty", func() {
			i := &FluidAppReconcilerImplement{Log: fake.NullLogger()}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test"}}
			fuseContainer := corev1.Container{Name: common.FuseContainerName + "-0"}

			patches = gomonkey.ApplyFunc(kubeclient.GetMountPathInContainer, func(corev1.Container) (string, error) {
				return "", nil
			})
			patches.ApplyFunc(kubeclient.ExecCommandInContainerWithContext, func(context.Context, string, string, string, []string) (string, string, error) {
				Fail("ExecCommandInContainerWithContext should not be called when mount path lookup returns empty")
				return "", "", nil
			})

			Expect(i.umountFuseSidecar(pod, fuseContainer)).To(Succeed())
		})

		It("uses the container prestop command when present", func() {
			patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithContext, func(_ context.Context, podName, containerName, namespace string, cmd []string) (string, string, error) {
				Expect(podName).To(Equal("test"))
				Expect(containerName).To(Equal(common.FuseContainerName + "-0"))
				Expect(namespace).To(BeEmpty())
				Expect(cmd).To(Equal([]string{"umount"}))
				return "", "", nil
			})

			i := &FluidAppReconcilerImplement{Log: fake.NullLogger()}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name: common.FuseContainerName + "-0",
					Lifecycle: &corev1.Lifecycle{PreStop: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{Command: []string{"umount"}},
					}},
				}}},
			}

			Expect(i.umountFuseSidecars(pod)).To(Succeed())
		})

		It("derives the mount path when the fuse sidecar has no prestop", func() {
			patches = gomonkey.ApplyFunc(kubeclient.ExecCommandInContainerWithContext, func(_ context.Context, _, _, _ string, cmd []string) (string, string, error) {
				Expect(cmd).To(Equal([]string{"umount", expectedJuiceFSMountCmd}))
				return "", "", nil
			})

			i := &FluidAppReconcilerImplement{Log: fake.NullLogger()}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name: common.FuseContainerName + "-0",
					VolumeMounts: []corev1.VolumeMount{{
						Name:      "juicefs-fuse-mount",
						MountPath: "/mnt/jfs",
					}},
				}}},
			}

			Expect(i.umountFuseSidecars(pod)).To(Succeed())
		})

		It("unmounts each fuse sidecar container", func() {
			containerNames := []string{}
			patches = gomonkey.ApplyFunc((*FluidAppReconcilerImplement).umountFuseSidecar, func(_ *FluidAppReconcilerImplement, _ *corev1.Pod, fuseContainer corev1.Container) error {
				containerNames = append(containerNames, fuseContainer.Name)
				return nil
			})

			i := &FluidAppReconcilerImplement{Log: fake.NullLogger()}
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "test"},
				Spec: corev1.PodSpec{Containers: []corev1.Container{
					{
						Name:         common.FuseContainerName + "-0",
						VolumeMounts: []corev1.VolumeMount{{Name: "juicefs-fuse-mount", MountPath: "/mnt/jfs"}},
					},
					{
						Name:         common.FuseContainerName + "-1",
						VolumeMounts: []corev1.VolumeMount{{Name: "juicefs-fuse-mount", MountPath: "/mnt/jfs"}},
					},
				}},
			}

			Expect(i.umountFuseSidecars(pod)).To(Succeed())
			Expect(containerNames).To(ConsistOf(common.FuseContainerName+"-0", common.FuseContainerName+"-1"))
			Expect(containerNames).To(HaveLen(2))
		})
	})
})
