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

package thin

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/thin/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/util/retry"
)

func (t *ThinEngine) SyncMetadata() (err error) {
	should, err := t.shouldSyncMetadata()
	if err != nil {
		t.Log.Error(err, "Failed to check if should sync metadata")
		return
	}
	// should sync metadata
	if should {
		return t.syncMetadataInternal()
	}
	return
}

// shouldSyncMetadata checks dataset's UfsTotal to decide whether should sync metadata
func (t *ThinEngine) shouldSyncMetadata() (should bool, err error) {
	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		should = false
		return should, err
	}

	runtime, err := utils.GetThinRuntime(t.Client, t.name, t.namespace)
	if err != nil {
		should = false
		return should, err
	}

	if !runtime.Spec.RuntimeManagement.MetadataSyncPolicy.AutoSyncEnabled() {
		t.Log.V(1).Info("Skip syncing metadta cause runtime.Spec.RuntimeManagement.MetadataSyncPolicy.AutoSync=false", "runtime name", runtime.Name, "runtime namespace", runtime.Namespace)
		should = false
		return should, nil
	}

	//todo(xuzhihao): option to enable/disable automatic metadata sync
	//todo: periodical metadata sync
	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != MetadataSyncNotDoneMsg {
		t.Log.V(1).Info("dataset ufs is ready",
			"dataset name", dataset.Name,
			"dataset namespace", dataset.Namespace,
			"ufstotal", dataset.Status.UfsTotal)
		should = false
		return should, nil
	}
	should = true
	return should, nil
}

// syncMetadataInternal does the actual work of metadata sync
// At any time, there is at most one goroutine working on metadata sync. First call to
// this function will start a goroutine including the following two steps:
//  1. load metadata
//  2. get total size of UFSs
//
// Any following calls to this function will try to get result of the working goroutine with a timeout, which
// ensures the function won't block the following Sync operations(e.g. CheckAndUpdateRuntimeStatus) for a long time.
func (t *ThinEngine) syncMetadataInternal() (err error) {
	if t.MetadataSyncDoneCh != nil {
		// Either get result from channel or timeout
		select {
		case result, ok := <-t.MetadataSyncDoneCh:
			defer func() {
				t.MetadataSyncDoneCh = nil
			}()
			if !ok {
				t.Log.Info("Get empty result from a closed MetadataSyncDoneCh")
				return
			}
			t.Log.Info("Get result from MetadataSyncDoneCh", "result", result)
			if result.Done {
				t.Log.Info("Metadata sync succeeded", "period", time.Since(result.StartTime))
				err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
					dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
					if err != nil {
						return
					}
					datasetToUpdate := dataset.DeepCopy()
					datasetToUpdate.Status.UfsTotal = result.UfsTotal
					datasetToUpdate.Status.FileNum = result.FileNum
					if !reflect.DeepEqual(datasetToUpdate, dataset) {
						err = t.Client.Status().Update(context.TODO(), datasetToUpdate)
						if err != nil {
							return
						}
						// Update dataset metrics after a suceessful status update
						base.RecordDatasetMetrics(result, datasetToUpdate.Namespace, datasetToUpdate.Name, t.Log)
					}
					return
				})
				if err != nil {
					t.Log.Error(err, "Failed to update UfsTotal and FileNum of the dataset")
					return err
				}
			} else {
				t.Log.Error(result.Err, "Metadata sync failed")
				return result.Err
			}
		case <-time.After(CheckMetadataSyncDoneTimeoutMillisec * time.Millisecond):
			t.Log.V(1).Info("Metadata sync still in progress")
		}
	} else {
		// Metadata sync haven't started
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
			if err != nil {
				return
			}
			datasetToUpdate := dataset.DeepCopy()
			datasetToUpdate.Status.UfsTotal = MetadataSyncNotDoneMsg
			datasetToUpdate.Status.FileNum = MetadataSyncNotDoneMsg
			if !reflect.DeepEqual(dataset, datasetToUpdate) {
				err = t.Client.Status().Update(context.TODO(), datasetToUpdate)
				if err != nil {
					return
				}
			}
			return
		})
		if err != nil {
			t.Log.Error(err, "Failed to set UfsTotal to METADATA_SYNC_NOT_DONE_MSG")
		}
		t.MetadataSyncDoneCh = make(chan base.MetadataSyncResult)
		go func(resultChan chan base.MetadataSyncResult) {
			defer base.SafeClose(resultChan)
			result := base.MetadataSyncResult{
				StartTime: time.Now(),
				UfsTotal:  "",
			}
			_, err := utils.GetDataset(t.Client, t.name, t.namespace)
			if err != nil {
				t.Log.Error(err, "Can't get dataset when syncing metadata", "name", t.name, "namespace", t.namespace)
				result.Err = err
				result.Done = false
				if closed := base.SafeSend(resultChan, result); closed {
					t.Log.Info("Recover from sending result to a closed channel", "result", result)
				}
				return
			}

			t.Log.Info("Metadata Sync starts", "dataset namespace", t.namespace, "dataset name", t.name)

			stsName := t.getFuseDaemonsetName()
			pods, err := t.GetRunningPodsOfDaemonset(stsName, t.namespace)
			if err != nil || len(pods) == 0 {
				result.UfsTotal = ""
				result.FileNum = ""
				result.Done = true
				if closed := base.SafeSend(resultChan, result); closed {
					t.Log.Info("Recover from sending result to a closed channel", "result", result)
				}
				return
			}
			for _, pod := range pods {
				fileUtils := operations.NewThinFileUtils(pod.Name, common.ThinFuseContainer, t.namespace, t.Log)

				// load metadata
				// ls -al /runtime-mnt/thin/namespace/name/thin-fuse/
				err = fileUtils.LoadMetadataWithoutTimeout(t.getTargetPath())
				if err != nil {
					t.Log.Error(err, "LoadMetadata failed when syncing metadata", "name", t.name, "namespace", t.namespace)
					result.Err = err
					result.Done = false
					if closed := base.SafeSend(resultChan, result); closed {
						t.Log.Info("Recover from sending result to a closed channel", "result", result)
					}
					return
				}

			}

			result.Done = true

			datasetUFSTotalBytes, err := t.TotalStorageBytes()
			if err != nil {
				t.Log.Error(err, "Get Ufs Total size failed when syncing metadata", "name", t.name, "namespace", t.namespace)
				result.Done = false
			} else {
				result.UfsTotal = utils.BytesSize(float64(datasetUFSTotalBytes))
			}
			fileNum, err := t.getDataSetFileNum()
			if err != nil {
				t.Log.Error(err, "Get File Num failed when syncing metadata", "name", t.name, "namespace", t.namespace)
				result.Done = false
			} else {
				result.FileNum = fileNum
			}

			if !result.Done {
				result.Err = errors.New("GetMetadataInfoFailed")
			} else {
				result.Err = nil
			}
			if closed := base.SafeSend(resultChan, result); closed {
				t.Log.Info("Recover from sending result to a closed channel", "result", result)
			}
		}(t.MetadataSyncDoneCh)
	}

	return
}
