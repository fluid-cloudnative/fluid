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
	"github.com/docker/go-units"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"
)

type UFSInitResult struct {
	Done      bool
	StartTime time.Time
	UfsTotal  string
	Err       error
}

func (e *AlluxioEngine) usedStorageBytesInternal() (value int64, err error) {
	return
}

func (e *AlluxioEngine) freeStorageBytesInternal() (value int64, err error) {
	return
}

func (e *AlluxioEngine) totalStorageBytesInternal() (total int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	_, _, total, err = fileUitls.Count("/")
	if err != nil {
		return
	}

	return
}

func (e *AlluxioEngine) totalFileNumsInternal() (fileCount int64, err error) {
	podName, containerName := e.getMasterPodInfo()

	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	fileCount, _, _, err = fileUitls.Count("/")
	if err != nil {
		return
	}

	return
}

// shouldMountUFS check if there's any UFS that need to be mounted
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

// mountUFS() mount all UFSs to Alluxio according to mount points in `dataset.Spec`. If a mount point is Fluid-native, mountUFS() will skip it.
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

	// Iterate all the mount points, do mount if the mount point is not Fluid-native(e.g. Hostpath or PVC)
	for _, mount := range dataset.Spec.Mounts {
		if e.isFluidNativeScheme(mount.MountPoint) {
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

//shouldInitializeUFS checks
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

// initializeUFS do UFS initialization. For Alluxio Engine, it includes the following steps:
// 1. Load Metadata
// 2. Get total size of all the UFSs
// At any time, there is only one goroutine actually doing UFS initialization. Multiple calls
func (e *AlluxioEngine) initializeUFS() (err error) {

	UFSInitDoneCh := make(chan UFSInitResult)
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
	}(UFSInitDoneCh)

	result := <-UFSInitDoneCh
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
			if !reflect.DeepEqual(datasetToUpdate, dataset) {
				err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
				if err != nil {
					return
				}
			}
			return
		})
		if err != nil {
			e.Log.Error(err, "Failed to update UfsTotal of the dataset")
			return result.Err
		}
	} else {
		e.Log.Error(result.Err, "ufs init failed")
		return result.Err
	}
	return nil
}
