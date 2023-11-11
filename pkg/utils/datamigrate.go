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

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// GetDataMigrate gets the DataMigrate given its name and namespace
func GetDataMigrate(client client.Client, name, namespace string) (*datav1alpha1.DataMigrate, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var dataMigrate datav1alpha1.DataMigrate
	if err := client.Get(context.TODO(), key, &dataMigrate); err != nil {
		return nil, err
	}
	return &dataMigrate, nil
}

// GetDataMigrateReleaseName returns DataMigrate helm release's name given the DataMigrate's name
func GetDataMigrateReleaseName(name string) string {
	return fmt.Sprintf("%s-migrate", name)
}

// GetDataMigrateJobName returns DataMigrate job(or cronjob)'s name given the DataMigrate helm release's name
func GetDataMigrateJobName(releaseName string) string {
	return fmt.Sprintf("%s-migrate", releaseName)
}

func GetTargetDatasetOfMigrate(client client.Client, dataMigrate *datav1alpha1.DataMigrate) (targetDataset *datav1alpha1.Dataset, err error) {
	var fromDataset, toDataset *datav1alpha1.Dataset
	var boundedRuntimeType = ""
	if dataMigrate.Spec.To.DataSet != nil && dataMigrate.Spec.To.DataSet.Name != "" {
		toDataset, err = GetDataset(client, dataMigrate.Spec.To.DataSet.Name, dataMigrate.Spec.To.DataSet.Namespace)
		if err != nil {
			return
		}

		// if runtimeType is not specified, we will use the toDataset as the targetDataset
		if dataMigrate.Spec.RuntimeType == "" {
			targetDataset = toDataset
			return
		}

		// if runtimeType is specified, check if toDataset's accelerate runtime type is the same as the runtimeType
		index, boundedRuntime := GetRuntimeByCategory(toDataset.Status.Runtimes, common.AccelerateCategory)
		if index == -1 {
			err = fmt.Errorf("bounded accelerate runtime not ready")
			return
		}
		if boundedRuntime.Type == dataMigrate.Spec.RuntimeType {
			targetDataset = toDataset
			return
		}
		boundedRuntimeType = boundedRuntime.Type
	}
	if dataMigrate.Spec.From.DataSet != nil && dataMigrate.Spec.From.DataSet.Name != "" {
		fromDataset, err = GetDataset(client, dataMigrate.Spec.From.DataSet.Name, dataMigrate.Spec.From.DataSet.Namespace)
		if err != nil {
			return
		}
		// if runtimeType is not specified, we will use the fromDataset as the targetDataset
		if dataMigrate.Spec.RuntimeType == "" {
			targetDataset = fromDataset
			return
		}

		// if runtimeType is specified, check if fromDataset's accelerate runtime type is the same as the runtimeType
		index, boundedRuntime := GetRuntimeByCategory(fromDataset.Status.Runtimes, common.AccelerateCategory)
		if index == -1 {
			err = fmt.Errorf("bounded accelerate runtime not ready")
			return
		}
		if boundedRuntime.Type == dataMigrate.Spec.RuntimeType {
			targetDataset = fromDataset
			return
		}
		boundedRuntimeType = boundedRuntime.Type
	}

	// DataMigrate has from/to dataset, but Spec.RuntimeType is different with target dataset' bounded runtime type;
	if boundedRuntimeType != "" {
		err = fmt.Errorf("the runtime type of the target dataset is %s, but the runtime type of the dataMigrate is %s",
			boundedRuntimeType, dataMigrate.Spec.RuntimeType)
		return nil, errors.Wrap(err, "Unable to get ddc runtime")
	}

	// DataMigrate has no from/to dataset
	return nil, apierrors.NewBadRequest("datamigrate should specify from or to dataset")
}
