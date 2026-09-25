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
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// tieredStoreVolumeNamePrefix prefixes every volume TransformRuntimeTieredStore
// generates, so the quota charged to a workload can later be recovered from it.
const tieredStoreVolumeNamePrefix = "tiered-store-level-"

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
	container.Resources = withTieredStoreMemoryQuota(container.Resources, totalQuota)

	// add an memory emptyDir for /dev/shm in the container
	volumeName := getMemoryTieredStoreVolumeName(levelIndex)
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
		volumeName := getTieredStoreVolumeName(levelIndex, i)
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

	volumeName := getTieredStoreVolumeName(levelIndex, 0)
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
		container.Resources = withTieredStoreMemoryQuota(container.Resources, quota)
	}

	return nil
}

// chargedTieredStoreMemoryQuota sums the tiered store memory quota already charged
// to a workload's container memory, recovering it from the tmpfs volumes the
// creation path wrote. It covers exactly the levels TransformRuntimeTieredStore
// charges -- process memory and Memory-medium emptyDir -- both of which carry the
// quota as the volume's size limit.
//
// The workload is authoritative rather than the CacheRuntime spec, because
// tieredStore is not a supported update field. Recomputing the quota from an edited
// spec would move the container's memory while the tmpfs volumes it must cover keep
// the size they were created with.
func chargedTieredStoreMemoryQuota(podSpec *corev1.PodSpec) resource.Quantity {
	total := *resource.NewQuantity(0, resource.BinarySI)
	for _, volume := range podSpec.Volumes {
		if !strings.HasPrefix(volume.Name, tieredStoreVolumeNamePrefix) {
			continue
		}
		emptyDir := volume.EmptyDir
		if emptyDir == nil || emptyDir.Medium != corev1.StorageMediumMemory || emptyDir.SizeLimit == nil {
			continue
		}
		total.Add(*emptyDir.SizeLimit)
	}
	return total
}

// withTieredStoreMemoryQuota returns base with quota added to its memory request
// and limit. Absent or zero constraints are left untouched, so a container
// without memory constraints stays unconstrained.
func withTieredStoreMemoryQuota(base corev1.ResourceRequirements, quota resource.Quantity) corev1.ResourceRequirements {
	result := *base.DeepCopy()
	if quota.IsZero() {
		return result
	}
	for _, list := range []corev1.ResourceList{result.Requests, result.Limits} {
		if list == nil {
			continue
		}
		if cur, exists := list[corev1.ResourceMemory]; exists && !cur.IsZero() {
			cur.Add(quota)
			list[corev1.ResourceMemory] = cur
		}
	}
	return result
}

// convertToLegacyTieredStore converts a CacheRuntime RuntimeTieredStore into the legacy
// TieredStore that base.BuildRuntimeInfo consumes. RuntimeInfo holds a single TieredStoreInfo
// and drives the cache capacity labels put on nodes, so it is fed from the worker tiered store
// only: labelling is keyed off nodes running worker pods (getDesiredNodesWithScheduleInfo), and
// a client tier would therefore be counted on nodes where a worker happens to co-reside and
// silently dropped everywhere else.
//
// The two APIs describe the storage medium differently. RuntimeTieredStoreLevel names it
// structurally through ProcessMemory, EmptyDir and HostPath, while the legacy Level carries an
// explicit MEM/SSD/HDD enum. Only the memory-versus-disk distinction survives, which is what the
// consumers need: tieredstore.GetLevelStorageMap buckets SSD and HDD together. Disk-backed levels
// are reported as HDD to match the medium type already written into the runtime config ConfigMap
// by extractTieredStoreLevels.
//
// The media are examined in the same order as TransformRuntimeTieredStore - memory, host path,
// empty dir - so that the two functions agree on which one wins should a level ever name more
// than the single medium the API allows.
//
// A level naming no medium is skipped rather than emitted. convertToTieredstoreInfo rejects a
// Level carrying neither Quota nor QuotaList, and that error propagates through WithTieredStore
// and BuildRuntimeInfo out of getRuntimeInfo, which would fail the whole reconcile.
func convertToLegacyTieredStore(tieredStore datav1alpha1.RuntimeTieredStore) datav1alpha1.TieredStore {
	legacyTieredStore := datav1alpha1.TieredStore{}

	for levelIndex, level := range tieredStore.Levels {
		legacyLevel := datav1alpha1.Level{
			High: level.High,
			Low:  level.Low,
		}

		switch {
		case level.ProcessMemory != nil:
			quota := level.ProcessMemory.Quota.DeepCopy()
			legacyLevel.MediumType = common.Memory
			legacyLevel.Path = GetMemoryTieredStoreMountPath(levelIndex)
			legacyLevel.Quota = &quota
		case level.HostPath != nil:
			// Paths and Quotas are parallel lists. The CRD requires both to be non-empty but
			// cannot express that they must have the same length, so guard here: a mismatch
			// would otherwise be rejected by convertToTieredstoreInfo.
			if len(level.HostPath.Paths) == 0 || len(level.HostPath.Paths) != len(level.HostPath.Quotas) {
				continue
			}
			paths := make([]string, 0, len(level.HostPath.Paths))
			quotas := make([]string, 0, len(level.HostPath.Quotas))
			for pathIndex := range level.HostPath.Paths {
				paths = append(paths, GetHostPathTieredStoreMountPath(levelIndex, pathIndex))
				quotas = append(quotas, level.HostPath.Quotas[pathIndex].String())
			}
			legacyLevel.MediumType = common.HDD
			// Per-path quotas must be kept apart: a single Quota would be divided equally over
			// the paths by convertToTieredstoreInfo and lose the declared distribution.
			legacyLevel.Path = strings.Join(paths, ",")
			legacyLevel.QuotaList = strings.Join(quotas, ",")
		case level.EmptyDir != nil:
			quota := level.EmptyDir.Quota.DeepCopy()
			legacyLevel.MediumType = common.HDD
			if level.EmptyDir.Medium == corev1.StorageMediumMemory {
				legacyLevel.MediumType = common.Memory
			}
			legacyLevel.Path = GetEmptyDirTieredStoreMountPath(levelIndex)
			legacyLevel.Quota = &quota
		default:
			continue
		}

		legacyTieredStore.Levels = append(legacyTieredStore.Levels, legacyLevel)
	}

	return legacyTieredStore
}
