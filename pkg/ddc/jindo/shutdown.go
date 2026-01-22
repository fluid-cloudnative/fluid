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

package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/pkg/errors"

	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// Shutdown shuts down the Jindo engine
func (e *JindoEngine) Shutdown() (err error) {

	if e.retryShutdown < e.gracefulShutdownLimits {
		err = e.invokeCleanCache()
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed",
				// "engine", e,
				"retry times", e.retryShutdown)
			return
		}
	}

	err = e.destroyWorkers()
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

// destroyMaster destroys the master
func (e *JindoEngine) destroyMaster() (err error) {
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

func (e *JindoEngine) releasePorts() (err error) {
	var valueConfigMapname = e.name + "-jindofs-config"

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		return errors.Wrap(err, "GetRuntimePortAllocator when releasePorts")
	}

	cm, err := kubeclient.GetConfigmapByName(e.Client, valueConfigMapname, e.namespace)
	if err != nil {
		return errors.Wrap(err, "GetConfigmapByName when releasePorts")
	}

	// The value configMap is not found
	if cm == nil {
		e.Log.Info("value configMap not found, there might be some unreleased ports", "valueConfigMapName", valueConfigMapname)
		return nil
	}

	portsToRelease, err := parsePortsFromConfigMap(cm)
	if err != nil {
		return errors.Wrap(err, "parsePortsFromConfigMap when releasePorts")
	}

	allocator.ReleaseReservedPorts(portsToRelease)
	return nil
}

// cleanAll cleans up the all
func (e *JindoEngine) cleanAll() (err error) {
	count, err := e.Helper.CleanUpFuse()
	if err != nil {
		e.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	e.Log.Info("clean up fuse count", "n", count)

	err = e.cleanConfigMap()
	if err != nil {
		e.Log.Error(err, "Err in cleaning configMap")
		return err
	}
	return
}

// cleanConfigmap cleans up the configmaps, such as:
// {dataset name}-jindo-values, {dataset name}-jindofs-client-config, {dataset name}-jindofs-config
func (e *JindoEngine) cleanConfigMap() (err error) {
	var (
		valueConfigmapName  = e.getHelmValuesConfigmapName()
		configmapName       = e.name + "-" + RuntimeFSType + "-config"
		clientConfigmapName = e.name + "-" + RuntimeFSType + "-client-config"
		namespace           = e.namespace
	)

	cms := []string{valueConfigmapName, configmapName, clientConfigmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(e.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// destroyWorkers will delete the workers by number of the workers, if workers is -1, it means all the workers are deleted
func (e *JindoEngine) destroyWorkers() (err error) {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	return e.Helper.TearDownWorkers(runtimeInfo)
}
