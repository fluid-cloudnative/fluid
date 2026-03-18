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

package dataload

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

var _ = Describe("IsTargetPathUnderFluidNativeMounts", func() {
	mockDataset := func(name, mountPoint, path string) datav1alpha1.Dataset {
		return datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						Name:       name,
						MountPoint: mountPoint,
						Path:       path,
					},
				},
			},
		}
	}

	DescribeTable("target path matching",
		func(targetPath string, dataset datav1alpha1.Dataset, want bool) {
			got := utils.IsTargetPathUnderFluidNativeMounts(targetPath, dataset)
			Expect(got).To(Equal(want))
		},
		Entry("no fluid native mount", "/imagenet",
			mockDataset("imagenet", "oss://imagenet-data/", ""), false),
		Entry("pvc mount", "/imagenet",
			mockDataset("imagenet", "pvc://nfs-imagenet", ""), true),
		Entry("hostpath mount", "/imagenet",
			mockDataset("imagenet", "local:///hostpath_imagenet", ""), true),
		Entry("target subpath", "/imagenet/data/train",
			mockDataset("imagenet", "pvc://nfs-imagenet", ""), true),
		Entry("mount path prefix", "/dataset/data/train",
			mockDataset("imagenet", "pvc://nfs-imagenet", "/dataset"), true),
		Entry("other path no match", "/dataset",
			mockDataset("imagenet", "pvc://nfs-imagenet", ""), false),
	)
})

