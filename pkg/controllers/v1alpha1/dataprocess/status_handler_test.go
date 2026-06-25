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

package dataprocess

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
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

func TestOnceGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{},
	}

	mockJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-processor-job",
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
			Name:      "test-processor-job",
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
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess, &testcase.job)
		onceStatusHandler := &OnceStatusHandler{Client: client, dataProcess: &mockDataProcess}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "",
			},
			Log: fake.NullLogger(),
		}
		opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataProcess.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}

func TestOnEventGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy: v1alpha1.OnEvent,
		},
	}

	mockJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-processor-job",
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
			Name:      "test-processor-job",
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
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess, &testcase.job)
		handler := &OnEventStatusHandler{Client: client, dataProcess: &mockDataProcess}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "",
			},
			Log: fake.NullLogger(),
		}
		opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}

func TestOnEventGetOperationStatusJobStillRunning(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy: v1alpha1.OnEvent,
		},
		Status: v1alpha1.OperationStatus{
			Phase: common.PhasePending,
		},
	}

	runningJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-processor-job",
			Namespace: "default",
		},
		Status: batchv1.JobStatus{},
	}

	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess, &runningJob)
	handler := &OnEventStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "",
		},
		Log: fake.NullLogger(),
	}
	opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
	if err != nil {
		t.Errorf("fail to GetOperationStatus with error %v", err)
	}
	if opStatus.Phase != common.PhasePending {
		t.Error("Failed to GetOperationStatus", "expected phase", common.PhasePending, "get", opStatus.Phase)
	}
}

func TestCronGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	patch := gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
		return true
	})
	defer patch.Reset()

	startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
	lastScheduleTime := v1.NewTime(startTime)
	lastSuccessfulTime := v1.NewTime(startTime.Add(time.Second * 10))

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy:   v1alpha1.Cron,
			Schedule: "* * * * *",
		},
		Status: v1alpha1.OperationStatus{
			Phase: common.PhaseComplete,
		},
	}

	releaseName := utils.GetDataProcessReleaseName(mockDataProcess.GetName())
	cronjobName := utils.GetDataProcessJobName(releaseName)

	mockCronJob := batchv1.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      cronjobName,
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
			Name:      cronjobName + "-1",
			Namespace: "default",
			Labels: map[string]string{
				"cronjob": cronjobName,
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

	mockFailedJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      cronjobName + "-1",
			Namespace: "default",
			Labels: map[string]string{
				"cronjob": cronjobName,
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

	runningJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      cronjobName + "-1",
			Namespace: "default",
			Labels: map[string]string{
				"cronjob": cronjobName,
			},
			CreationTimestamp: lastScheduleTime,
		},
		Status: batchv1.JobStatus{},
	}

	testcases := []struct {
		name          string
		job           *batchv1.Job
		expectedPhase common.Phase
	}{
		{
			name:          "job success yields PhaseComplete",
			job:           &mockJob,
			expectedPhase: common.PhaseComplete,
		},
		{
			name:          "job failed yields PhaseFailed",
			job:           &mockFailedJob,
			expectedPhase: common.PhaseFailed,
		},
		{
			name:          "job still running yields PhasePending",
			job:           &runningJob,
			expectedPhase: common.PhasePending,
		},
	}

	for _, testcase := range testcases {
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess, &mockCronJob, testcase.job)
		handler := &CronStatusHandler{Client: client, dataProcess: &mockDataProcess}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

		opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
		if err != nil {
			t.Errorf("%s: fail to GetOperationStatus with error %v", testcase.name, err)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Errorf("%s: expected phase %s, got %s", testcase.name, testcase.expectedPhase, opStatus.Phase)
		}
	}
}

