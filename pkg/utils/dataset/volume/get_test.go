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

package volume

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Get helpers related tests", Label("pkg.utils.dataset.volume.get_test.go"), func() {
	var (
		scheme    *runtime.Scheme
		clientObj client.Client
		resources []runtime.Object
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = v1.AddToScheme(scheme)
		_ = datav1alpha1.AddToScheme(scheme)
		resources = nil
	})

	JustBeforeEach(func() {
		clientObj = fake.NewFakeClientWithScheme(scheme, resources...)
	})

	Context("Test GetPVCByVolumeId()", func() {
		When("pv is bound to a fluid pvc", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "ns-name"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns", Name: "name"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "ns", Labels: map[string]string{common.LabelAnnotationStorageCapacityPrefix + "ns-name": ""}}},
				)
			})
			It("should return the pvc", func() {
				got, err := GetPVCByVolumeId(clientObj, "ns-name")
				Expect(err).To(BeNil())
				Expect(got).NotTo(BeNil())
				Expect(got.Name).To(Equal("name"))
			})
		})

		When("pv is bound to a non-fluid pvc", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "x"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns", Name: "n"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "n", Namespace: "ns"}},
				)
			})
			It("should return error", func() {
				_, err := GetPVCByVolumeId(clientObj, "x")
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Context("Test GetNamespacedNameByVolumeId()", func() {
		When("pv has claimRef and pvc is managed by fluid", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "ns2-n"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns2", Name: "n"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "n", Namespace: "ns2", Labels: map[string]string{common.LabelAnnotationStorageCapacityPrefix + "ns2-n": "true"}}},
				)
			})
			It("should return namespace and name", func() {
				ns, name, err := GetNamespacedNameByVolumeId(clientObj, "ns2-n")
				Expect(err).To(BeNil())
				Expect(ns).To(Equal("ns2"))
				Expect(name).To(Equal("n"))
			})
		})

		When("pv has nil claimRef", func() {
			BeforeEach(func() {
				resources = append(resources, &v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "v"}})
			})
			It("should return error", func() {
				_, _, err := GetNamespacedNameByVolumeId(clientObj, "v")
				Expect(err).NotTo(BeNil())
			})
		})
	})
})
