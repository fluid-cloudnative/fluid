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

package goosefs

import (
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// transform dataset which has ufsPaths and ufsVolumes
func (e *GooseFSEngine) transformDatasetToVolume(runtime *datav1alpha1.GooseFSRuntime, dataset *datav1alpha1.Dataset, value *GooseFS) {

	mounts := dataset.Spec.Mounts
	for _, mount := range mounts {
		// if mount.MountPoint
		if strings.HasPrefix(mount.MountPoint, common.PathScheme.String()) {
			if len(value.UFSPaths) == 0 {
				value.UFSPaths = []UFSPath{}
			}

			ufsPath := UFSPath{}
			ufsPath.Name = mount.Name
			ufsPath.ContainerPath = utils.UFSPathBuilder{}.GenLocalStoragePath(mount)
			ufsPath.HostPath = strings.TrimPrefix(mount.MountPoint, common.PathScheme.String())
			value.UFSPaths = append(value.UFSPaths, ufsPath)

		} else if strings.HasPrefix(mount.MountPoint, common.VolumeScheme.String()) {
			if len(value.UFSVolumes) == 0 {
				value.UFSVolumes = []UFSVolume{}
			}

			// Split MountPoint into PVC name and subpath (if it contains a subpath)
			parts := strings.SplitN(strings.TrimPrefix(mount.MountPoint, common.VolumeScheme.String()), "/", 2)

			if len(parts) > 1 {
				// MountPoint contains subpath
				value.UFSVolumes = append(value.UFSVolumes, UFSVolume{
					Name:          parts[0],
					SubPath:       parts[1],
					ContainerPath: utils.UFSPathBuilder{}.GenLocalStoragePath(mount),
				})
			} else {
				// MountPoint does not contain subpath
				value.UFSVolumes = append(value.UFSVolumes, UFSVolume{
					Name:          parts[0],
					ContainerPath: utils.UFSPathBuilder{}.GenLocalStoragePath(mount),
				})
			}
		}
	}

	if len(value.UFSPaths) > 0 {
		// fmt.Println("UFSPaths length 1")
		if dataset.Spec.NodeAffinity != nil {
			value.Master.Affinity = Affinity{
				NodeAffinity: translateCacheToNodeAffinity(dataset.Spec.NodeAffinity),
			}
		}
	}

}
