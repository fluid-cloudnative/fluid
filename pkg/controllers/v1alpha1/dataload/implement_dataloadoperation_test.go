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

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func newTestDataLoadOperation(dataLoad *datav1alpha1.DataLoad, objs ...runtime.Object) *dataLoadOperation {
	testScheme := runtime.NewScheme()
	if err := datav1alpha1.AddToScheme(testScheme); err != nil {
		panic(err)
	}
	allObjs := append([]runtime.Object{dataLoad}, objs...)
	c := fake.NewFakeClientWithScheme(testScheme, allObjs...)
	return &dataLoadOperation{
		Client:   c,
		Log:      fake.NullLogger(),
		Recorder: record.NewFakeRecorder(32),
		dataLoad: dataLoad,
	}
}

var _ = Describe("dataLoadOperation", func() {
	var mockDataLoad *datav1alpha1.DataLoad

	BeforeEach(func() {
		mockDataLoad = &datav1alpha1.DataLoad{
			ObjectMeta: metav1.ObjectMeta{
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
	})

	Describe("GetOperationObject", func() {
		It("returns the dataLoad object", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetOperationObject()).To(Equal(mockDataLoad))
		})
	})

	Describe("HasPrecedingOperation", func() {
		It("returns false when RunAfter is nil", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.HasPrecedingOperation()).To(BeFalse())
		})

		It("returns true when RunAfter is set", func() {
			mockDataLoad.Spec.RunAfter = &datav1alpha1.OperationRef{}
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.HasPrecedingOperation()).To(BeTrue())
		})
	})

	Describe("GetPossibleTargetDatasetNamespacedNames", func() {
		It("returns the dataset namespaced name", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			names := op.GetPossibleTargetDatasetNamespacedNames()
			Expect(names).To(HaveLen(1))
			Expect(names[0]).To(Equal(types.NamespacedName{Namespace: "default", Name: "hadoop"}))
		})
	})

	Describe("GetReleaseNameSpacedName", func() {
		It("returns the release namespaced name", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			nn := op.GetReleaseNameSpacedName()
			Expect(nn.Namespace).To(Equal("default"))
			Expect(nn.Name).NotTo(BeEmpty())
		})
	})

	Describe("GetChartsDirectory", func() {
		It("returns a non-empty charts directory", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetChartsDirectory()).NotTo(BeEmpty())
		})
	})

	Describe("GetOperationType", func() {
		It("returns DataLoadType", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetOperationType()).To(Equal(dataoperation.DataLoadType))
		})
	})

	Describe("UpdateStatusInfoForCompleted", func() {
		It("returns nil (no-op)", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.UpdateStatusInfoForCompleted(nil)).To(Succeed())
		})
	})

	Describe("SetTargetDatasetStatusInProgress and RemoveTargetDatasetStatusInProgress", func() {
		It("are no-ops and do not panic", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			dataset := &datav1alpha1.Dataset{}
			Expect(func() { op.SetTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
			Expect(func() { op.RemoveTargetDatasetStatusInProgress(dataset) }).NotTo(Panic())
		})
	})

	Describe("GetTargetDataset", func() {
		It("returns dataset when it exists", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hadoop", Namespace: "default"},
			}
			op := newTestDataLoadOperation(mockDataLoad, dataset)
			ds, err := op.GetTargetDataset()
			Expect(err).NotTo(HaveOccurred())
			Expect(ds.Name).To(Equal("hadoop"))
		})
	})

	Describe("UpdateOperationApiStatus", func() {
		It("succeeds when dataload exists in fake client", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			opStatus := mockDataLoad.Status.DeepCopy()
			// fake client supports Status().Update when the object exists
			err := op.UpdateOperationApiStatus(opStatus)
			// expect no error — fake client handles status update
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("GetParallelTaskNumber", func() {
		It("returns 1", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetParallelTaskNumber()).To(Equal(int32(1)))
		})
	})

	Describe("GetStatusHandler", func() {
		It("returns OnceStatusHandler for Once policy", func() {
			mockDataLoad.Spec.Policy = datav1alpha1.Once
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetStatusHandler()).To(BeAssignableToTypeOf(&OnceStatusHandler{}))
		})

		It("returns CronStatusHandler for Cron policy", func() {
			mockDataLoad.Spec.Policy = datav1alpha1.Cron
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetStatusHandler()).To(BeAssignableToTypeOf(&CronStatusHandler{}))
		})

		It("returns OnEventStatusHandler for OnEvent policy", func() {
			mockDataLoad.Spec.Policy = datav1alpha1.OnEvent
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetStatusHandler()).To(BeAssignableToTypeOf(&OnEventStatusHandler{}))
		})

		It("returns nil for unknown policy", func() {
			mockDataLoad.Spec.Policy = "Unknown"
			op := newTestDataLoadOperation(mockDataLoad)
			Expect(op.GetStatusHandler()).To(BeNil())
		})
	})

	Describe("GetTTL", func() {
		It("returns nil TTL for Once policy with no TTL set", func() {
			mockDataLoad.Spec.Policy = datav1alpha1.Once
			op := newTestDataLoadOperation(mockDataLoad)
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("returns TTL for Once policy with TTL set", func() {
			ttlVal := int32(300)
			mockDataLoad.Spec.Policy = datav1alpha1.Once
			mockDataLoad.Spec.TTLSecondsAfterFinished = &ttlVal
			op := newTestDataLoadOperation(mockDataLoad)
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).NotTo(BeNil())
			Expect(*ttl).To(Equal(int32(300)))
		})

		It("returns nil TTL for Cron policy", func() {
			mockDataLoad.Spec.Policy = datav1alpha1.Cron
			op := newTestDataLoadOperation(mockDataLoad)
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("returns nil TTL for OnEvent policy", func() {
			mockDataLoad.Spec.Policy = datav1alpha1.OnEvent
			op := newTestDataLoadOperation(mockDataLoad)
			ttl, err := op.GetTTL()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNil())
		})

		It("returns error for unknown policy", func() {
			mockDataLoad.Spec.Policy = "Unknown"
			op := newTestDataLoadOperation(mockDataLoad)
			_, err := op.GetTTL()
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Validate", func() {
		It("returns nil when namespace matches dataset namespace", func() {
			op := newTestDataLoadOperation(mockDataLoad)
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(conditions).To(BeNil())
		})

		It("returns error when namespace does not match dataset namespace", func() {
			mockDataLoad.Namespace = "other"
			mockDataLoad.Spec.Dataset.Namespace = "default"
			op := newTestDataLoadOperation(mockDataLoad)
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			conditions, err := op.Validate(ctx)
			Expect(err).To(HaveOccurred())
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0].Type).To(Equal(common.Failed))
		})
	})
})

