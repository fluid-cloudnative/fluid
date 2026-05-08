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

package dataprocess

import (
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("OnceStatusHandler", func() {
	var (
		testScheme      *runtime.Scheme
		mockDataProcess *datav1alpha1.DataProcess
		ctx             cruntime.ReconcileRequestContext
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		mockDataProcess = &datav1alpha1.DataProcess{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: datav1alpha1.DataProcessSpec{},
		}

		ctx = cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "test",
			},
			Log: fake.NullLogger(),
		}
	})

	Describe("GetOperationStatus", func() {
		Context("when job completes successfully", func() {
			It("should return PhaseComplete", func() {
				job := &batchv1.Job{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-processor-job",
						Namespace: "default",
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobComplete,
								Status:             corev1.ConditionTrue,
								LastProbeTime:      v1.NewTime(time.Now()),
								LastTransitionTime: v1.NewTime(time.Now()),
							},
						},
					},
				}

				fakeClient := fake.NewFakeClientWithScheme(testScheme, mockDataProcess, job)
				handler := &OnceStatusHandler{
					Client:      fakeClient,
					dataProcess: mockDataProcess,
				}

				opStatus := &datav1alpha1.OperationStatus{}
				result, err := handler.GetOperationStatus(ctx, opStatus)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Phase).To(Equal(common.PhaseComplete))
			})
		})

		Context("when job fails", func() {
			It("should return PhaseFailed", func() {
				job := &batchv1.Job{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-processor-job",
						Namespace: "default",
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobFailed,
								Status:             corev1.ConditionTrue,
								LastProbeTime:      v1.NewTime(time.Now()),
								LastTransitionTime: v1.NewTime(time.Now()),
							},
						},
					},
				}

				fakeClient := fake.NewFakeClientWithScheme(testScheme, mockDataProcess, job)
				handler := &OnceStatusHandler{
					Client:      fakeClient,
					dataProcess: mockDataProcess,
				}

				opStatus := &datav1alpha1.OperationStatus{}
				result, err := handler.GetOperationStatus(ctx, opStatus)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Phase).To(Equal(common.PhaseFailed))
			})
		})

		Context("when job is still running (no finished condition)", func() {
			It("should return current opStatus phase unchanged", func() {
				job := &batchv1.Job{
					ObjectMeta: v1.ObjectMeta{
						Name:      "test-processor-job",
						Namespace: "default",
					},
					Status: batchv1.JobStatus{
						// No conditions — job is still running
						Conditions: []batchv1.JobCondition{},
					},
				}

				fakeClient := fake.NewFakeClientWithScheme(testScheme, mockDataProcess, job)
				handler := &OnceStatusHandler{
					Client:      fakeClient,
					dataProcess: mockDataProcess,
				}

				opStatus := &datav1alpha1.OperationStatus{
					Phase: common.PhaseExecuting,
				}
				result, err := handler.GetOperationStatus(ctx, opStatus)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				// Phase should remain as it was before — handler returns early
				Expect(result.Phase).To(Equal(common.PhaseExecuting))
			})
		})

		Context("when the job is not found", func() {
			It("should delete the helm release and return the unchanged status copy without error", func() {
				// Patch helm.DeleteReleaseIfExists to avoid exec'ing ddc-helm in a unit test.
				patch := gomonkey.ApplyFunc(helm.DeleteReleaseIfExists,
					func(_ string, _ string) error { return nil })
				defer patch.Reset()

				fakeClient := fake.NewFakeClientWithScheme(testScheme, mockDataProcess)
				handler := &OnceStatusHandler{
					Client:      fakeClient,
					dataProcess: mockDataProcess,
				}

				opStatus := &datav1alpha1.OperationStatus{Phase: common.PhaseExecuting}
				result, err := handler.GetOperationStatus(ctx, opStatus)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result).NotTo(BeIdenticalTo(opStatus))
				Expect(result.Phase).To(Equal(common.PhaseExecuting))
			})
		})

		Context("when job conditions are out of order", func() {
			It("should use the finished condition instead of the first condition", func() {
				createdAt := time.Now().Add(-2 * time.Minute)
				suspendedAt := createdAt.Add(30 * time.Second)
				failedAt := createdAt.Add(90 * time.Second)

				job := &batchv1.Job{
					ObjectMeta: v1.ObjectMeta{
						Name:              "test-processor-job",
						Namespace:         "default",
						CreationTimestamp: v1.NewTime(createdAt),
					},
					Status: batchv1.JobStatus{
						Conditions: []batchv1.JobCondition{
							{
								Type:               batchv1.JobSuspended,
								Status:             corev1.ConditionTrue,
								LastProbeTime:      v1.NewTime(suspendedAt),
								LastTransitionTime: v1.NewTime(suspendedAt),
							},
							{
								Type:               batchv1.JobFailed,
								Status:             corev1.ConditionTrue,
								Reason:             "FailedReason",
								Message:            "failed after resume",
								LastProbeTime:      v1.NewTime(failedAt),
								LastTransitionTime: v1.NewTime(failedAt),
							},
						},
					},
				}

				fakeClient := fake.NewFakeClientWithScheme(testScheme, mockDataProcess, job)
				handler := &OnceStatusHandler{
					Client:      fakeClient,
					dataProcess: mockDataProcess,
				}

				result, err := handler.GetOperationStatus(ctx, &datav1alpha1.OperationStatus{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Phase).To(Equal(common.PhaseFailed))
				Expect(result.Conditions).To(HaveLen(1))
				Expect(result.Conditions[0].Type).To(Equal(common.ConditionType(batchv1.JobFailed)))
				Expect(result.Conditions[0].Reason).To(Equal("FailedReason"))
				Expect(result.Duration).To(Equal(failedAt.Sub(createdAt).String()))
			})
		})
	})
})
