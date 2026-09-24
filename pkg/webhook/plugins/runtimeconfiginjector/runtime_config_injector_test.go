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

package runtimeconfiginjector

import (
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testNamespace = "fluid-test"

// configMapFor builds the config map that fluid's reconcile would have created for a dataset.
func configMapFor(datasetName string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.GetCacheRuntimeConfigConfigMapName(datasetName),
			Namespace: testNamespace,
		},
		Data: map[string]string{
			common.RuntimeConfigJSONFileName:  `{"master":{"name":"demo-master"}}`,
			common.RuntimeConfigShellFileName: "#!/bin/sh\nexport MASTER_NAME='demo-master'\n",
		},
	}
}

// runtimeInfosFor stands in for what the handler resolves from the pod's annotation.
func runtimeInfosFor(datasetNames ...string) map[string]base.RuntimeInfoInterface {
	runtimeInfos := map[string]base.RuntimeInfoInterface{}
	for _, datasetName := range datasetNames {
		runtimeInfo, err := base.BuildRuntimeInfo(datasetName, testNamespace, "")
		Expect(err).NotTo(HaveOccurred())
		runtimeInfos[datasetName] = runtimeInfo
	}
	return runtimeInfos
}

func newInjector(objs ...runtime.Object) *RuntimeConfigInjector {
	plugin, err := NewPlugin(fake.NewFakeClient(objs...), "")
	Expect(err).NotTo(HaveOccurred())

	injector, ok := plugin.(*RuntimeConfigInjector)
	Expect(ok).To(BeTrue())
	return injector
}

func podWithContainers(containerNames ...string) *corev1.Pod {
	containers := make([]corev1.Container, 0, len(containerNames))
	for _, containerName := range containerNames {
		containers = append(containers, corev1.Container{Name: containerName})
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: testNamespace},
		Spec:       corev1.PodSpec{Containers: containers},
	}
}

