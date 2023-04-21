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

package dataload

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DataLoadReconciler) GetTargetDataset(object client.Object) (*datav1alpha1.Dataset, error) {
	typeObject, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		return nil, fmt.Errorf("object %v is not a DataLoad", object)
	}

	dataLoad := *typeObject

	return utils.GetDataset(r.Client, dataLoad.Spec.Dataset.Name, dataLoad.Spec.Dataset.Namespace)
}

func (r *DataLoadReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      utils.GetDataLoadReleaseName(object.GetName()),
	}
}

func (r *DataLoadReconciler) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdataload.DataloadChart
}

func (r *DataLoadReconciler) GetOperationType() dataoperation.OperationType {
	return dataoperation.DataLoad
}

func (r *DataLoadReconciler) UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	dataLoad, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		return fmt.Errorf("%+v is not a type of DataLoad", object)
	}
	var dataLoadCpy = dataLoad.DeepCopy()
	dataLoadCpy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.Background(), dataLoadCpy)
}

func (r *DataLoadReconciler) Validate(ctx cruntime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error) {
	dataLoad, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		return []datav1alpha1.Condition{}, fmt.Errorf("object %v is not a DataLoad", object)
	}

	// 1. Check dataLoad namespace and dataset namespace need to be same
	if dataLoad.Namespace != dataLoad.Spec.Dataset.Namespace {
		r.Recorder.Eventf(dataLoad,
			v1.EventTypeWarning,
			common.TargetDatasetNamespaceNotSame,
			"dataLoad(%s) namespace is not same as dataset",
			dataLoad.Name)
		err := fmt.Errorf("dataLoad(%s) namespace is not same as dataset", dataLoad.Name)

		return []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             v1.ConditionTrue,
				Reason:             common.TargetDatasetNamespaceNotSame,
				Message:            "dataLoad namespace is not same as dataset",
				LastProbeTime:      metav1.NewTime(time.Now()),
				LastTransitionTime: metav1.NewTime(time.Now()),
			},
		}, err
	}
	return nil, nil
}

func (r *DataLoadReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	// DataLoad does not need to update OperationStatus's Infos field.
	return nil
}

func (r *DataLoadReconciler) UpdateStatusByHelmStatus(ctx cruntime.ReconcileRequestContext, object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	// 2. Check running status of the DataLoad job
	releaseName := utils.GetDataLoadReleaseName(object.GetName())
	jobName := utils.GetDataLoadJobName(releaseName)

	ctx.Log.V(1).Info("DataLoad chart already existed, check its running status")
	job, err := utils.GetDataLoadJob(r.Client, jobName, ctx.Namespace)
	if err != nil {
		// helm release found but job missing, delete the helm release and requeue
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.Info("Related Job missing, will delete helm chart and retry", "namespace", ctx.Namespace, "jobName", jobName)
			if err = helm.DeleteReleaseIfExists(releaseName, ctx.Namespace); err != nil {
				ctx.Log.Error(err, "can't delete dataload release", "namespace", ctx.Namespace, "releaseName", releaseName)
				return err
			}
		}
		// other error
		ctx.Log.Error(err, "can't get dataload job", "namespace", ctx.Namespace, "jobName", jobName)
		return err
	}

	if len(job.Status.Conditions) != 0 {
		if job.Status.Conditions[0].Type == batchv1.JobFailed ||
			job.Status.Conditions[0].Type == batchv1.JobComplete {
			// job either failed or complete, update DataLoad's phase status
			jobCondition := job.Status.Conditions[0]

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
	ctx.Log.V(1).Info("DataLoad job still running", "namespace", ctx.Namespace, "jobName", jobName)
	return nil
}

func (r *DataLoadReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataLoad does not need to update Dataset other field except for DataOperationRef.
}

func (r *DataLoadReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataLoad does not need to update Dataset other field except for DataOperationRef.
}
