/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j JuiceFSEngine) UsedStorageBytes() (int64, error) {
	panic("implement me")
}

func (j JuiceFSEngine) FreeStorageBytes() (int64, error) {
	panic("implement me")
}

func (j JuiceFSEngine) TotalStorageBytes() (int64, error) {
	return j.totalStorageBytesInternal()
}

func (j JuiceFSEngine) TotalFileNums() (int64, error) {
	return j.totalFileNumsInternal()
}

func (j JuiceFSEngine) ShouldCheckUFS() (should bool, err error) {
	return false, nil
}

func (j JuiceFSEngine) PrepareUFS() (err error) {
	//// 1. Mount UFS (Synchronous Operation)
	//shouldMountUfs, err := j.shouldMountUFS()
	//if err != nil {
	//	return
	//}
	//j.Log.Info("shouldMountUFS", "should", shouldMountUfs)
	//
	//if shouldMountUfs {
	//	err = j.mountUFS()
	//	if err != nil {
	//		return
	//	}
	//}
	//j.Log.Info("mountUFS")

	return
}

// shouldMountUFS checks if there's any UFS that need to be mounted
func (j *JuiceFSEngine) shouldMountUFS() (should bool, err error) {
	//dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
	//j.Log.Info("get dataset info", "dataset", dataset)
	//if err != nil {
	//	return should, err
	//}
	//
	//fuseName := j.getFuseDaemonsetName()
	//pods, err := j.getRunningPodsOfDaemonset(fuseName, j.namespace)
	//if err != nil {
	//	return should, err
	//}
	//
	//for _, pod := range pods {
	//	fileUtils := operations.NewJuiceFileUtils(pod.Name, common.JuiceFSFuseContainer, j.namespace, j.Log)
	//
	//	// Check subpath
	//	for _, mount := range dataset.Spec.Mounts {
	//		subpath := ""
	//		if mount.Path == "" {
	//			subpath = mount.Name
	//		} else {
	//			subpath = mount.Path
	//		}
	//		juiceFSSubPath := fmt.Sprintf("%s/%s", j.getMountPoint(), subpath)
	//		existed, err := fileUtils.IsExist(juiceFSSubPath)
	//		if err != nil {
	//			should = false
	//			return should, err
	//		}
	//		if !existed {
	//			j.Log.Info("Found dataset subpath not existed.", "dataset", dataset)
	//			should = true
	//			return should, err
	//		}
	//	}
	//}
	return should, err
}

// mountUFS
func (j *JuiceFSEngine) mountUFS() (err error) {
	dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
	j.Log.Info("get dataset info", "dataset", dataset)
	if err != nil {
		return err
	}

	fuseName := j.getFuseDaemonsetName()
	pods, err := j.getRunningPodsOfDaemonset(fuseName, j.namespace)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		fileUtils := operations.NewJuiceFileUtils(pod.Name, common.JuiceFSFuseContainer, j.namespace, j.Log)

		// Check subpath
		for _, mount := range dataset.Spec.Mounts {
			subpath := ""
			if mount.Path == "" {
				subpath = mount.Name
			} else {
				subpath = mount.Path
			}
			juiceFSSubPath := fmt.Sprintf("%s/%s", j.getMountPoint(), subpath)
			err := fileUtils.Mkdir(juiceFSSubPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// JuiceFSEngine hasn't support UpdateOnUFSChange
func (j JuiceFSEngine) ShouldUpdateUFS() (ufsToUpdate *utils.UFSToUpdate) {
	return nil
}

func (j JuiceFSEngine) UpdateOnUFSChange(ufsToUpdate *utils.UFSToUpdate) (ready bool, err error) {
	panic("implement me")
}
