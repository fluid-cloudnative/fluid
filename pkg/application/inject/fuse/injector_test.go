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

package fuse

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/google/go-cmp/cmp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Injector Related Tests", Label("pkg.application.inject.fuse.injector_test.go"), func() {
	type testCaseContext struct {
		in             *corev1.Pod
		datasets       []*datav1alpha1.Dataset
		pvs            []*corev1.PersistentVolume
		pvcs           []*corev1.PersistentVolumeClaim
		fuseDaemonsets []*appsv1.DaemonSet
	}

	mockTestCaseContextFn := func(datasetNames []string, namespace string) *testCaseContext {
		mockedDatasets := []*datav1alpha1.Dataset{}
		for _, name := range datasetNames {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			}
			mockedDatasets = append(mockedDatasets, dataset)
		}

		mockedPVs := []*corev1.PersistentVolume{}
		for _, dataset := range mockedDatasets {
			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", dataset.Namespace, dataset.Name),
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", dataset.Namespace, dataset.Name),
								common.VolumeAttrMountType: common.ThinRuntime,
							},
						},
					},
				},
			}

			mockedPVs = append(mockedPVs, pv)
		}

		mockedPVCs := []*corev1.PersistentVolumeClaim{}
		for _, dataset := range mockedDatasets {
			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataset.Name,
					Namespace: dataset.Namespace,
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: fmt.Sprintf("%s-%s", dataset.Namespace, dataset.Name),
				},
			}
			mockedPVCs = append(mockedPVCs, pvc)
		}

		hostPathCharDev := corev1.HostPathCharDev
		hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
		mockedFuses := []*appsv1.DaemonSet{}
		for _, dataset := range mockedDatasets {
			fuseDs := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-fuse", dataset.Name),
					Namespace: dataset.Namespace,
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "thin-fuse",
									Args: []string{
										"-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "test-thin-fuse:v1.0.0",
									SecurityContext: &corev1.SecurityContext{
										Privileged: ptr.To(true),
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
											Name:      "thin-fuse-mount",
											MountPath: fmt.Sprintf("/runtime-mnt/thin/%s/%s/", dataset.Namespace, dataset.Name),
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/dataset-conflict",
										},
									}},
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
									Name: "thin-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s", dataset.Namespace, dataset.Name),
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}
			mockedFuses = append(mockedFuses, fuseDs)
		}

		inPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: namespace,
				Labels: map[string]string{
					common.InjectServerless: common.True,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: "test-image",
						Name:  "test-container",
					},
				},
			},
		}

		for _, dataset := range mockedDatasets {
			inPod.Spec.Volumes = append(inPod.Spec.Volumes, corev1.Volume{
				Name: dataset.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: dataset.Name,
					},
				},
			})
			inPod.Spec.Containers[0].VolumeMounts = append(inPod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
				Name:      dataset.Name,
				MountPath: "/data-" + dataset.Name,
			})
		}

		return &testCaseContext{
			in:             inPod,
			datasets:       mockedDatasets,
			pvs:            mockedPVs,
			pvcs:           mockedPVCs,
			fuseDaemonsets: mockedFuses,
		}
	}

	var (
		testCtx     *testCaseContext
		injector    *Injector
		client      client.Client
		runtimeObjs []runtime.Object
	)

	JustBeforeEach(func() {
		runtimeObjs = []runtime.Object{}
		for _, obj := range testCtx.datasets {
			runtimeObjs = append(runtimeObjs, obj)
		}
		for _, obj := range testCtx.pvs {
			runtimeObjs = append(runtimeObjs, obj)
		}
		for _, obj := range testCtx.pvcs {
			runtimeObjs = append(runtimeObjs, obj)
		}
		for _, obj := range testCtx.fuseDaemonsets {
			runtimeObjs = append(runtimeObjs, obj)
		}

		client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, runtimeObjs...)
		injector = NewInjector(client)
	})

	expectInjectionFn := func(out *corev1.Pod, fuseDs *appsv1.DaemonSet, dataset *datav1alpha1.Dataset) {
		// find out which injected container is related to the dataset
		var containerIdx int = -1
		var containerName string
		for k, v := range out.Labels {
			if strings.HasPrefix(k, common.LabelContainerDatasetNameKeyPrefix) && v == dataset.Name {
				containerName = strings.TrimPrefix(k, common.LabelContainerDatasetNameKeyPrefix)
				containerIdx = slices.IndexFunc(out.Spec.Containers, func(c corev1.Container) bool {
					return c.Name == containerName
				})
				break
			}
		}
		Expect(containerName).NotTo(BeEmpty())
		Expect(containerIdx).NotTo(Equal(-1))

		// Verify the fuse container image
		Expect(out.Spec.Containers[containerIdx].Image).To(Equal(fuseDs.Spec.Template.Spec.Containers[0].Image))

		// Verify the fuse container security context
		Expect(out.Spec.Containers[containerIdx].SecurityContext.Privileged).To(Equal(ptr.To(true)))

		// Verify the fuse container command and args
		Expect(out.Spec.Containers[containerIdx].Command).To(ContainElements(fuseDs.Spec.Template.Spec.Containers[0].Command))
		Expect(out.Spec.Containers[containerIdx].Args).To(ContainElements(fuseDs.Spec.Template.Spec.Containers[0].Args))
		Expect(out.Spec.Containers[containerIdx].Env).To(ContainElements(fuseDs.Spec.Template.Spec.Containers[0].Env))

		fuseContainerSuffix := strings.TrimPrefix(containerName, common.FuseContainerName)
		// Verify post start hook
		Expect(out.Spec.Containers[containerIdx].Lifecycle.PostStart).To(Equal(&corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"bash", "-c", fmt.Sprintf("time /check-mount.sh /runtime-mnt/thin/%s/%s/ thin ", dataset.Namespace, dataset.Name)},
			},
		}))

		// Verify the fuse container volume mounts
		for _, vm := range fuseDs.Spec.Template.Spec.Containers[0].VolumeMounts {
			vm.Name = vm.Name + fuseContainerSuffix
			Expect(out.Spec.Containers[containerIdx].VolumeMounts).To(ContainElement(vm))
		}
		Expect(out.Spec.Containers[containerIdx].VolumeMounts).To(ContainElement(corev1.VolumeMount{
			Name:      "default-check-mount" + fuseContainerSuffix,
			ReadOnly:  true,
			MountPath: "/check-mount.sh",
			SubPath:   "check-mount.sh",
		}))

		// Verify the fuse container volumes
		for _, v := range fuseDs.Spec.Template.Spec.Volumes {
			v.Name = v.Name + fuseContainerSuffix
			Expect(out.Spec.Volumes).To(ContainElement(v))
		}
		Expect(out.Spec.Volumes).To(ContainElement(corev1.Volume{
			Name: "default-check-mount" + fuseContainerSuffix,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: "thin-default-check-mount"},
					DefaultMode:          ptr.To[int32](0755),
				},
			},
		}))

		Expect(out.Spec.Volumes).To(ContainElement(corev1.Volume{
			Name:         dataset.Name,
			VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", dataset.Namespace, dataset.Name)}}},
		))

		hostToContainerMountPropagation := corev1.MountPropagationHostToContainer
		Expect(out.Spec.Containers[len(out.Spec.Containers)-1].VolumeMounts).To(ContainElement(corev1.VolumeMount{
			Name:             dataset.Name,
			MountPath:        "/data-" + dataset.Name,
			MountPropagation: &hostToContainerMountPropagation,
		}))
	}

	Context("Inject Pod mounting only one Fluid Dataset", func() {
		BeforeEach(func() {
			testCtx = mockTestCaseContextFn([]string{"dataset-1"}, "fluid-test")
		})

		It("should inject Pod successfully with one Fluid Dataset PVC", func() {
			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			for _, dataset := range testCtx.datasets {
				info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
				info.SetAPIReader(client)
				Expect(err).NotTo(HaveOccurred())
				runtimeInfos[dataset.Name] = info
			}

			out, err := injector.InjectPod(testCtx.in, runtimeInfos)

			fuseDs := testCtx.fuseDaemonsets[0]
			Expect(err).NotTo(HaveOccurred())
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0", "fluid-test"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0", "dataset-1"))
			Expect(out.Spec.Containers).To(HaveLen(2))

			expectInjectionFn(out, fuseDs, testCtx.datasets[0])

		})

		When("fuse daemonset has user-specified fields (env, volumes, volume mounts)", func() {
			BeforeEach(func() {
				testCtx.fuseDaemonsets[0].Spec.Template.Spec.Volumes = append(testCtx.fuseDaemonsets[0].Spec.Template.Spec.Volumes, corev1.Volume{Name: "new-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}})
				testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].VolumeMounts = append(testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: "new-vol", MountPath: "/new-vol"})
				testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Env = append(testCtx.fuseDaemonsets[0].Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "new-env", Value: "new-env-value"})
			})

			It("should inject Pod successfully with the user-specified fields", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{}
				for _, dataset := range testCtx.datasets {
					info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
					info.SetAPIReader(client)
					Expect(err).NotTo(HaveOccurred())
					runtimeInfos[dataset.Name] = info
				}

				out, err := injector.InjectPod(testCtx.in, runtimeInfos)

				fuseDs := testCtx.fuseDaemonsets[0]
				Expect(err).NotTo(HaveOccurred())
				Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0", "fluid-test"))
				Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0", "dataset-1"))
				Expect(out.Spec.Containers).To(HaveLen(2))

				expectInjectionFn(out, fuseDs, testCtx.datasets[0])
			})
		})

		When("pod has a label that indicates it has been injected", func() {
			BeforeEach(func() {
				testCtx.in.ObjectMeta.Labels[common.InjectSidecarDone] = common.True
			})

			It("should not inject anything", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{}
				for _, dataset := range testCtx.datasets {
					info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
					info.SetAPIReader(client)
					Expect(err).NotTo(HaveOccurred())
					runtimeInfos[dataset.Name] = info
				}

				out, err := injector.InjectPod(testCtx.in, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(out).To(Equal(testCtx.in))
			})
		})

		When("pod is mounting same Fluid PVC several times", func() {
			BeforeEach(func() {
				testCtx.in.Spec.Volumes = append(testCtx.in.Spec.Volumes, corev1.Volume{
					Name: "duplicate-pvc",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: testCtx.datasets[0].Name,
						},
					},
				})
				testCtx.in.Spec.Containers[0].VolumeMounts = append(testCtx.in.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
					Name:      "duplicate-pvc",
					MountPath: "/duplicate-pvc",
				})
			})

			It("should inject pod successfully, but only one sidecar will be injected", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{}
				for _, dataset := range testCtx.datasets {
					info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
					info.SetAPIReader(client)
					Expect(err).NotTo(HaveOccurred())
					runtimeInfos[dataset.Name] = info
				}

				out, err := injector.InjectPod(testCtx.in, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
				Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0", "fluid-test"))
				Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0", "dataset-1"))
				Expect(out.Spec.Containers).To(HaveLen(2))
				expectInjectionFn(out, testCtx.fuseDaemonsets[0], testCtx.datasets[0])

				Expect(out.Spec.Volumes).To(ContainElement(corev1.Volume{
					Name: "duplicate-pvc",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", testCtx.datasets[0].Namespace, testCtx.datasets[0].Name),
						},
					},
				}))

				hostToContainerMountPropagation := corev1.MountPropagationHostToContainer
				Expect(out.Spec.Containers[1].VolumeMounts).To(ContainElement(corev1.VolumeMount{
					Name:             "duplicate-pvc",
					MountPath:        "/duplicate-pvc",
					MountPropagation: &hostToContainerMountPropagation,
				}))
			})
		})

		When("inject pod with unprivileged mutator", func() {
			BeforeEach(func() {
				testCtx.in.Labels[common.InjectUnprivilegedFuseSidecar] = common.True
			})

			It("should inject pod with unprivileged mutator successfully", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{}
				for _, dataset := range testCtx.datasets {
					info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
					info.SetAPIReader(client)
					Expect(err).NotTo(HaveOccurred())
					runtimeInfos[dataset.Name] = info
				}

				out, err := injector.InjectPod(testCtx.in, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())

				// Check that the pod has the expected labels
				Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0", "fluid-test"))
				Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0", "dataset-1"))
				Expect(out.Spec.Containers).To(HaveLen(2))

				// Check that fuse container is not privileged
				fuseContainer := out.Spec.Containers[0]
				Expect(fuseContainer.Name).To(HavePrefix(common.FuseContainerName))
				Expect(fuseContainer.SecurityContext).NotTo(BeNil())
				if fuseContainer.SecurityContext.Privileged != nil {
					Expect(*fuseContainer.SecurityContext.Privileged).To(BeFalse())
				}

				// Check that capabilities don't include SYS_ADMIN
				if fuseContainer.SecurityContext.Capabilities != nil {
					Expect(fuseContainer.SecurityContext.Capabilities.Add).ShouldNot(ContainElement(corev1.Capability("SYS_ADMIN")))
				}

			})
		})

		When("pod has a init container mounting Fluid dataset's PVC", func() {
			BeforeEach(func() {
				testCtx.in.Spec.InitContainers = append(testCtx.in.Spec.InitContainers, corev1.Container{
					Name: "init-container",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      testCtx.datasets[0].Name,
							MountPath: "/init-data-" + testCtx.datasets[0].Name,
						},
					},
				})
			})

			It("should inject pod successfully", func() {
				runtimeInfos := map[string]base.RuntimeInfoInterface{}
				for _, dataset := range testCtx.datasets {
					info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
					info.SetAPIReader(client)
					Expect(err).NotTo(HaveOccurred())
					runtimeInfos[dataset.Name] = info
				}

				out, err := injector.InjectPod(testCtx.in, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())

				// Check that init container is preserved and has the correct volume mount
				fuseDs := testCtx.fuseDaemonsets[0]
				Expect(out.Spec.InitContainers).To(HaveLen(2))
				Expect(out.Spec.InitContainers[0].Name).To(HavePrefix(common.InitFuseContainerName))

				for _, vm := range fuseDs.Spec.Template.Spec.Containers[0].VolumeMounts {
					vm.Name = vm.Name + "-0"
					Expect(out.Spec.Containers[0].VolumeMounts).To(ContainElement(vm))
				}

				Expect(out.Spec.InitContainers[0].Command).To(ContainElements("sleep"))
				Expect(out.Spec.InitContainers[0].Args).To(ContainElements("2s"))

			})
		})
	})

	Context("Inject Pod mounting multiple Fluid Datasets", func() {
		BeforeEach(func() {
			testCtx = mockTestCaseContextFn([]string{"dataset-1", "dataset-2", "dataset-3"}, "fluid-test")
		})

		It("should inject Pod successfully with multiple Fluid Dataset PVC", func() {
			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			for _, dataset := range testCtx.datasets {
				info, err := base.BuildRuntimeInfo(dataset.Name, dataset.Namespace, common.ThinRuntime)
				info.SetAPIReader(client)
				Expect(err).NotTo(HaveOccurred())
				runtimeInfos[dataset.Name] = info
			}

			out, err := injector.InjectPod(testCtx.in, runtimeInfos)

			Expect(err).NotTo(HaveOccurred())
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0", "fluid-test"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0", "dataset-1"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-1", "fluid-test"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-1", "dataset-2"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-2", "fluid-test"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-2", "dataset-3"))
			Expect(out.Spec.Containers).To(HaveLen(4))

			for i := 0; i < len(testCtx.datasets); i++ {
				expectInjectionFn(out, testCtx.fuseDaemonsets[i], testCtx.datasets[i])
			}
		})
	})

	Context("Inject Pod mounting a ref dataset in another namespace", func() {
		BeforeEach(func() {
			testCtx = mockTestCaseContextFn([]string{"root-dataset"}, "fluid-test")
			testCtx.datasets = append(testCtx.datasets, &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sub-dataset",
					Namespace: "ref",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://fluid-test/root-dataset/path-a",
						},
					},
				},
			})

			testCtx.pvcs = append(testCtx.pvcs, &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "sub-dataset",
					Namespace: "ref",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "ref-sub-dataset",
				},
			})

			testCtx.pvs = append(testCtx.pvs, &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ref-sub-dataset",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath:    "/runtime-mnt/thin/fluid-test/root-dataset/thin-fuse",
								common.VolumeAttrMountType:    common.ThinRuntime,
								common.VolumeAttrFluidSubPath: "path-a",
							},
						},
					},
				},
			})

			testCtx.in.Namespace = "ref"
			testCtx.in.Spec.Volumes = []corev1.Volume{
				{
					Name: "sub-dataset",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "sub-dataset",
						},
					},
				},
			}
			testCtx.in.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
				{
					Name:      "sub-dataset",
					MountPath: "/data",
				},
			}

			subDatasetFuseDs := testCtx.fuseDaemonsets[0].DeepCopy()
			subDatasetFuseDs.Name = "sub-dataset-fuse"
			subDatasetFuseDs.Namespace = "ref"
			testCtx.fuseDaemonsets = append(testCtx.fuseDaemonsets, subDatasetFuseDs)
		})

		It("should inject pod successfully", func() {
			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			// when injecting pod mounting a reference dataset, only the reference dataset's runtime info is needed
			info, err := base.BuildRuntimeInfo(testCtx.datasets[1].Name, testCtx.datasets[1].Namespace, common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())
			info.SetAPIReader(client)
			runtimeInfos[testCtx.datasets[1].Name] = info
			out, err := injector.InjectPod(testCtx.in, runtimeInfos)

			fuseDs := testCtx.fuseDaemonsets[1]
			containerIdx := 0
			refDataset := testCtx.datasets[1]
			rootDataset := testCtx.datasets[0]
			Expect(err).NotTo(HaveOccurred())
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNamespaceKeyPrefix+"fluid-fuse-0", "ref"))
			Expect(out.ObjectMeta.Labels).To(HaveKeyWithValue(common.LabelContainerDatasetNameKeyPrefix+"fluid-fuse-0", "sub-dataset"))
			Expect(out.Spec.Containers).To(HaveLen(2))

			// Verify the fuse container command and args
			Expect(out.Spec.Containers[containerIdx].Command).To(ContainElements(fuseDs.Spec.Template.Spec.Containers[0].Command))
			Expect(out.Spec.Containers[containerIdx].Args).To(ContainElements(fuseDs.Spec.Template.Spec.Containers[0].Args))
			Expect(out.Spec.Containers[containerIdx].Env).To(ContainElements(fuseDs.Spec.Template.Spec.Containers[0].Env))

			for _, v := range fuseDs.Spec.Template.Spec.Volumes {
				v.Name = v.Name + "-0"
				Expect(out.Spec.Volumes).To(ContainElement(v))
			}
			Expect(out.Spec.Volumes).To(ContainElement(corev1.Volume{
				Name: "default-check-mount" + "-0",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: "thin-default-check-mount"},
						DefaultMode:          ptr.To[int32](0755),
					},
				},
			}))

			Expect(out.Spec.Volumes).To(ContainElement(corev1.Volume{
				Name:         refDataset.Name,
				VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse/%s", rootDataset.Namespace, rootDataset.Name, testCtx.pvs[1].Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath])}}},
			))
		})
	})
})

