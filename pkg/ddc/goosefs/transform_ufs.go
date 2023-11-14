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
