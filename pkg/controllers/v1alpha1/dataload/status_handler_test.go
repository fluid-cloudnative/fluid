/*
Copyright 2023 The Fluid Authors.

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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
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
		func(job batchv1.Job, expectedPhase common.Phase) {
			client := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &job)
			handler := &OnceStatusHandler{Client: client, dataLoad: &mockDataload}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{Namespace: "default", Name: ""},
				Log:            fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(expectedPhase))
		},
		Entry("job success yields PhaseComplete",
			batchv1.Job{
				ObjectMeta: v1.ObjectMeta{Name: "test-dataload-loader-job", Namespace: "default"},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobComplete,
							LastProbeTime:      v1.NewTime(time.Now()),
							LastTransitionTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			common.PhaseComplete,
		),
		Entry("job failed yields PhaseFailed",
			batchv1.Job{
				ObjectMeta: v1.ObjectMeta{Name: "test-dataload-loader-job", Namespace: "default"},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobFailed,
							LastProbeTime:      v1.NewTime(time.Now()),
							LastTransitionTime: v1.NewTime(time.Now()),
						},
					},
				},
			},
			common.PhaseFailed,
		),
	)

	It("GetOperationStatus returns current status when job is still running (no finished condition)", func() {
		runningJob := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{Name: "test-dataload-loader-job", Namespace: "default"},
			Status:     batchv1.JobStatus{}, // no conditions = still running
		}
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataload, &runningJob)
		handler := &OnceStatusHandler{Client: client, dataLoad: &mockDataload}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: ""},
			Log:            fake.NullLogger(),
		}

		opStatus, err := handler.GetOperationStatus(ctx, &mockDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		// when job is still running, status is returned as-is (no phase change)
		Expect(opStatus).NotTo(BeNil())
	})
})

var _ = Describe("CronStatusHandler", func() {
	var (
		testScheme         *runtime.Scheme
		mockCronDataload   v1alpha1.DataLoad
		lastScheduleTime   v1.Time
		lastSuccessfulTime v1.Time
		mockCronJob        batchv1.CronJob
		patches            *gomonkey.Patches
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
		lastScheduleTime = v1.NewTime(startTime)
		lastSuccessfulTime = v1.NewTime(startTime.Add(time.Second * 10))

		mockCronDataload = v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{Name: "test-dataload", Namespace: "default"},
			Spec: v1alpha1.DataLoadSpec{
				Dataset:  v1alpha1.TargetDataset{Name: "hadoop", Namespace: "default"},
				Policy:   "Cron",
				Schedule: "* * * * *",
			},
			Status: v1alpha1.OperationStatus{
				Phase: common.PhaseComplete,
			},
		}

		mockCronJob = batchv1.CronJob{
			ObjectMeta: v1.ObjectMeta{Name: "test-dataload-loader-job", Namespace: "default"},
			Spec:       batchv1.CronJobSpec{Schedule: "* * * * *"},
			Status: batchv1.CronJobStatus{
				LastScheduleTime:   &lastScheduleTime,
				LastSuccessfulTime: &lastSuccessfulTime,
			},
		}

		// Patch IsBatchV1CronJobSupported to avoid ctrl.GetConfigOrDie() panic in unit tests.
		patches = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
			return true
		})
	})

	AfterEach(func() {
		patches.Reset()
	})

	It("completed job yields PhaseComplete", func() {
		job := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-dataload-loader-job-1",
				Namespace:         "default",
				Labels:            map[string]string{"cronjob": "test-dataload-loader-job"},
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
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &job)
		handler := &CronStatusHandler{Client: client, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
		Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
	})

	It("running job yields PhasePending", func() {
		job := batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-dataload-loader-job-1",
				Namespace:         "default",
				Labels:            map[string]string{"cronjob": "test-dataload-loader-job"},
				CreationTimestamp: lastScheduleTime,
			},
			Status: batchv1.JobStatus{},
		}
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob, &job)
		handler := &CronStatusHandler{Client: client, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
		Expect(opStatus.Phase).To(Equal(common.PhasePending))
	})

	It("returns status with timestamps when no current job matches schedule time", func() {
		// no jobs in the fake client that match lastScheduleTime
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataload, &mockCronJob)
		handler := &CronStatusHandler{Client: client, dataLoad: &mockCronDataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataload.Status)
		Expect(err).NotTo(HaveOccurred())
		// should still update the schedule/successful times from the cronjob status
		Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
		Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
	})
})

var _ = Describe("OnEventStatusHandler", func() {
	It("GetOperationStatus returns nil result and nil error (stub)", func() {
		testScheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		dataload := &v1alpha1.DataLoad{
			ObjectMeta: v1.ObjectMeta{Name: "test-dataload", Namespace: "default"},
		}
		c := fake.NewFakeClientWithScheme(testScheme, dataload)
		handler := &OnEventStatusHandler{Client: c, dataLoad: dataload}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		result, err := handler.GetOperationStatus(ctx, &dataload.Status)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeNil())
	})
})
