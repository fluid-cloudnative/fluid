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
limitations under the License
*/

package utils

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("GetAlluxioRuntime", func() {
	var (
		s              *runtime.Scheme
		runtimeName    = "alluxio-runtime-1"
		runtimeNs      = "default"
		alluxioRuntime *datav1alpha1.AlluxioRuntime
	)

	BeforeEach(func() {
		alluxioRuntime = &datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: runtimeNs,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, alluxioRuntime)
	})

	DescribeTable("should handle runtime lookup",
		func(name, namespace, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, alluxioRuntime)
			gotRuntime, err := GetAlluxioRuntime(fakeClient, name, namespace)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(gotRuntime).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotRuntime.Name).To(Equal(wantName))
			}
		},
		Entry("existing runtime", "alluxio-runtime-1", "default", "alluxio-runtime-1", false),
		Entry("non-existent name", "alluxio-runtime-1not-exist", "default", "", true),
		Entry("non-existent namespace", "alluxio-runtime-1", "defaultnot-exist", "", true),
	)
})

var _ = Describe("GetJuiceFSRuntime", func() {
	var (
		s              *runtime.Scheme
		runtimeName    = "juicefs-runtime-1"
		runtimeNs      = "default"
		juicefsRuntime *datav1alpha1.JuiceFSRuntime
	)

	BeforeEach(func() {
		juicefsRuntime = &datav1alpha1.JuiceFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: runtimeNs,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, juicefsRuntime)
	})

	DescribeTable("should handle runtime lookup",
		func(name, namespace, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, juicefsRuntime)
			gotRuntime, err := GetJuiceFSRuntime(fakeClient, name, namespace)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(gotRuntime).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotRuntime.Name).To(Equal(wantName))
			}
		},
		Entry("existing runtime", "juicefs-runtime-1", "default", "juicefs-runtime-1", false),
		Entry("non-existent name", "juicefs-runtime-1not-exist", "default", "", true),
		Entry("non-existent namespace", "juicefs-runtime-1", "defaultnot-exist", "", true),
	)
})

var _ = Describe("GetJindoRuntime", func() {
	var (
		s            *runtime.Scheme
		runtimeName  = "jindo-runtime-1"
		runtimeNs    = "default"
		jindoRuntime *datav1alpha1.JindoRuntime
	)

	BeforeEach(func() {
		jindoRuntime = &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: runtimeNs,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, jindoRuntime)
	})

	DescribeTable("should handle runtime lookup",
		func(name, namespace, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, jindoRuntime)
			gotRuntime, err := GetJindoRuntime(fakeClient, name, namespace)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(gotRuntime).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotRuntime.Name).To(Equal(wantName))
			}
		},
		Entry("existing runtime", "jindo-runtime-1", "default", "jindo-runtime-1", false),
		Entry("non-existent name", "jindo-runtime-1not-exist", "default", "", true),
		Entry("non-existent namespace", "jindo-runtime-1", "defaultnot-exist", "", true),
	)
})

var _ = Describe("GetGooseFSRuntime", func() {
	var (
		s              *runtime.Scheme
		runtimeName    = "goosefs-runtime-1"
		runtimeNs      = "default"
		goosefsRuntime *datav1alpha1.GooseFSRuntime
	)

	BeforeEach(func() {
		goosefsRuntime = &datav1alpha1.GooseFSRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: runtimeNs,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, goosefsRuntime)
	})

	DescribeTable("should handle runtime lookup",
		func(name, namespace, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, goosefsRuntime)
			gotRuntime, err := GetGooseFSRuntime(fakeClient, name, namespace)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(gotRuntime).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotRuntime.Name).To(Equal(wantName))
			}
		},
		Entry("existing runtime", "goosefs-runtime-1", "default", "goosefs-runtime-1", false),
		Entry("non-existent name", "goosefs-runtime-1not-exist", "default", "", true),
		Entry("non-existent namespace", "goosefs-runtime-1", "defaultnot-exist", "", true),
	)
})

var _ = Describe("GetThinRuntime", func() {
	var (
		s           *runtime.Scheme
		runtimeName = "thin-runtime-1"
		runtimeNs   = "default"
		thinRuntime *datav1alpha1.ThinRuntime
	)

	BeforeEach(func() {
		thinRuntime = &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: runtimeNs,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, thinRuntime)
	})

	DescribeTable("should handle runtime lookup",
		func(name, namespace, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, thinRuntime)
			gotRuntime, err := GetThinRuntime(fakeClient, name, namespace)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(gotRuntime).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotRuntime.Name).To(Equal(wantName))
			}
		},
		Entry("existing runtime", "thin-runtime-1", "default", "thin-runtime-1", false),
		Entry("non-existent name", "thin-runtime-1not-exist", "default", "", true),
		Entry("non-existent namespace", "thin-runtime-1", "defaultnot-exist", "", true),
	)
})

