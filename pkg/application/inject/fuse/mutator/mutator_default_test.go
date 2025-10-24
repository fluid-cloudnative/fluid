package mutator

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	applicationspod "github.com/fluid-cloudnative/fluid/pkg/utils/applications/pod"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Default mutator related unit tests", Label("pkg.application.inject.fuse.mutator.mutator_default_test.go"), func() {
	var scheme *runtime.Scheme

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = datav1alpha1.AddToScheme(scheme)
		_ = corev1.AddToScheme(scheme)
		_ = appsv1.AddToScheme(scheme)
	})

	When("the mutator injects a pod mounting one Fluid PVC", func() {
		var (
			datasetName      string = "test-dataset"
			datasetNamespace string = "fluid"
			dataset          *datav1alpha1.Dataset
			runtime          *datav1alpha1.ThinRuntime
			daemonSet        *appsv1.DaemonSet
			pv               *corev1.PersistentVolume
			podToMutate      *corev1.Pod

			client  client.Client
			mutator Mutator
			args    MutatorBuildArgs
		)
		BeforeEach(func() {
			dataset, runtime, daemonSet, pv = test_buildFluidResources(datasetName, datasetNamespace)
			podToMutate = test_buildPodToMutate([]string{datasetName})
		})

		JustBeforeEach(func() {
			client = fake.NewFakeClientWithScheme(scheme, dataset, runtime, daemonSet, pv)
			pod, err := applicationspod.NewApplication(podToMutate).GetPodSpecs()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pod).To(HaveLen(1))

			specs, err := CollectFluidObjectSpecs(pod[0])
			Expect(err).NotTo(HaveOccurred())

			args = MutatorBuildArgs{
				Client: client,
				Log:    fake.NullLogger(),
				Options: common.FuseSidecarInjectOption{
					EnableCacheDir:             false,
					SkipSidecarPostStartInject: false,
				},
				Specs: specs,
			}
		})

		It("should successfully mutate the pod and one fuse sidecar container will be injected", func() {
			By("mutate Pod", func() {
				mutator = NewDefaultMutator(args)
				runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
				Expect(err).NotTo(HaveOccurred())

				err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
				Expect(err).To(BeNil())

				err = mutator.PostMutate()
				Expect(err).To(BeNil())
			})

			By("check mutated Pod", func() {
				podSpecs := mutator.GetMutatedPodSpecs()
				Expect(podSpecs).NotTo(BeNil())

				mountPropagationBidirectional := corev1.MountPropagationBidirectional
				mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
				Expect(podSpecs.Containers).To(HaveLen(2))
				Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))
				Expect(podSpecs.Containers[0].VolumeMounts).To(ContainElement(
					corev1.VolumeMount{
						Name:             "thin-fuse-mount-0",
						MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
						MountPropagation: &mountPropagationBidirectional,
					}))
				Expect(podSpecs.Containers[0].VolumeMounts).To(ContainElement(
					corev1.VolumeMount{
						Name:      "default-check-mount-0",
						ReadOnly:  true,
						MountPath: "/check-mount.sh",
						SubPath:   "check-mount.sh",
					}))

				Expect(podSpecs.Containers[1].VolumeMounts).To(ContainElement(
					corev1.VolumeMount{
						Name:             "data-vol-0",
						MountPath:        "/data0",
						MountPropagation: &mountPropagationHostToContainer,
					},
				))
				Expect(podSpecs.Volumes).To(ContainElements(
					corev1.Volume{Name: "data-vol-0", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", datasetNamespace, datasetName)}}},
					corev1.Volume{Name: "default-check-mount-0", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-default-check-mount", pv.Spec.CSI.VolumeAttributes[common.VolumeAttrMountType])}, DefaultMode: ptr.To[int32](0755)}}},
				))
			})
		})

		When("SkipSidecarPostStartInject is true", func() {
			It("should inject a fuse sidecar container without postStart hook", func() {
				By("mutate Pod", func() {
					args.Options.SkipSidecarPostStartInject = true
					mutator = NewDefaultMutator(args)
					runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
					Expect(err).NotTo(HaveOccurred())

					err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
					Expect(err).To(BeNil())

					err = mutator.PostMutate()
					Expect(err).To(BeNil())
				})

				By("check Pod", func() {
					podSpecs := mutator.GetMutatedPodSpecs()
					Expect(podSpecs).NotTo(BeNil())

					Expect(podSpecs.Containers).To(HaveLen(2))
					Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))

					Expect(podSpecs.Containers[0].VolumeMounts).Should(Not(ContainElement(WithTransform(func(volumeMount corev1.VolumeMount) string { return volumeMount.Name }, HavePrefix("default-check-mount-0")))))
					Expect(podSpecs.Volumes).Should(Not(ContainElement(WithTransform(func(volume corev1.Volume) string { return volume.Name }, HavePrefix("default-check-mount-0")))))
				})
			})
		})

		When("FUSE daemonset contains customized fields", func() {
			BeforeEach(func() {
				daemonSet.Spec.Template.Spec.Containers[0].Env = append(daemonSet.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: "FOO", Value: "BAR"})
				daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: "myvol", MountPath: "/tmp/myvol"})
				daemonSet.Spec.Template.Spec.Volumes = append(daemonSet.Spec.Template.Spec.Volumes, corev1.Volume{Name: "myvol", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/myvol"}}})

				daemonSet.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2"), corev1.ResourceMemory: resource.MustParse("4Gi")},
					Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("4000m"), corev1.ResourceMemory: resource.MustParse("8Gi")},
				}
			})

			It("should inject a fuse sidecar container with customized fields", func() {
				By("mutate Pod", func() {
					mutator = NewDefaultMutator(args)
					runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
					Expect(err).NotTo(HaveOccurred())

					err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
					Expect(err).To(BeNil())

					err = mutator.PostMutate()
					Expect(err).To(BeNil())
				})

				By("check Pod", func() {
					podSpecs := mutator.GetMutatedPodSpecs()
					Expect(podSpecs).NotTo(BeNil())

					Expect(podSpecs.Containers).To(HaveLen(2))
					Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))

					Expect(podSpecs.Containers[0].VolumeMounts).Should(ContainElement(corev1.VolumeMount{Name: "myvol-0", MountPath: "/tmp/myvol"}))
					Expect(podSpecs.Volumes).Should(ContainElement(corev1.Volume{Name: "myvol-0", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/myvol"}}}))

					Expect(podSpecs.Containers[0].Env).Should(ContainElement(corev1.EnvVar{Name: "FOO", Value: "BAR"}))

					Expect(podSpecs.Containers[0].Resources.Requests).To(HaveKeyWithValue(corev1.ResourceCPU, resource.MustParse("2")))
					Expect(podSpecs.Containers[0].Resources.Requests).To(HaveKeyWithValue(corev1.ResourceMemory, resource.MustParse("4Gi")))
					Expect(podSpecs.Containers[0].Resources.Limits).To(HaveKeyWithValue(corev1.ResourceCPU, resource.MustParse("4")))
					Expect(podSpecs.Containers[0].Resources.Limits).To(HaveKeyWithValue(corev1.ResourceMemory, resource.MustParse("8Gi")))
				})
			})
		})

		When("EnableCacheDir is true", func() {
			BeforeEach(func() {
				daemonSet.Spec.Template.Spec.Volumes = append(daemonSet.Spec.Template.Spec.Volumes, corev1.Volume{Name: "cache-dir", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/cache-dir"}}})
				daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(daemonSet.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: "cache-dir", MountPath: "/tmp/cache-dir"})
			})
			It("should inject a fuse sidecar container with cache dir volume", func() {
				By("mutate Pod", func() {
					args.Options.EnableCacheDir = true
					mutator = NewDefaultMutator(args)
					runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
					Expect(err).NotTo(HaveOccurred())

					err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
					Expect(err).To(BeNil())

					err = mutator.PostMutate()
					Expect(err).To(BeNil())
				})

				By("check Pod", func() {
					podSpecs := mutator.GetMutatedPodSpecs()
					Expect(podSpecs).NotTo(BeNil())

					Expect(podSpecs.Containers).To(HaveLen(2))
					Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))

					// When EnableCacheDir is true, cache related volumes should be kept
					// Check if cache-dir volume mount exists (should be kept)
					Expect(podSpecs.Containers[0].VolumeMounts).To(ContainElement(
						WithTransform(func(vm corev1.VolumeMount) string { return vm.Name }, ContainSubstring("cache-dir")),
					))

					// Check if cache-dir volume exists (should be kept)
					Expect(podSpecs.Volumes).To(ContainElement(
						WithTransform(func(v corev1.Volume) string { return v.Name }, ContainSubstring("cache-dir")),
					))
				})
			})
		})

		When("fluid pvc has defined subpath", func() {
			BeforeEach(func() {
				pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath] = "path-a"
			})

			It("should inject a fuse sidecar container and mutate the pvc to a hostpath volume with subpath", func() {
				By("mutate Pod", func() {
					mutator = NewDefaultMutator(args)
					runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
					Expect(err).NotTo(HaveOccurred())

					err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
					Expect(err).To(BeNil())

					err = mutator.PostMutate()
					Expect(err).To(BeNil())
				})

				By("check mutated Pod", func() {
					podSpecs := mutator.GetMutatedPodSpecs()
					Expect(podSpecs).NotTo(BeNil())

					// Check the hostPath is correctly constructed with subpath
					expectedHostPath := fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse/path-a", datasetNamespace, datasetName)
					Expect(podSpecs.Volumes).To(ContainElement(
						corev1.Volume{
							Name: "data-vol-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: expectedHostPath,
								},
							},
						},
					))
				})
			})
		})

		When("fluid pvc is also mounted on the pod's init container", func() {
			BeforeEach(func() {
				// Add an init container that also mounts the Fluid PVC
				podToMutate.Spec.InitContainers = []corev1.Container{
					{
						Name:  "init-container",
						Image: "init-image",
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "data-vol-0",
								MountPath: "/data0",
							},
						},
					},
				}
			})

			It("should inject a fuse sidecar container into both app container and init container", func() {
				By("mutate Pod", func() {
					mutator = NewDefaultMutator(args)
					runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
					Expect(err).NotTo(HaveOccurred())

					err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
					Expect(err).To(BeNil())

					err = mutator.PostMutate()
					Expect(err).To(BeNil())
				})

				By("check mutated Pod", func() {
					podSpecs := mutator.GetMutatedPodSpecs()
					Expect(podSpecs).NotTo(BeNil())

					// Should have 2 containers (fuse sidecar + app container) and 1 init container (fuse sidecar)
					Expect(podSpecs.Containers).To(HaveLen(2))
					Expect(podSpecs.InitContainers).To(HaveLen(2))
					Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))
					Expect(podSpecs.InitContainers[0].Name).To(HavePrefix(common.InitFuseContainerName))

					// Both containers should have the dataset volume mount with proper mount propagation
					mountPropagationBidirectional := corev1.MountPropagationBidirectional
					mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

					// App container fuse sidecar should have bidirectional mount propagation
					Expect(podSpecs.Containers[0].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             "thin-fuse-mount-0",
							MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
							MountPropagation: &mountPropagationBidirectional,
						}))

					// Init container fuse sidecar should not have mount propagation or post start hook
					Expect(podSpecs.InitContainers[0].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             "thin-fuse-mount-0",
							MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
							MountPropagation: &mountPropagationBidirectional,
						}))
					Expect(podSpecs.InitContainers[0].Lifecycle).To(BeNil())

					// App container should have the dataset volume mount with host to container propagation
					Expect(podSpecs.Containers[1].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             "data-vol-0",
							MountPath:        "/data0",
							MountPropagation: &mountPropagationHostToContainer,
						}))

					// Init container should have the dataset volume mount with host to container propagation
					Expect(podSpecs.InitContainers[1].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             "data-vol-0",
							MountPath:        "/data0",
							MountPropagation: &mountPropagationHostToContainer,
						}))
				})
			})
		})

		When("the pod to mutate specifies random-suffix host path mode", func() {
			BeforeEach(func() {
				podToMutate.ObjectMeta.Annotations = map[string]string{
					common.HostMountPathModeOnDefaultPlatformKey: string(common.HostPathModeRandomSuffix),
				}
			})

			When("the pod to mutate has a non-empty name", func() {
				It("should add or mutate host path volumes with random suffix when pod has non-empty name", func() {
					By("mutate Pod", func() {
						mutator = NewDefaultMutator(args)
						runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
						Expect(err).NotTo(HaveOccurred())

						err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
						Expect(err).To(BeNil())

						err = mutator.PostMutate()
						Expect(err).To(BeNil())
					})

					By("check mutated Pod", func() {
						podSpecs := mutator.GetMutatedPodSpecs()
						Expect(podSpecs).NotTo(BeNil())

						Expect(podSpecs.Containers).To(HaveLen(2))
						Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))

						matchedVolume := corev1.Volume{}
						Expect(podSpecs.Volumes).To(ContainElement(WithTransform(func(volume corev1.Volume) string { return volume.Name }, Equal("thin-fuse-mount-0")), &matchedVolume))
						Expect(matchedVolume.HostPath.Path).To(MatchRegexp("^/runtime-mnt/thin/fluid/test-dataset//test-pod/\\d+-[0-9a-z]{1,8}$"))

						matchedMutatedVolume := corev1.Volume{}
						Expect(podSpecs.Volumes).To(ContainElement(WithTransform(func(volume corev1.Volume) string { return volume.Name }, Equal("data-vol-0")), &matchedMutatedVolume))
						Expect(matchedMutatedVolume.HostPath.Path).To(MatchRegexp("^/runtime-mnt/thin/fluid/test-dataset/test-pod/\\d+-[0-9a-z]{1,8}/thin-fuse$"))
					})
				})
			})

			When("pod has generate name", func() {
				BeforeEach(func() {
					podToMutate.ObjectMeta.GenerateName = "mypod-"
					podToMutate.ObjectMeta.Name = ""
				})

				It("should add or mutate host path volumes with random suffix and generate name when pod has generate name", func() {
					By("mutate Pod", func() {
						mutator = NewDefaultMutator(args)
						runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
						Expect(err).NotTo(HaveOccurred())

						err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
						Expect(err).To(BeNil())

						err = mutator.PostMutate()
						Expect(err).To(BeNil())
					})

					By("check mutated Pod", func() {
						podSpecs := mutator.GetMutatedPodSpecs()
						Expect(podSpecs).NotTo(BeNil())

						Expect(podSpecs.Containers).To(HaveLen(2))
						Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))

						matchedVolume := corev1.Volume{}
						Expect(podSpecs.Volumes).To(ContainElement(WithTransform(func(volume corev1.Volume) string { return volume.Name }, Equal("thin-fuse-mount-0")), &matchedVolume))
						Expect(matchedVolume.HostPath.Path).To(MatchRegexp("^/runtime-mnt/thin/fluid/test-dataset//mypod---generate-name/\\d+-[0-9a-z]{1,8}$"))

						matchedMutatedVolume := corev1.Volume{}
						Expect(podSpecs.Volumes).To(ContainElement(WithTransform(func(volume corev1.Volume) string { return volume.Name }, Equal("data-vol-0")), &matchedMutatedVolume))
						Expect(matchedMutatedVolume.HostPath.Path).To(MatchRegexp("^/runtime-mnt/thin/fluid/test-dataset/mypod---generate-name/\\d+-[0-9a-z]{1,8}/thin-fuse$"))
					})
				})
			})
		})

		When("the mutator uses native-sidecar injection mode", func() {
			When("pod.spec.containers mounts a Fluid PVC", func() {
				It("should inject a native fuse sidecar container into init container", func() {
					By("mutate Pod", func() {
						args.Options.SidecarInjectionMode = common.SidecarInjectionMode_NativeSidecar
						mutator = NewDefaultMutator(args)
						runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
						Expect(err).NotTo(HaveOccurred())

						err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
						Expect(err).To(BeNil())

						err = mutator.PostMutate()
						Expect(err).To(BeNil())
					})

					By("check mutated Pod", func() {
						podSpecs := mutator.GetMutatedPodSpecs()
						Expect(podSpecs).NotTo(BeNil())

						Expect(podSpecs.Containers).To(HaveLen(1))
						Expect(podSpecs.InitContainers).To(HaveLen(1))
						Expect(podSpecs.InitContainers[0].Name).To(HavePrefix(common.FuseContainerName))
						containerRestartPolicyAlways := corev1.ContainerRestartPolicyAlways
						Expect(podSpecs.InitContainers[0].RestartPolicy).To(Equal(&containerRestartPolicyAlways))
					})
				})
			})

			When("both pod.spec.containers and pod.spec.initContainers mount the same Fluid PVC", func() {
				BeforeEach(func() {
					// Add an init container that also mounts the Fluid PVC
					podToMutate.Spec.InitContainers = []corev1.Container{
						{
							Name:  "init-container",
							Image: "init-image",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-vol-0",
									MountPath: "/data0",
								},
							},
						},
					}
				})
				It("should inject ONLY ONE native fuse sidecar container into init container", func() {
					By("mutate Pod", func() {
						args.Options.SidecarInjectionMode = common.SidecarInjectionMode_NativeSidecar
						mutator = NewDefaultMutator(args)
						runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
						Expect(err).NotTo(HaveOccurred())

						err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, "-0")
						Expect(err).To(BeNil())

						err = mutator.PostMutate()
						Expect(err).To(BeNil())
					})

					By("check mutated Pod", func() {
						podSpecs := mutator.GetMutatedPodSpecs()
						Expect(podSpecs).NotTo(BeNil())

						Expect(podSpecs.Containers).To(HaveLen(1))
						Expect(podSpecs.InitContainers).To(HaveLen(2)) // one native sidecar + one app init container
						Expect(podSpecs.InitContainers[0].Name).To(HavePrefix(common.FuseContainerName))

						containerRestartPolicyAlways := corev1.ContainerRestartPolicyAlways
						Expect(podSpecs.InitContainers[0].RestartPolicy).To(Equal(&containerRestartPolicyAlways))
					})
				})
			})
		})
	})

	When("the mutator injects a Pod with multiple Fluid PVCs", func() {
		var (
			datasetNum int = 3
		)

		var (
			datasetNames      []string
			datasetNamespaces []string
			datasets          []*datav1alpha1.Dataset
			runtimes          []*datav1alpha1.ThinRuntime
			daemonSets        []*appsv1.DaemonSet
			pvs               []*corev1.PersistentVolume

			objs []runtime.Object
		)

		var (
			podToMutate *corev1.Pod

			client  client.Client
			mutator Mutator
			args    MutatorBuildArgs
		)
		BeforeEach(func() {
			for i := 0; i < datasetNum; i++ {
				datasetName := fmt.Sprintf("test-dataset-%d", i)
				datasetNamespace := "fluid"
				dataset, runtime, daemonSet, pv := test_buildFluidResources(datasetName, datasetNamespace)

				datasetNames = append(datasetNames, datasetName)
				datasetNamespaces = append(datasetNamespaces, datasetNamespace)
				datasets = append(datasets, dataset)
				runtimes = append(runtimes, runtime)
				daemonSets = append(daemonSets, daemonSet)
				pvs = append(pvs, pv)

				objs = append(objs, dataset, runtime, daemonSet, pv)
			}

			podToMutate = test_buildPodToMutate(datasetNames)
		})

		JustBeforeEach(func() {
			client = fake.NewFakeClientWithScheme(scheme, objs...)
			pod, err := applicationspod.NewApplication(podToMutate).GetPodSpecs()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pod).To(HaveLen(1))

			specs, err := CollectFluidObjectSpecs(pod[0])
			Expect(err).NotTo(HaveOccurred())

			args = MutatorBuildArgs{
				Client: client,
				Log:    fake.NullLogger(),
				Options: common.FuseSidecarInjectOption{
					EnableCacheDir:             false,
					SkipSidecarPostStartInject: false,
				},
				Specs: specs,
			}
		})

		It("should successfully mutate the pod and multiple fuse sidecar containers should be injected", func() {
			mutator = NewDefaultMutator(args)

			for i := 0; i < datasetNum; i++ {
				By(fmt.Sprintf("mutate pvc No.%d", i), func() {
					datasetName := datasetNames[i]
					datasetNamespace := datasetNamespaces[i]
					runtimeInfo, err := base.GetRuntimeInfo(client, datasetName, datasetNamespace)
					Expect(err).NotTo(HaveOccurred())

					err = mutator.MutateWithRuntimeInfo(datasetName, runtimeInfo, fmt.Sprintf("-%d", i))
					Expect(err).To(BeNil())
				})

				By(fmt.Sprintf("check pvc No.%d is mutated successfully", i), func() {
					podSpecs := mutator.GetMutatedPodSpecs()
					Expect(podSpecs).NotTo(BeNil())

					mountPropagationBidirectional := corev1.MountPropagationBidirectional
					mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
					Expect(podSpecs.Containers).To(HaveLen(i + 2))
					Expect(podSpecs.Containers[0].Name).To(HavePrefix(common.FuseContainerName))
					Expect(podSpecs.Containers[0].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             fmt.Sprintf("thin-fuse-mount-%d", i),
							MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespaces[i], datasetNames[i]),
							MountPropagation: &mountPropagationBidirectional,
						}))
					Expect(podSpecs.Containers[0].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:      fmt.Sprintf("default-check-mount-%d", i),
							ReadOnly:  true,
							MountPath: "/check-mount.sh",
							SubPath:   "check-mount.sh",
						}))

					Expect(podSpecs.Containers[len(podSpecs.Containers)-1].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             fmt.Sprintf("data-vol-%d", i),
							MountPath:        fmt.Sprintf("/data%d", i),
							MountPropagation: &mountPropagationHostToContainer,
						},
					))
					Expect(podSpecs.Volumes).To(ContainElements(
						corev1.Volume{Name: fmt.Sprintf("data-vol-%d", i), VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: fmt.Sprintf("/runtime-mnt/thin/%s/%s/thin-fuse", datasetNamespaces[i], datasetNames[i])}}},
						corev1.Volume{Name: fmt.Sprintf("default-check-mount-%d", i), VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-default-check-mount", pvs[i].Spec.CSI.VolumeAttributes[common.VolumeAttrMountType])}, DefaultMode: ptr.To[int32](0755)}}},
					))
				})
			}

			By("PostMutate should successfully", func() {
				err := mutator.PostMutate()
				Expect(err).To(BeNil())
			})
		})
	})
})
