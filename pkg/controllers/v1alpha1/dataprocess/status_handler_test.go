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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

func TestOnceGetOperationStatusIgnoresMissingJobAfterHelmCleanup(t *testing.T) {
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

	patch := gomonkey.ApplyFunc(helm.DeleteReleaseIfExists, func(_ string, _ string) error { return nil })
	defer patch.Reset()

	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess)
	handler := &OnceStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test"},
		Log:            fake.NullLogger(),
	}

	opStatus := &v1alpha1.OperationStatus{Phase: common.PhaseExecuting}
	result, err := handler.GetOperationStatus(ctx, opStatus)
	if err != nil {
		t.Fatalf("expected missing job to be handled without error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result == opStatus {
		t.Fatal("expected deep-copied status result")
	}
	if result.Phase != common.PhaseExecuting {
		t.Fatalf("expected phase %q, got %q", common.PhaseExecuting, result.Phase)
	}
}

func TestOnceGetOperationStatusUsesFinishedConditionWhenConditionsOutOfOrder(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	createdAt := time.Now().Add(-2 * time.Minute)
	suspendedAt := createdAt.Add(30 * time.Second)
	failedAt := createdAt.Add(90 * time.Second)

	mockDataProcess := v1alpha1.DataProcess{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataProcessSpec{},
	}

	mockJob := batchv1.Job{
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

	client := fake.NewFakeClientWithScheme(testScheme, &mockDataProcess, &mockJob)
	handler := &OnceStatusHandler{Client: client, dataProcess: &mockDataProcess}
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "test"},
		Log:            fake.NullLogger(),
	}

	result, err := handler.GetOperationStatus(ctx, &v1alpha1.OperationStatus{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Phase != common.PhaseFailed {
		t.Fatalf("expected phase %q, got %q", common.PhaseFailed, result.Phase)
	}
	if len(result.Conditions) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(result.Conditions))
	}
	if result.Conditions[0].Type != common.ConditionType(batchv1.JobFailed) {
		t.Fatalf("expected failed condition type, got %q", result.Conditions[0].Type)
	}
	if result.Conditions[0].Reason != "FailedReason" {
		t.Fatalf("expected failed reason, got %q", result.Conditions[0].Reason)
	}
	if result.Duration != failedAt.Sub(createdAt).String() {
		t.Fatalf("expected duration %q, got %q", failedAt.Sub(createdAt).String(), result.Duration)
	}
}
