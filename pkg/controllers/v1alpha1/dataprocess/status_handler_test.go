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
