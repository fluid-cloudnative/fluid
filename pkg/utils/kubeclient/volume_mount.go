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
	"path"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func PVCNames(volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) (pvcNames []string) {
	return pvcNamesFromVolumes(volumeNamesFromMounts(volumeMounts),
		volumes)
}

// volumeNamesFromMounts gets all the volume names refered by given volumeMounts
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

// pvcNamesFromVolumes gets the pvcNames from existing volume names and volume specs
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

func GetFuseMountInContainer(mountType string, container corev1.Container) (volumeMount corev1.VolumeMount, err error) {
	kv := map[string]string{
		common.JindoMountType:   common.JindoChartName,
		common.JindoRuntime:     common.JindoChartName,
		common.AlluxioMountType: common.AlluxioChart,
		common.AlluxioRuntime:   common.AlluxioChart,
		common.GooseFSMountType: common.GooseFSChart,
		common.JuiceFSMountType: common.JuiceFSChart,
		common.JuiceFSRuntime:   common.JuiceFSChart,
	}

	volumeMountName := ""
	if prefix, found := kv[mountType]; found {
		volumeMountName = prefix + "-fuse-mount"
	} else {
		for _, vm := range container.VolumeMounts {
			if vm.Name == common.ThinMountType {
				volumeMountName = common.ThinMountType
				break
			}
		}
	}
	if len(volumeMountName) == 0 {
		err = fmt.Errorf("failed to find the prefix by mountType %s", mountType)
		return
	}

	for _, vm := range container.VolumeMounts {
		if vm.Name == volumeMountName {
			volumeMount = vm
			return
		}
	}

	err = fmt.Errorf("failed to find the volumeMount from slice %v by the name %s", container.VolumeMounts, volumeMountName)
	return
}

func GetMountPathInContainer(container corev1.Container) (string, error) {
	kv := map[string]string{
		common.JindoChartName: "jindofs-fuse",
		common.AlluxioChart:   "alluxio-fuse",
		common.GooseFSChart:   "goosefs-fuse",
		common.JuiceFSChart:   "juicefs-fuse",
	}
	// consider the env FLUID_FUSE_MOUNTPOINT
	if len(container.Env) > 0 {
		for _, env := range container.Env {
			if env.Name == common.FuseMountEnv {
				return env.Value, nil
			}
		}
	}
	for _, vm := range container.VolumeMounts {
		if strings.HasSuffix(vm.Name, "-fuse-mount") {
			mountType := vm.Name[:len(vm.Name)-11]
			volumePathSuffix := ""
			if suffix, found := kv[mountType]; found {
				volumePathSuffix = suffix
			} else {
				return "", fmt.Errorf("failed to find the suffix by mountType %s", mountType)
			}
			return path.Join(vm.MountPath, volumePathSuffix), nil
		}
	}
	return "", fmt.Errorf("failed to find fluid fuse mount path in container")
}
