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
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OnceHandler struct {
	dataBackup *v1alpha1.DataBackup
}

var _ dataoperation.StatusHandler = &OnceHandler{}

// UpdateStatusByHelmStatus update the operation status according to helm job status
func (o *OnceHandler) GetOperationStatus(ctx runtime.ReconcileRequestContext, opStatus *v1alpha1.OperationStatus) (result *v1alpha1.OperationStatus, err error) {
	result = opStatus.DeepCopy()
	object := o.dataBackup
	// 1. gdt pod name
	backupPodName := utils.GetDataBackupPodName(object.GetName())
	backupPod, err := kubeclient.GetPodByName(ctx.Client, backupPodName, object.GetNamespace())
	if err != nil {
		ctx.Log.Error(err, "Failed to get databackup-pod")
		return
	}

	// 2. only update status if finished
	if !kubeclient.IsFinishedPod(backupPod) {
		return
	}

	var finishTime time.Time
	if len(backupPod.Status.Conditions) != 0 {
		finishTime = backupPod.Status.Conditions[0].LastTransitionTime.Time
	} else {
		// fail to get finishTime, use current time as default
		finishTime = time.Now()
	}
	result.Duration = utils.CalculateDuration(backupPod.CreationTimestamp.Time, finishTime)

	if kubeclient.IsSucceededPod(backupPod) {
		result.Phase = common.PhaseComplete
		result.Conditions = []v1alpha1.Condition{
			{
				Type:               common.Complete,
				Status:             v1.ConditionTrue,
				Reason:             "BackupSuccessful",
				Message:            "Backup Pod exec successfully and finish",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(finishTime),
			},
		}
	} else if kubeclient.IsFailedPod(backupPod) {
		result.Phase = common.PhaseFailed
		result.Conditions = []v1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "BackupFailed",
				Message:            "Backup Pod exec failed and exit",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(finishTime),
			},
		}
	}
	return
}
