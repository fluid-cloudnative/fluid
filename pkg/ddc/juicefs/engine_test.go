/*
Copyright 2021 The Fluid Authors.

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
	base "github.com/fluid-cloudnative/fluid/pkg/ddc/base"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("JuiceFS engine", func() {
	const (
		runtimeName      = "hbase"
		runtimeNamespace = "fluid"
	)

	var (
		key     types.NamespacedName
		runtime *datav1alpha1.JuiceFSRuntime
	)

	BeforeEach(func() {
		key = types.NamespacedName{
			Name:      runtimeName,
			Namespace: runtimeNamespace,
		}
		runtime = &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
		}
	})

	Describe("Build", func() {
		When("the runtime exists and the context runtime is valid", func() {
			It("builds a template engine", func() {
				ctx := cruntime.ReconcileRequestContext{
					NamespacedName: key,
					Client:         fake.NewFakeClientWithScheme(testScheme, runtime.DeepCopy()),
					Log:            fake.NullLogger(),
					RuntimeType:    "juicefs",
					Runtime:        runtime.DeepCopy(),
				}

				engine, err := Build("test-id", ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(engine).NotTo(BeNil())
				Expect(engine).To(BeAssignableToTypeOf(&base.TemplateEngine{}))
			})
		})

		When("the context runtime is nil", func() {
			It("returns a parse error", func() {
				ctx := cruntime.ReconcileRequestContext{
					NamespacedName: key,
					Client:         fake.NewFakeClientWithScheme(testScheme, runtime.DeepCopy()),
					Log:            fake.NullLogger(),
					RuntimeType:    "juicefs",
				}

				engine, err := Build("test-id", ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse"))
				Expect(engine).To(BeNil())
			})
		})

		When("the context runtime has the wrong concrete type", func() {
			It("returns a parse error", func() {
				ctx := cruntime.ReconcileRequestContext{
					NamespacedName: key,
					Client:         fake.NewFakeClientWithScheme(testScheme, runtime.DeepCopy()),
					Log:            fake.NullLogger(),
					RuntimeType:    "juicefs",
					Runtime: &datav1alpha1.AlluxioRuntime{
						ObjectMeta: metav1.ObjectMeta{
							Name:      key.Name,
							Namespace: key.Namespace,
						},
					},
				}

				engine, err := Build("test-id", ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse"))
				Expect(engine).To(BeNil())
			})
		})

		When("runtime info cannot be loaded from the client", func() {
			It("returns a runtime info error", func() {
				ctx := cruntime.ReconcileRequestContext{
					NamespacedName: key,
					Client:         fake.NewFakeClientWithScheme(testScheme),
					Log:            fake.NullLogger(),
					RuntimeType:    "juicefs",
					Runtime:        runtime.DeepCopy(),
				}

				engine, err := Build("test-id", ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get runtime info"))
				Expect(engine).To(BeNil())
			})
		})
	})

	Describe("Precheck", func() {
		When("the runtime exists", func() {
			It("returns found", func() {
				client := fake.NewFakeClientWithScheme(testScheme, runtime.DeepCopy())

				found, err := Precheck(client, key)

				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		When("the runtime does not exist", func() {
			It("returns not found without error", func() {
				client := fake.NewFakeClientWithScheme(testScheme)

				found, err := Precheck(client, key)

				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})
