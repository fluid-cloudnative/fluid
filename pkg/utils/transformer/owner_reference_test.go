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
var _ = Describe("GenerateOwnerReferenceFromObject", func() {
	DescribeTable("when generating owner reference from dataset",
		func(dataset *datav1alpha1.Dataset, expected *common.OwnerReference) {
			testScheme := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
			testScheme.AddKnownTypes(datav1alpha1.GroupVersion, dataset)
			testObjs := []runtime.Object{dataset.DeepCopy()}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, testObjs...)
			obj := &datav1alpha1.Dataset{}

			err := fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: dataset.Namespace,
				Name:      dataset.Name,
			}, obj)

			Expect(err).NotTo(HaveOccurred())
			result := GenerateOwnerReferenceFromObject(obj)
			Expect(result).To(Equal(expected))
		},

		Entry("should handle standard dataset with all fields",
			&datav1alpha1.Dataset{
=======
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
>>>>>>> f14ce29b (test(transformer): Migrate transformer to ginko/gomega and Improve coverage)
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
<<<<<<< HEAD
					Name:      "test-dataset",
					Namespace: "fluid",
					UID:       "12345",
				},
			},
			&common.OwnerReference{
=======
					Name:      name,
					Namespace: namespace,
					UID:       "12345",
				},
			}
			expect = &common.OwnerReference{
>>>>>>> f14ce29b (test(transformer): Migrate transformer to ginko/gomega and Improve coverage)
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "12345",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1alpha1",
<<<<<<< HEAD
				Name:               "test-dataset",
			},
		),

		Entry("should handle dataset with different name and namespace",
			&datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-dataset",
					Namespace: "production",
					UID:       "abcdef",
				},
			},
			&common.OwnerReference{
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "abcdef",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1alpha1",
				Name:               "my-dataset",
			},
		),

		Entry("should handle dataset with empty UID",
			&datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-uid-dataset",
					Namespace: "default",
					UID:       "",
				},
			},
			&common.OwnerReference{
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1alpha1",
				Name:               "empty-uid-dataset",
			},
		),

		Entry("should handle dataset with different API version format",
			&datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1beta1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "beta-dataset",
					Namespace: "test",
					UID:       "xyz123",
				},
			},
			&common.OwnerReference{
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "xyz123",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1beta1",
				Name:               "beta-dataset",
			},
		),

		Entry("should handle dataset with long name",
			&datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "very-long-dataset-name-with-many-characters-for-testing-purposes",
					Namespace: "default",
					UID:       "long-uid-12345",
				},
			},
			&common.OwnerReference{
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "long-uid-12345",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1alpha1",
				Name:               "very-long-dataset-name-with-many-characters-for-testing-purposes",
			},
		),

		Entry("should handle dataset with special characters in name",
			&datav1alpha1.Dataset{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Dataset",
					APIVersion: "data.fluid.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset-test-123",
					Namespace: "ns-test-456",
					UID:       "uid-789-abc",
				},
			},
			&common.OwnerReference{
				Enabled:            true,
				Controller:         true,
				BlockOwnerDeletion: false,
				UID:                "uid-789-abc",
				Kind:               "Dataset",
				APIVersion:         "data.fluid.io/v1alpha1",
				Name:               "dataset-test-123",
			},
		),
	)
})

var _ = Describe("FilterOwnerByKind", func() {
	DescribeTable("when filtering owner references",
		func(ownerReferences []metav1.OwnerReference, ownerKind string, expected []metav1.OwnerReference) {
			result := FilterOwnerByKind(ownerReferences, ownerKind)
			Expect(result).To(Equal(expected))
		},

		Entry("should filter single matching kind",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
			},
		),

		Entry("should filter multiple matching kinds",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-1", UID: "11111"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-2", UID: "22222"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-3", UID: "33333"},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-1", UID: "11111"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-2", UID: "22222"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-3", UID: "33333"},
			},
		),

		Entry("should filter mixed kinds - only return matching",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
				{APIVersion: "v1", Kind: "ConfigMap", Name: "test-config", UID: "abcde"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "another-dataset", UID: "fghij"},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "another-dataset", UID: "fghij"},
			},
		),

		Entry("should return empty slice when no matching kinds",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
				{APIVersion: "v1", Kind: "ConfigMap", Name: "test-config", UID: "abcde"},
			},
			"Dataset",
			[]metav1.OwnerReference{},
		),

		Entry("should return empty slice when owner references list is empty",
			[]metav1.OwnerReference{},
			"Dataset",
			[]metav1.OwnerReference{},
		),

		Entry("should return empty slice when owner references list is nil",
			nil,
			"Dataset",
			[]metav1.OwnerReference{},
		),

		Entry("should filter by different kind",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
			},
			"Deployment",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
			},
		),

		Entry("should handle case sensitive kind matching",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
			},
			"dataset",
			[]metav1.OwnerReference{},
		),

		Entry("should filter with owner references having all fields populated",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345", Controller: boolPtr(true), BlockOwnerDeletion: boolPtr(true)},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890", Controller: boolPtr(false), BlockOwnerDeletion: boolPtr(false)},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345", Controller: boolPtr(true), BlockOwnerDeletion: boolPtr(true)},
			},
		),

		Entry("should filter StatefulSet kind",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
				{APIVersion: "apps/v1", Kind: "StatefulSet", Name: "test-statefulset", UID: "sts123"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
			},
			"StatefulSet",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "StatefulSet", Name: "test-statefulset", UID: "sts123"},
			},
		),

		Entry("should filter DaemonSet kind",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "DaemonSet", Name: "test-daemonset", UID: "ds123"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
			},
			"DaemonSet",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "DaemonSet", Name: "test-daemonset", UID: "ds123"},
			},
		),

		Entry("should handle multiple kinds with same name but different types",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "same-name", UID: "12345"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "same-name", UID: "67890"},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "same-name", UID: "12345"},
			},
		),

		Entry("should preserve order when filtering",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-3", UID: "33333"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "deployment-1", UID: "11111"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-1", UID: "22222"},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-3", UID: "33333"},
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "dataset-1", UID: "22222"},
			},
		),

		Entry("should handle empty kind string",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "test-dataset", UID: "12345"},
			},
			"",
			[]metav1.OwnerReference{},
		),

		Entry("should filter ReplicaSet kind",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "test-rs", UID: "rs123"},
				{APIVersion: "apps/v1", Kind: "Deployment", Name: "test-deployment", UID: "67890"},
			},
			"ReplicaSet",
			[]metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "test-rs", UID: "rs123"},
			},
		),

		Entry("should handle owner references with only required fields",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "minimal-dataset", UID: "min123"},
			},
			"Dataset",
			[]metav1.OwnerReference{
				{APIVersion: "data.fluid.io/v1alpha1", Kind: "Dataset", Name: "minimal-dataset", UID: "min123"},
			},
		),
	)
=======
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
>>>>>>> f14ce29b (test(transformer): Migrate transformer to ginko/gomega and Improve coverage)
})

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
