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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unprivileged mutator related unit tests", Label("pkg.application.inject.fuse.mutator.mutator_unprivileged_test.go"), func() {
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

			mutator = NewUnprivilegedMutator(args)
		})

		It("should successfully mutate the pod and one fuse sidecar container will be injected", func() {
			By("mutate Pod", func() {
				mutator = NewUnprivilegedMutator(args)
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
				Expect(podSpecs.Containers[0].SecurityContext.Privileged).To(Equal(ptr.To(false)))
				Expect(podSpecs.Containers[0].SecurityContext.Capabilities.Add).ShouldNot(ContainElement("SYS_ADMIN"))
				Expect(podSpecs.Containers[0].VolumeMounts).ShouldNot(ContainElement(WithTransform(func(vm corev1.VolumeMount) *corev1.MountPropagationMode { return vm.MountPropagation }, Equal(&mountPropagationBidirectional))))
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

		When("fluid pvc has defined subpath", func() {
			BeforeEach(func() {
				pv.Spec.CSI.VolumeAttributes[common.VolumeAttrFluidSubPath] = "path-a"
			})

			It("should inject a fuse sidecar container and mutate the pvc to a hostpath volume with subpath", func() {
				By("mutate Pod", func() {
					mutator = NewUnprivilegedMutator(args)
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
					Expect(podSpecs.Containers[0].SecurityContext.Privileged).To(Equal(ptr.To(false)))
					Expect(podSpecs.Containers[0].SecurityContext.Capabilities.Add).ShouldNot(ContainElement("SYS_ADMIN"))

					mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
					// Check the hostPath is correctly constructed with subpath
					Expect(podSpecs.Containers[1].VolumeMounts).To(ContainElement(
						corev1.VolumeMount{
							Name:             "data-vol-0",
							MountPath:        "/data0",
							MountPropagation: &mountPropagationHostToContainer,
						},
					))
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
					mutator = NewUnprivilegedMutator(args)
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

					Expect(podSpecs.Containers[0].SecurityContext.Privileged).To(Equal(ptr.To(false)))
					Expect(podSpecs.Containers[0].SecurityContext.Capabilities.Add).ShouldNot(ContainElement(corev1.Capability("SYS_ADMIN")))
					Expect(podSpecs.InitContainers[0].SecurityContext.Privileged).To(Equal(ptr.To(false)))
					Expect(podSpecs.InitContainers[0].SecurityContext.Capabilities.Add).ShouldNot(ContainElement(corev1.Capability("SYS_ADMIN")))

					// Both containers should have the dataset volume mount with proper mount propagation
					mountPropagationBidirectional := corev1.MountPropagationBidirectional
					mountPropagationHostToContainer := corev1.MountPropagationHostToContainer

					// App container fuse sidecar should have bidirectional mount propagation
					Expect(podSpecs.Containers[0].VolumeMounts).ShouldNot(ContainElement(
						corev1.VolumeMount{
							Name:             "thin-fuse-mount-0",
							MountPath:        fmt.Sprintf("/runtime-mnt/thin/%s/%s/", datasetNamespace, datasetName),
							MountPropagation: &mountPropagationBidirectional,
						}))

					// Init container fuse sidecar should not have mount propagation or post start hook
					Expect(podSpecs.InitContainers[0].VolumeMounts).ShouldNot(ContainElement(
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
	})
})
