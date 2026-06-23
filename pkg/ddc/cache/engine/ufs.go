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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (e *CacheEngine) PrepareUFS(archApi ArchitectureApi) (mountOutput *datav1alpha1.CacheRuntimeMountUfsOutput, err error) {
	entries := archApi.GetExecutionEntries()
	if entries == nil || entries.MountUFS == nil {
		e.Log.Info("No mount ufs command found in runtime class")
		return nil, nil
	}
	mountUfs := entries.MountUFS

	// execute mount command in master pod
	podName, containerName, err := archApi.GetExecutionPodInfo()
	if err != nil {
		return nil, err
	}

	fileUtil := NewCacheFileUtil(podName, containerName, e.namespace, e.Log)
	// at least 20 seconds
	timeoutSeconds := max(mountUfs.TimeoutSeconds, common.MinExecutionTimeoutSeconds)
	stdout, err := fileUtil.Mount(mountUfs.Command, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		return nil, err
	}

	// parse mount output, and sync dataset mounts
	stdout = strings.TrimSpace(stdout)
	if stdout == "" {
		return nil, errors.New("mount ufs command produced empty output")
	}

	mountOutput = &datav1alpha1.CacheRuntimeMountUfsOutput{}
	err = json.Unmarshal([]byte(stdout), mountOutput)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse mount ufs output as CacheRuntimeMountUfsOutput, output: %q", stdout)
	}
	return mountOutput, nil
}

// shouldUpdateUFS determines whether the UFS configuration needs to be updated.
// It analyzes path differences between the provided dataset spec and status to identify
// which UFS entries require updates. It also checks if the master pod has restarted
// since the last mount operation.
// Returns true if either mount paths have changed or remount is required due to pod restart.
func (e *CacheEngine) shouldUpdateUFS(dataset *datav1alpha1.Dataset, archApi ArchitectureApi,
	runtime *datav1alpha1.CacheRuntime) bool {
	// whether the ufs mount paths need to update
	ufsToUpdate := utils.NewUFSToUpdate(dataset)
	ufsToUpdate.AnalyzePathsDelta()
	if ufsToUpdate != nil && ufsToUpdate.ShouldUpdate() {
		e.Log.Info("Detected UFS changes, updating mount points", "toAdd", ufsToUpdate.ToAdd(), "toRemove", ufsToUpdate.ToRemove())
		return true
	}

	// check if master pod restart after the latest mount time
	restart := e.checkIfRemountRequired(archApi, runtime)
	if restart {
		e.Log.Info("Master pod restart after the latest mount time, need to remount")
	}

	return restart
}

// UpdateOnUFSChange handles changes to the UFS configuration by updating mount points dynamically.
// When dataset mount information changes, this method ensures the changes take effect without requiring a restart.
func (e *CacheEngine) UpdateOnUFSChange(runtime *datav1alpha1.CacheRuntime) (err error) {
	runtimeClass, err := e.getRuntimeClass(runtime.Spec.RuntimeClassName)
	if err != nil {
		return
	}

	// update only when architecture supports mount ufs.
	archApi := resolveArchitectureApi(e.name, e.namespace, runtime, runtimeClass)
	if !archApi.IsMountUFSSupported() {
		return
	}

	// 1. get the dataset
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get the dataset")
		return
	}

	// update when dataset mount paths change or master pod restart after the latest mount time
	shouldUpdate := e.shouldUpdateUFS(dataset, archApi, runtime)
	if !shouldUpdate {
		return
	}

	// 2. set update status to updating
	err = utils.UpdateMountStatus(e.Client, e.name, e.namespace, datav1alpha1.UpdatingDatasetPhase)
	if err != nil {
		e.Log.Error(err, "Failed to update dataset status to updating")
		return
	}

	// 3. use the same mount script to add or remove mount points
	mountOutput, err := e.PrepareUFS(archApi)
	if err != nil {
		return errors.Wrapf(err, "failed to execute mount ufs command")
	}

	// 5. sync dataset mounts, to prevent the runtime config not updated in pods in time.
	err = e.syncDatasetMounts(dataset, mountOutput)
	if err != nil {
		e.Log.Error(err, "Failed to sync dataset mounts")
		return err
	}

	// 6. update latest mount time
	err = e.updateMountTime()
	if err != nil {
		return
	}

	return
}

// syncDatasetMounts synchronizes the dataset mount points with the current runtime state.
// This ensures that any changes to dataset.spec.mounts are reflected in the running system.
func (e *CacheEngine) syncDatasetMounts(dataset *datav1alpha1.Dataset, mountOutput *datav1alpha1.CacheRuntimeMountUfsOutput) (err error) {

	// Update MountPoints based on current dataset mounts
	var mountedPaths = map[string]bool{}
	for _, path := range mountOutput.Mounted {
		mountedPaths[path] = true
	}

	// check mounted path is the same as dataset spec mounts
	for _, mount := range dataset.Spec.Mounts {
		if common.IsFluidNativeScheme(mount.MountPoint) {
			continue
		}
		datasetMountPath := utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		if !mountedPaths[datasetMountPath] {
			e.Log.Info("Waiting for mount point to be mounted", "Mount point", datasetMountPath)
			return fmt.Errorf("mount point %s is not yet mounted", datasetMountPath)
		}
		delete(mountedPaths, datasetMountPath)
	}
	if len(mountedPaths) != 0 {
		e.Log.Info("Waiting for mount point to be unmounted", "Mount points", mountedPaths)
		return fmt.Errorf("unexpected mounted paths remain: %v", mountedPaths)
	}

	// update dataset status mount and phase status with retry
	err = utils.UpdateMountStatus(e.Client, e.name, e.namespace, datav1alpha1.BoundDatasetPhase)
	if err != nil {
		return err
	}

	return nil
}

func (e *CacheEngine) checkIfRemountRequired(archApi ArchitectureApi, runtime *datav1alpha1.CacheRuntime) bool {
	masterPodName, masterContainerName, err := archApi.GetExecutionPodInfo()
	if err != nil {
		e.Log.Error(err, "get runtime pod container name failed", "method", "checkIfRemountRequired", "runtimeClass name", e.name)
		return false
	}

	masterPod, err := kubeclient.GetPodByName(e.Client, masterPodName, e.namespace)
	if err != nil {
		e.Log.Error(err, "Got master pod failed, skip remount check", "pod name", masterPodName)
		return false
	}
	if masterPod == nil {
		e.Log.Info("Master pod not found, skip remount check", "pod name", masterPodName)
		return false
	}

	// Check pod phase to ensure it is actually running
	if masterPod.Status.Phase != corev1.PodRunning {
		e.Log.Info("Master pod is not in Running phase, skip remount check",
			"pod", masterPodName, "phase", masterPod.Status.Phase)
		return false
	}

	var startedAt *v1.Time
	for _, containerStatus := range masterPod.Status.ContainerStatuses {
		if containerStatus.Name == masterContainerName {
			if containerStatus.State.Running == nil {
				e.Log.Info("Container not running, skip remount check",
					"container", masterContainerName, "pod", masterPodName)
				return false
			}

			startedAt = &containerStatus.State.Running.StartedAt
			break
		}
	}
	if startedAt == nil {
		e.Log.Info("Cannot get container start time, skip remount check", "pod name", masterPodName,
			"container name", masterContainerName)
		return false
	}

	// If mount time is earlier than master container start time, remount is necessary
	if runtime.Status.MountTime == nil || runtime.Status.MountTime.Before(startedAt) {
		return true
	}

	return false
}
