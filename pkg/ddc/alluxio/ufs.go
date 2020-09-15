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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"strings"

	units "github.com/docker/go-units"

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

func (e *AlluxioEngine) preparePersistVolumeClaimSubdir(mount datav1alpha1.Mount, fileUtils operations.AlluxioFileUtils) (err error) {
	// When all the mounts are pvc with subdirectory specified,
	// alluxio's local storage will not be created.
	// This will cause error when later creating soft link.
	// Thus we create it regardless of whether it exists.
	if err = fileUtils.LocalMkdir(e.getLocalStorageDirectory()); err != nil {
		return
	}
	_, subdir, basename, ok := e.parsePersistVolumeClaimSubdir(mount.MountPoint)
	if !ok {
		return fmt.Errorf("failed to parse subdirectory of PVC")
	}
	target := fmt.Sprintf("%s/%s/%s", e.getPersistVolumeClaimDirectory(), mount.Name, subdir)
	linkName := fmt.Sprintf("%s/%s", e.getLocalStorageDirectory(), basename)
	if err = fileUtils.SoftLink(target, linkName); err != nil {
		return
	}
	if err = fileUtils.SyncLocalDir(linkName); err != nil {
		return
	}
	return
}

// Prepare the mounts and metadata
func (e *AlluxioEngine) PrepareUFS() (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	if dataset.Status.UfsTotal != "" {
		e.Log.V(1).Info("dataset ufs is ready",
			"dataset name", dataset.Name,
			"dataset namespace", dataset.Namespace,
			"ufstotal", dataset.Status.UfsTotal)
		return nil
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		return fmt.Errorf("The UFS is not ready")
	}

	//1. make mount
	for _, mount := range dataset.Spec.Mounts {
		if e.isFluidNativeScheme(mount.MountPoint) {
			if e.containsPersistVolumeClaimSubdir(mount.MountPoint) {
				// create soft link
				err = e.preparePersistVolumeClaimSubdir(mount, fileUitls)
				if err != nil {
					return
				}
				continue
			} else {
				err = fileUitls.SyncLocalDir(fmt.Sprintf("%s/%s", e.getLocalStorageDirectory(), mount.Name))
				if err != nil {
					return
				}
				continue
			}
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

	//2. load the metadata
	err = fileUitls.LoadMetaData("/", true)
	if err != nil {
		return
	}

	//3. update the status of dataset
	datasetToUpdate := dataset.DeepCopy()
	datasetUFSTotalBytes, err := e.TotalStorageBytes()
	if err != nil {
		return
	}
	ufsTotal := units.BytesSize(float64(datasetUFSTotalBytes))
	e.Log.Info("UFS storage", "storage", ufsTotal)
	datasetToUpdate.Status.UfsTotal = ufsTotal
	err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
	if err != nil {
		e.Log.Error(err, "Failed to update the ufs of the dataset")
		return err
	}

	return
}

// du the ufs
func (e *AlluxioEngine) du() (ufs int64, cached int64, cachedPercentage string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewAlluxioFileUtils(podName, containerName, e.namespace, e.Log)
	return fileUitls.Du("/")
}

// ShouldCheckUFS checks if it requires checking UFS
func (e *AlluxioEngine) ShouldCheckUFS() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	e.Log.V(1).Info("get dataset info", "dataset", dataset)
	if err != nil {
		return
	}

	// TODO(iluoeli): this will cause error if UfsTotal is stale and not properly cleaned.
	// maybe we should check mounts other than UfsTotal
	if dataset.Status.UfsTotal == "" {
		should = true
	}

	return

}

func (e *AlluxioEngine) PrepareUFSForWorkers() (err error) {
	workerName := e.getWorkerDaemonsetName()
	pods, err := e.getRunningPodsOfDaemonset(workerName, e.namespace)
	if err != nil {
		return err
	}
	e.Log.V(1).Info("PrepareUFSForWorkers", "Running worker pods are", pods)
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return
	}
	for _, pod := range pods {
		podName := pod.Name
		mounts := dataset.Spec.Mounts
		for _, mount := range mounts {
			if strings.HasPrefix(mount.MountPoint, volumeScheme) &&
				e.containsPersistVolumeClaimSubdir(mount.MountPoint) {
				e.Log.Info("Create soft link", "PodName", podName)
				fileUtils := operations.NewAlluxioFileUtils(podName, e.getWorkerContainerName(), e.namespace, e.Log)
				err = e.preparePersistVolumeClaimSubdir(mount, fileUtils)
				if err != nil {
					return
				}
			}
		}
	}
	return
}
