/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/
package databackup

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestOnceGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = corev1.AddToScheme(testScheme)

	mockDataBackup := v1alpha1.DataBackup{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataBackupSpec{},
	}

	mockPod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}

	mockFailedPod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodFailed,
		},
	}

	testcases := []struct {
		name          string
		pod           corev1.Pod
		expectedPhase common.Phase
	}{
		{
			name:          "job success",
			pod:           mockPod,
			expectedPhase: common.PhaseComplete,
		},
		{
			name:          "job failed",
			pod:           mockFailedPod,
			expectedPhase: common.PhaseFailed,
		},
	}

	for _, testcase := range testcases {
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataBackup, &testcase.pod)
		onceStatusHandler := &OnceHandler{dataBackup: &mockDataBackup}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "",
			},
			Client: client,
			Log:    fake.NullLogger(),
		}
		opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataBackup.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}