var _ = Describe("DataLoadReconciler", func() {
	var patches *gomonkey.Patches

	BeforeEach(func() {
		patches = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
			return true
		})
	})

	AfterEach(func() {
		patches.Reset()
	})

	Describe("Build", func() {
		It("returns dataLoadOperation for a DataLoad object", func() {
			testScheme := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
			c := fake.NewFakeClientWithScheme(testScheme)
			r := NewDataLoadReconciler(c, fake.NullLogger(), testScheme, record.NewFakeRecorder(32))

			dl := &datav1alpha1.DataLoad{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"},
			}
			op, err := r.Build(dl)
			Expect(err).NotTo(HaveOccurred())
			Expect(op).NotTo(BeNil())
		})

		It("returns error for a non-DataLoad object", func() {
			testScheme := runtime.NewScheme()
			c := fake.NewFakeClientWithScheme(testScheme)
			r := NewDataLoadReconciler(c, fake.NullLogger(), testScheme, record.NewFakeRecorder(32))

			pod := &corev1.Pod{}
			_, err := r.Build(pod)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ControllerName", func() {
		It("returns DataLoadReconciler", func() {
			testScheme := runtime.NewScheme()
			c := fake.NewFakeClientWithScheme(testScheme)
			r := NewDataLoadReconciler(c, fake.NullLogger(), testScheme, record.NewFakeRecorder(32))
			Expect(r.ControllerName()).To(Equal("DataLoadReconciler"))
		})
	})

	Describe("Reconcile", func() {
		It("returns no-requeue when DataLoad not found", func() {
			testScheme := runtime.NewScheme()
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
			c := fake.NewFakeClientWithScheme(testScheme)
			r := NewDataLoadReconciler(c, fake.NullLogger(), testScheme, record.NewFakeRecorder(32))

			result, err := r.Reconcile(context.Background(), ctrl.Request{
				NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
		})
	})
})
