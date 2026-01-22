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

package efc

import (
	"fmt"
	"path/filepath"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/efc/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
)

// Shutdown shuts down the EFC engine
func (e *EFCEngine) Shutdown() (err error) {
	if e.retryShutdown < e.gracefulShutdownLimits {
		err = e.cleanupCache()
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed",
				// "engine", e,
				"retry times", e.retryShutdown)
			return
		}
	}

	_, err = e.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = e.releasePorts()
	if err != nil {
		return
	}

	err = e.destroyMaster()
	if err != nil {
		return
	}

	err = e.cleanAll()
	return err
}

// cleanupCache cleans up the cache
func (e *EFCEngine) cleanupCache() (err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}
	e.Log.Info("get runtime info", "runtime", runtime)

	configMapName := e.getHelmValuesConfigMapName()
	configMap, err := kubeclient.GetConfigmapByName(e.Client, configMapName, runtime.Namespace)
	if err != nil {
		return errors.Wrap(err, "GetConfigMapByName fail when cleanupCache")
	}

	// The value configMap is not found
	if configMap == nil {
		e.Log.Info("value configMap not found, there might be some uncleaned up cache", "valueConfigMapName", configMapName)
		return nil
	}

	cacheDir, cacheType, err := parseCacheDirFromConfigMap(configMap)
	if err != nil {
		return errors.Wrap(err, "parseCacheDirFromConfigMap fail when cleanupCache")
	}

	if cacheType == common.VolumeTypeEmptyDir {
		e.Log.Info("cache in emptyDir, skip clean up cache")
		return
	}

	workerPods, err := e.getWorkerRunningPods()
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			e.Log.Info(fmt.Sprintf("worker of runtime %s namespace %s has been shutdown.", runtime.Name, runtime.Namespace))
			return nil
		} else {
			return err
		}
	}

	for _, pod := range workerPods {
		fileUtils := operations.NewEFCFileUtils(pod.Name, "efc-worker", e.namespace, e.Log)

		e.Log.Info("Remove cache in worker pod", "pod", pod.Name, "cache", cacheDir)
		cacheDirToBeDeleted := filepath.Join(cacheDir, "tier_dadi")
		err := fileUtils.DeleteDir(cacheDirToBeDeleted)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EFCEngine) releasePorts() (err error) {
	var valueConfigMapName = e.getHelmValuesConfigMapName()

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		return errors.Wrap(err, "GetRuntimePortAllocator when releasePorts")
	}

	cm, err := kubeclient.GetConfigmapByName(e.Client, valueConfigMapName, e.namespace)
	if err != nil {
		return errors.Wrap(err, "GetConfigmapByName when releasePorts")
	}

	// The value configMap is not found
	if cm == nil {
		e.Log.Info("value configMap not found, there might be some unreleased ports", "valueConfigMapName", valueConfigMapName)
		return nil
	}

	portsToRelease, err := parsePortsFromConfigMap(cm)
	if err != nil {
		return errors.Wrap(err, "parsePortsFromConfigMap when releasePorts")
	}

	allocator.ReleaseReservedPorts(portsToRelease)
	return nil
}

// destroyMaster destroys the master
func (e *EFCEngine) destroyMaster() (err error) {
	var found bool
	found, err = helm.CheckRelease(e.name, e.namespace)
	if err != nil {
		return err
	}

	if found {
		err = helm.DeleteRelease(e.name, e.namespace)
		if err != nil {
			return
		}
	}
	return
}

func (e *EFCEngine) cleanAll() (err error) {
	count, err := e.Helper.CleanUpFuse()
	if err != nil {
		e.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	e.Log.Info("clean up fuse count", "n", count)

	var (
		valueConfigmapName = e.getHelmValuesConfigMapName()
		namespace          = e.namespace
	)

	cms := []string{valueConfigmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(e.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// destroyWorkers attempts to delete the workers until worker num reaches the given expectedWorkers, if expectedWorkers is -1, it means all the workers should be deleted
// This func returns currentWorkers representing how many workers are left after this process.
func (e *EFCEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return currentWorkers, err
	}

	return e.Helper.TearDownWorkers(runtimeInfo)
}
