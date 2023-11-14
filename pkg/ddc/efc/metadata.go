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
