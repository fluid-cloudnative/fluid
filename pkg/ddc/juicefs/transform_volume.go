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

package juicefs

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// transform worker volumes
func (j *JuiceFSEngine) transformWorkerVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	if len(runtime.Spec.Worker.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Worker.VolumeMounts {
			var volume *corev1.Volume

			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Worker.VolumeMounts) == 0 {
				value.Worker.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Worker.VolumeMounts = append(value.Worker.VolumeMounts, volumeMount)

			if len(value.Worker.Volumes) == 0 {
				value.Worker.Volumes = []corev1.Volume{}
			}
			value.Worker.Volumes = append(value.Worker.Volumes, *volume)
		}
	}

	return err
}

// transform worker cache volumes
// after genValue & genMount function
func (j *JuiceFSEngine) transformWorkerCacheVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS, options map[string]string) (err error) {
	cacheDir := ""

	// if cache-dir is set in worker option, it will override the cache-dir of worker in runtime
	cacheDir = options["cache-dir"]
	cacheValueMap := map[string]string{}
	for _, v := range value.CacheDirs {
		cacheValueMap[v.Path] = v.Path
	}
	// set tiredstore cache as volume also, for clear cache when shut down
	caches := MapDeepCopy(value.CacheDirs)
	index := len(caches)
	if cacheDir != "" {
		originPath := strings.Split(cacheDir, ":")
		for i, p := range originPath {
			if _, ok := cacheValueMap[p]; ok {
				continue
			}
			var volumeType = common.VolumeTypeHostPath
			caches[strconv.Itoa(index+i+1)] = cache{
				Path: p,
				Type: string(volumeType),
			}
		}
	}

	// set volumes & volumeMounts for cache
	volumeMap := map[string]corev1.VolumeMount{}
	for _, v := range runtime.Spec.Worker.VolumeMounts {
		volumeMap[v.MountPath] = v
	}
	for i, cache := range caches {
		if _, ok := volumeMap[cache.Path]; ok {
			// cache path is already in volumeMounts
			continue
		}
		value.Worker.VolumeMounts = append(value.Worker.VolumeMounts, corev1.VolumeMount{
			Name:      "cache-dir-" + i,
			MountPath: cache.Path,
		})
		v := corev1.Volume{
			Name: "cache-dir-" + i,
		}
		switch cache.Type {
		case string(common.VolumeTypeHostPath):
			dir := corev1.HostPathDirectoryOrCreate
			v.VolumeSource = corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: cache.Path,
					Type: &dir,
				},
			}
		case string(common.VolumeTypeEmptyDir):
			v.VolumeSource = corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}
			if cache.VolumeSource != nil && cache.VolumeSource.EmptyDir != nil {
				v.VolumeSource = cache.VolumeSource.VolumeSource
			}
			// todo: support volume template
		}
		value.Worker.Volumes = append(value.Worker.Volumes, v)
	}
	return err
}

// transform fuse volumes
func (j *JuiceFSEngine) transformFuseVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	if len(runtime.Spec.Fuse.VolumeMounts) > 0 {
		for _, volumeMount := range runtime.Spec.Fuse.VolumeMounts {
			var volume *corev1.Volume
			for _, v := range runtime.Spec.Volumes {
				if v.Name == volumeMount.Name {
					volume = &v
					break
				}
			}

			if volume == nil {
				return fmt.Errorf("failed to find the volume for volumeMount %s", volumeMount.Name)
			}

			if len(value.Fuse.VolumeMounts) == 0 {
				value.Fuse.VolumeMounts = []corev1.VolumeMount{}
			}
			value.Fuse.VolumeMounts = append(value.Fuse.VolumeMounts, volumeMount)

			if len(value.Fuse.Volumes) == 0 {
				value.Fuse.Volumes = []corev1.Volume{}
			}
			value.Fuse.Volumes = append(value.Fuse.Volumes, *volume)
		}
	}

	return err
}

// transform fuse cache volumes
// after genValue & genMount function
func (j *JuiceFSEngine) transformFuseCacheVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS, options map[string]string) (err error) {
	cacheDir := ""

	// if cache-dir is set in fuse option, it will override the cache-dir of worker in runtime
	cacheDir = options["cache-dir"]

	cacheValueMap := map[string]string{}
	for _, v := range value.CacheDirs {
		cacheValueMap[v.Path] = v.Path
	}
	caches := MapDeepCopy(value.CacheDirs)
	index := len(caches)
	if cacheDir != "" {
		originPath := strings.Split(cacheDir, ":")
		for i, p := range originPath {
			if _, ok := cacheValueMap[p]; ok {
				continue
			}
			var volumeType = common.VolumeTypeHostPath
			caches[strconv.Itoa(index+i+1)] = cache{
				Path: p,
				Type: string(volumeType),
			}
		}
	}
	// set volumes & volumeMounts for cache
	volumeMap := map[string]corev1.VolumeMount{}
	for _, v := range runtime.Spec.Fuse.VolumeMounts {
		volumeMap[v.MountPath] = v
	}
	for i, cache := range caches {
		if _, ok := volumeMap[cache.Path]; ok {
			// cache path is already in volumeMounts
			continue
		}
		value.Fuse.VolumeMounts = append(value.Fuse.VolumeMounts, corev1.VolumeMount{
			Name:      "cache-dir-" + i,
			MountPath: cache.Path,
		})
		v := corev1.Volume{
			Name: "cache-dir-" + i,
		}
		switch cache.Type {
		case string(common.VolumeTypeHostPath):
			dir := corev1.HostPathDirectoryOrCreate
			v.VolumeSource = corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: cache.Path,
					Type: &dir,
				},
			}
		case string(common.VolumeTypeEmptyDir):
			v.VolumeSource = corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}
			if cache.VolumeSource != nil && cache.VolumeSource.EmptyDir != nil {
				v.VolumeSource = cache.VolumeSource.VolumeSource
			}
			// todo: support volume template
		}
		value.Fuse.Volumes = append(value.Fuse.Volumes, v)
	}
	return err
}

func (j *JuiceFSEngine) transformFuseDownwardAPIVolumes(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) {
	if runtime.Spec.Fuse.CleanPolicy != datav1alpha1.OnFuseChangedCleanPolicy {
		return
	}

	volumeName := "fuse-labels-downward-api-volume"
	var mode int32 = 0755
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{
				Items: []corev1.DownwardAPIVolumeFile{
					{
						Path: utils.MetaDataFuseLabelFileName,
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.labels",
						},
					},
				},
				DefaultMode: ptr.To(mode),
			},
		},
	}
	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  true,
		MountPath: utils.GetRuntimeFuseMetadataPath(runtime.Namespace, runtime.Name, common.JuiceFSRuntime),
	}
	value.Fuse.Volumes = append(value.Fuse.Volumes, volume)
	value.Fuse.VolumeMounts = append(value.Fuse.VolumeMounts, volumeMount)
}
