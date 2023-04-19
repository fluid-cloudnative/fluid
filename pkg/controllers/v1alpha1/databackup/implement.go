/*
Copyright 2022 The Fluid Authors.

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
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// UpdateStatusByHelmStatus update the operation status according to helm job status
func (r *DataBackupReconciler) UpdateStatusByHelmStatus(ctx runtime.ReconcileRequestContext, object client.Object, opStatus *v1alpha1.OperationStatus) (err error) {
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
	opStatus.Duration = utils.CalculateDuration(object.GetCreationTimestamp().Time, finishTime)

	if kubeclient.IsSucceededPod(backupPod) {
		opStatus.Phase = common.PhaseComplete
		opStatus.Conditions = []v1alpha1.Condition{
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
		opStatus.Phase = common.PhaseFailed
		opStatus.Conditions = []v1alpha1.Condition{
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

func (r *DataBackupReconciler) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdatabackup.DatabackupChart
}

func (r *DataBackupReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	dataBackup, ok := object.(*v1alpha1.DataBackup)
	if !ok {
		return fmt.Errorf("object %v is not a DataBackup", object)
	}

	infos[cdatabackup.BackupLocationPath] = dataBackup.Spec.BackupPath

	if strings.HasPrefix(dataBackup.Spec.BackupPath, common.PathScheme.String()) {
		podName := object.GetName() + "-pod"
		backupPod, err := kubeclient.GetPodByName(r.Client, podName, object.GetNamespace())
		if err != nil {
			r.Log.Error(err, "Failed to get backup pod")
			return fmt.Errorf("failed to get backup pod")
		}
		infos[cdatabackup.BackupLocationNodeName] = backupPod.Spec.NodeName
	} else {
		infos[cdatabackup.BackupLocationNodeName] = "NA"
	}

	return nil
}
func (r *DataBackupReconciler) Validate(ctx runtime.ReconcileRequestContext, object client.Object) ([]v1alpha1.Condition, error) {
	dataBackup, ok := object.(*v1alpha1.DataBackup)
	if !ok {
		return []v1alpha1.Condition{}, fmt.Errorf("object %v is not a DataBackup", object)
	}

	// 0. check the supported backup path format
	if !strings.HasPrefix(dataBackup.Spec.BackupPath, common.PathScheme.String()) && !strings.HasPrefix(dataBackup.Spec.BackupPath, common.VolumeScheme.String()) {
		err := fmt.Errorf("don't support path in this form, path: %s", dataBackup.Spec.BackupPath)
		return []v1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             "PathNotSupported",
				Message:            "Only support pvc and local path now",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}, err
	}
	return nil, nil
}

func (r *DataBackupReconciler) SetTargetDatasetStatusInProgress(dataset *v1alpha1.Dataset) {
}

func (r *DataBackupReconciler) RemoveTargetDatasetStatusInProgress(dataset *v1alpha1.Dataset) {
}

func (r *DataBackupReconciler) GetOperationType() dataoperation.OperationType {
	return dataoperation.DataBackup
}

func (r *DataBackupReconciler) GetTargetDatasetNamespacedName(object client.Object) (*types.NamespacedName, error) {
	typeObject, ok := object.(*v1alpha1.DataBackup)
	if !ok {
		return nil, fmt.Errorf("object %v is not a DataBackup", object)
	}

	targetDataBackup := *typeObject

	return &types.NamespacedName{
		Name:      targetDataBackup.Spec.Dataset,
		Namespace: object.GetNamespace(),
	}, nil
}

func (r *DataBackupReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      utils.GetDataBackupReleaseName(object.GetName()),
	}
}

// UpdateOperationStatus update the DataBackup Status
func (r *DataBackupReconciler) UpdateOperationApiStatus(object client.Object, opStatus *v1alpha1.OperationStatus) error {
	dataBackup, ok := object.(*v1alpha1.DataBackup)
	if !ok {
		return fmt.Errorf("%+v is not a type of DataBackup", object)
	}
	var dataBackupCpy = dataBackup.DeepCopy()
	dataBackupCpy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.Background(), dataBackupCpy)
}
