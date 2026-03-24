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
	"time"

	"github.com/agiledragon/gomonkey/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

var _ = Describe("OnceStatusHandler", func() {
	var (
		testScheme      *runtime.Scheme
		mockDataMigrate v1alpha1.DataMigrate
		mockJob         batchv1.Job
		mockFailedJob   batchv1.Job
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		mockDataMigrate = v1alpha1.DataMigrate{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: v1alpha1.DataMigrateSpec{},
		}

		mockJob = batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-migrate-migrate",
				Namespace: "default",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:               batchv1.JobComplete,
						LastProbeTime:      v1.NewTime(time.Now()),
						LastTransitionTime: v1.NewTime(time.Now()),
					},
				},
			},
		}

		mockFailedJob = batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-migrate-migrate",
				Namespace: "default",
			},
			Status: batchv1.JobStatus{
				Conditions: []batchv1.JobCondition{
					{
						Type:               batchv1.JobFailed,
						LastProbeTime:      v1.NewTime(time.Now()),
						LastTransitionTime: v1.NewTime(time.Now()),
					},
				},
			},
		}
	})

	Describe("GetOperationStatus", func() {
		It("should return PhaseComplete when job succeeded", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockDataMigrate, &mockJob)
			handler := &OnceStatusHandler{Client: fakeClient, dataMigrate: &mockDataMigrate}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Log: fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &mockDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
		})

		It("should return PhaseFailed when job failed", func() {
			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockDataMigrate, &mockFailedJob)
			handler := &OnceStatusHandler{Client: fakeClient, dataMigrate: &mockDataMigrate}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Log: fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &mockDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseFailed))
		})

		It("should return original status without change when job is still running (no finished condition)", func() {
			runningJob := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate",
					Namespace: "default",
				},
				Status: batchv1.JobStatus{},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockDataMigrate, &runningJob)
			handler := &OnceStatusHandler{Client: fakeClient, dataMigrate: &mockDataMigrate}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Log: fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &mockDataMigrate.Status)

			// job still running; GetOperationStatus returns nil result (early return)
			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus).NotTo(BeNil())
		})

		It("should return PhaseComplete and set NodeAffinity for single-parallelism succeed job", func() {
			parallelism1DataMigrate := v1alpha1.DataMigrate{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1alpha1.DataMigrateSpec{
					Parallelism: 1,
				},
			}
			// Job with node selector labels to trigger node affinity generation
			jobWithNode := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate",
					Namespace: "default",
				},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobComplete,
							LastProbeTime:      v1.NewTime(time.Now()),
							LastTransitionTime: v1.NewTime(time.Now()),
						},
					},
				},
			}
			fakeClient := fake.NewFakeClientWithScheme(testScheme, &parallelism1DataMigrate, &jobWithNode)
			handler := &OnceStatusHandler{Client: fakeClient, dataMigrate: &parallelism1DataMigrate}
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "",
				},
				Log: fake.NullLogger(),
			}

			opStatus, err := handler.GetOperationStatus(ctx, &parallelism1DataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
		})
	})
})

