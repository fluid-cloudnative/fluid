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

package cacheruntime

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// baseSpec is a CacheRuntime spec with every field populated, so that a test
// can change any one of them and watch the validator react.
func baseSpec() datav1alpha1.CacheRuntimeSpec {
	common := func(options map[string]string) datav1alpha1.RuntimeComponentCommonSpec {
		return datav1alpha1.RuntimeComponentCommonSpec{
			RuntimeVersion: datav1alpha1.VersionSpec{Image: "registry/cache", ImageTag: "v1", ImagePullPolicy: "IfNotPresent"},
			Options:        options,
			PodMetadata:    datav1alpha1.PodMetadata{Labels: map[string]string{"tier": "cache"}},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("2Gi")},
			},
			Env:          []corev1.EnvVar{{Name: "LOG_LEVEL", Value: "info"}},
			VolumeMounts: []corev1.VolumeMount{{Name: "logs", MountPath: "/var/log/cache"}},
			Args:         []string{"--foreground"},
			NodeSelector: map[string]string{"cache": "true"},
			Tolerations:  []corev1.Toleration{{Key: "cache", Operator: corev1.TolerationOpExists}},
		}
	}
	return datav1alpha1.CacheRuntimeSpec{
		RuntimeClassName: "curvine",
		Master: datav1alpha1.CacheRuntimeMasterSpec{
			RuntimeComponentCommonSpec: common(map[string]string{"rpc_port": "8995"}),
			Replicas:                   1,
		},
		Worker: datav1alpha1.CacheRuntimeWorkerSpec{
			RuntimeComponentCommonSpec: common(map[string]string{"rpc_port": "8997"}),
			Replicas:                   2,
			TieredStore: datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
					ProcessMemory: &datav1alpha1.ProcessMemoryMediumSource{Quota: resource.MustParse("2Gi")},
				}},
			},
		},
		Client: datav1alpha1.CacheRuntimeClientSpec{
			RuntimeComponentCommonSpec: common(map[string]string{"debug": "false"}),
			TieredStore: datav1alpha1.RuntimeTieredStore{
				Levels: []datav1alpha1.RuntimeTieredStoreLevel{{
					EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("1Gi")},
				}},
			},
			CleanPolicy: datav1alpha1.OnRuntimeDeletedCleanPolicy,
		},
		Options:          map[string]string{"global": "a"},
		PodMetadata:      datav1alpha1.PodMetadata{Annotations: map[string]string{"owner": "team-a"}},
		ImagePullSecrets: []corev1.LocalObjectReference{{Name: "registry-secret"}},
		Volumes:          []corev1.Volume{{Name: "logs", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
	}
}

