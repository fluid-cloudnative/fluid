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
	// EFCRuntime stays in parity with the other built-in runtime helper checks below.
	DescribeTable("Replicas returns the runtime spec replicas",
		func(runtimeObj interface{ Replicas() int32 }, expected int32) {
			Expect(runtimeObj.Replicas()).To(Equal(expected))
		},
		Entry("AlluxioRuntime", &AlluxioRuntime{Spec: AlluxioRuntimeSpec{Replicas: 1}}, int32(1)),
		Entry("EFCRuntime", &EFCRuntime{Spec: EFCRuntimeSpec{Replicas: 2}}, int32(2)),
		Entry("GooseFSRuntime", &GooseFSRuntime{Spec: GooseFSRuntimeSpec{Replicas: 2}}, int32(2)),
		Entry("JindoRuntime", &JindoRuntime{Spec: JindoRuntimeSpec{Replicas: 3}}, int32(3)),
		Entry("JuiceFSRuntime", &JuiceFSRuntime{Spec: JuiceFSRuntimeSpec{Replicas: 4}}, int32(4)),
		Entry("ThinRuntime", &ThinRuntime{Spec: ThinRuntimeSpec{Replicas: 5}}, int32(5)),
		Entry("VineyardRuntime", &VineyardRuntime{Spec: VineyardRuntimeSpec{Replicas: 6}}, int32(6)),
	)

	Describe("GetStatus", func() {
		It("returns the Alluxio runtime status pointer", func() {
			runtimeObj := &AlluxioRuntime{}
			runtimeObj.Status.WorkerPhase = RuntimePhaseReady

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().WorkerPhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the EFCRuntime status pointer", func() {
			runtimeObj := &EFCRuntime{}
			runtimeObj.Status.WorkerPhase = RuntimePhaseReady

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().WorkerPhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the GooseFS runtime status pointer", func() {
			runtimeObj := &GooseFSRuntime{}
			runtimeObj.Status.MasterPhase = RuntimePhaseReady

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().MasterPhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the Jindo runtime status pointer", func() {
			runtimeObj := &JindoRuntime{}
			runtimeObj.Status.FusePhase = RuntimePhaseReady

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().FusePhase).To(Equal(RuntimePhaseReady))
		})

		It("returns the JuiceFS runtime status pointer", func() {
			runtimeObj := &JuiceFSRuntime{}
			runtimeObj.Status.Selector = "app=juicefs"

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().Selector).To(Equal("app=juicefs"))
		})

		It("returns the Thin runtime status pointer", func() {
			runtimeObj := &ThinRuntime{}
			runtimeObj.Status.SetupDuration = "1m"

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().SetupDuration).To(Equal("1m"))
		})

		It("returns the Vineyard runtime status pointer", func() {
			runtimeObj := &VineyardRuntime{}
			runtimeObj.Status.ValueFileConfigmap = "values"

			Expect(runtimeObj.GetStatus()).To(BeIdenticalTo(&runtimeObj.Status))
			Expect(runtimeObj.GetStatus().ValueFileConfigmap).To(Equal("values"))
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
		Entry("EFCRuntime", &EFCRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: EFCRuntimeKind}}, &EFCRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "EFCRuntimeList"}}, EFCRuntimeKind, "EFCRuntimeList"),
		Entry("GooseFSRuntime", &GooseFSRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "GooseFSRuntime"}}, &GooseFSRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "GooseFSRuntimeList"}}, "GooseFSRuntime", "GooseFSRuntimeList"),
		Entry("JindoRuntime", &JindoRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: JindoRuntimeKind}}, &JindoRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "JindoRuntimeList"}}, JindoRuntimeKind, "JindoRuntimeList"),
		Entry("JuiceFSRuntime", &JuiceFSRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: JuiceFSRuntimeKind}}, &JuiceFSRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "JuiceFSRuntimeList"}}, JuiceFSRuntimeKind, "JuiceFSRuntimeList"),
		Entry("ThinRuntime", &ThinRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: ThinRuntimeKind}}, &ThinRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "ThinRuntimeList"}}, ThinRuntimeKind, "ThinRuntimeList"),
		Entry("VineyardRuntime", &VineyardRuntime{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: VineyardRuntimeKind}}, &VineyardRuntimeList{TypeMeta: metav1.TypeMeta{APIVersion: GroupVersion.String(), Kind: "VineyardRuntimeList"}}, VineyardRuntimeKind, "VineyardRuntimeList"),
	)
})
