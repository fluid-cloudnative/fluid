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

package jindocache

import (
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindocache/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// shouldMountUFS checks if there's any UFS that need to be mounted
func (e *JindoCacheEngine) shouldMountUFS() (should bool, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return should, err
	}
	e.Log.Info("get dataset info", "dataset", dataset)

	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUtils.Ready()
	if !ready {
		should = false
		err = fmt.Errorf("the UFS is not ready")
		return should, err
	}

	// Check if any of the Mounts has not been mounted in Alluxio
	for _, mount := range dataset.Spec.Mounts {
		mounted, err := fileUtils.IsMounted("/" + mount.Name)
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

	return should, err
}

// mountUFS() mount all UFSs to Alluxio according to mount points in `dataset.Spec`. If a mount point is Fluid-native, mountUFS() will skip it.
func (e *JindoCacheEngine) mountUFS() (err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return err
	}

	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		return fmt.Errorf("the UFS is not ready")
	}

	// Iterate all the mount points, do mount if the mount point is not Fluid-native(e.g. Hostpath or PVC)
	for _, mount := range dataset.Spec.Mounts {

		// first to check the path isMounted
		mounted := false
		if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			ufsVolumesPath := utils.UFSPathBuilder{}.GenLocalStoragePath(mount)
			mount.MountPoint = "local://" + ufsVolumesPath
		}
		if !mounted {
			if mount.Path != "" {
				err = fileUitls.Mount(mount.Path, mount.MountPoint)
				if err != nil {
					return err
				}
				continue
			}
			err = fileUitls.Mount(mount.Name, mount.MountPoint)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (e *JindoCacheEngine) ShouldRefreshCacheSet() (shouldRefresh bool, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		shouldRefresh = false
		return shouldRefresh, fmt.Errorf("the UFS is not ready")
	}

	refreshed, err := fileUitls.IsRefreshed()
	if err != nil {
		return
	}
	return !refreshed, err
}

func (e *JindoCacheEngine) RefreshCacheSet() (err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUitls := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)

	ready := fileUitls.Ready()
	if !ready {
		return fmt.Errorf("the UFS is not ready")
	}

	err = fileUitls.RefreshCacheSet()
	return
}
