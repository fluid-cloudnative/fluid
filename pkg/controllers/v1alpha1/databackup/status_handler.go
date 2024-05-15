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

package databackup

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/dataflow"
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
	// set the node labels in status when job finished
	if result.NodeAffinity == nil {
		// generate the node labels
		result.NodeAffinity, err = dataflow.GenerateNodeAffinity(ctx.Client, backupPod)
		if err != nil {
			return nil, fmt.Errorf("error to generate the node labels: %v", err)
		}
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
