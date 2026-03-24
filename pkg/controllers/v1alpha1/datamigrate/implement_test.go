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

package datamigrate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("dataMigrateOperation", func() {
	Describe("Validate", func() {
		It("should error when SSH secret is not set for parallel migrate", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Parallelism:     2,
						ParallelOptions: map[string]string{},
					},
				},
			}
			ctx := runtime.ReconcileRequestContext{
				Dataset: nil,
			}

			got, err := op.Validate(ctx)

			Expect(err).To(HaveOccurred())
			Expect(got).NotTo(BeEmpty())
			Expect(got[0].Reason).To(Equal(common.TargetSSHSecretNameNotSet))
		})

		It("should error when datamigrate namespace differs from dataset namespace", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Parallelism: 1,
					},
				},
			}
			op.dataMigrate.Namespace = "ns-a"

			dataset := &datav1alpha1.Dataset{}
			dataset.Namespace = "ns-b"

			ctx := runtime.ReconcileRequestContext{
				Dataset: dataset,
			}

			got, err := op.Validate(ctx)

			Expect(err).To(HaveOccurred())
			Expect(got).NotTo(BeEmpty())
			Expect(got[0].Reason).To(Equal(common.TargetDatasetNamespaceNotSame))
		})

		It("should return nil when validation passes", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Parallelism: 1,
					},
				},
			}
			op.dataMigrate.Namespace = "default"

			dataset := &datav1alpha1.Dataset{}
			dataset.Namespace = "default"

			ctx := runtime.ReconcileRequestContext{
				Dataset: dataset,
			}

			got, err := op.Validate(ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})
	})

	Describe("GetOperationObject", func() {
		It("should return the dataMigrate object", func() {
			dm := &datav1alpha1.DataMigrate{}
			dm.Name = "test-migrate"
			op := &dataMigrateOperation{dataMigrate: dm}

			Expect(op.GetOperationObject()).To(Equal(dm))
		})
	})

	Describe("HasPrecedingOperation", func() {
		It("should return false when RunAfter is nil", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{},
				},
			}
			Expect(op.HasPrecedingOperation()).To(BeFalse())
		})

		It("should return true when RunAfter is set", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						RunAfter: &datav1alpha1.OperationRef{},
					},
				},
			}
			Expect(op.HasPrecedingOperation()).To(BeTrue())
		})
	})

	Describe("GetPossibleTargetDatasetNamespacedNames", func() {
		It("should return empty when no datasets are set", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{},
				},
			}

			result := op.GetPossibleTargetDatasetNamespacedNames()

			Expect(result).To(BeEmpty())
		})

		It("should return To dataset when set", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						To: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name:      "dest-dataset",
								Namespace: "ns-dest",
							},
						},
					},
				},
			}

			result := op.GetPossibleTargetDatasetNamespacedNames()

			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(types.NamespacedName{Namespace: "ns-dest", Name: "dest-dataset"}))
		})

		It("should fall back to dataMigrate namespace when dataset namespace is empty", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						From: datav1alpha1.DataToMigrate{
							DataSet: &datav1alpha1.DatasetToMigrate{
								Name: "src-dataset",
							},
						},
					},
				},
			}
			op.dataMigrate.Namespace = "default"

			result := op.GetPossibleTargetDatasetNamespacedNames()

			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(types.NamespacedName{Namespace: "default", Name: "src-dataset"}))
		})
	})

	Describe("GetReleaseNameSpacedName", func() {
		It("should return correct namespaced name for release", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{},
			}
			op.dataMigrate.Name = "my-migrate"
			op.dataMigrate.Namespace = "my-ns"

			result := op.GetReleaseNameSpacedName()

			expected := utils.GetDataMigrateReleaseName("my-migrate")
			Expect(result.Namespace).To(Equal("my-ns"))
			Expect(result.Name).To(Equal(expected))
		})
	})

	Describe("GetChartsDirectory", func() {
		It("should return charts directory containing datamigrate chart name", func() {
			op := &dataMigrateOperation{dataMigrate: &datav1alpha1.DataMigrate{}}
			result := op.GetChartsDirectory()
			Expect(result).To(ContainSubstring(cdatamigrate.DataMigrateChart))
		})
	})

	Describe("GetOperationType", func() {
		It("should return DataMigrateType", func() {
			op := &dataMigrateOperation{dataMigrate: &datav1alpha1.DataMigrate{}}
			Expect(op.GetOperationType()).To(Equal(dataoperation.DataMigrateType))
		})
	})

	Describe("UpdateStatusInfoForCompleted", func() {
		It("should return nil (no-op)", func() {
			op := &dataMigrateOperation{dataMigrate: &datav1alpha1.DataMigrate{}}
			Expect(op.UpdateStatusInfoForCompleted(nil)).To(Succeed())
		})
	})

	Describe("SetTargetDatasetStatusInProgress", func() {
		It("should set dataset phase to DataMigrating", func() {
			op := &dataMigrateOperation{dataMigrate: &datav1alpha1.DataMigrate{}}
			dataset := &datav1alpha1.Dataset{}

			op.SetTargetDatasetStatusInProgress(dataset)

			Expect(dataset.Status.Phase).To(Equal(datav1alpha1.DataMigrating))
		})
	})

	Describe("RemoveTargetDatasetStatusInProgress", func() {
		It("should set dataset phase to BoundDatasetPhase", func() {
			op := &dataMigrateOperation{dataMigrate: &datav1alpha1.DataMigrate{}}
			dataset := &datav1alpha1.Dataset{}

			op.RemoveTargetDatasetStatusInProgress(dataset)

			Expect(dataset.Status.Phase).To(Equal(datav1alpha1.BoundDatasetPhase))
		})
	})

	Describe("GetStatusHandler", func() {
		It("should return OnceStatusHandler for Once policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: datav1alpha1.Once,
					},
				},
			}
			handler := op.GetStatusHandler()
			Expect(handler).To(BeAssignableToTypeOf(&OnceStatusHandler{}))
		})

		It("should return CronStatusHandler for Cron policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: datav1alpha1.Cron,
					},
				},
			}
			handler := op.GetStatusHandler()
			Expect(handler).To(BeAssignableToTypeOf(&CronStatusHandler{}))
		})

		It("should return OnEventStatusHandler for OnEvent policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: datav1alpha1.OnEvent,
					},
				},
			}
			handler := op.GetStatusHandler()
			Expect(handler).To(BeAssignableToTypeOf(&OnEventStatusHandler{}))
		})

		It("should return nil for unknown policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: "Unknown",
					},
				},
			}
			handler := op.GetStatusHandler()
			Expect(handler).To(BeNil())
		})
	})

	Describe("GetTTL", func() {
		It("should return TTL for Once policy", func() {
			ttlVal := int32(300)
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy:                  datav1alpha1.Once,
						TTLSecondsAfterFinished: &ttlVal,
					},
				},
			}
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).NotTo(BeNil())
			Expect(*ttl).To(Equal(ttlVal))
		})

		It("should return nil TTL for Cron policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: datav1alpha1.Cron,
					},
				},
			}
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("should return nil TTL for OnEvent policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: datav1alpha1.OnEvent,
					},
				},
			}
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("should return error for unknown policy", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Policy: "UnknownPolicy",
					},
				},
			}
			ttl, err := op.GetTTL()
			Expect(err).To(HaveOccurred())
			Expect(ttl).To(BeNil())
		})
	})

	Describe("GetParallelTaskNumber", func() {
		It("should return Parallelism from spec", func() {
			op := &dataMigrateOperation{
				dataMigrate: &datav1alpha1.DataMigrate{
					Spec: datav1alpha1.DataMigrateSpec{
						Parallelism: 5,
					},
				},
			}
			Expect(op.GetParallelTaskNumber()).To(Equal(int32(5)))
		})
	})

	Describe("GetTargetDataset", func() {
		It("should return the dataset when it exists and is referenced by datamigrate", func() {
			testScheme := k8sruntime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())

			dataset := &datav1alpha1.Dataset{}
			dataset.Name = "my-dataset"
			dataset.Namespace = "default"

			dm := &datav1alpha1.DataMigrate{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate",
					Namespace: "default",
				},
				Spec: datav1alpha1.DataMigrateSpec{
					To: datav1alpha1.DataToMigrate{
						DataSet: &datav1alpha1.DatasetToMigrate{
							Name:      "my-dataset",
							Namespace: "default",
						},
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, dataset, dm)
			op := &dataMigrateOperation{
				Client:      fakeClient,
				dataMigrate: dm,
			}

			result, err := op.GetTargetDataset()

			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("my-dataset"))
		})
	})

	Describe("UpdateOperationApiStatus", func() {
		It("should update status without error", func() {
			testScheme := k8sruntime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())

			dm := &datav1alpha1.DataMigrate{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate",
					Namespace: "default",
				},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, dm)
			op := &dataMigrateOperation{
				Client:      fakeClient,
				dataMigrate: dm,
			}

			newStatus := &datav1alpha1.OperationStatus{
				Phase: common.PhaseComplete,
			}
			err := op.UpdateOperationApiStatus(newStatus)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