var _ = Describe("ValidateSpecUpdate", func() {
	type mutate func(spec *datav1alpha1.CacheRuntimeSpec)

	// paths returns the field paths of errs, in order.
	paths := func(errs field.ErrorList) []string {
		var got []string
		for _, err := range errs {
			got = append(got, err.Field)
		}
		return got
	}

	It("accepts an unchanged spec", func() {
		oldSpec, newSpec := baseSpec(), baseSpec()
		Expect(ValidateSpecUpdate(&oldSpec, &newSpec)).To(BeEmpty())
	})

	DescribeTable("accepts a change to a field the controller syncs in place",
		func(m mutate) {
			oldSpec, newSpec := baseSpec(), baseSpec()
			m(&newSpec)
			Expect(ValidateSpecUpdate(&oldSpec, &newSpec)).To(BeEmpty())
		},
		Entry("master replicas", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.Replicas = 3 })),
		Entry("worker replicas", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Worker.Replicas = 5 })),
		Entry("master runtimeVersion", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.RuntimeVersion.ImageTag = "v2" })),
		Entry("worker runtimeVersion", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Worker.RuntimeVersion.Image = "registry/other" })),
		Entry("master resources", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Master.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
		})),
		Entry("worker resources", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Worker.Resources.Requests[corev1.ResourceCPU] = resource.MustParse("1")
		})),
		Entry("all of them at once", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Master.Replicas, s.Worker.Replicas = 3, 5
			s.Master.RuntimeVersion.ImageTag, s.Worker.RuntimeVersion.ImageTag = "v2", "v2"
			s.Master.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
			s.Worker.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
		})),
	)

	DescribeTable("accepts a rewrite that does not change the field value",
		func(prepare func(oldSpec, newSpec *datav1alpha1.CacheRuntimeSpec)) {
			oldSpec, newSpec := baseSpec(), baseSpec()
			prepare(&oldSpec, &newSpec)
			Expect(ValidateSpecUpdate(&oldSpec, &newSpec)).To(BeEmpty())
		},
		Entry("the same tieredStore quota in another unit", func(_, newSpec *datav1alpha1.CacheRuntimeSpec) {
			newSpec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("2048Mi")
		}),
		Entry("a nil list replaced by an empty list", func(oldSpec, newSpec *datav1alpha1.CacheRuntimeSpec) {
			oldSpec.Master.Env, newSpec.Master.Env = nil, []corev1.EnvVar{}
			oldSpec.Master.Args, newSpec.Master.Args = nil, []string{}
			oldSpec.Options, newSpec.Options = nil, map[string]string{}
		}),
	)

	DescribeTable("rejects a change to a field that is read only at creation",
		func(m mutate, wantPath string) {
			oldSpec, newSpec := baseSpec(), baseSpec()
			m(&newSpec)
			errs := ValidateSpecUpdate(&oldSpec, &newSpec)
			Expect(paths(errs)).To(Equal([]string{wantPath}))
			Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
		},
		Entry("runtimeClassName", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.RuntimeClassName = "other" }), "spec.runtimeClassName"),
		Entry("options", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Options["global"] = "b" }), "spec.options"),
		Entry("podMetadata", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.PodMetadata.Labels = map[string]string{"x": "y"} }), "spec.podMetadata"),
		Entry("imagePullSecrets", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.ImagePullSecrets = nil }), "spec.imagePullSecrets"),
		Entry("volumes", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Volumes = append(s.Volumes, corev1.Volume{Name: "data", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}})
		}), "spec.volumes"),

		Entry("master disabled", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.Disabled = true }), "spec.master.disabled"),
		Entry("master options", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.Options["rpc_port"] = "9000" }), "spec.master.options"),
		Entry("master podMetadata", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.PodMetadata.Labels["tier"] = "x" }), "spec.master.podMetadata"),
		Entry("master env", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.Env[0].Value = "debug" }), "spec.master.env"),
		Entry("master volumeMounts", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.VolumeMounts[0].MountPath = "/logs" }), "spec.master.volumeMounts"),
		Entry("master args", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.Args = append(s.Master.Args, "--verbose") }), "spec.master.args"),
		Entry("master nodeSelector", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.NodeSelector = nil }), "spec.master.nodeSelector"),
		Entry("master tolerations", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Master.Tolerations = nil }), "spec.master.tolerations"),

		Entry("worker tieredStore quota", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("4Gi")
		}), "spec.worker.tieredStore"),
		Entry("worker tieredStore medium", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Worker.TieredStore.Levels[0] = datav1alpha1.RuntimeTieredStoreLevel{
				EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("2Gi")},
			}
		}), "spec.worker.tieredStore"),
		Entry("worker tieredStore level added", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Worker.TieredStore.Levels = append(s.Worker.TieredStore.Levels, datav1alpha1.RuntimeTieredStoreLevel{
				EmptyDir: &datav1alpha1.EmptyDirMediumSource{Quota: resource.MustParse("10Gi")},
			})
		}), "spec.worker.tieredStore"),
		Entry("worker env", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Worker.Env = nil }), "spec.worker.env"),

		Entry("client disabled", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Client.Disabled = true }), "spec.client.disabled"),
		Entry("client runtimeVersion", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Client.RuntimeVersion.ImageTag = "v2" }), "spec.client.runtimeVersion"),
		Entry("client resources", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Client.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
		}), "spec.client.resources"),
		Entry("client options", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Client.Options["debug"] = "true" }), "spec.client.options"),
		Entry("client tieredStore", mutate(func(s *datav1alpha1.CacheRuntimeSpec) {
			s.Client.TieredStore.Levels[0].EmptyDir.Quota = resource.MustParse("2Gi")
		}), "spec.client.tieredStore"),
		Entry("client cleanPolicy", mutate(func(s *datav1alpha1.CacheRuntimeSpec) { s.Client.CleanPolicy = datav1alpha1.OnDemandCleanPolicy }), "spec.client.cleanPolicy"),
	)

	It("reports every changed immutable field, and none of the updatable ones", func() {
		oldSpec, newSpec := baseSpec(), baseSpec()
		newSpec.Worker.Replicas = 5
		newSpec.Worker.RuntimeVersion.ImageTag = "v2"
		newSpec.Worker.TieredStore.Levels[0].ProcessMemory.Quota = resource.MustParse("4Gi")
		newSpec.Client.Options["debug"] = "true"
		newSpec.Options = nil

		errs := ValidateSpecUpdate(&oldSpec, &newSpec)
		Expect(paths(errs)).To(ConsistOf("spec.worker.tieredStore", "spec.client.options", "spec.options"))
	})
})
