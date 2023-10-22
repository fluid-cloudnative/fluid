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
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
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

type dataLoadReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder

	dataLoad *datav1alpha1.DataLoad
}

var _ dataoperation.OperationReconcilerInterface = &dataLoadReconciler{}

func (r *dataLoadReconciler) GetObject() client.Object {
	return r.dataLoad
}

func (r *dataLoadReconciler) GetTargetDataset() (*datav1alpha1.Dataset, error) {
	return utils.GetDataset(r.Client, r.dataLoad.Spec.Dataset.Name, r.dataLoad.Spec.Dataset.Namespace)
}

func (r *dataLoadReconciler) GetReleaseNameSpacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: r.dataLoad.GetNamespace(),
		Name:      utils.GetDataLoadReleaseName(r.dataLoad.GetName()),
	}
}

func (r *dataLoadReconciler) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdataload.DataloadChart
}

func (r *dataLoadReconciler) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataLoadType
}

func (r *dataLoadReconciler) UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error {
	var dataLoadCpy = r.dataLoad.DeepCopy()
	dataLoadCpy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.Background(), dataLoadCpy)
}

func (r *dataLoadReconciler) Validate(ctx cruntime.ReconcileRequestContext) ([]datav1alpha1.Condition, error) {
	dataLoad := r.dataLoad

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

func (r *dataLoadReconciler) UpdateStatusInfoForCompleted(infos map[string]string) error {
	// DataLoad does not need to update OperationStatus's Infos field.
	return nil
}

func (r *dataLoadReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataLoad does not need to update Dataset other field except for DataOperationRef.
}

func (r *dataLoadReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataLoad does not need to update Dataset other field except for DataOperationRef.
}

func (r *dataLoadReconciler) GetStatusHandler() dataoperation.StatusHandler {
	policy := r.dataLoad.Spec.Policy

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
func (r *dataLoadReconciler) GetTTL() (ttl *int32, err error) {
	dataLoad := r.dataLoad

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
