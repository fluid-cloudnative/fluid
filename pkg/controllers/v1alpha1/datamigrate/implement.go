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
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cdatamigrate "github.com/fluid-cloudnative/fluid/pkg/datamigrate"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"

	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (r *DataMigrateReconciler) GetTargetDataset(object client.Object) (*datav1alpha1.Dataset, error) {
	return utils.GetTargetDatasetOfMigrate(r.Client, *object.(*datav1alpha1.DataMigrate))
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

func (r *DataMigrateReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	dataset.Status.Phase = datav1alpha1.DataMigrating
}

func (r *DataMigrateReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
}

func (r *DataMigrateReconciler) GetStatusHandler(obj client.Object) dataoperation.StatusHandler {
	dataMigrate := obj.(*datav1alpha1.DataMigrate)
	policy := dataMigrate.Spec.Policy

	switch policy {
	case datav1alpha1.Once:
		return &OnceStatusHandler{Client: r.Client, Log: r.Log}
	case datav1alpha1.Cron:
		return &CronStatusHandler{Client: r.Client, Log: r.Log}
	case datav1alpha1.OnEvent:
		return &OnEventStatusHandler{Client: r.Client, Log: r.Log}
	default:
		return nil
	}
}
