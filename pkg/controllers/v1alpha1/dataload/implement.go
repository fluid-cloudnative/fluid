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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdataload "github.com/fluid-cloudnative/fluid/pkg/dataload"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

func (r *DataLoadReconciler) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataLoadType
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

func (r *DataLoadReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataLoad does not need to update Dataset other field except for DataOperationRef.
}

func (r *DataLoadReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataLoad does not need to update Dataset other field except for DataOperationRef.
}

func (r *DataLoadReconciler) GetStatusHandler(object client.Object) dataoperation.StatusHandler {
	dataLoad := object.(*datav1alpha1.DataLoad)
	policy := dataLoad.Spec.Policy

	switch policy {
	case datav1alpha1.Once:
		return &OnceStatusHandler{Client: r.Client}
	case datav1alpha1.Cron:
		return &CronStatusHandler{Client: r.Client}
	case datav1alpha1.OnEvent:
		return &OnEventStatusHandler{Client: r.Client}
	default:
		return nil
	}
}

// GetTTL implements dataoperation.OperationReconcilerInterface.
func (*DataLoadReconciler) GetTTL(object client.Object) (ttl *int32, err error) {
	dataLoad, ok := object.(*datav1alpha1.DataLoad)
	if !ok {
		err = fmt.Errorf("%+v is not a type of DataBackup", object)
		return
	}

	policy := dataLoad.Spec.Policy
	switch policy {
	case datav1alpha1.Once:
		ttl = dataLoad.Spec.TTLSecondsAfterFinished
	case datav1alpha1.Cron, datav1alpha1.OnEvent:
		// For Cron and OnEvent policies, no TTL is provided
		ttl = nil
	default:
		err = fmt.Errorf("unknown policy type: %s", policy)
	}

	return
}
