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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ThinEngine transform", Label("pkg.ddc.thin.transform_ginkgo_test.go"), func() {
	It("returns an error when the runtime is nil", func() {
		engine := &ThinEngine{Log: fake.NullLogger()}

		value, err := engine.transform(nil, nil)

		Expect(err).To(MatchError("the thinRuntime is null"))
		Expect(value).To(BeNil())
	})

	It("merges runtime and profile state while enabling worker orchestration", func() {
		dataset, runtime, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
		engine := mockThinEngineForTests(dataset, runtime, profile)
		engine.Client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, dataset, runtime, profile)
		engine.runtimeInfo.SetOwnerDatasetUID(dataset.UID)

		runtime.Spec.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "runtime-secret"}}
		runtime.Spec.Worker.Enabled = true
		runtime.Spec.Worker.Image = "runtime-worker-image"
		runtime.Spec.Worker.ImageTag = "runtime-worker-tag"
		runtime.Spec.Worker.ImagePullPolicy = string(corev1.PullIfNotPresent)
		runtime.Spec.Worker.ImagePullSecrets = []corev1.LocalObjectReference{{Name: "runtime-worker-secret"}}
		runtime.Spec.Worker.Env = []corev1.EnvVar{{Name: "RUNTIME_ENV", Value: "runtime-value"}}
		runtime.Spec.Worker.Ports = []corev1.ContainerPort{{Name: "rpc", ContainerPort: 19998}}
		runtime.Spec.Worker.NodeSelector = map[string]string{"disk": "ssd"}
		runtime.Spec.Worker.NetworkMode = datav1alpha1.ContainerNetworkMode
		runtime.Spec.TieredStore.Levels = []datav1alpha1.Level{{Path: "/runtime/cache"}}
		runtime.Spec.Fuse.Image = "runtime-fuse-image"

		profile.Spec.Worker = datav1alpha1.ThinCompTemplateSpec{
			Image:            "profile-worker-image",
			ImageTag:         "profile-worker-tag",
			ImagePullPolicy:  string(corev1.PullAlways),
			ImagePullSecrets: []corev1.LocalObjectReference{{Name: "profile-worker-secret"}},
			Env:              []corev1.EnvVar{{Name: "PROFILE_ENV", Value: "profile-value"}},
			Ports:            []corev1.ContainerPort{{Name: "metrics", ContainerPort: 8080}},
			NodeSelector:     map[string]string{"from": "profile"},
			NetworkMode:      datav1alpha1.HostNetworkMode,
		}

		value, err := engine.transform(runtime, profile)

		Expect(err).NotTo(HaveOccurred())
		Expect(value.ImagePullSecrets).To(Equal(runtime.Spec.ImagePullSecrets))
		Expect(value.Worker.Image).To(Equal(runtime.Spec.Worker.Image))
		Expect(value.Worker.ImageTag).To(Equal(runtime.Spec.Worker.ImageTag))
		Expect(value.Worker.ImagePullPolicy).To(Equal(runtime.Spec.Worker.ImagePullPolicy))
		Expect(value.Worker.ImagePullSecrets).To(Equal(runtime.Spec.Worker.ImagePullSecrets))
		Expect(value.Worker.Envs).To(ContainElement(runtime.Spec.Worker.Env[0]))
		Expect(value.Worker.Ports).To(ContainElement(runtime.Spec.Worker.Ports[0]))
		Expect(value.Worker.NodeSelector).To(Equal(runtime.Spec.Worker.NodeSelector))
		Expect(value.Worker.HostNetwork).To(BeFalse())
		Expect(value.Worker.CacheDir).To(Equal("/runtime/cache"))
		Expect(value.RuntimeIdentity.Namespace).To(Equal(runtime.Namespace))
		Expect(value.RuntimeIdentity.Name).To(Equal(runtime.Name))
		Expect(value.OwnerDatasetId).To(Equal(utils.GetDatasetId(engine.namespace, engine.name, string(dataset.UID))))
		Expect(value.Owner).NotTo(BeNil())
		Expect(value.Owner.Name).To(Equal(runtime.Name))
		Expect(value.Fuse.Image).To(Equal(runtime.Spec.Fuse.Image))
		Expect(value.Fuse.NodeSelector).To(HaveKeyWithValue(utils.GetFuseLabelName(runtime.Namespace, runtime.Name, string(dataset.UID)), "true"))
	})

	It("defaults placement mode to exclusive when the dataset does not set one", func() {
		dataset, runtime, profile := mockFluidObjectsForTests(types.NamespacedName{Name: "test-dataset", Namespace: "default"})
		dataset.Spec.PlacementMode = ""
		engine := mockThinEngineForTests(dataset, runtime, profile)
		engine.Client = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, dataset, runtime, profile)
		engine.runtimeInfo.SetOwnerDatasetUID(dataset.UID)
		runtime.Spec.Fuse.Image = "runtime-fuse-image"

		value, err := engine.transform(runtime, profile)

		Expect(err).NotTo(HaveOccurred())
		Expect(value.PlacementMode).To(Equal(string(datav1alpha1.ExclusiveMode)))
		Expect(value.Tolerations).To(BeEmpty())
		Expect(value.ImagePullSecrets).To(Equal(profile.Spec.ImagePullSecrets))
		Expect(value.Fuse.Enabled).To(BeTrue())
		Expect(value.OwnerDatasetId).To(Equal(utils.GetDatasetId(engine.namespace, engine.name, string(dataset.UID))))
		Expect(value.FullnameOverride).To(Equal(engine.name))
		Expect(value.Owner.Name).To(Equal(runtime.Name))
		Expect(value.Owner.UID).To(Equal(string(runtime.UID)))
	})
})
