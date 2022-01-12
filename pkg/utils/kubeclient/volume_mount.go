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

package kubeclient

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func PVCNames(volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) (pvcNames []string) {
	return pvcNamesFromVolumes(volumeNamesFromMounts(volumeMounts),
		volumes)
}

func volumeNamesFromMounts(volumeMounts []corev1.VolumeMount) (volumeNames []string) {
	volumeNameMap := map[string]bool{}

	for _, volumeMount := range volumeMounts {
		name := volumeMount.Name
		if len(name) > 0 {
			if !volumeNameMap[name] {
				volumeNameMap[name] = true
			}
		}
	}

	volumeNames = []string{}
	for key := range volumeNameMap {
		volumeNames = append(volumeNames, key)
	}

	return

}

// pvcNamesFromVolumes gets the pvcNames from names of volumeMounts and volumes
func pvcNamesFromVolumes(knownVolumeNames []string, volumes []corev1.Volume) (pvcNames []string) {
	vMap := map[string]corev1.Volume{}
	for _, v := range volumes {
		vMap[v.Name] = v
	}

	for _, name := range knownVolumeNames {
		if volume, found := vMap[name]; found {
			if volume.PersistentVolumeClaim != nil && len(volume.PersistentVolumeClaim.ClaimName) > 0 {
				pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
			}
		} else {
			log.Info("Not able to find volume by name", "name", name, "volume", volumes)
		}
	}

	return
}

func GetFuseMountInContainer(mountType string, volumeMounts []corev1.VolumeMount) (volumeMount corev1.VolumeMount, err error) {
	kv := map[string]string{
		common.JindoMountType:     common.JindoChartName,
		common.JindoRuntime:       common.JindoChartName,
		common.ALLUXIO_MOUNT_TYPE: common.ALLUXIO_CHART,
		common.ALLUXIO_RUNTIME:    common.ALLUXIO_CHART,
		common.GooseFSMountType:   common.GooseFSChart,
	}

	volumeMountName := ""
	switch mountType {
	case common.JuiceFSMountType, common.JuiceFSRuntime:
		volumeMountName = "jfs-dir"
	default:
		if prefix, found := kv[mountType]; found {
			volumeMountName = prefix + "-fuse-mount"
		} else {
			err = fmt.Errorf("failed to find the prefix by mountType %s", mountType)
			return
		}
	}

	for _, vm := range volumeMounts {
		if vm.Name == volumeMountName {
			volumeMount = vm
			return
		}
	}

	err = fmt.Errorf("failed to find the volumeMount from slice %v by the name %s", volumeMounts, volumeMountName)
	return
}
