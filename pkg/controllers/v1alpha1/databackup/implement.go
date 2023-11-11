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
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

type dataBackupOperation struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder

	// object for reconciler
	dataBackup *datav1alpha1.DataBackup
}

var _ dataoperation.OperationInterface = &dataBackupOperation{}

func (r *dataBackupOperation) GetOperationObject() client.Object {
	return r.dataBackup
}

func (r *dataBackupOperation) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdatabackup.DatabackupChart
}

func (r *dataBackupOperation) HasPrecedingOperation() bool {
	return r.dataBackup.Spec.RunAfter != nil
}

func (r *dataBackupOperation) UpdateStatusInfoForCompleted(infos map[string]string) error {
	dataBackup := r.dataBackup

	infos[cdatabackup.BackupLocationPath] = dataBackup.Spec.BackupPath

	if strings.HasPrefix(dataBackup.Spec.BackupPath, common.PathScheme.String()) {
		podName := dataBackup.GetName() + "-pod"
		backupPod, err := kubeclient.GetPodByName(r.Client, podName, dataBackup.GetNamespace())
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

func (r *dataBackupOperation) Validate(ctx runtime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	dataBackup := r.dataBackup

	// 0. check the supported backup path format
	if !strings.HasPrefix(dataBackup.Spec.BackupPath, common.PathScheme.String()) && !strings.HasPrefix(dataBackup.Spec.BackupPath, common.VolumeScheme.String()) {
		err := fmt.Errorf("don't support path in this form, path: %s", dataBackup.Spec.BackupPath)
		return []datav1alpha1.Condition{
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

func (r *dataBackupOperation) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
}

func (r *dataBackupOperation) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
}

func (r *dataBackupOperation) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataBackupType
}

func (r *dataBackupOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return utils.GetDataset(r.Client, r.dataBackup.Spec.Dataset, r.dataBackup.Namespace)
}

func (r *dataBackupOperation) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.dataBackup.GetNamespace(),
		Name:      utils.GetDataBackupReleaseName(r.dataBackup.GetName()),
	}
}

// UpdateOperationApiStatus update the DataBackup Status
func (r *dataBackupOperation) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	var dataBackupCpy = r.dataBackup.DeepCopy()
	dataBackupCpy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.Background(), dataBackupCpy)
}

func (r *dataBackupOperation) GetStatusHandler() dataoperation.StatusHandler {
	return &OnceHandler{}
}

// GetTTL implements dataoperation.OperationInterface.
func (r *dataBackupOperation) GetTTL() (ttl *int32, err error) {
	ttl = r.dataBackup.Spec.TTLSecondsAfterFinished
	return
}
