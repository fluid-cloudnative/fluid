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

package transformer

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Transformer", func() {
	Describe("GenerateOwnerReferenceFromObject", func() {
		var (
			name      string
			namespace string
			dataset   *datav1alpha1.Dataset
			expect    *common.OwnerReference
		)

		BeforeEach(func() {
			name = "test-dataset"
			namespace = "fluid"
			dataset = &datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					UID:       "12345",
				},
			}
			expect = &common.OwnerReference{
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "12345",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1alpha1",
				Name:               name,
			}
		})

		It("should generate correct owner reference from dataset object", func() {
			testScheme := runtime.NewScheme()
			err := datav1alpha1.AddToScheme(testScheme)
			Expect(err).NotTo(HaveOccurred())

			testScheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			testObjs := []runtime.Object{dataset.DeepCopy()}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			obj := &datav1alpha1.Dataset{}

			err = fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: namespace,
				Name:      name,
			}, obj)
			Expect(err).NotTo(HaveOccurred())

			result := GenerateOwnerReferenceFromObject(obj)
			Expect(result).To(Equal(expect))
		})

		It("should handle objects with different API groups correctly", func() {
			dataset := &datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "CustomResource",
					APIVersion: "custom.example.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "custom-resource",
					Namespace: "custom-ns",
					UID:       "custom-uid-123",
				},
			}

			testScheme := runtime.NewScheme()
			err := datav1alpha1.AddToScheme(testScheme)
			Expect(err).NotTo(HaveOccurred())

			testScheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			testObjs := []runtime.Object{dataset.DeepCopy()}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			obj := &datav1alpha1.Dataset{}

			err = fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "custom-ns",
				Name:      "custom-resource",
			}, obj)
			Expect(err).NotTo(HaveOccurred())

			result := GenerateOwnerReferenceFromObject(obj)

			Expect(result).NotTo(BeNil())
			Expect(result.Kind).To(Equal("CustomResource"))
			Expect(result.APIVersion).To(Equal("custom.example.io/v1beta1"))
			Expect(result.UID).To(Equal("custom-uid-123"))
			Expect(result.Name).To(Equal("custom-resource"))
			Expect(result.Enabled).To(BeTrue())
			Expect(result.Controller).To(BeTrue())
			Expect(result.BlockOwnerDeletion).To(BeFalse())
		})
	})

	Describe("FilterOwnerByKind", func() {
		Context("when filtering owner references by kind", func() {
			It("should return only owners matching the specified kind", func() {
				ownerReferences := []metav1.OwnerReference{
					{
						Kind:       "Dataset",
						Name:       "dataset-1",
						UID:        "uid-1",
						APIVersion: "data.fluid.io/v1alpha1",
					},
					{
						Kind:       "Pod",
						Name:       "pod-1",
						UID:        "uid-2",
						APIVersion: "v1",
					},
					{
						Kind:       "Dataset",
						Name:       "dataset-2",
						UID:        "uid-3",
						APIVersion: "data.fluid.io/v1alpha1",
					},
					{
						Kind:       "Deployment",
						Name:       "deployment-1",
						UID:        "uid-4",
						APIVersion: "apps/v1",
					},
				}

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).To(HaveLen(2))
				Expect(result[0].Kind).To(Equal("Dataset"))
				Expect(result[0].Name).To(Equal("dataset-1"))
				Expect(result[1].Kind).To(Equal("Dataset"))
				Expect(result[1].Name).To(Equal("dataset-2"))
			})

			It("should return empty slice when no owners match", func() {
				ownerReferences := []metav1.OwnerReference{
					{
						Kind:       "Pod",
						Name:       "pod-1",
						UID:        "uid-1",
						APIVersion: "v1",
					},
					{
						Kind:       "Deployment",
						Name:       "deployment-1",
						UID:        "uid-2",
						APIVersion: "apps/v1",
					},
				}

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).To(BeEmpty())
			})

			It("should handle empty owner references slice", func() {
				ownerReferences := []metav1.OwnerReference{}

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).To(BeEmpty())
			})

			It("should handle nil owner references slice", func() {
				var ownerReferences []metav1.OwnerReference

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).NotTo(BeNil())
				Expect(result).To(BeEmpty())
			})

			It("should filter correctly with single matching owner", func() {
				ownerReferences := []metav1.OwnerReference{
					{
						Kind:       "Dataset",
						Name:       "dataset-1",
						UID:        "uid-1",
						APIVersion: "data.fluid.io/v1alpha1",
					},
				}

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).To(HaveLen(1))
				Expect(result[0].Kind).To(Equal("Dataset"))
				Expect(result[0].Name).To(Equal("dataset-1"))
			})

			It("should be case-sensitive when filtering by kind", func() {
				ownerReferences := []metav1.OwnerReference{
					{
						Kind:       "Dataset",
						Name:       "dataset-1",
						UID:        "uid-1",
						APIVersion: "data.fluid.io/v1alpha1",
					},
					{
						Kind:       "dataset",
						Name:       "dataset-2",
						UID:        "uid-2",
						APIVersion: "data.fluid.io/v1alpha1",
					},
				}

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).To(HaveLen(1))
				Expect(result[0].Kind).To(Equal("Dataset"))
				Expect(result[0].Name).To(Equal("dataset-1"))
			})

			It("should preserve all owner reference fields when filtering", func() {
				ownerReferences := []metav1.OwnerReference{
					{
						Kind:               "Dataset",
						Name:               "dataset-1",
						UID:                "uid-1",
						APIVersion:         "data.fluid.io/v1alpha1",
						Controller:         boolPtr(true),
						BlockOwnerDeletion: boolPtr(true),
					},
				}

				result := FilterOwnerByKind(ownerReferences, "Dataset")

				Expect(result).To(HaveLen(1))
				Expect(result[0].Kind).To(Equal("Dataset"))
				Expect(result[0].Name).To(Equal("dataset-1"))
				Expect(result[0].UID).To(Equal(types.UID("uid-1")))
				Expect(result[0].APIVersion).To(Equal("data.fluid.io/v1alpha1"))
				Expect(*result[0].Controller).To(BeTrue())
				Expect(*result[0].BlockOwnerDeletion).To(BeTrue())
			})
		})
	})
})

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
