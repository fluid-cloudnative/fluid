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

package dataoperation

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/runtime"
)

type OperationInterfaceBuilder interface {
	Build(object client.Object) (OperationInterface, error)
}

// OperationInterface the interface of data operation crd
type OperationInterface interface {
	// HasPrecedingOperation check if current data operation depends on another data operation
	HasPrecedingOperation() bool

	// GetOperationObject get the data operation object
	GetOperationObject() client.Object

	// GetTargetDataset get the target dataset of the data operation, implementor should return the newest target dataset.
	GetTargetDataset() (*datav1alpha1.Dataset, error)

	// GetReleaseNameSpacedName get the installed helm chart name
	GetReleaseNameSpacedName() types.NamespacedName

	// GetChartsDirectory get the helm charts directory of data operation
	GetChartsDirectory() string

	// GetOperationType get the data operation type
	GetOperationType() datav1alpha1.OperationType

	// UpdateOperationApiStatus update the data operation status, object is the data operation crd instance.
	UpdateOperationApiStatus(opStatus *datav1alpha1.OperationStatus) error

	// Validate check the data operation spec is valid or not, if not valid return error with conditions
	Validate(ctx runtime.ReconcileRequestContext) ([]datav1alpha1.Condition, error)

	// UpdateStatusInfoForCompleted update the status infos field for phase completed, the parameter infos is not nil
	UpdateStatusInfoForCompleted(infos map[string]string) error

	// SetTargetDatasetStatusInProgress set the dataset status for certain field when data operation executing.
	SetTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset)

	// RemoveTargetDatasetStatusInProgress remove the dataset status for certain field when data operation finished.
	RemoveTargetDatasetStatusInProgress(dataset *datav1alpha1.Dataset)

	GetStatusHandler() StatusHandler

	// GetTTL gets timeToLive
	GetTTL() (ttl *int32, err error)
}

type StatusHandler interface {
	// GetOperationStatus get operation status according to helm chart status
	GetOperationStatus(ctx runtime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus) (result *datav1alpha1.OperationStatus, err error)
}
