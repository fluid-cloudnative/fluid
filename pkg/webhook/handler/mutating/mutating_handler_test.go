/*
Copyright 2023 The Fluid Authors.

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

package mutating

import (
	"context"
	"os"

	"github.com/agiledragon/gomonkey/v2"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var (
	pluginsProfile = `
plugins:
  serverful:
    withDataset:
    - RequireNodeWithFuse
    - NodeAffinityWithCache
    - MountPropagationInjector
    withoutDataset:
    - PreferNodesWithoutCache
  serverless:
    withDataset:
    - FuseSidecar
    withoutDataset:
    - PreferNodesWithoutCache
pluginConfig:
  - name: NodeAffinityWithCache
    args: |
      preferred:
      - name: fluid.io/node
        weight: 100
      required:
      - fluid.io/node
`
)

var _ = Describe("MutatePod", func() {
	var (
		hostPathCharDev           = corev1.HostPathCharDev
		hostPathDirectoryOrCreate = corev1.HostPathDirectoryOrCreate
		bTrue                     = true
		s                         *runtime.Scheme
		patch                     *gomonkey.Patches
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(appsv1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	Context("serverless pod without dataset", func() {
		It("should return error when dataset does not exist", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "noexist",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectServerless: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "noexist",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-noexist",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/noexist/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-noexist",
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "noexist-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/noexist",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/noexist",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid",
					Namespace: "big-data",
				},
			}

			objs := []runtime.Object{fuse, pv, pvc, dataset, jindoRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("serverless pod with dataset", func() {
		It("should mutate pod successfully", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "exist",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Runtimes: []datav1alpha1.Runtime{
						{
							Name:      "exist",
							Namespace: "big-data",
							Type:      common.JindoRuntime,
						},
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectServerless: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "exist",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-exist",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/exist/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "exist",
					Namespace: "big-data",
					Labels: map[string]string{
						"fluid.io/s-big-data-exist": "true",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-exist",
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "exist-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/exist",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/exist",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "exist",
					Namespace: "big-data",
				},
			}

			objs := []runtime.Object{fuse, pv, pvc, dataset, jindoRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("serverless pod done", func() {
		It("should skip mutation when sidecar already injected", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectServerless:  common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "done",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-done",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/done/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-done",
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/done",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/done",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid",
					Namespace: "big-data",
				},
			}

			objs := []runtime.Object{fuse, pv, pvc, dataset, jindoRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("no serverless pod without runtime", func() {
		It("should return error when runtime does not exist", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-csi",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels:    map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "pod-with-csi",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-pod-with-csi",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/pod-with-csi/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-csi",
					Namespace: "big-data",
					Labels: map[string]string{
						common.LabelAnnotationStorageCapacityPrefix + "big-data" + "-" + "pod-with-csi": common.True,
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-pod-with-csi",
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-csi-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/pod-with-csi",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/pod-with-csi",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid",
					Namespace: "big-data",
				},
			}

			objs := []runtime.Object{fuse, pv, pvc, dataset, jindoRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("no serverless pod with runtime", func() {
		It("should mutate pod successfully", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Runtimes: []datav1alpha1.Runtime{
						{
							Type: common.JindoRuntime,
						},
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels:    map[string]string{},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "pod-with-fluid",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-pod-with-fluid",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/pod-with-fluid/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid",
					Namespace: "big-data",
					Labels: map[string]string{
						common.LabelAnnotationStorageCapacityPrefix + "big-data" + "-" + "pod-with-fluid": common.True,
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-pod-with-fluid",
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/pod-with-fluid",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/pod-with-fluid",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-with-fluid",
					Namespace: "big-data",
				},
			}

			objs := []runtime.Object{fuse, pv, pvc, dataset, jindoRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Handle", func() {
	var (
		decoder *admission.Decoder
		s       *runtime.Scheme
		patch   *gomonkey.Patches
	)

	BeforeEach(func() {
		decoder = admission.NewDecoder(scheme.Scheme)
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	Context("when handling admission requests", func() {
		It("should allow pod with namespace in pod", func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Namespace: "default",
					Object: runtime.RawExtension{
						Raw: []byte(
							`{
                                "apiVersion": "v1",
                                "kind": "Pod",
                                "metadata": {
                                    "name": "foo"
                                },
                                "spec": {
                                    "containers": [
                                        {
                                            "image": "bar:v2",
                                            "name": "bar"
                                        }
                                    ]
                                }
                            }`),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(s)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{}
			handler.Setup(fakeClient, fakeClient, decoder)

			resp := handler.Handle(context.TODO(), req)
			Expect(resp.AdmissionResponse.Allowed).To(BeTrue())
		})

		It("should allow pod with namespace not in pod", func() {
			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Namespace: "default",
					Object: runtime.RawExtension{
						Raw: []byte(
							`{
                                "apiVersion": "v1",
                                "kind": "Pod",
                                "metadata": {
                                    "name": "foo"
                                },
                                "spec": {
                                    "containers": [
                                        {
                                            "image": "bar:v2",
                                            "name": "bar"
                                        }
                                    ]
                                }
                            }`),
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(s)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{}
			handler.Setup(fakeClient, fakeClient, decoder)

			resp := handler.Handle(context.TODO(), req)
			Expect(resp.AdmissionResponse.Allowed).To(BeTrue())
		})
	})
})

var _ = Describe("MutatePodWithReferencedDataset", func() {
	var (
		hostPathCharDev           = corev1.HostPathCharDev
		hostPathDirectoryOrCreate = corev1.HostPathDirectoryOrCreate
		bTrue                     = true
		s                         *runtime.Scheme
		patch                     *gomonkey.Patches
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(appsv1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	Context("no serverless pod with runtime", func() {
		It("should mutate pod and inject cache affinity", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Runtimes: []datav1alpha1.Runtime{
						{
							Type: common.JindoRuntime,
						},
					},
				},
			}

			refDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "ref",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://big-data/done",
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Runtimes: []datav1alpha1.Runtime{
						{
							Type: common.ThinRuntime,
						},
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ref",
					Labels: map[string]string{
						common.InjectServerfulFuse:    common.True,
						"fluid.io/dataset.done.sched": "required",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "done",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-done",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/done/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-done",
				},
			}

			refPvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "ref",
					Labels: map[string]string{
						common.LabelAnnotationStorageCapacityPrefix + "ref-done": "true",
						common.LabelAnnotationDatasetReferringName:               "done",
						common.LabelAnnotationDatasetReferringNameSpace:          "big-data",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-done",
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/done-without-ref-pvc",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/done",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "big-data",
				},
			}

			refRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done",
					Namespace: "ref",
				},
			}

			objs := []runtime.Object{fuse, pv, pvc, dataset, refDataset, refPvc, jindoRuntime, refRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).NotTo(HaveOccurred())

			injectMountPropagation := pod.Spec.Containers[0].VolumeMounts[0].MountPropagation
			Expect(*injectMountPropagation).To(Equal(corev1.MountPropagationHostToContainer))

			cacheAffinity := pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
			Expect(cacheAffinity).NotTo(BeNil())
		})
	})

	Context("no serverless pod without ref pvc", func() {
		It("should return error when referenced pvc does not exist", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-without-ref-pvc",
					Namespace: "big-data",
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Runtimes: []datav1alpha1.Runtime{
						{
							Type: common.JindoRuntime,
						},
					},
				},
			}

			refDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-without-ref-pvc",
					Namespace: "ref",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://big-data/done-without-ref-pvc", // Fix: Match the actual dataset name
						},
					},
				},
				Status: datav1alpha1.DatasetStatus{
					Phase: datav1alpha1.BoundDatasetPhase,
					Runtimes: []datav1alpha1.Runtime{
						{
							Type: common.ThinRuntime,
						},
					},
				},
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ref",
					Labels: map[string]string{
						common.InjectServerfulFuse:                    common.True,
						"fluid.io/dataset.done-without-ref-pvc.sched": "required", // Fix: Match dataset name
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "done-without-ref-pvc",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-done-without-ref-pvc",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/done-without-ref-pvc/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-without-ref-pvc-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/done-without-ref-pvc",
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/done-without-ref-pvc",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			jindoRuntime := &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-without-ref-pvc",
					Namespace: "big-data",
				},
			}

			refRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "done-without-ref-pvc",
					Namespace: "ref",
				},
			}

			// Don't include the referencing PVC in the objects to trigger the error
			objs := []runtime.Object{fuse, pv, dataset, refDataset, jindoRuntime, refRuntime}
			fakeClient := fake.NewFakeClientWithScheme(s, objs...)
			Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

			handler := &FluidMutatingHandler{
				Client: fakeClient,
			}

			err := handler.MutatePod(pod, false)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("FluidMutatingHandler Setup", func() {
	It("should setup handler with client, reader and decoder", func() {
		s := runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())

		fakeClient := fake.NewFakeClientWithScheme(s)
		decoder := admission.NewDecoder(scheme.Scheme)

		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		Expect(handler.Client).To(Equal(fakeClient))
		Expect(handler.Reader).To(Equal(fakeClient))
		Expect(handler.decoder).To(Equal(decoder))
	})
})

var _ = Describe("Handle - Global Injection Disabled", func() {
	var (
		decoder *admission.Decoder
		s       *runtime.Scheme
		patch   *gomonkey.Patches
	)

	BeforeEach(func() {
		decoder = admission.NewDecoder(scheme.Scheme)
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	It("should skip mutation when global injection is disabled", func() {
		patch = gomonkey.ApplyFunc(utils.GetBoolValueFromEnv, func(key string, defaultValue bool) bool {
			if key == common.EnvDisableInjection {
				return true
			}
			return defaultValue
		})

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: "default",
				Object: runtime.RawExtension{
					Raw: []byte(
						`{
                            "apiVersion": "v1",
                            "kind": "Pod",
                            "metadata": {
                                "name": "test-pod",
                                "namespace": "default"
                            },
                            "spec": {
                                "containers": [
                                    {
                                        "image": "test:v1",
                                        "name": "test"
                                    }
                                ]
                            }
                        }`),
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.AdmissionResponse.Allowed).To(BeTrue())
		Expect(resp.Result).NotTo(BeNil())
		Expect(resp.Result.Message).To(ContainSubstring("global injection is disabled"))
	})
})

var _ = Describe("Handle - Decoder Error", func() {
	var (
		decoder *admission.Decoder
		s       *runtime.Scheme
	)

	BeforeEach(func() {
		decoder = admission.NewDecoder(scheme.Scheme)
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	It("should return error when decoder fails", func() {
		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: "default",
				Object: runtime.RawExtension{
					Raw: []byte(`invalid json`),
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.AdmissionResponse.Allowed).To(BeFalse())
	})
})

var _ = Describe("Handle - Namespace Validation", func() {
	var (
		decoder *admission.Decoder
		s       *runtime.Scheme
	)

	BeforeEach(func() {
		decoder = admission.NewDecoder(scheme.Scheme)
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
	})

	It("should deny pod with mismatched namespace", func() {
		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: "default",
				Object: runtime.RawExtension{
					Raw: []byte(
						`{
                            "apiVersion": "v1",
                            "kind": "Pod",
                            "metadata": {
                                "name": "test-pod",
                                "namespace": "different"
                            },
                            "spec": {
                                "containers": [
                                    {
                                        "image": "test:v1",
                                        "name": "test"
                                    }
                                ]
                            }
                        }`),
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.AdmissionResponse.Allowed).To(BeFalse())
		Expect(resp.Result).NotTo(BeNil())
		Expect(resp.Result.Message).To(ContainSubstring("invalid pod.metadata.namespace"))
	})

	It("should return error when both namespaces are empty", func() {
		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: "",
				Object: runtime.RawExtension{
					Raw: []byte(
						`{
                            "apiVersion": "v1",
                            "kind": "Pod",
                            "metadata": {
                                "name": "test-pod"
                            },
                            "spec": {
                                "containers": [
                                    {
                                        "image": "test:v1",
                                        "name": "test"
                                    }
                                ]
                            }
                        }`),
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.AdmissionResponse.Allowed).To(BeFalse())
	})
})

var _ = Describe("Handle - Injection Flags", func() {
	var (
		decoder *admission.Decoder
		s       *runtime.Scheme
		patch   *gomonkey.Patches
	)

	BeforeEach(func() {
		decoder = admission.NewDecoder(scheme.Scheme)
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	It("should skip mutation when injection is disabled via label", func() {
		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: "default",
				Object: runtime.RawExtension{
					Raw: []byte(
						`{
                            "apiVersion": "v1",
                            "kind": "Pod",
                            "metadata": {
                                "name": "test-pod",
                                "namespace": "default",
                                "labels": {
                                    "fluid.io/enable-injection": "false"
                                }
                            },
                            "spec": {
                                "containers": [
                                    {
                                        "image": "test:v1",
                                        "name": "test"
                                    }
                                ]
                            }
                        }`),
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.AdmissionResponse.Allowed).To(BeTrue())
		Expect(resp.Result).NotTo(BeNil())
		Expect(resp.Result.Message).To(ContainSubstring("injection is disabled"))
	})

})

var _ = Describe("Handle - Retry with API Reader", func() {
	var (
		decoder *admission.Decoder
		s       *runtime.Scheme
		patch   *gomonkey.Patches
	)

	BeforeEach(func() {
		decoder = admission.NewDecoder(scheme.Scheme)
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(appsv1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	It("should return error when retry with API reader also fails", func() {
		dataset := &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-dataset",
				Namespace: "default",
			},
			Status: datav1alpha1.DatasetStatus{
				Phase: datav1alpha1.BoundDatasetPhase,
			},
		}

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: "default",
				Object: runtime.RawExtension{
					Raw: []byte(
						`{
                            "apiVersion": "v1",
                            "kind": "Pod",
                            "metadata": {
                                "name": "test-pod",
                                "namespace": "default",
                                "labels": {
                                    "serverless.fluid.io/inject": "true"
                                }
                            },
                            "spec": {
                                "containers": [
                                    {
                                        "image": "test:v1",
                                        "name": "test",
                                        "volumeMounts": [
                                            {
                                                "name": "dataset",
                                                "mountPath": "/data"
                                            }
                                        ]
                                    }
                                ],
                                "volumes": [
                                    {
                                        "name": "dataset",
                                        "persistentVolumeClaim": {
                                            "claimName": "nonexistent-pvc"
                                        }
                                    }
                                ]
                            }
                        }`),
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s, dataset)
		Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

		handler := &FluidMutatingHandler{}
		handler.Setup(fakeClient, fakeClient, decoder)

		resp := handler.Handle(context.TODO(), req)
		Expect(resp.AdmissionResponse.Allowed).To(BeFalse())
	})
})

var _ = Describe("MutatePod - useDirectReader flag", func() {
	var (
		s     *runtime.Scheme
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(appsv1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	It("should use direct reader when useDirectReader is true", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels:    map[string]string{},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "test",
						Name:  "test",
					},
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

		handler := &FluidMutatingHandler{
			Client: fakeClient,
			Reader: fakeClient,
		}

		err := handler.MutatePod(pod, true)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("MutatePod - Plugin Execution", func() {
	var (
		s     *runtime.Scheme
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
		Expect(appsv1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	It("should handle pod without any special labels", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels:    map[string]string{},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "test",
						Name:  "test",
					},
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

		handler := &FluidMutatingHandler{
			Client: fakeClient,
		}

		err := handler.MutatePod(pod, false)
		Expect(err).NotTo(HaveOccurred())
	})
})

var _ = Describe("MutatePod - PVC Collection Error", func() {
	var (
		s     *runtime.Scheme
		patch *gomonkey.Patches
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(corev1.AddToScheme(s)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())

		mockReadFile := func(content string) ([]byte, error) {
			return []byte(pluginsProfile), nil
		}
		patch = gomonkey.ApplyFunc(os.ReadFile, mockReadFile)
	})

	AfterEach(func() {
		patch.Reset()
	})

	It("should return error when PVC collection fails", func() {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Labels: map[string]string{
					common.InjectServerless: common.True,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "test",
						Name:  "test",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "dataset",
								MountPath: "/data",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "dataset",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "nonexistent",
								ReadOnly:  true,
							},
						},
					},
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(s)
		Expect(plugins.RegisterMutatingHandlers(fakeClient)).To(Succeed())

		handler := &FluidMutatingHandler{
			Client: fakeClient,
		}

		err := handler.MutatePod(pod, false)
		Expect(err).To(HaveOccurred())
	})
})
