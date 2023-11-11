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

package jindofsx

import (
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// shouldMountUFS checks if there's any UFS that need to be mounted
func (e *JindoFSxEngine) shouldMountUFS() (should bool, err error) {
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
func (e *JindoFSxEngine) mountUFS() (err error) {
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
