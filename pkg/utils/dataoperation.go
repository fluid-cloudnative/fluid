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
package utils

import (
	"context"
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
)

// ListDataOperationJobByCronjob gets the DataOperation(i.e. DataMigrate, DataLoad) job by cronjob given its name and namespace
func ListDataOperationJobByCronjob(c client.Client, cronjobNamespacedName types.NamespacedName) ([]batchv1.Job, error) {
	jobLabelSelector, err := labels.Parse(fmt.Sprintf("cronjob=%s", cronjobNamespacedName.Name))
	if err != nil {
		return nil, err
	}
	var jobList batchv1.JobList
	if err := c.List(context.TODO(), &jobList, &client.ListOptions{
		LabelSelector: jobLabelSelector,
		Namespace:     cronjobNamespacedName.Namespace,
	}); err != nil {
		return nil, err
	}
	return jobList.Items, nil
}

func GetOperationStatus(obj client.Object) (*datav1alpha1.OperationStatus, error) {
	if obj == nil {
		return nil, nil
	}

	if dataLoad, ok := obj.(*datav1alpha1.DataLoad); ok {
		return dataLoad.Status.DeepCopy(), nil
	} else if dataMigrate, ok := obj.(*datav1alpha1.DataMigrate); ok {
		return dataMigrate.Status.DeepCopy(), nil
	} else if dataBackup, ok := obj.(*datav1alpha1.DataBackup); ok {
		return dataBackup.Status.DeepCopy(), nil
	} else if dataProcess, ok := obj.(*datav1alpha1.DataProcess); ok {
		return dataProcess.Status.DeepCopy(), nil
	}

	return nil, fmt.Errorf("obj is not of any data operation type")
}

func GetPrecedingOperationStatus(client client.Client, opRef *datav1alpha1.OperationRef, opRefNamespace string) (*datav1alpha1.OperationStatus, error) {
	if opRef == nil {
		return nil, nil
	}

	switch opRef.Kind {
	case string(datav1alpha1.DataBackupType):
		object, err := GetDataBackup(client, opRef.Name, opRefNamespace)
		if err != nil {
			return nil, err
		}
		return &object.Status, nil
	case string(datav1alpha1.DataLoadType):
		object, err := GetDataLoad(client, opRef.Name, opRefNamespace)
		if err != nil {
			return nil, err
		}
		return &object.Status, nil
	case string(datav1alpha1.DataMigrateType):
		object, err := GetDataMigrate(client, opRef.Name, opRefNamespace)
		if err != nil {
			return nil, err
		}
		return &object.Status, nil
	case string(datav1alpha1.DataProcessType):
		object, err := GetDataProcess(client, opRef.Name, opRefNamespace)
		if err != nil {
			return nil, err
		}
		return &object.Status, nil
	default:
		// TODO: Support non-builtin Kind
		return nil, fmt.Errorf("kind %v is currently not supported for runAfter", opRef.Kind)
	}
}

func NeedCleanUp(opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) bool {
	if len(opStatus.Conditions) == 0 {
		// data operation has no completion time, no need to clean up
		return false
	}
	ttl, err := operation.GetTTL()
	if err != nil {
		return false
	}
	return ttl != nil
}

// Timeleft return not nil remaining time if data operation has completion time and set ttlAfterFinished
func Timeleft(opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (*time.Duration, error) {
	if len(opStatus.Conditions) == 0 {
		// data operation has no completion time
		return nil, nil
	}
	if opStatus.Conditions[0].Type != common.Complete && opStatus.Conditions[0].Type != common.Failed {
		// only clean up complete or failed data operation
		return nil, nil
	}

	ttl, err := operation.GetTTL()
	if err != nil {
		return nil, err
	}
	if ttl == nil {
		// if data operation does not set ttlAfterFinished, return nil
		return nil, nil
	}

	curTime := time.Now()
	// completionTime := opStatus.Conditions[0].LastProbeTime
	completionTime := opStatus.Conditions[0].LastTransitionTime
	expireTime := completionTime.Add(time.Duration(*ttl) * time.Second)
	// calculate remainint time to clean up data operation
	remaining := expireTime.Sub(curTime)
	return &remaining, nil
}
