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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("runtime API helpers", func() {
	DescribeTable("Replicas returns the runtime spec replicas",
		func(runtime interface{ Replicas() int32 }, expected int32) {
			Expect(runtime.Replicas()).To(Equal(expected))
		},
		Entry("AlluxioRuntime", &AlluxioRuntime{Spec: AlluxioRuntimeSpec{Replicas: 1}}, int32(1)),
		Entry("GooseFSRuntime", &GooseFSRuntime{Spec: GooseFSRuntimeSpec{Replicas: 2}}, int32(2)),
		Entry("JindoRuntime", &JindoRuntime{Spec: JindoRuntimeSpec{Replicas: 3}}, int32(3)),
		Entry("JuiceFSRuntime", &JuiceFSRuntime{Spec: JuiceFSRuntimeSpec{Replicas: 4}}, int32(4)),
		Entry("ThinRuntime", &ThinRuntime{Spec: ThinRuntimeSpec{Replicas: 5}}, int32(5)),
		Entry("VineyardRuntime", &VineyardRuntime{Spec: VineyardRuntimeSpec{Replicas: 6}}, int32(6)),
	)

	Describe("GetStatus", func() {
		It("returns the Alluxio runtime status pointer", func() {
			runtime := &AlluxioRuntime{}
			runtime.Status.WorkerPhase = RuntimePhaseReady

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().WorkerPhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the GooseFS runtime status pointer", func() {
			runtime := &GooseFSRuntime{}
			runtime.Status.MasterPhase = RuntimePhaseReady

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().MasterPhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the Jindo runtime status pointer", func() {
			runtime := &JindoRuntime{}
			runtime.Status.FusePhase = RuntimePhaseReady

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().FusePhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the JuiceFS runtime status pointer", func() {
			runtime := &JuiceFSRuntime{}
			runtime.Status.Selector = "app=juicefs"

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().Selector).To(Equal("app=juicefs"))
		})

		It("returns the Thin runtime status pointer", func() {
			runtime := &ThinRuntime{}
			runtime.Status.SetupDuration = "1m"

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().SetupDuration).To(Equal("1m"))
		})

		It("returns the Vineyard runtime status pointer", func() {
			runtime := &VineyardRuntime{}
			runtime.Status.ValueFileConfigmap = "values"

			Expect(runtime.GetStatus()).To(BeIdenticalTo(&runtime.Status))
			Expect(runtime.GetStatus().ValueFileConfigmap).To(Equal("values"))
		})
	})

	DescribeTable("runtime objects are registered in the scheme with the expected GVKs",
		func(obj runtime.Object, list runtime.Object, expectedKind string, expectedListKind string) {
			gvk := schema.GroupVersionKind{Group: GroupVersion.Group, Version: GroupVersion.Version, Kind: expectedKind}
			listGVK := schema.GroupVersionKind{Group: GroupVersion.Group, Version: GroupVersion.Version, Kind: expectedListKind}

			kinds, _, err := UnitTestScheme.ObjectKinds(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(kinds).To(ContainElement(gvk))

			listKinds, _, err := UnitTestScheme.ObjectKinds(list)
			Expect(err).NotTo(HaveOccurred())
			Expect(listKinds).To(ContainElement(listGVK))
		},
		Entry("AlluxioRuntime", &AlluxioRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "AlluxioRuntime"}}, &AlluxioRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "AlluxioRuntimeList"}}, "AlluxioRuntime", "AlluxioRuntimeList"),
		Entry("GooseFSRuntime", &GooseFSRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "GooseFSRuntime"}}, &GooseFSRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "GooseFSRuntimeList"}}, "GooseFSRuntime", "GooseFSRuntimeList"),
		Entry("JindoRuntime", &JindoRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "JindoRuntime"}}, &JindoRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "JindoRuntimeList"}}, "JindoRuntime", "JindoRuntimeList"),
		Entry("JuiceFSRuntime", &JuiceFSRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "JuiceFSRuntime"}}, &JuiceFSRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "JuiceFSRuntimeList"}}, "JuiceFSRuntime", "JuiceFSRuntimeList"),
		Entry("ThinRuntime", &ThinRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "ThinRuntime"}}, &ThinRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "ThinRuntimeList"}}, "ThinRuntime", "ThinRuntimeList"),
		Entry("VineyardRuntime", &VineyardRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "VineyardRuntime"}}, &VineyardRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "VineyardRuntimeList"}}, "VineyardRuntime", "VineyardRuntimeList"),
	)
})
