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

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("CacheRuntime API helpers", func() {
	Describe("GetStatus", func() {
		It("returns the cache runtime status pointer", func() {
			runtime := &CacheRuntime{}
			runtime.Status.Selector = "app=cacheruntime"

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().Selector).To(Equal("app=cacheruntime"))
		})
	})

	Describe("scheme registration", func() {
		It("registers CacheRuntime and CacheRuntimeList with the api group version", func() {
			runtime := &CacheRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: CacheRuntimeKind}}
			runtimeList := &CacheRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "CacheRuntimeList"}}

			runtimeKinds, _, err := UnitTestScheme.ObjectKinds(runtime)
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeKinds).To(ContainElement(schema.GroupVersionKind{Group: GroupVersion.Group, Version: GroupVersion.Version, Kind: CacheRuntimeKind}))

			listKinds, _, err := UnitTestScheme.ObjectKinds(runtimeList)
			Expect(err).NotTo(HaveOccurred())
			Expect(listKinds).To(ContainElement(schema.GroupVersionKind{Group: GroupVersion.Group, Version: GroupVersion.Version, Kind: "CacheRuntimeList"}))
		})
	})
})
