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

package juicefs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("JuiceFSEngine imagePullSecrets", func() {
	secrets := func(names ...string) []corev1.LocalObjectReference {
		refs := make([]corev1.LocalObjectReference, 0, len(names))
		for _, name := range names {
			refs = append(refs, corev1.LocalObjectReference{Name: name})
		}
		return refs
	}

	newDataset := func() *datav1alpha1.Dataset {
		return &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "juicefs:///mnt/test",
					Name:       "test",
					EncryptOptions: []datav1alpha1.EncryptOption{{
						Name: "metaurl",
						ValueFrom: datav1alpha1.EncryptOptionSource{
							SecretKeyRef: datav1alpha1.SecretKeySelector{Name: "test", Key: "metaurl"},
						},
					}},
				}},
			},
		}
	}

	newEngine := func(jfsRuntime *datav1alpha1.JuiceFSRuntime) *JuiceFSEngine {
		juicefsSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "fluid"},
			Data:       map[string][]byte{"metaurl": []byte("test")},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, []runtime.Object{juicefsSecret}...)
		runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "juicefs")
		Expect(err).NotTo(HaveOccurred())

		return &JuiceFSEngine{
			name:        "test",
			namespace:   "fluid",
			Client:      fakeClient,
			Log:         fake.NullLogger(),
			runtime:     jfsRuntime,
			runtimeInfo: runtimeInfo,
		}
	}

	Describe("transformWorkers", func() {
		It("uses the controller default when the runtime sets no secrets", func() {
			GinkgoT().Setenv(common.EnvImagePullSecretsKey, "default-secret")
			jfsRuntime := &datav1alpha1.JuiceFSRuntime{}
			value := &JuiceFS{Edition: CommunityEdition}

			Expect(newEngine(jfsRuntime).transformWorkers(jfsRuntime, newDataset(), value)).To(Succeed())

			Expect(value.ImagePullSecrets).To(Equal(secrets("default-secret")))
			Expect(value.Worker.ImagePullSecrets).To(BeEmpty())
		})

		It("prefers the runtime and worker secrets", func() {
			GinkgoT().Setenv(common.EnvImagePullSecretsKey, "default-secret")
			jfsRuntime := &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					ImagePullSecrets: secrets("runtime-secret"),
					Worker:           datav1alpha1.JuiceFSCompTemplateSpec{ImagePullSecrets: secrets("worker-secret")},
				},
			}
			value := &JuiceFS{Edition: CommunityEdition}

			Expect(newEngine(jfsRuntime).transformWorkers(jfsRuntime, newDataset(), value)).To(Succeed())

			Expect(value.ImagePullSecrets).To(Equal(secrets("runtime-secret")))
			Expect(value.Worker.ImagePullSecrets).To(Equal(secrets("worker-secret")))
		})
	})

	Describe("transformFuse", func() {
		It("sets the fuse secrets only when the fuse spec has them", func() {
			jfsRuntime := &datav1alpha1.JuiceFSRuntime{
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{ImagePullSecrets: secrets("fuse-secret")},
				},
			}
			value := &JuiceFS{}

			Expect(newEngine(jfsRuntime).transformFuse(jfsRuntime, newDataset(), value)).To(Succeed())
			Expect(value.Fuse.ImagePullSecrets).To(Equal(secrets("fuse-secret")))

			jfsRuntime.Spec.Fuse.ImagePullSecrets = nil
			value = &JuiceFS{}
			Expect(newEngine(jfsRuntime).transformFuse(jfsRuntime, newDataset(), value)).To(Succeed())
			Expect(value.Fuse.ImagePullSecrets).To(BeEmpty())
		})
	})
})