var _ = Describe("AddRuntimesIfNotExist", func() {
	var (
		runtime1 datav1alpha1.Runtime
		runtime2 datav1alpha1.Runtime
		runtime3 datav1alpha1.Runtime
	)

	BeforeEach(func() {
		runtime1 = datav1alpha1.Runtime{
			Name:     "imagenet",
			Category: common.AccelerateCategory,
		}
		runtime2 = datav1alpha1.Runtime{
			Name:     "mock-name",
			Category: "mock-category",
		}
		runtime3 = datav1alpha1.Runtime{
			Name:     "cifar10",
			Category: common.AccelerateCategory,
		}
	})

	runtimeSliceEqual := func(a, b []datav1alpha1.Runtime) bool {
		if len(a) != len(b) || (a == nil) != (b == nil) {
			return false
		}
		for i, s := range a {
			if s != b[i] {
				return false
			}
		}
		return true
	}

	It("should add runtime to an empty slice successfully", func() {
		result := AddRuntimesIfNotExist([]datav1alpha1.Runtime{}, runtime1)
		Expect(runtimeSliceEqual(result, []datav1alpha1.Runtime{runtime1})).To(BeTrue())
	})

	It("should not add duplicate runtime", func() {
		result := AddRuntimesIfNotExist([]datav1alpha1.Runtime{runtime1}, runtime1)
		Expect(runtimeSliceEqual(result, []datav1alpha1.Runtime{runtime1})).To(BeTrue())
	})

	It("should add runtime of different name and category successfully", func() {
		result := AddRuntimesIfNotExist([]datav1alpha1.Runtime{runtime1}, runtime2)
		Expect(runtimeSliceEqual(result, []datav1alpha1.Runtime{runtime1, runtime2})).To(BeTrue())
	})

	It("should not add runtime of the same category but different name", func() {
		result := AddRuntimesIfNotExist([]datav1alpha1.Runtime{runtime1}, runtime3)
		Expect(runtimeSliceEqual(result, []datav1alpha1.Runtime{runtime1})).To(BeTrue())
	})
})

var _ = Describe("GetThinRuntimeProfile", func() {
	var (
		s                  *runtime.Scheme
		profileName        = "test-profile"
		thinRuntimeProfile *datav1alpha1.ThinRuntimeProfile
	)

	BeforeEach(func() {
		thinRuntimeProfile = &datav1alpha1.ThinRuntimeProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name: profileName,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, thinRuntimeProfile)
	})

	DescribeTable("should handle profile lookup",
		func(name, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, thinRuntimeProfile)
			got, err := GetThinRuntimeProfile(fakeClient, name)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(got.Name).To(Equal(wantName))
			}
		},
		Entry("existing profile", "test-profile", "test-profile", false),
		Entry("non-existent profile", "test-profilenot-exist", "", true),
	)
})

var _ = Describe("GetEFCRuntime", func() {
	var (
		s           *runtime.Scheme
		runtimeName = "efc-runtime-1"
		runtimeNs   = "default"
		efcRuntime  *datav1alpha1.EFCRuntime
	)

	BeforeEach(func() {
		efcRuntime = &datav1alpha1.EFCRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      runtimeName,
				Namespace: runtimeNs,
			},
		}
		s = runtime.NewScheme()
		s.AddKnownTypes(datav1alpha1.GroupVersion, efcRuntime)
	})

	DescribeTable("should handle runtime lookup",
		func(name, namespace, wantName string, notFound bool) {
			fakeClient := fake.NewFakeClientWithScheme(s, efcRuntime)
			gotRuntime, err := GetEFCRuntime(fakeClient, name, namespace)

			if notFound {
				Expect(err).To(HaveOccurred())
				Expect(gotRuntime).To(BeNil())
				Expect(apierrs.IsNotFound(err)).To(BeTrue())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(gotRuntime.Name).To(Equal(wantName))
			}
		},
		Entry("existing runtime", "efc-runtime-1", "default", "efc-runtime-1", false),
		Entry("non-existent name", "efc-runtime-1not-exist", "default", "", true),
		Entry("non-existent namespace", "efc-runtime-1", "defaultnot-exist", "", true),
	)
})
