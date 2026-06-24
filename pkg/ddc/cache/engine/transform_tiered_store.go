/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
)

// TransformRuntimeTieredStore transforms the tiered store configuration to worker pod spec
func (e *CacheEngine) TransformRuntimeTieredStore(tieredStore *datav1alpha1.RuntimeTieredStore, podSpec *corev1.PodSpec) error {
	if len(tieredStore.Levels) == 0 {
		return nil
	}

	if len(podSpec.Containers) == 0 {
		return fmt.Errorf("no containers found in worker pod spec")
	}

	container := &podSpec.Containers[0]

	// validate then set
	memoryLevelCount := 0
	for idx, level := range tieredStore.Levels {
		// order: memory, host path, empty. only one can be specified per level
		mediaCount := 0
		if level.ProcessMemory != nil {
			mediaCount++
			memoryLevelCount++
			if memoryLevelCount > 1 {
				return fmt.Errorf("RuntimeTieredStore should have only one ProcessMemoryMediumSource for all levels")
			}
		}

		if level.HostPath != nil {
			mediaCount++
		}

		if level.EmptyDir != nil {
			mediaCount++
		}
		if mediaCount > 1 {
			return fmt.Errorf("only one storage medium can be specified per level at index %d, but found %d", idx, mediaCount)
		}
	}

	// Process each tier level
	for idx, level := range tieredStore.Levels {
		// Process memory: add resource requests and limits
		if level.ProcessMemory != nil {
			err := e.handleProcessMemory(podSpec, container, level.ProcessMemory, idx)
			if err != nil {
				return err
			}
		}

		// Volume-based storage: create volumes and volume mounts
		if level.HostPath != nil {
			err := e.handleHostPath(podSpec, container, level.HostPath, idx)
			if err != nil {
				return err
			}
		}

		// EmptyDir: add volume and volume mount
		if level.EmptyDir != nil {
			err := e.handleEmptyDir(podSpec, container, level.EmptyDir, idx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// handleProcessMemory adds memory resources to container for process memory medium
func (e *CacheEngine) handleProcessMemory(podSpec *corev1.PodSpec, container *corev1.Container, memoryMediumSource *datav1alpha1.ProcessMemoryMediumSource, levelIndex int) error {
	if memoryMediumSource.Quota.IsZero() {
		return fmt.Errorf("process memory quota cannot be zero at level index %d", levelIndex)
	}

	// Calculate total memory quota across all paths
	totalQuota := memoryMediumSource.Quota.DeepCopy()

	// add totalQuota to memory resources only when memory is restricted.
	if container.Resources.Requests != nil {
		if currentRequest, exists := container.Resources.Requests[corev1.ResourceMemory]; exists && !currentRequest.IsZero() {
			currentRequest.Add(totalQuota)
			container.Resources.Requests[corev1.ResourceMemory] = currentRequest
		}
	}
	if container.Resources.Limits != nil {
		if currentLimit, exists := container.Resources.Limits[corev1.ResourceMemory]; exists && !currentLimit.IsZero() {
			currentLimit.Add(totalQuota)
			container.Resources.Limits[corev1.ResourceMemory] = currentLimit
		}
	}

	// add an memory emptyDir for /dev/shm in the container
	volumeName := fmt.Sprintf("tiered-store-level-%d-memory", levelIndex)
	mountPath := GetMemoryTieredStoreMountPath(levelIndex)
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium:    corev1.StorageMediumMemory,
				SizeLimit: &totalQuota,
			},
		},
	}

	// Add volume to pod spec
	podSpec.Volumes = utils.AppendOrOverrideVolume(podSpec.Volumes, volume)

	// Add volume mount to container
	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	}
	container.VolumeMounts = utils.AppendOrOverrideVolumeMounts(container.VolumeMounts, volumeMount)

	return nil
}

// handleHostPath adds volume and volume mount for volume-based medium
func (e *CacheEngine) handleHostPath(podSpec *corev1.PodSpec, container *corev1.Container,
	hostPathMediumSource *datav1alpha1.HostPathMediumSource, levelIndex int) error {

	if len(hostPathMediumSource.Paths) != len(hostPathMediumSource.Quotas) {
		return fmt.Errorf("number of paths and quotas must be equal at level index %d", levelIndex)
	}

	// Process each path and corresponding quota
	for i, hostPath := range hostPathMediumSource.Paths {
		volumeName := fmt.Sprintf("tiered-store-level-%d-index-%d", levelIndex, i)
		mountPath := GetHostPathTieredStoreMountPath(levelIndex, i)

		volume := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
					Type: hostPathMediumSource.Type,
				},
			},
		}

		// Add volume to pod spec
		podSpec.Volumes = utils.AppendOrOverrideVolume(podSpec.Volumes, volume)

		// Add volume mount to container
		volumeMount := corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		}
		container.VolumeMounts = utils.AppendOrOverrideVolumeMounts(container.VolumeMounts, volumeMount)
	}

	return nil
}

func (e *CacheEngine) handleEmptyDir(podSpec *corev1.PodSpec, container *corev1.Container,
	emptyDirMediumSource *datav1alpha1.EmptyDirMediumSource, levelIndex int) error {

	if emptyDirMediumSource.Quota.IsZero() {
		return fmt.Errorf("emptyDir quota cannot be zero for empty dir medium source at level index %d", levelIndex)
	}

	volumeName := fmt.Sprintf("tiered-store-level-%d-index-%d", levelIndex, 0)
	mountPath := GetEmptyDirTieredStoreMountPath(levelIndex)

	quota := emptyDirMediumSource.Quota.DeepCopy()
	volume := corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium:    emptyDirMediumSource.Medium,
				SizeLimit: &quota,
			},
		},
	}

	// Add volume to pod spec
	podSpec.Volumes = utils.AppendOrOverrideVolume(podSpec.Volumes, volume)

	// Add volume mount to container
	volumeMount := corev1.VolumeMount{
		Name:      volumeName,
		MountPath: mountPath,
	}
	container.VolumeMounts = utils.AppendOrOverrideVolumeMounts(container.VolumeMounts, volumeMount)

	// For Memory-backed EmptyDir (tmpfs), add quota to container memory resources
	// This ensures proper resource accounting and prevents excessive memory usage
	if emptyDirMediumSource.Medium == corev1.StorageMediumMemory {
		// Only add to resources if the container already has memory constraints
		// If no memory resources are set, the container is unconstrained and we don't need to add
		if container.Resources.Requests != nil {
			if currentRequest, exists := container.Resources.Requests[corev1.ResourceMemory]; exists && !currentRequest.IsZero() {
				currentRequest.Add(quota)
				container.Resources.Requests[corev1.ResourceMemory] = currentRequest
			}
		}
		if container.Resources.Limits != nil {
			if currentLimit, exists := container.Resources.Limits[corev1.ResourceMemory]; exists && !currentLimit.IsZero() {
				currentLimit.Add(quota)
				container.Resources.Limits[corev1.ResourceMemory] = currentLimit
			}
		}
	}

	return nil
}
