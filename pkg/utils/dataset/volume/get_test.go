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

		When("pv does not exist", func() {
			It("should return error", func() {
				_, _, err := GetNamespacedNameByVolumeId(clientObj, "non-existent-pv")
				Expect(err).NotTo(BeNil())
			})
		})

		When("pv has nil claimRef", func() {
			BeforeEach(func() {
				resources = append(resources, &v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "v"}})
			})
			It("should return error", func() {
				_, _, err := GetNamespacedNameByVolumeId(clientObj, "v")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("has unexpected nil claimRef"))
			})
		})

		When("pvc does not exist", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv-orphan"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns", Name: "missing-pvc"}}},
				)
			})
			It("should return error", func() {
				_, _, err := GetNamespacedNameByVolumeId(clientObj, "pv-orphan")
				Expect(err).NotTo(BeNil())
			})
		})

		When("pvc is not a fluid dataset", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv-regular"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns", Name: "regular-pvc"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "regular-pvc", Namespace: "ns"}},
				)
			})
			It("should return error", func() {
				_, _, err := GetNamespacedNameByVolumeId(clientObj, "pv-regular")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("is not bounded with a fluid pvc"))
			})
		})
	})

	Context("Test GetVolumePairByVolumeId()", func() {
		When("pv is bound to a fluid pvc", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "ns-name"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns", Name: "name"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "ns", Labels: map[string]string{common.LabelAnnotationStorageCapacityPrefix + "ns-name": ""}}},
				)
			})
			It("should return the pvc and pv", func() {
				pvc, pv, err := GetVolumePairByVolumeId(clientObj, "ns-name")
				Expect(err).To(BeNil())
				Expect(pvc).NotTo(BeNil())
				Expect(pv).NotTo(BeNil())
				Expect(pvc.Name).To(Equal("name"))
				Expect(pvc.Namespace).To(Equal("ns"))
				Expect(pv.Name).To(Equal("ns-name"))
			})
		})

		When("pv does not exist", func() {
			It("should return error and nil pointers", func() {
				pvc, pv, err := GetVolumePairByVolumeId(clientObj, "non-existent")
				Expect(err).NotTo(BeNil())
				Expect(pvc).To(BeNil())
				Expect(pv).To(BeNil())
			})
		})

		When("pv has nil claimRef", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv-no-claim"}},
				)
			})
			It("should return error and nil pointers", func() {
				pvc, pv, err := GetVolumePairByVolumeId(clientObj, "pv-no-claim")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("has unexpected nil claimRef"))
				Expect(pvc).To(BeNil())
				Expect(pv).To(BeNil())
			})
		})

		When("pvc does not exist", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "pv-with-ref"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "test-ns", Name: "ghost-pvc"}}},
				)
			})
			It("should return error and nil pointers", func() {
				pvc, pv, err := GetVolumePairByVolumeId(clientObj, "pv-with-ref")
				Expect(err).NotTo(BeNil())
				Expect(pvc).To(BeNil())
				Expect(pv).To(BeNil())
			})
		})

		When("pv is bound to a non-fluid pvc", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "x"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "ns", Name: "n"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "n", Namespace: "ns"}},
				)
			})
			It("should return error and nil pointers", func() {
				pvc, pv, err := GetVolumePairByVolumeId(clientObj, "x")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("is not bounded with a fluid pvc"))
				Expect(pvc).To(BeNil())
				Expect(pv).To(BeNil())
			})
		})

		When("pvc has fluid labels with different format", func() {
			BeforeEach(func() {
				resources = append(resources,
					&v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: "test-pv"}, Spec: v1.PersistentVolumeSpec{ClaimRef: &v1.ObjectReference{Namespace: "default", Name: "test-pvc"}}},
					&v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "test-pvc", Namespace: "default", Labels: map[string]string{common.LabelAnnotationStorageCapacityPrefix + "test-pv": "100Gi"}}},
				)
			})
			It("should successfully return pvc and pv", func() {
				pvc, pv, err := GetVolumePairByVolumeId(clientObj, "test-pv")
				Expect(err).To(BeNil())
				Expect(pvc).NotTo(BeNil())
				Expect(pv).NotTo(BeNil())
				Expect(pvc.Name).To(Equal("test-pvc"))
				Expect(pv.Name).To(Equal("test-pv"))
			})
		})
	})
})
