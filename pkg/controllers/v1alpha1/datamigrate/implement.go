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
	"context"
	"fmt"
	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

// TODO(lzq) 将 返回信息变成 Dataset
func (r *DataMigrateReconciler) GetTargetDatasetNamespacedName(object client.Object) (*types.NamespacedName, error) {

	dataset, err := utils.GetTargetDatasetOfMigrate(r.Client, *object.(*datav1alpha1.DataMigrate))
	if err != nil {
		return nil, err
	}
	return &types.NamespacedName{
		Namespace: dataset.Namespace,
		Name:      dataset.Name,
	}, nil
}

func (r *DataMigrateReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	releaseName := utils.GetDataMigrateReleaseName(object.GetName())
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      releaseName,
	}
}

func (r *DataMigrateReconciler) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdatamigrate.DataMigrateChart
}

func (r *DataMigrateReconciler) GetOperationType() dataoperation.OperationType {
	return dataoperation.DataMigrate
}

func (r *DataMigrateReconciler) UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	dataMigrate, ok := object.(*datav1alpha1.DataMigrate)
	if !ok {
		return fmt.Errorf("%+v is not a type of DataMigrate", object)
	}
	var dataMigrateCpy = dataMigrate.DeepCopy()
	dataMigrateCpy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.Background(), dataMigrateCpy)
}

func (r *DataMigrateReconciler) Validate(ctx cruntime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error) {
	targetDataSet := ctx.Dataset

	if object.GetNamespace() != targetDataSet.Namespace {
		err := fmt.Errorf("DataMigrate(%s) namespace is not same as dataset", object.GetName())
		return []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             common.TargetDatasetNamespaceNotSame,
				Message:            "DataMigrate namespace is not same as dataset",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}, err
	}
	return nil, nil
}

func (r *DataMigrateReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	// DataMigrate does not need to update OperationStatus's Infos field.
	return nil
}

func (r *DataMigrateReconciler) UpdateStatusByHelmStatus(ctx cruntime.ReconcileRequestContext,
	object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	// 1. Check running status of the DataMigrate job
	releaseName := utils.GetDataMigrateReleaseName(object.GetName())
	jobName := utils.GetDataMigrateJobName(releaseName)
	job, err := utils.GetDataMigrateJob(r.Client, jobName, object.GetNamespace())

	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related Job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				r.Log.Error(err, "can't delete DataMigrate release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return err
			}
			return err
		}
		// other error
		ctx.Log.Error(err, "can't get DataMigrate job", "namespace", ctx.Namespace, "jobName", jobName)
		return err
	}

	if len(job.Status.Conditions) != 0 {
		if job.Status.Conditions[0].Type == batchv1.JobFailed ||
			job.Status.Conditions[0].Type == batchv1.JobComplete {
			jobCondition := job.Status.Conditions[0]
			// job either failed or complete, update DataMigrate's phase status
			opStatus.Conditions = []datav1alpha1.Condition{
				{
					Type:               common.ConditionType(jobCondition.Type),
					Status:             jobCondition.Status,
					Reason:             jobCondition.Reason,
					Message:            jobCondition.Message,
					LastProbeTime:      jobCondition.LastProbeTime,
					LastTransitionTime: jobCondition.LastTransitionTime,
				},
			}
			if jobCondition.Type == batchv1.JobFailed {
				opStatus.Phase = common.PhaseFailed
			} else {
				opStatus.Phase = common.PhaseComplete
			}
			opStatus.Duration = utils.CalculateDuration(object.GetCreationTimestamp().Time, jobCondition.LastTransitionTime.Time)
			return nil
		}
	}

	ctx.Log.V(1).Info("DataMigrate job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return nil
}

func (r *DataMigrateReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	dataset.Status.Phase = datav1alpha1.DataMigrating
}

func (r *DataMigrateReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
}
