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

package datamigrate

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	"reflect"
	"testing"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestOnceGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	mockDataMigrate := v1alpha1.DataMigrate{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataMigrateSpec{},
	}

	mockJob := batchv1.Job{
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

	mockFailedJob := batchv1.Job{
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

	testcases := []struct {
		name          string
		job           batchv1.Job
		expectedPhase common.Phase
	}{
		{
			name:          "job success",
			job:           mockJob,
			expectedPhase: common.PhaseComplete,
		},
		{
			name:          "job failed",
			job:           mockFailedJob,
			expectedPhase: common.PhaseFailed,
		},
	}

	for _, testcase := range testcases {
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataMigrate, &testcase.job)
		onceStatusHandler := &OnceStatusHandler{Client: client, dataMigrate: &mockDataMigrate, Reader: client}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "",
			},
			Log: fake.NullLogger(),
		}
		opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataMigrate.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}

func TestCronGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
	lastScheduleTime := v1.NewTime(startTime)
	lastSuccessfulTime := v1.NewTime(startTime.Add(time.Second * 10))

	mockCronDataMigrate := v1alpha1.DataMigrate{
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

	mockCronJob := batchv1.CronJob{
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

	testcases := []struct {
		name          string
		job           batchv1.Job
		expectedPhase common.Phase
	}{
		{
			name:          "job complete",
			job:           mockJob,
			expectedPhase: common.PhaseComplete,
		},
		{
			name:          "job running",
			job:           mockRunningJob,
			expectedPhase: common.PhasePending,
		},
	}

	for _, testcase := range testcases {
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &testcase.job)
		cronStatusHandler := &CronStatusHandler{Client: client, dataMigrate: &mockCronDataMigrate}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		opStatus, err := cronStatusHandler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if !reflect.DeepEqual(opStatus.LastScheduleTime, &lastScheduleTime) || !reflect.DeepEqual(opStatus.LastSuccessfulTime, &lastSuccessfulTime) {
			t.Error("fail to get correct Operation Status", "expected LastScheduleTime", lastScheduleTime, "expected LastSuccessfulTime", lastSuccessfulTime, "get", opStatus)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}

func TestCronGetOperationStatusWithParallelTasks(t1 *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)

	startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
	lastScheduleTime := v1.NewTime(startTime)
	lastSuccessfulTime := v1.NewTime(startTime.Add(time.Second * 10))

	mockCronDataMigrate := v1alpha1.DataMigrate{
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

	mockCronJob := batchv1.CronJob{
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

	defaultStsReplicas := mockCronDataMigrate.Spec.Parallelism - 1
	sts := appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      utils.GetParallelOperationWorkersName(utils.GetDataMigrateReleaseName(mockCronDataMigrate.Name)),
			Namespace: mockCronDataMigrate.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &defaultStsReplicas,
		},
		Status: appsv1.StatefulSetStatus{},
	}

	trueFlag := true

	testcases := []struct {
		name          string
		job           batchv1.Job
		expectedPhase common.Phase
		stsReplicas   int32
	}{
		{
			name: "job complete, suspend is false",
			job: batchv1.Job{
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
			},
			expectedPhase: common.PhaseComplete,
			stsReplicas:   *sts.Spec.Replicas,
		},
		{
			name: "job start, suspend is true",
			job: batchv1.Job{
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
			},
			expectedPhase: common.PhasePending,
			stsReplicas:   *sts.Spec.Replicas,
		},
	}

	for _, testcase := range testcases {
		t1.Run(testcase.name, func(t *testing.T) {
			client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &testcase.job, &sts)
			cronStatusHandler := &CronStatusHandler{Client: client, dataMigrate: &mockCronDataMigrate}
			ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
			opStatus, err := cronStatusHandler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)
			if err != nil {
				t.Errorf("fail to GetOperationStatus with error %v", err)
			}
			if !reflect.DeepEqual(opStatus.LastScheduleTime, &lastScheduleTime) || !reflect.DeepEqual(opStatus.LastSuccessfulTime, &lastSuccessfulTime) {
				t.Error("fail to get correct Operation Status", "expected LastScheduleTime", lastScheduleTime, "expected LastSuccessfulTime", lastSuccessfulTime, "get", opStatus)
			}
			if opStatus.Phase != testcase.expectedPhase {
				t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
			}

			updatedSts, err := kubeclient.GetStatefulSet(client, sts.Name, sts.Namespace)
			if err != nil {
				t.Error("Failed to GetStatefulSet", err)
			}

			if *updatedSts.Spec.Replicas != testcase.stsReplicas {
				t.Error("Failed to GetOperationStatus", "expected replicas", testcase.stsReplicas, "get", *sts.Spec.Replicas)
			}
		})
	}
}
