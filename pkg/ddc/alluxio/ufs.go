/*

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

package alluxio

import (
	"context"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (e *AlluxioEngine) UsedStorageBytes() (value int64, err error) {
	// return e.usedStorageBytesInternal()
	return e.usedStorageBytesInternal()
}

func (e *AlluxioEngine) FreeStorageBytes() (value int64, err error) {
	// return e.freeStorageBytesInternal()
	return e.freeStorageBytesInternal()
}
func (e *AlluxioEngine) TotalStorageBytes() (value int64, err error) {
	// return e.totalStorageBytesInternal()
	return e.totalStorageBytesInternal()
}
func (e *AlluxioEngine) TotalFileNums() (value int64, err error) {
	// return e.totalFileNumsInternal()
	return e.totalFileNumsInternal()
}

// ShouldCheckUFS checks if it requires checking UFS
func (e *AlluxioEngine) ShouldCheckUFS() (should bool, err error) {
	return !e.UFSChecked, nil
}

// PrepareUFS do all the UFS preparations
func (e *AlluxioEngine) PrepareUFS() (err error) {
	// 1. Mount UFS (Synchronous Operation)
	shouldMountUfs, err := e.shouldMountUFS()
	if err != nil {
		e.Log.Error(err, "Failed to check if should mount ufs")
		return err
	}

	if shouldMountUfs {
		err := e.mountUFS()
		if err != nil {
			e.Log.Error(err, "Fail to mount UFS")
			return err
		}
	}

	// 2. Initialize UFS(Asynchronous Operation), including:
	// 	 2.1 Load metadata of all files in mounted ufs
	//   2.2 Get total size of all files in mounted ufs
	go func() {
		shouldInitializeUFS, err := e.shouldInitializeUFS()
		if err != nil {
			e.Log.Error(err, "Can't check if should initialize UFS")
			return
		}
		if shouldInitializeUFS {
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
				dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
				if err != nil {
					return
				}
				datasetToUpdate := dataset.DeepCopy()
				datasetToUpdate.Status.UfsTotal = UFS_INIT_NOT_DONE_MSG
				if !reflect.DeepEqual(dataset, datasetToUpdate) {
					err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
					if err != nil {
						return
					}
				}
				return
			})
			if err != nil {
				e.Log.Error(err, "Failed to set UfsTotal to UFS_INIT_NOT_DONE_MSG")
			}
			for {
				// If anything unexpected happened, will retry after 20s.
				time.Sleep(20 * time.Second)
				err = e.initializeUFS()
				if err != nil {
					e.Log.Error(err, "Can't initialize dataset")
				}
				shouldInitializeUFS, err = e.shouldInitializeUFS()
				if err != nil {
					if apierrs.IsNotFound(err) {
						e.Log.Info("Can't find dataset when checking if should initialize UFS, abort UFS Initialization")
						break
					}
					e.Log.Error(err, "Can't check if should initialize dataset")
					continue
				}
				if !shouldInitializeUFS {
					break
				}
			}
		}
	}()
	// UFS preparation successfully done, no need to retry
	e.UFSChecked = true
	return
}

// report alluxio summary
func (e *AlluxioEngine) reportSummary() (summary string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUtils.ReportSummary()
}

// du the ufs
func (e *AlluxioEngine) du() (ufs int64, cached int64, cachedPercentage string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUitls.Du("/")
}
