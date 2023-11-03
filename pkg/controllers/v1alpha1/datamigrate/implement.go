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
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
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

type dataMigrateOperation struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder

	dataMigrate *datav1alpha1.DataMigrate
}

var _ dataoperation.OperationInterface = &dataMigrateOperation{}

func (r *dataMigrateOperation) GetOperationObject() client.Object {
	return r.dataMigrate
}

func (r *dataMigrateOperation) HasPrecedingOperation() bool {
	return r.dataMigrate.Spec.RunAfter != nil
}

func (r *dataMigrateOperation) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return utils.GetTargetDatasetOfMigrate(r.Client, r.dataMigrate)
}

func (r *dataMigrateOperation) GetReleaseNameSpacedName() types.NamespacedName {
	releaseName := utils.GetDataMigrateReleaseName(r.dataMigrate.GetName())
	return types.NamespacedName{
		Namespace: r.dataMigrate.GetNamespace(),
		Name:      releaseName,
	}
}

func (r *dataMigrateOperation) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdatamigrate.DataMigrateChart
}

func (r *dataMigrateOperation) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataMigrateType
}

func (r *dataMigrateOperation) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	var dataMigrateCpy = r.dataMigrate.DeepCopy()
	dataMigrateCpy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.Background(), dataMigrateCpy)
}

func (r *dataMigrateOperation) Validate(ctx cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	targetDataSet := ctx.Dataset

	if r.dataMigrate.GetNamespace() != targetDataSet.Namespace {
		err := fmt.Errorf("DataMigrate(%s) namespace is not same as dataset", r.dataMigrate.GetName())
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

func (r *dataMigrateOperation) UpdateStatusInfoForCompleted(infos map[string]string) error {
	// DataMigrate does not need to update OperationStatus's Infos field.
	return nil
}

func (r *dataMigrateOperation) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	dataset.Status.Phase = datav1alpha1.DataMigrating
}

func (r *dataMigrateOperation) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	dataset.Status.Phase = datav1alpha1.BoundDatasetPhase
}

func (r *dataMigrateOperation) GetStatusHandler() dataoperation.StatusHandler {
	policy := r.dataMigrate.Spec.Policy

	switch policy {
	case datav1alpha1.Once:
		return &OnceStatusHandler{Client: r.Client, Log: r.Log, dataMigrate: r.dataMigrate}
	case datav1alpha1.Cron:
		return &CronStatusHandler{Client: r.Client, Log: r.Log, dataMigrate: r.dataMigrate}
	case datav1alpha1.OnEvent:
		return &OnEventStatusHandler{Client: r.Client, Log: r.Log, dataMigrate: r.dataMigrate}
	default:
		return nil
	}
}

// GetTTL implements dataoperation.OperationInterface.
func (r *dataMigrateOperation) GetTTL() (ttl *int32, err error) {
	dataMigrate := r.dataMigrate

	policy := dataMigrate.Spec.Policy
	switch policy {
	case datav1alpha1.Once:
		ttl = dataMigrate.Spec.TTLSecondsAfterFinished
	case datav1alpha1.Cron, datav1alpha1.OnEvent:
		// For Cron and OnEvent policies, no TTL is provided
		ttl = nil
	default:
		err = fmt.Errorf("unknown policy type: %s", policy)
	}

	return
}
