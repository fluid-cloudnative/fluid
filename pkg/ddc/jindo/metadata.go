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

package jindo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/util/retry"
)

func (e *JindoEngine) SyncMetadata() (err error) {
	defer utils.TimeTrack(time.Now(), "JindoEngine.SyncMetadata", "name", e.name, "namespace", e.namespace)
	defer e.Log.V(1).Info("End to sync metadata", "name", e.name, "namespace", e.namespace)
	e.Log.V(1).Info("Start to sync metadata", "name", e.name, "namespace", e.namespace)
	should, err := e.shouldSyncMetadata()
	if err != nil {
		e.Log.Error(err, "Failed to check if should sync metadata")
		return
	}
	// should sync metadata
	if should {
		// load metadata again
		return e.syncMetadataInternal()
	}
	return
}

// shouldSyncMetadata checks dataset's UfsTotal to decide whether should sync metadata
func (e *JindoEngine) shouldSyncMetadata() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		should = false
		return should, err
	}

	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != METADATA_SYNC_NOT_DONE_MSG {
		e.Log.V(1).Info("dataset ufs is ready",
			"dataset name", dataset.Name,
			"dataset namespace", dataset.Namespace,
			"ufstotal", dataset.Status.UfsTotal)
		should = false
		return should, nil
	}
	should = true
	return should, nil
}

func (e *JindoEngine) syncMetadataInternal() (err error) {
	if e.MetadataSyncDoneCh != nil {
		// Either get result from channel or timeout
		select {
		case result, ok := <-e.MetadataSyncDoneCh:
			defer func() {
				e.MetadataSyncDoneCh = nil
			}()
			if !ok {
				e.Log.Info("Get empty result from a closed MetadataSyncDoneCh")
				return
			}
			e.Log.Info("Get result from MetadataSyncDoneCh", "result", result)
			if result.Done {
				e.Log.Info("Metadata sync succeeded", "period", time.Since(result.StartTime))
				err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
					dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
					if err != nil {
						return
					}
					datasetToUpdate := dataset.DeepCopy()
					datasetToUpdate.Status.UfsTotal = result.UfsTotal
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
			} else {
				e.Log.Error(result.Err, "Metadata sync failed")
				return result.Err
			}
		case <-time.After(CHECK_METADATA_SYNC_DONE_TIMEOUT_MILLISEC * time.Millisecond):
			e.Log.V(1).Info("Metadata sync still in progress")
		}
	} else {
		// Metadata sync haven't started
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
			if err != nil {
				return
			}
			datasetToUpdate := dataset.DeepCopy()
			datasetToUpdate.Status.UfsTotal = METADATA_SYNC_NOT_DONE_MSG
			datasetToUpdate.Status.FileNum = METADATA_SYNC_NOT_DONE_MSG
			if !reflect.DeepEqual(dataset, datasetToUpdate) {
				err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
				if err != nil {
					return
				}
			}
			return
		})
		if err != nil {
			e.Log.Error(err, "Failed to set UfsTotal to METADATA_SYNC_NOT_DONE_MSG")
		}
		e.MetadataSyncDoneCh = make(chan base.MetadataSyncResult)
		go func(resultChan chan base.MetadataSyncResult) {
			defer base.SafeClose(resultChan)
			result := base.MetadataSyncResult{
				StartTime: time.Now(),
				UfsTotal:  "",
			}

			if err != nil {
				e.Log.Error(err, "Can't get dataset when syncing metadata", "name", e.name, "namespace", e.namespace)
				result.Err = err
				result.Done = false
				if closed := base.SafeSend(resultChan, result); closed {
					e.Log.Info("Recover from sending result to a closed channel", "result", result)
				}
				return
			}

			/*e.Log.Info("Metadata Sync starts", "dataset namespace", e.namespace, "dataset name", e.name)

			podName, containerName := e.getMasterPodInfo()
			fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
			// load metadata
			err = fileUtils.LoadMetadataWithoutTimeout("/")
			if err != nil {
				e.Log.Error(err, "LoadMetadata failed when syncing metadata", "name", e.name, "namespace", e.namespace)
				result.Err = err
				result.Done = false
				resultChan <- result
				return
			}*/
			result.Done = true

			useStsSecret := false
			if len(e.runtime.Spec.Secret) != 0 {
				useStsSecret = true
			}
			datasetUFSTotalBytes, err := e.TotalJindoStorageBytes(useStsSecret)
			if err != nil {
				e.Log.Error(err, "Get Ufs Total size failed when syncing metadata", "name", e.name, "namespace", e.namespace)
				result.Done = false
			} else {
				result.UfsTotal = utils.BytesSize(float64(datasetUFSTotalBytes))
			}

			if !result.Done {
				result.Err = errors.New("GetMetadataInfoFailed")
			} else {
				result.Err = nil
			}
			if closed := base.SafeSend(resultChan, result); closed {
				e.Log.Info("Recover from sending result to a closed channel", "result", result)
			}
		}(e.MetadataSyncDoneCh)
	}
	return
}