func TestInjectPodWithEnabledFUSEMetrics(t *testing.T) {
	type runtimeInfo struct {
		name         string
		namespace    string
		runtimeType  string
		scrapeTarget string
	}
	type testCase struct {
		name    string
		in      *corev1.Pod
		dataset *datav1alpha1.Dataset
		pv      *corev1.PersistentVolume
		pvc     *corev1.PersistentVolumeClaim
		fuse    *appsv1.DaemonSet
		infos   map[string]runtimeInfo
		want    *corev1.Pod
		wantErr error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
	bTrue := true
	bFalse := false
	var mode int32 = 0755

	testcases := []testCase{
		{
			name: "inject_pod_with_enabled_fuse_metrics",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "duplicate",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
									},
									Ports: []corev1.ContainerPort{
										{Name: "jindo-metrics", ContainerPort: 15000},
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "duplicate-pvc-name",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									}, VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "duplicate",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "duplicate",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/mnt/disk1",
											Type: &hostPathDirectoryOrCreate,
										},
									}},
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
											Path: "/runtime-mnt/jindo/big-data/duplicate",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"duplicate": {
					name:         "duplicate",
					namespace:    "big-data",
					runtimeType:  common.JindoRuntime,
					scrapeTarget: base.MountModeSelectAll,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s_%s", "big-data", "duplicate"),
					},
					Annotations: map[string]string{
						common.AnnotationPrometheusFuseMetricsScrapeKey: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
							},
							Ports: []corev1.ContainerPort{
								{Name: "jindo-metrics", ContainerPort: 15000},
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo ",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						}, {
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "duplicate",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_scrape_target_mount_pod_only",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate2",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate2",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate2/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate2",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate2",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "duplicate2",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate2-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
									},
									Ports: []corev1.ContainerPort{
										{Name: "jindo-metrics", ContainerPort: 15000},
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "duplicate-pvc-name",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									}, VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "duplicate",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "duplicate",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/mnt/disk1",
											Type: &hostPathDirectoryOrCreate,
										},
									}},
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
											Path: "/runtime-mnt/jindo/big-data/duplicate2",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"duplicate2": {
					name:         "duplicate2",
					namespace:    "big-data",
					runtimeType:  common.JindoRuntime,
					scrapeTarget: string(base.MountPodMountMode),
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s_%s", "big-data", "duplicate2"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo ",
										},
									},
								},
							},
							Ports:   []corev1.ContainerPort{},
							Command: []string{"/entrypoint.sh"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						}, {
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "duplicate",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate2/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate2",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_enabled_fuse_metrics_unprivileged",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate3",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate3",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate3/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate3",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate3",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
					},
					Annotations: nil,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "duplicate3",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate3-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
									},
									Ports: []corev1.ContainerPort{
										{Name: "jindo-metrics", ContainerPort: 15000},
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "duplicate-pvc-name",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									}, VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "duplicate",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "duplicate",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/mnt/disk1",
											Type: &hostPathDirectoryOrCreate,
										},
									}},
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
											Path: "/runtime-mnt/jindo/big-data/duplicate3",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"duplicate3": {
					name:         "duplicate3",
					namespace:    "big-data",
					runtimeType:  common.JindoRuntime,
					scrapeTarget: string(base.SidecarMountMode),
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s_%s", "big-data", "duplicate3"),
					},
					Annotations: map[string]string{
						common.AnnotationPrometheusFuseMetricsScrapeKey: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo ",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{Name: "jindo-metrics", ContainerPort: 15000},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						}, {
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "duplicate",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate3/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	for _, testcase := range testcases {
		objs = append(objs, testcase.fuse, testcase.pv, testcase.pvc, testcase.dataset)
	}

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	for _, testcase := range testcases {
		injector := NewInjector(fakeClient)

		runtimeInfos := map[string]base.RuntimeInfoInterface{}

		for pvc, info := range testcase.infos {
			opts := []base.RuntimeInfoOption{
				base.WithClientMetrics(datav1alpha1.ClientMetrics{
					Enabled:      true,
					ScrapeTarget: info.scrapeTarget,
				}),
			}
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType, opts...)
			if err != nil {
				t.Errorf("testcase %s failed due to error %v", testcase.name, err)
			}
			runtimeInfo.SetAPIReader(fakeClient)
			runtimeInfos[pvc] = runtimeInfo
		}

		out, err := injector.InjectPod(testcase.in, runtimeInfos)
		if err != nil {
			if testcase.wantErr == nil {
				t.Errorf("testcase %s failed, Got error %v", testcase.name, err)
			} else {
				continue
			}
		}

		gotMetaObj := out.ObjectMeta
		wantMetaObj := testcase.want.ObjectMeta

		if !reflect.DeepEqual(gotMetaObj, wantMetaObj) {
			t.Errorf("testcase %s failed, diff between want and got is: %v", testcase.name, cmp.Diff(gotMetaObj, wantMetaObj))
		}

		gotContainers := out.Spec.Containers
		gotVolumes := out.Spec.Volumes
		// gotContainers := out.
		// , gotVolumes, err := getInjectPiece(out)
		// if err != nil {
		// 	t.Errorf("testcase %s failed due to inject error %v", testcase.name, err)
		// }

		wantContainers := testcase.want.Spec.Containers
		wantVolumes := testcase.want.Spec.Volumes

		gotContainerMap := makeContainerMap(gotContainers)
		wantContainerMap := makeContainerMap(wantContainers)

		if len(gotContainerMap) != len(wantContainerMap) {
			t.Errorf("testcase %s failed, want containers length %d, Got containers length  %d", testcase.name, len(gotContainerMap), len(wantContainerMap))
		}

		for k, wantContainer := range wantContainerMap {
			if gotContainer, found := gotContainerMap[k]; found {
				if !reflect.DeepEqual(wantContainer, gotContainer) {
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the container %s", testcase.name, k)
			}
		}

		gotVolumeMap := makeVolumeMap(gotVolumes)
		wantVolumeMap := makeVolumeMap(wantVolumes)
		if len(gotVolumeMap) != len(wantVolumeMap) {
			gotVolumeKeys := keys(gotVolumeMap)
			wantVolumeKeys := keys(wantVolumeMap)
			t.Errorf("testcase %s failed, got volumes length %d with keys %v, want volumes length  %d with keys %v", testcase.name, len(gotVolumeMap),
				gotVolumeKeys, len(wantVolumeMap), wantVolumeKeys)
		}

		for k, wantVolume := range wantVolumeMap {
			if gotVolume, found := gotVolumeMap[k]; found {
				if !reflect.DeepEqual(wantVolume, gotVolume) {
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantVolume, gotVolume))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the volume %s", testcase.name, k)
			}
		}
	}
}

func makeContainerMap(containers []corev1.Container) (containerMap map[string]corev1.Container) {
	containerMap = map[string]corev1.Container{}
	for _, c := range containers {
		containerMap[c.Name] = c
	}
	return
}

func makeVolumeMap(volumes []corev1.Volume) (volumeMap map[string]corev1.Volume) {
	volumeMap = map[string]corev1.Volume{}
	for _, v := range volumes {
		volumeMap[v.Name] = v
	}
	return
}

func keys(vMap interface{}) (keys []string) {
	switch v := vMap.(type) {
	case map[string]corev1.Volume:
		for k := range v {
			keys = append(keys, k)
		}
	case map[string]corev1.Container:
		for k := range v {
			keys = append(keys, k)
		}
	}

	return
}