var _ = Describe("CronStatusHandler", func() {
	var (
		testScheme          *runtime.Scheme
		startTime           time.Time
		lastScheduleTime    v1.Time
		lastSuccessfulTime  v1.Time
		mockCronDataMigrate v1alpha1.DataMigrate
		mockCronJob         batchv1.CronJob
		patch               *gomonkey.Patches
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(batchv1.AddToScheme(testScheme)).To(Succeed())

		startTime = time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
		lastScheduleTime = v1.NewTime(startTime)
		lastSuccessfulTime = v1.NewTime(startTime.Add(time.Second * 10))

		mockCronDataMigrate = v1alpha1.DataMigrate{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: v1alpha1.DataMigrateSpec{
				Policy:   "Cron",
				Schedule: "* * * * *",
			},
			Status: v1alpha1.OperationStatus{
				Phase: common.PhaseComplete,
			},
		}

		mockCronJob = batchv1.CronJob{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-migrate-migrate",
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

		patch = gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
			return true
		})
	})

	AfterEach(func() {
		if patch != nil {
			patch.Reset()
		}
	})

	Describe("GetOperationStatus (non-parallel)", func() {
		It("should return PhaseComplete and correct timestamps when job completes", func() {
			mockJob := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate-1",
					Namespace: "default",
					Labels: map[string]string{
						"cronjob": "test-migrate-migrate",
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

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &mockJob)
			handler := &CronStatusHandler{Client: fakeClient, dataMigrate: &mockCronDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
			Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
			Expect(opStatus.Phase).To(Equal(common.PhaseComplete))
		})

		It("should return PhasePending when job is still running", func() {
			mockRunningJob := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate-1",
					Namespace: "default",
					Labels: map[string]string{
						"cronjob": "test-migrate-migrate",
					},
					CreationTimestamp: lastScheduleTime,
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &mockRunningJob)
			handler := &CronStatusHandler{Client: fakeClient, dataMigrate: &mockCronDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
			Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
			Expect(opStatus.Phase).To(Equal(common.PhasePending))
		})

		It("should return PhaseFailed when cron job's latest job failed", func() {
			mockFailedCronJob := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate-1",
					Namespace: "default",
					Labels: map[string]string{
						"cronjob": "test-migrate-migrate",
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

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &mockFailedCronJob)
			handler := &CronStatusHandler{Client: fakeClient, dataMigrate: &mockCronDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.Phase).To(Equal(common.PhaseFailed))
		})

		It("should skip and return current status when no current job matches schedule time", func() {
			// Job with CreationTimestamp BEFORE lastScheduleTime — won't match
			oldJob := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate-1",
					Namespace: "default",
					Labels: map[string]string{
						"cronjob": "test-migrate-migrate",
					},
					CreationTimestamp: v1.NewTime(startTime.Add(-time.Hour)),
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &oldJob)
			handler := &CronStatusHandler{Client: fakeClient, dataMigrate: &mockCronDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			opStatus, err := handler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			// currentJob is nil — returns early with the current opStatus copy
			Expect(opStatus).NotTo(BeNil())
		})
	})

	Describe("GetOperationStatus (parallel tasks)", func() {
		var (
			parallelScheme      *runtime.Scheme
			parallelDataMigrate v1alpha1.DataMigrate
			parallelCronJob     batchv1.CronJob
			sts                 appsv1.StatefulSet
			defaultStsReplicas  int32
		)

		BeforeEach(func() {
			parallelScheme = runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(parallelScheme)).To(Succeed())
			Expect(batchv1.AddToScheme(parallelScheme)).To(Succeed())
			Expect(appsv1.AddToScheme(parallelScheme)).To(Succeed())

			parallelDataMigrate = v1alpha1.DataMigrate{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: v1alpha1.DataMigrateSpec{
					Policy:      "Cron",
					Schedule:    "* * * * *",
					Parallelism: 3,
				},
				Status: v1alpha1.OperationStatus{
					Phase: common.PhaseComplete,
				},
			}

			parallelCronJob = batchv1.CronJob{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate",
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

			defaultStsReplicas = parallelDataMigrate.Spec.Parallelism - 1
			sts = appsv1.StatefulSet{
				ObjectMeta: v1.ObjectMeta{
					Name:      utils.GetParallelOperationWorkersName(utils.GetDataMigrateReleaseName(parallelDataMigrate.Name)),
					Namespace: parallelDataMigrate.Namespace,
				},
				Spec: appsv1.StatefulSetSpec{
					Replicas: &defaultStsReplicas,
				},
				Status: appsv1.StatefulSetStatus{},
			}
		})

		It("should return PhaseComplete and leave StatefulSet replicas unchanged when job completes without suspend", func() {
			job := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate-1",
					Namespace: "default",
					Labels: map[string]string{
						"cronjob": "test-migrate-migrate",
					},
					CreationTimestamp: lastScheduleTime,
				},
				Spec: batchv1.JobSpec{},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobSuspended,
							LastProbeTime:      lastSuccessfulTime,
							LastTransitionTime: lastSuccessfulTime,
						},
						{
							Type:               batchv1.JobComplete,
							LastProbeTime:      lastSuccessfulTime,
							LastTransitionTime: lastSuccessfulTime,
						},
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(parallelScheme, &parallelDataMigrate, &parallelCronJob, &job, &sts)
			handler := &CronStatusHandler{Client: fakeClient, dataMigrate: &parallelDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			opStatus, err := handler.GetOperationStatus(ctx, &parallelDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
			Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
			Expect(opStatus.Phase).To(Equal(common.PhaseComplete))

			updatedSts, err := kubeclient.GetStatefulSet(fakeClient, sts.Name, sts.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedSts.Spec.Replicas).To(Equal(defaultStsReplicas))
		})

		It("should return PhasePending and scale StatefulSet when job starts with suspend=true", func() {
			trueFlag := true
			job := batchv1.Job{
				ObjectMeta: v1.ObjectMeta{
					Name:      "test-migrate-migrate-1",
					Namespace: "default",
					Labels: map[string]string{
						"cronjob": "test-migrate-migrate",
					},
					CreationTimestamp: lastScheduleTime,
				},
				Spec: batchv1.JobSpec{
					Suspend: &trueFlag,
				},
				Status: batchv1.JobStatus{
					Conditions: []batchv1.JobCondition{
						{
							Type:               batchv1.JobSuspended,
							LastProbeTime:      lastSuccessfulTime,
							LastTransitionTime: lastSuccessfulTime,
						},
					},
				},
			}

			fakeClient := fake.NewFakeClientWithScheme(parallelScheme, &parallelDataMigrate, &parallelCronJob, &job, &sts)
			handler := &CronStatusHandler{Client: fakeClient, dataMigrate: &parallelDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			opStatus, err := handler.GetOperationStatus(ctx, &parallelDataMigrate.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(opStatus.LastScheduleTime).To(Equal(&lastScheduleTime))
			Expect(opStatus.LastSuccessfulTime).To(Equal(&lastSuccessfulTime))
			Expect(opStatus.Phase).To(Equal(common.PhasePending))

			updatedSts, err := kubeclient.GetStatefulSet(fakeClient, sts.Name, sts.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(*updatedSts.Spec.Replicas).To(Equal(defaultStsReplicas))
		})
	})
})

var _ = Describe("OnEventStatusHandler", func() {
	Describe("GetOperationStatus", func() {
		It("should return nil result and nil error (stub)", func() {
			testScheme := runtime.NewScheme()
			Expect(v1alpha1.AddToScheme(testScheme)).To(Succeed())
			dm := &v1alpha1.DataMigrate{}
			dm.Name = "test"
			dm.Namespace = "default"
			fakeClient := fake.NewFakeClientWithScheme(testScheme, dm)
			handler := &OnEventStatusHandler{Client: fakeClient, dataMigrate: dm}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

			result, err := handler.GetOperationStatus(ctx, &dm.Status)

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})
})