var _ = Describe("dataLoadOperation", func() {
	var (
		testScheme   *runtime.Scheme
		mockDataload *datav1alpha1.DataLoad
		op           *dataLoadOperation
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())

		mockDataload = &datav1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload",
				Namespace: "default",
			},
			Spec: datav1alpha1.DataLoadSpec{
				Dataset: datav1alpha1.TargetDataset{
					Name:      "hadoop",
					Namespace: "default",
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, mockDataload)
		op = &dataLoadOperation{
			Client:   fakeClient,
			Log:      fake.NullLogger(),
			Recorder: record.NewFakeRecorder(10),
			dataLoad: mockDataload,
		}
	})

	Describe("GetOperationObject", func() {
		It("returns the DataLoad object", func() {
			obj := op.GetOperationObject()
			Expect(obj).To(Equal(mockDataload))
		})
	})

	Describe("HasPrecedingOperation", func() {
		It("returns false when RunAfter is nil", func() {
			mockDataload.Spec.RunAfter = nil
			Expect(op.HasPrecedingOperation()).To(BeFalse())
		})

		It("returns true when RunAfter is set", func() {
			mockDataload.Spec.RunAfter = &datav1alpha1.OperationRef{
				ObjectRef: datav1alpha1.ObjectRef{
					Kind: "DataLoad",
					Name: "prev-op",
				},
			}
			Expect(op.HasPrecedingOperation()).To(BeTrue())
		})
	})

	Describe("GetPossibleTargetDatasetNamespacedNames", func() {
		It("returns the dataset namespaced name", func() {
			names := op.GetPossibleTargetDatasetNamespacedNames()
			Expect(names).To(HaveLen(1))
			Expect(names[0]).To(Equal(types.NamespacedName{
				Namespace: "default",
				Name:      "hadoop",
			}))
		})
	})

	Describe("GetTargetDataset", func() {
		It("returns the dataset when it exists", func() {
			mockDataset := &datav1alpha1.Dataset{
				ObjectMeta: v1.ObjectMeta{
					Name:      "hadoop",
					Namespace: "default",
				},
			}
			testSchemeWithDataset := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testSchemeWithDataset)).To(Succeed())
			fakeClientWithDataset := fake.NewFakeClientWithScheme(testSchemeWithDataset, mockDataload, mockDataset)
			opWithDataset := &dataLoadOperation{
				Client:   fakeClientWithDataset,
				Log:      fake.NullLogger(),
				Recorder: record.NewFakeRecorder(10),
				dataLoad: mockDataload,
			}
			dataset, err := opWithDataset.GetTargetDataset()
			Expect(err).NotTo(HaveOccurred())
			Expect(dataset).NotTo(BeNil())
			Expect(dataset.Name).To(Equal("hadoop"))
		})

		It("returns error when dataset does not exist", func() {
			_, err := op.GetTargetDataset()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetReleaseNameSpacedName", func() {
		It("returns the release namespaced name", func() {
			nsn := op.GetReleaseNameSpacedName()
			Expect(nsn.Namespace).To(Equal("default"))
			Expect(nsn.Name).To(Equal(utils.GetDataLoadReleaseName("test-dataload")))
		})
	})

	Describe("GetChartsDirectory", func() {
		It("returns the charts directory for dataload", func() {
			dir := op.GetChartsDirectory()
			Expect(dir).To(ContainSubstring(cdataload.DataloadChart))
		})
	})

	Describe("GetOperationType", func() {
		It("returns DataLoadType", func() {
			Expect(op.GetOperationType()).To(Equal(dataoperation.DataLoadType))
		})
	})

	Describe("GetParallelTaskNumber", func() {
		It("returns 1", func() {
			Expect(op.GetParallelTaskNumber()).To(Equal(int32(1)))
		})
	})

	Describe("UpdateStatusInfoForCompleted", func() {
		It("returns nil (no-op)", func() {
			err := op.UpdateStatusInfoForCompleted(map[string]string{"key": "val"})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UpdateOperationApiStatus", func() {
		It("updates the DataLoad status via the client", func() {
			newStatus := &datav1alpha1.OperationStatus{
				Phase: common.PhaseComplete,
			}
			err := op.UpdateOperationApiStatus(newStatus)
			Expect(err).NotTo(HaveOccurred())

			persisted := &datav1alpha1.DataLoad{}
			err = op.Client.Get(context.Background(), types.NamespacedName{
				Namespace: mockDataload.Namespace,
				Name:      mockDataload.Name,
			}, persisted)
			Expect(err).NotTo(HaveOccurred())
			Expect(persisted.Status.Phase).To(Equal(newStatus.Phase))
		})
	})

	Describe("SetTargetDatasetStatusInProgress", func() {
		It("does not panic (no-op)", func() {
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.SetTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("RemoveTargetDatasetStatusInProgress", func() {
		It("does not panic (no-op)", func() {
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.RemoveTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("Validate", func() {
		It("returns nil when dataload and dataset namespaces match", func() {
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(BeNil())
		})

		It("returns error when dataload and dataset namespaces differ", func() {
			mockDataload.Namespace = "other-ns"
			mockDataload.Spec.Dataset.Namespace = "default"
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).To(HaveOccurred())
			Expect(conditions).To(HaveLen(1))
		})
	})

	Describe("GetStatusHandler", func() {
		It("returns OnceStatusHandler for Once policy", func() {
			mockDataload.Spec.Policy = datav1alpha1.Once
			handler := op.GetStatusHandler()
			Expect(handler).To(BeAssignableToTypeOf(&OnceStatusHandler{}))
		})

		It("returns CronStatusHandler for Cron policy", func() {
			mockDataload.Spec.Policy = datav1alpha1.Cron
			handler := op.GetStatusHandler()
			Expect(handler).To(BeAssignableToTypeOf(&CronStatusHandler{}))
		})

		It("returns OnEventStatusHandler for OnEvent policy", func() {
			mockDataload.Spec.Policy = datav1alpha1.OnEvent
			handler := op.GetStatusHandler()
			Expect(handler).To(BeAssignableToTypeOf(&OnEventStatusHandler{}))
		})

		It("returns nil for unknown policy", func() {
			mockDataload.Spec.Policy = "Unknown"
			handler := op.GetStatusHandler()
			Expect(handler).To(BeNil())
		})
	})

	Describe("GetTTL", func() {
		It("returns TTLSecondsAfterFinished for Once policy", func() {
			ttlVal := int32(300)
			mockDataload.Spec.Policy = datav1alpha1.Once
			mockDataload.Spec.TTLSecondsAfterFinished = &ttlVal
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(&ttlVal))
		})

		It("returns nil TTL for Cron policy", func() {
			mockDataload.Spec.Policy = datav1alpha1.Cron
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("returns nil TTL for OnEvent policy", func() {
			mockDataload.Spec.Policy = datav1alpha1.OnEvent
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("returns error for unknown policy", func() {
			mockDataload.Spec.Policy = "Unknown"
			_, err := op.GetTTL()
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("DataLoadReconciler", func() {
	Describe("NewDataLoadReconciler", func() {
		It("creates a reconciler with populated fields", func() {
			testScheme := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
			fakeClient := fake.NewFakeClientWithScheme(testScheme)
			recorder := record.NewFakeRecorder(10)
			r := NewDataLoadReconciler(fakeClient, fake.NullLogger(), testScheme, recorder)
			Expect(r).NotTo(BeNil())
			Expect(r.Scheme).To(Equal(testScheme))
			Expect(r.OperationReconciler).NotTo(BeNil())
		})
	})

	Describe("Build", func() {
		var r *DataLoadReconciler

		BeforeEach(func() {
			testScheme := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
			fakeClient := fake.NewFakeClientWithScheme(testScheme)
			recorder := record.NewFakeRecorder(10)
			r = &DataLoadReconciler{}
			r.OperationReconciler = controllers.NewDataOperationReconciler(r, fakeClient, fake.NullLogger(), recorder)
		})

		It("returns error when object is not a DataLoad", func() {
			_, err := r.Build(&datav1alpha1.Dataset{})
			Expect(err).To(HaveOccurred())
		})

		It("returns dataLoadOperation when object is a DataLoad", func() {
			dataLoad := &datav1alpha1.DataLoad{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			}
			result, err := r.Build(dataLoad)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})
	})

	Describe("ControllerName", func() {
		It("returns the expected controller name", func() {
			r := &DataLoadReconciler{}
			Expect(r.ControllerName()).To(Equal("DataLoadReconciler"))
		})
	})
})