func TestCronGetOperationStatusNotScheduledYet(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	patch := gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
		return true
	})
	defer patch.Reset()

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy:   v1alpha1.Cron,
			Schedule: "* * * * *",
		},
	}

	releaseName := utils.GetDataProcessReleaseName(mockDataProcess.GetName())
	cronjobName := utils.GetDataProcessJobName(releaseName)

	// CronJob exists but has not been scheduled yet (LastScheduleTime is nil)
	mockCronJob := batchv1.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      cronjobName,
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "* * * * *",
		},
		Status: batchv1.CronJobStatus{},
	}

	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess, &mockCronJob)
	handler := &CronStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

	opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
	if err != nil {
		t.Errorf("fail to GetOperationStatus with error %v", err)
	}
	if opStatus == nil {
		t.Error("expected non-nil opStatus")
	}
}

func TestOnceGetOperationStatusJobNotFound(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	// Patch helm.DeleteReleaseIfExists to avoid shelling out to ddc-helm binary
	helmPatch := gomonkey.ApplyFunc(helm.DeleteReleaseIfExists, func(name, namespace string) error {
		return nil
	})
	defer helmPatch.Reset()

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy: v1alpha1.Once,
		},
		Status: v1alpha1.OperationStatus{
			Phase: common.PhasePending,
		},
	}

	// No job in fake client - simulates NotFound
	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess)
	handler := &OnceStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: ""},
		Log:            fake.NullLogger(),
	}

	// When job is not found, helm release is deleted and we get early return with unchanged status
	opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
	if err != nil {
		t.Errorf("unexpected error on NotFound path: %v", err)
	}
	if opStatus == nil {
		t.Error("expected non-nil opStatus")
	}
	if opStatus.Phase != common.PhasePending {
		t.Errorf("expected phase %s, got %s", common.PhasePending, opStatus.Phase)
	}
}

func TestOnEventGetOperationStatusJobNotFound(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	// Patch helm.DeleteReleaseIfExists to avoid shelling out to ddc-helm binary
	helmPatch := gomonkey.ApplyFunc(helm.DeleteReleaseIfExists, func(name, namespace string) error {
		return nil
	})
	defer helmPatch.Reset()

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy: v1alpha1.OnEvent,
		},
		Status: v1alpha1.OperationStatus{
			Phase: common.PhasePending,
		},
	}

	// No job in fake client - simulates NotFound
	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess)
	handler := &OnEventStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: ""},
		Log:            fake.NullLogger(),
	}

	// When job is not found, helm release is deleted and we get early return with unchanged status
	opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
	if err != nil {
		t.Errorf("unexpected error on NotFound path: %v", err)
	}
	if opStatus == nil {
		t.Error("expected non-nil opStatus")
	}
	if opStatus.Phase != common.PhasePending {
		t.Errorf("expected phase %s, got %s", common.PhasePending, opStatus.Phase)
	}
}

func TestCronGetOperationStatusCronJobNotFound(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	patch := gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
		return true
	})
	defer patch.Reset()

	// Patch helm.DeleteReleaseIfExists to avoid shelling out to ddc-helm binary
	helmPatch := gomonkey.ApplyFunc(helm.DeleteReleaseIfExists, func(name, namespace string) error {
		return nil
	})
	defer helmPatch.Reset()

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{
			Policy:   v1alpha1.Cron,
			Schedule: "* * * * *",
		},
		Status: v1alpha1.OperationStatus{
			Phase: common.PhasePending,
		},
	}

	// No CronJob in fake client - simulates NotFound
	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess)
	handler := &CronStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}

	// When CronJob is not found, helm release is deleted and we get early return with unchanged status
	opStatus, err := handler.GetOperationStatus(ctx, &mockDataProcess.Status)
	if err != nil {
		t.Errorf("unexpected error on NotFound path: %v", err)
	}
	if opStatus == nil {
		t.Error("expected non-nil opStatus")
	}
	if opStatus.Phase != common.PhasePending {
		t.Errorf("expected phase %s, got %s", common.PhasePending, opStatus.Phase)
	}
}
