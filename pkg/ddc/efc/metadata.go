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

package efc

import (
	"context"
	"reflect"
	"strconv"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/util/retry"
)

func (e *EFCEngine) SyncMetadata() (err error) {
	should, err := e.ShouldCheckUFS()
	if err != nil {
		e.Log.Error(err, "Failed to check if should sync metadata")
		return
	}
	if should {
		return e.syncMetadataInternal()
	}
	return
}

func (e *EFCEngine) syncMetadataInternal() (err error) {
	datasetUFSTotalBytes, err := e.TotalStorageBytes()
	if err != nil {
		e.Log.Error(err, "Failed to get UfsTotal")
		return err
	}
	ufsTotal := utils.BytesSize(float64(datasetUFSTotalBytes))

	fileCount, err := e.TotalFileNums()
	if err != nil {
		e.Log.Error(err, "Failed to get FileNum")
		return err
	}
	fileNum := strconv.FormatInt(fileCount, 10)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return
		}
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.Status.UfsTotal = ufsTotal
		datasetToUpdate.Status.FileNum = fileNum
		if !reflect.DeepEqual(datasetToUpdate, dataset) {
			err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
			if err != nil {
				return
			}
		}
		return
	})
	if err != nil {
		e.Log.Error(err, "Failed to update UfsTotal and FileNum of the dataset")
		return err
	}
	return
}