var _ = Describe("RuntimeConfigInjector", func() {
	Describe("GetName", func() {
		It("should report the name the plugin is registered under", func() {
			Expect(newInjector().GetName()).To(Equal(Name))
		})
	})

	Describe("Mutate", func() {
		Context("when the pod is already marked as injected", func() {
			It("should leave the pod untouched", func() {
				injector := newInjector(configMapFor("demo"))
				pod := podWithContainers("app")
				pod.Labels = map[string]string{common.InjectSidecarDone: common.True}

				shouldStop, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())
				Expect(pod.Spec.Volumes).To(BeEmpty())
				Expect(pod.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})

		Context("when the pod uses no dataset", func() {
			It("should leave the pod untouched", func() {
				injector := newInjector()
				pod := podWithContainers("app")

				shouldStop, err := injector.Mutate(pod, map[string]base.RuntimeInfoInterface{})
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())
				Expect(pod.Spec.Volumes).To(BeEmpty())
				Expect(pod.Labels).NotTo(HaveKey(common.InjectSidecarDone))
			})
		})

		Context("when one dataset is declared", func() {
			It("should inject the volume, the mount and the env pointer", func() {
				injector := newInjector(configMapFor("demo"))
				pod := podWithContainers("app")

				shouldStop, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).NotTo(HaveOccurred())
				Expect(shouldStop).To(BeFalse())

				volumeName := common.GetCacheRuntimeConfigConfigMapName("demo")

				Expect(pod.Spec.Volumes).To(HaveLen(1))
				Expect(pod.Spec.Volumes[0].Name).To(Equal(volumeName))
				Expect(pod.Spec.Volumes[0].ConfigMap).NotTo(BeNil())
				Expect(pod.Spec.Volumes[0].ConfigMap.Name).To(Equal(volumeName))

				volumeMounts := pod.Spec.Containers[0].VolumeMounts
				Expect(volumeMounts).To(HaveLen(1))
				// The mount has to reference the volume by the very same name, or the pod
				// is rejected by the api server.
				Expect(volumeMounts[0].Name).To(Equal(volumeName))
				Expect(volumeMounts[0].MountPath).To(Equal(common.RuntimeConfigDir + "/demo"))
				Expect(volumeMounts[0].ReadOnly).To(BeTrue())

				envs := pod.Spec.Containers[0].Env
				Expect(envs).To(HaveLen(1))
				Expect(envs[0].Name).To(Equal(common.EnvRuntimeConfigPathPrefix + "DEMO"))
				Expect(envs[0].Value).To(Equal(
					common.RuntimeConfigDir + "/demo/" + common.RuntimeConfigShellFileName))

				Expect(pod.Labels).To(HaveKeyWithValue(common.InjectSidecarDone, common.True))
			})
		})

		Context("when the dataset name contains a hyphen", func() {
			It("should render an env name a shell can reference", func() {
				injector := newInjector(configMapFor("mooncake-demo"))
				pod := podWithContainers("app")

				_, err := injector.Mutate(pod, runtimeInfosFor("mooncake-demo"))
				Expect(err).NotTo(HaveOccurred())

				// $FLUID_RUNTIME_CONFIG_PATH_mooncake-demo would not parse in a shell,
				// since a variable name cannot hold a hyphen.
				Expect(pod.Spec.Containers[0].Env[0].Name).To(
					Equal(common.EnvRuntimeConfigPathPrefix + "MOONCAKE_DEMO"))
			})
		})

		Context("when several datasets are declared", func() {
			It("should inject each of them in a stable order", func() {
				injector := newInjector(configMapFor("alpha"), configMapFor("beta"))

				// Iterating a map has no defined order, so the same input has to be run
				// more than once to catch an unstable rendering.
				for i := 0; i < 5; i++ {
					pod := podWithContainers("app")

					_, err := injector.Mutate(pod, runtimeInfosFor("beta", "alpha"))
					Expect(err).NotTo(HaveOccurred())

					Expect(pod.Spec.Volumes).To(HaveLen(2))
					Expect(pod.Spec.Volumes[0].Name).To(ContainSubstring("alpha"))
					Expect(pod.Spec.Volumes[1].Name).To(ContainSubstring("beta"))

					envs := pod.Spec.Containers[0].Env
					Expect(envs).To(HaveLen(2))
					Expect(envs[0].Name).To(Equal(common.EnvRuntimeConfigPathPrefix + "ALPHA"))
					Expect(envs[1].Name).To(Equal(common.EnvRuntimeConfigPathPrefix + "BETA"))
				}
			})
		})

		Context("when the pod has init containers and a fuse sidecar", func() {
			It("should inject into the app and init containers but not the fuse sidecar", func() {
				injector := newInjector(configMapFor("demo"))
				pod := podWithContainers("app", common.FuseContainerName+"-0")
				pod.Spec.InitContainers = []corev1.Container{{Name: "init"}}

				_, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).NotTo(HaveOccurred())

				Expect(pod.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(pod.Spec.Containers[0].Env).To(HaveLen(1))
				// fluid's own fuse sidecar has no use for the config and is skipped.
				Expect(pod.Spec.Containers[1].VolumeMounts).To(BeEmpty())
				Expect(pod.Spec.Containers[1].Env).To(BeEmpty())
				// An app may well read the config from an init container.
				Expect(pod.Spec.InitContainers[0].VolumeMounts).To(HaveLen(1))
				Expect(pod.Spec.InitContainers[0].Env).To(HaveLen(1))
			})
		})

		Context("when the config map does not exist", func() {
			It("should fail and leave the pod untouched", func() {
				injector := newInjector()
				pod := podWithContainers("app")

				shouldStop, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
				Expect(shouldStop).To(BeTrue())

				// A plain plugin error does not stop the pod from being admitted, so a
				// failure must leave nothing behind rather than half an injection.
				Expect(pod.Spec.Volumes).To(BeEmpty())
				Expect(pod.Spec.Containers[0].VolumeMounts).To(BeEmpty())
				Expect(pod.Spec.Containers[0].Env).To(BeEmpty())
				Expect(pod.Labels).NotTo(HaveKey(common.InjectSidecarDone))
			})
		})

		Context("when one of several config maps is missing", func() {
			It("should inject nothing at all", func() {
				injector := newInjector(configMapFor("alpha"))
				pod := podWithContainers("app")

				_, err := injector.Mutate(pod, runtimeInfosFor("alpha", "beta"))
				Expect(err).To(HaveOccurred())

				// "alpha" resolves and "beta" does not, which is exactly the case that
				// would leave a partially injected pod if validation and injection were
				// interleaved instead of separated.
				Expect(pod.Spec.Volumes).To(BeEmpty())
				Expect(pod.Spec.Containers[0].VolumeMounts).To(BeEmpty())
			})
		})

		Context("when the config map carries no shell rendering", func() {
			It("should fail with a message naming the missing key", func() {
				configMap := configMapFor("demo")
				delete(configMap.Data, common.RuntimeConfigShellFileName)

				injector := newInjector(configMap)
				pod := podWithContainers("app")

				shouldStop, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(common.RuntimeConfigShellFileName))
				Expect(shouldStop).To(BeTrue())
				Expect(pod.Spec.Volumes).To(BeEmpty())
			})
		})

		Context("when the dataset name is too long for a volume name", func() {
			It("should fail before touching the pod", func() {
				longName := strings.Repeat("a", maxDatasetNameLen+1)

				injector := newInjector(configMapFor(longName))
				pod := podWithContainers("app")

				shouldStop, err := injector.Mutate(pod, runtimeInfosFor(longName))
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too long"))
				Expect(shouldStop).To(BeTrue())
				Expect(pod.Spec.Volumes).To(BeEmpty())
			})
		})

		Context("when the same pod is mutated twice", func() {
			It("should not duplicate the injected volume, mount or env", func() {
				injector := newInjector(configMapFor("demo"))
				pod := podWithContainers("app")

				_, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).NotTo(HaveOccurred())

				// The done label makes a second call a no-op, so drop it to exercise the
				// append-or-override helpers themselves.
				delete(pod.Labels, common.InjectSidecarDone)

				_, err = injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).NotTo(HaveOccurred())

				Expect(pod.Spec.Volumes).To(HaveLen(1))
				Expect(pod.Spec.Containers[0].VolumeMounts).To(HaveLen(1))
				Expect(pod.Spec.Containers[0].Env).To(HaveLen(1))
			})
		})

		Context("when the pod already has volumes and env of its own", func() {
			It("should keep them", func() {
				injector := newInjector(configMapFor("demo"))
				pod := podWithContainers("app")
				pod.Spec.Volumes = []corev1.Volume{{
					Name:         "workdir",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}}
				pod.Spec.Containers[0].Env = []corev1.EnvVar{{Name: "POD_IP", Value: "10.0.0.1"}}
				pod.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{
					Name:      "workdir",
					MountPath: "/work",
				}}

				_, err := injector.Mutate(pod, runtimeInfosFor("demo"))
				Expect(err).NotTo(HaveOccurred())

				Expect(pod.Spec.Volumes).To(HaveLen(2))
				Expect(pod.Spec.Volumes[0].Name).To(Equal("workdir"))
				Expect(pod.Spec.Containers[0].Env).To(HaveLen(2))
				Expect(pod.Spec.Containers[0].Env[0].Name).To(Equal("POD_IP"))
				Expect(pod.Spec.Containers[0].VolumeMounts).To(HaveLen(2))
			})
		})
	})
})

var _ = Describe("envSuffix", func() {
	It("should upper-case the name and replace hyphens", func() {
		// A dataset name is a DNS-1035 label: lower case, digits and hyphens, starting
		// with a letter. Hyphens are the only character a shell variable name rejects,
		// and since '_' is not a legal dataset name character the mapping cannot make
		// two different datasets collide.
		Expect(envSuffix("demo")).To(Equal("DEMO"))
		Expect(envSuffix("mooncake-demo")).To(Equal("MOONCAKE_DEMO"))
		Expect(envSuffix("ds-1-a")).To(Equal("DS_1_A"))
	})
})
