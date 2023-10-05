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

package dataprocess

import (
	"context"
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cdataprocess "github.com/fluid-cloudnative/fluid/pkg/dataprocess"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *DataProcessReconciler) GetTargetDataset(object client.Object) (*datav1alpha1.Dataset, error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		return nil, fmt.Errorf("object %v is not of type DataProcess", object)
	}

	return utils.GetDataset(r.Client, dataProcess.Spec.Dataset.Name, dataProcess.Spec.Dataset.Namespace)
}

// GetReleaseNameSpacedName get the installed helm chart name
func (r *DataProcessReconciler) GetReleaseNameSpacedName(object client.Object) types.NamespacedName {
	return types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      utils.GetDataProcessReleaseName(object.GetName()),
	}
}

// GetChartsDirectory get the helm charts directory of data operation
func (r *DataProcessReconciler) GetChartsDirectory() string {
	return utils.GetChartsDirectory() + "/" + cdataprocess.DataProcessChart
}

// GetOperationType get the data operation type
func (r *DataProcessReconciler) GetOperationType() datav1alpha1.OperationType {
	return datav1alpha1.DataProcessType
}

// UpdateOperationApiStatus update the data operation status, object is the data operation crd instance.
func (r *DataProcessReconciler) UpdateOperationApiStatus(object client.Object, opStatus *datav1alpha1.OperationStatus) error {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		return fmt.Errorf("%+v is not of type DataProcess", object)
	}

	var dataProcessCopy = dataProcess.DeepCopy()
	dataProcessCopy.Status = *opStatus.DeepCopy()
	return r.Status().Update(context.TODO(), dataProcessCopy)
}

// Validate check the data operation spec is valid or not, if not valid return error with conditions
func (r *DataProcessReconciler) Validate(ctx runtime.ReconcileRequestContext, object client.Object) ([]datav1alpha1.Condition, error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		return []datav1alpha1.Condition{}, fmt.Errorf("object %v is not of type DataProcess", object)
	}

	// DataProcess's targetDataset must be in the same namespace.
	if dataProcess.Namespace != dataProcess.Spec.Dataset.Namespace {
		r.Recorder.Eventf(dataProcess,
			corev1.EventTypeWarning,
			common.TargetDatasetNamespaceNotSame,
			"DataProcess(%s)'s namespace is not same as its spec.dataset",
			dataProcess.Name,
		)
		err := fmt.Errorf("DataProcess(%s/%s)'s namespace is not same as its spec.dataset", dataProcess.Namespace, dataProcess.Name)

		now := time.Now()
		return []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             corev1.ConditionTrue,
				Reason:             common.TargetDatasetNamespaceNotSame,
				Message:            "DataProcess's namespace is not same as its spec.dataset",
				LastProbeTime:      metav1.NewTime(now),
				LastTransitionTime: metav1.NewTime(now),
			},
		}, err
	}

	// DataProcess should not have unset processor
	if dataProcess.Spec.Processor.Job == nil && dataProcess.Spec.Processor.Script == nil {
		r.Recorder.Eventf(dataProcess,
			corev1.EventTypeWarning,
			common.DataProcessProcessorNotSpecified,
			"DataProcess(%s)'s processor is missing",
			dataProcess.Name,
		)
		err := fmt.Errorf("DataProcess(%s/%s)'s spec.processor is not specified", dataProcess.Namespace, dataProcess.Name)

		now := time.Now()
		return []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             corev1.ConditionTrue,
				Reason:             common.DataProcessProcessorNotSpecified,
				Message:            "DataProcess's spec.processor is not specified",
				LastProbeTime:      metav1.NewTime(now),
				LastTransitionTime: metav1.NewTime(now),
			},
		}, err
	}

	// DataProcess should only set exactly one processor
	if dataProcess.Spec.Processor.Job != nil && dataProcess.Spec.Processor.Script != nil {
		r.Recorder.Eventf(dataProcess,
			corev1.EventTypeWarning,
			common.DataProcessMultipleProcessorSpecified,
			"DataProcess(%s)'s has specified multiple processors, only one processor is allowed",
			dataProcess.Name,
		)
		err := fmt.Errorf("DataProcess(%s/%s) has specified multiple processors", dataProcess.Namespace, dataProcess.Name)

		now := time.Now()
		return []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             corev1.ConditionTrue,
				Reason:             common.DataProcessMultipleProcessorSpecified,
				Message:            "DataProcess has specified multiple processors",
				LastProbeTime:      metav1.NewTime(now),
				LastTransitionTime: metav1.NewTime(now),
			},
		}, err
	}

	processorImpl := cdataprocess.GetProcessorImpl(dataProcess)
	if processorImpl == nil {
		return []datav1alpha1.Condition{}, fmt.Errorf("neither jobProcessor or scriptProcessor is set for DataProcess (%s/%s)", dataProcess.Namespace, dataProcess.Name)
	}

	// DataProcess should not have conflict volume mountPath with targetDataset's mountPath
	if ok, conflictVolName, conflictCtrName := processorImpl.ValidateDatasetMountPath(dataProcess.Spec.Dataset.MountPath); !ok {
		r.Recorder.Eventf(dataProcess,
			corev1.EventTypeWarning,
			common.DataProcessConflictMountPath,
			"DataProcess(%s) has conflict mountPath (%s) with spec.dataset.mountPath for volume %s mounting on container %s",
			dataProcess.Name,
			dataProcess.Spec.Dataset.MountPath,
			conflictVolName,
			conflictCtrName,
		)

		err := fmt.Errorf("DataProcess(%s/%s) has conflict mountPath (%s) with spec.dataset.mountPath for volume %s mounting on container %s",
			dataProcess.Namespace,
			dataProcess.Name,
			dataProcess.Spec.Dataset.MountPath,
			conflictVolName,
			conflictCtrName,
		)

		now := time.Now()
		return []datav1alpha1.Condition{
			{
				Type:               common.Failed,
				Status:             corev1.ConditionTrue,
				Reason:             common.DataProcessConflictMountPath,
				Message:            "DataProcess has conflict volume mount paths",
				LastProbeTime:      metav1.NewTime(now),
				LastTransitionTime: metav1.NewTime(now),
			},
		}, err
	}

	return nil, nil
}

// UpdateStatusInfoForCompleted update the status infos field for phase completed, the parameter infos is not nil
func (r *DataProcessReconciler) UpdateStatusInfoForCompleted(object client.Object, infos map[string]string) error {
	return nil
}

// SetTargetDatasetStatusInProgress set the dataset status for certain field when data operation executing.
func (r *DataProcessReconciler) SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataProcess does not need to update Dataset status before execution.
}

// RemoveTargetDatasetStatusInProgress remove the dataset status for certain field when data operation finished.
func (r *DataProcessReconciler) RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset) {
	// DataProcess does not need to recover Dataset status after execution.
}

func (r *DataProcessReconciler) GetStatusHandler(object client.Object) dataoperation.StatusHandler {
	// TODO: Support dataProcess.Spec.Policy
	return &OnceStatusHandler{Client: r.Client}
}

// GetTTL implements dataoperation.OperationReconcilerInterface.
func (*DataProcessReconciler) GetTTL(object client.Object) (ttl *int32, err error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		err = fmt.Errorf("%+v is not a type of DataProcess", object)
		return
	}

	ttl = dataProcess.Spec.TTLSecondsAfterFinished
	return
}
