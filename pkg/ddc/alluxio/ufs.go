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
	"fmt"
	units "github.com/docker/go-units"
	"k8s.io/client-go/util/retry"
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

func (e *AlluxioEngine) ShouldCheckUFS() (should bool, err error) {
	// PrepareUFS() should be called exactly once
	// todo: use mutex to protect it from race condition
	if !e.UFSChecked {
		e.UFSChecked = true
		return true, nil
	}
	return false, nil
}

// Prepare the mounts and metadata
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
			e.Log.Error(err, "Can't check if should initialize dataset")
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
				err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
				if err != nil {
					return
				}
				return
			})
			if err != nil {
				e.Log.Error(err, "Failed to update UfsTotal to default Value")
			}
		}
		for shouldInitializeUFS {
			err = e.initializeUFS()
			if err != nil {
				e.Log.Error(err, "Can't initialize dataset")
			}
			// Check if ufs initialization is done every ten seconds
			time.Sleep(10 * time.Second)
			shouldInitializeUFS, err = e.shouldInitializeUFS()
			if err != nil {
				e.Log.Error(err, "Can't check if should initialize dataset")
				break
			}
		}

	}()

	return
}

// ShouldCheckUFS checks if it requires checking UFS
func (e *AlluxioEngine) shouldMountUFS() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.V(1).Info("get dataset info", "dataset", dataset)
	if err != nil {
		return should, err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		should = false
		err = fmt.Errorf("The UFS is not ready")
		return should, err
	}

	// Check if any of the Mounts has not been mounted in Alluxio
	for _, mount := range dataset.Spec.Mounts {
		if e.isFluidNativeScheme(mount.MountPoint) {
			// No need for a mount point with Fluid native scheme('local://' and 'pvc://') to be mounted
			continue
		}
		alluxioPath := fmt.Sprintf("/%s", mount.Name)
		mounted, err := fileUitls.IsMounted(alluxioPath)
		if err != nil {
			should = false
			return should, err
		}
		if !mounted {
			e.Log.Info("Found dataset that is not mounted.", "dataset", dataset)
			should = true
			return should, err
		}
	}

	// maybe we should check mounts other than UfsTotal
	//if dataset.Status.UfsTotal == "" {
	//	should = true
	//}

	return should, err

}

func (e *AlluxioEngine) mountUFS() (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		return fmt.Errorf("The UFS is not ready")
	}

	// make mount
	for _, mount := range dataset.Spec.Mounts {
		if e.isFluidNativeScheme(mount.MountPoint) {
			//err = fileUitls.SyncLocalDir(fmt.Sprintf("%s/%s", e.getLocalStorageDirectory(), mount.Name))
			//if err != nil {
			//	return
			//}
			continue
		}

		alluxioPath := fmt.Sprintf("/%s", mount.Name)
		mounted, err := fileUitls.IsMounted(alluxioPath)
		e.Log.Info("Check if the alluxio path is mounted.", "alluxioPath", alluxioPath, "mounted", mounted)
		if err != nil {
			return err
		}

		if !mounted {
			err = fileUitls.Mount(alluxioPath, mount.MountPoint, mount.Options, mount.ReadOnly, mount.Shared)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (e *AlluxioEngine) shouldInitializeUFS() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		should = false
		return should, err
	}

	//todo(trafalgarzzz) configurable knob for load metadata
	if dataset.Status.UfsTotal != "" && dataset.Status.UfsTotal != UFS_INIT_NOT_DONE_MSG {
		e.Log.V(1).Info("dataset ufs is ready",
			"dataset name", dataset.Name,
			"dataset namespace", dataset.Namespace,
			"ufstotal", dataset.Status.UfsTotal)
		should = false
		return should, err
	}
	should = true
	return should, err
}

func (e *AlluxioEngine) initializeUFS() (err error) {
	if e.UFSInitDoneCh != nil {
		// UFS init in progress
		select {
		case result := <-e.UFSInitDoneCh:
			defer func() { e.UFSInitDoneCh = nil }()
			e.Log.V(1).Info("get result from ufsInitDoneCh", "result", result)
			if result.Done {
				e.Log.Info("ufs init done", "period", time.Since(result.StartTime))
				// update `dataset.Status.UfsTotal`
				err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
					dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
					if err != nil {
						return
					}
					datasetToUpdate := dataset.DeepCopy()
					datasetToUpdate.Status.UfsTotal = result.UfsTotal
					err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
					if err != nil {
						return
					}
					return
				})
				if err != nil {
					e.Log.Error(err, "Failed to update UfsTotal of the dataset")
				}
			} else {
				e.Log.Error(result.Err, "ufs init failed")
			}
		case <-time.After(500 * time.Millisecond):
			e.Log.V(1).Info("ufs init still in progress")
		}
	} else {
		// UFS init not started
		e.UFSInitDoneCh = make(chan UFSInitResult)
		go func(resultChan chan UFSInitResult) {
			result := UFSInitResult{
				StartTime: time.Now(),
				UfsTotal:  "",
			}
			e.Log.Info("UFS init starts", "dataset namespace", e.namespace, "dataset name", e.name)

			podName, containerName := e.getMasterPodInfo()
			fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
			// 2.1 load metadata
			err = fileUitls.LoadMetadataWithoutTimeout("/", true)
			if err != nil {
				result.Err = err
				result.Done = false
				resultChan <- result
				return
			}
			// 2.2 get total size
			datasetUFSTotalBytes, err := e.TotalStorageBytes()
			if err != nil {
				result.Err = err
				result.Done = false
				resultChan <- result
				return
			}
			ufsTotal := units.BytesSize(float64(datasetUFSTotalBytes))
			result.Err = nil
			result.UfsTotal = ufsTotal
			result.Done = true
			resultChan <- result
		}(e.UFSInitDoneCh)
	}
	return nil
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
