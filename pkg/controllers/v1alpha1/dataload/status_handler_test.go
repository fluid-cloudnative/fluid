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
	"errors"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("OnceStatusHandler", func() {
	var (
		testScheme   *runtime.Scheme
		mockDataload v1alpha1.DataLoad
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		mockDataload = v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload",
				Namespace: "default",
			},
			Spec: v1alpha1.DataLoadSpec{
				Dataset: v1alpha1.TargetDataset{
					Name:      "hadoop",
					Namespace: "default",
				},
			},
		}
	})

	DescribeTable("GetOperationStatus",
		func(jobConditionType batchv1.JobConditionType, expectedPhase common.Phase) {
			mockJob := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-dataload-loader-job",
					Namespace: "default",
				},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               jobConditionType,
							LastProbeTime:      v1.NewTime(time.Now()),
							LastTransitionTime: v1.NewTime(time.Now()),
						},
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &mockJob)
			handler := &OnceStatusHandler{Client: fakeClient, dataLoad: &mockDataload}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Log: fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(expectedPhase))
		},
		Entry("job success returns PhaseComplete", batchv1.JobComplete, common.PhaseComplete),
		Entry("job failed returns PhaseFailed", batchv1.JobFailed, common.PhaseFailed),
	)

	It("returns current status when job is still running (no finished condition)", func() {
		runningJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload-loader-job",
				Namespace: "default",
			},
			Status: batchv1.JobStatus{
				// No conditions — job is still running
			},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &runningJob)
		handler := &OnceStatusHandler{Client: fakeClient, dataLoad: &mockDataload}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{Namespace: "default"},
			Log:            fake.NullLogger(),
		}

		opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		// Status should be returned unchanged when job is still running
		Expect(opStatus).NotTo(BeNil())
		Expect(opStatus.Phase).To(Equal(mockDataload.Status.Phase))
	})

	It("returns error when job is missing and helm release cleanup fails", func() {
		sentinelErr := errors.New("cleanup failed")
		expectedReleaseName := utils.GetDataLoadReleaseName(mockDataload.Name)
		var cleanupReleaseName string
		var cleanupNamespace string
		patches := gomonkey.ApplyFunc(helm.DeleteReleaseIfExists, func(name, namespace string) error {
			cleanupReleaseName = name
			cleanupNamespace = namespace
			return sentinelErr
		})
		defer patches.Reset()

		fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockDataload)
		handler := &OnceStatusHandler{Client: fakeClient, dataLoad: &mockDataload}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{Namespace: "default"},
			Log:            fake.NullLogger(),
		}

		_, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
		Expect(err).To(MatchError(sentinelErr))
		Expect(cleanupReleaseName).To(Equal(expectedReleaseName))
		Expect(cleanupNamespace).To(Equal(mockDataload.Namespace))
	})
})

var _ = Describe("CronStatusHandler", func() {
	var (
		testScheme         *runtime.Scheme
		mockCronDataload   v1alpha1.DataLoad
		lastScheduleTime   v1.Time
		lastSuccessfulTime v1.Time
		mockCronJob        batchv1.CronJob
		patch              *gomonkey.Patches
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
		lastScheduleTime = v1.NewTime(startTime)
		lastSuccessfulTime = v1.NewTime(startTime.Add(time.Second * 10))

		mockCronDataload = v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload",
				Namespace: "default",
			},
			Spec: v1alpha1.DataLoadSpec{
				Dataset: v1alpha1.TargetDataset{
					Name:      "hadoop",
					Namespace: "default",
				},
				Policy:   "Cron",
				Schedule: "* * * * *",
			},
			Status: v1alpha1.OperationStatus{
				Phase: common.PhaseComplete,
			},
		}

		mockCronJob = batchv1.CronJob{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload-loader-job",
				Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "* * * * *",
			},
			Status: batchv1.CronJobStatus{
				LastScheduleTime:   &lastScheduleTime,
				LastSuccessfulTime: &lastSuccessfulTime,
			},
		}

		// Patch IsBatchV1CronJobSupported to avoid real cluster API discovery in unit tests.
		patch = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
			return true
		})
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	It("returns PhaseComplete when job is complete", func() {
		mockJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload-loader-job-1",
				Namespace: "default",
				Labels: map[string]string{
					"cronjob": "test-dataload-loader-job",
				},
				CreationTimestamp: lastScheduleTime,
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:               batchv1.JobComplete,
						LastProbeTime:      lastSuccessfulTime,
						LastTransitionTime: lastSuccessfulTime,
					},
				},
			},
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &mockJob)
		handler := &CronStatusHandler{Client: fakeClient, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
		Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
	})

	It("returns PhasePending when job is still running", func() {
		mockRunningJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload-loader-job-1",
				Namespace: "default",
				Labels: map[string]string{
					"cronjob": "test-dataload-loader-job",
				},
				CreationTimestamp: lastScheduleTime,
			},
			Status: batchv1.JobStatus{},
		}

		fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &mockRunningJob)
		handler := &CronStatusHandler{Client: fakeClient, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
		Expect(opStatus.Phase).To(Equal(common.PhasePending))
	})

	It("returns PhaseFailed when job failed", func() {
		mockFailedJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload-loader-job-1",
				Namespace: "default",
				Labels: map[string]string{
					"cronjob": "test-dataload-loader-job",
				},
				CreationTimestamp: lastScheduleTime,
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:               batchv1.JobFailed,
						LastProbeTime:      lastSuccessfulTime,
						LastTransitionTime: lastSuccessfulTime,
					},
				},
			},
		}

		// Start with pending phase so it can transition to failed
		mockCronDataload.Status.Phase = common.PhasePending
		fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &mockFailedJob)
		handler := &CronStatusHandler{Client: fakeClient, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.Phase).To(Equal(common.PhaseFailed))
	})

	It("returns current status when no current job matches schedule", func() {
		// CronJob with LastScheduleTime but no jobs matching
		fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob)
		handler := &CronStatusHandler{Client: fakeClient, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		// When no current job is found, the status is returned unchanged
		Expect(opStatus).NotTo(BeNil())
	})
})

var _ = Describe("OnEventStatusHandler", func() {
	It("GetOperationStatus returns nil (not yet implemented)", func() {
		testScheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())

		dataLoad := &v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-dataload",
				Namespace: "default",
			},
		}
		fakeClient := fake.NewFakeClientWithScheme(testScheme, dataLoad)
		handler := &OnEventStatusHandler{Client: fakeClient, dataLoad: dataLoad}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		result, err := handler.GetOperationStatus(ctx, &dataLoad.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})
})
