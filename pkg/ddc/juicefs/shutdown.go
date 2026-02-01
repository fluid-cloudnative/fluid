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

package juicefs

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/juicefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (j *JuiceFSEngine) Shutdown() (err error) {
	if j.retryShutdown < j.gracefulShutdownLimits {
		err = j.cleanupCache()
		if err != nil {
			j.retryShutdown = j.retryShutdown + 1
			j.Log.Info("clean cache failed",
				"retry times", j.retryShutdown)
			return
		}
	}

	err = j.destroyWorkers()
	if err != nil {
		return
	}

	err = j.releasePorts()
	if err != nil {
		return
	}

	err = j.destroyMaster()
	if err != nil {
		return
	}

	err = j.cleanAll()
	return err
}

// destroyMaster Destroy the master
func (j *JuiceFSEngine) destroyMaster() (err error) {
	var found bool
	found, err = helm.CheckRelease(j.name, j.namespace)
	if err != nil {
		return err
	}

	if found {
		err = helm.DeleteRelease(j.name, j.namespace)
		if err != nil {
			return
		}
	} else {
		// clean residual resources
		j.Log.Info("delete residual resources")
		err = j.cleanResidualResources()
		if err != nil {
			return
		}
	}
	return
}

func (j *JuiceFSEngine) releasePorts() (err error) {
	var valueConfigMapName = j.getHelmValuesConfigMapName()

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		return errors.Wrap(err, "GetRuntimePortAllocator when releasePorts")
	}

	cm, err := kubeclient.GetConfigmapByName(j.Client, valueConfigMapName, j.namespace)
	if err != nil {
		return errors.Wrap(err, "GetConfigmapByName when releasePorts")
	}

	// The value configMap is not found
	if cm == nil {
		j.Log.Info("value configMap not found, there might be some unreleased ports", "valueConfigMapName", valueConfigMapName)
		return nil
	}

	portsToRelease, err := parsePortsFromConfigMap(cm)
	if err != nil {
		return errors.Wrap(err, "parsePortsFromConfigMap when releasePorts")
	}

	allocator.ReleaseReservedPorts(portsToRelease)
	return nil
}

// cleanupCache cleans up the cache
func (j *JuiceFSEngine) cleanupCache() (err error) {
	runtime, err := j.getRuntime()
	if err != nil {
		return err
	}
	j.Log.Info("get runtime info", "runtime", runtime)

	cacheDirs := j.getCacheDirs(runtime)
	workerName := j.getWorkerName()
	pods, err := j.GetRunningPodsOfStatefulSet(workerName, j.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) != nil {
			return err
		}
		j.Log.Info("worker of runtime %s namespace %s has been shutdown.", runtime.Name, runtime.Namespace)
	}

	var uuid string
	if len(pods) > 0 {
		uuid, err = j.getUUID(pods[0], common.JuiceFSWorkerContainer)
		if err != nil {
			return err
		}

		for _, pod := range pods {
			fileUtils := operations.NewJuiceFileUtils(pod.Name, common.JuiceFSWorkerContainer, j.namespace, j.Log)
			j.Log.Info("Remove cache in worker pod", "pod", pod.Name, "cache", cacheDirs)

			cacheDirsToBeDeleted := []string{}
			for _, cacheDir := range cacheDirs {
				cacheDirsToBeDeleted = append(cacheDirsToBeDeleted, filepath.Join(cacheDir, uuid, "raw/chunks"))
			}
			err := fileUtils.DeleteCacheDirs(cacheDirsToBeDeleted)
			if err != nil {
				return err
			}
		}
	}

	fuseName := j.getFuseName()
	fusePods, err := j.GetRunningPodsOfDaemonset(fuseName, j.namespace)
	if err != nil {
		if utils.IgnoreNotFound(err) != nil {
			return err
		}
		j.Log.Info("fuse of runtime %s namespace %s has been shutdown.", runtime.Name, runtime.Namespace)
	}

	if len(fusePods) > 0 {
		// If UUID was not found from workers (e.g. workers already down), try to get it from fuse
		if uuid == "" {
			uuid, err = j.getUUID(fusePods[0], common.JuiceFSFuseContainer)
			if err != nil {
				return err
			}
		}

		fuseCacheDirs := j.getFuseCacheDirs(runtime)
		for _, pod := range fusePods {
			fileUtils := operations.NewJuiceFileUtils(pod.Name, common.JuiceFSFuseContainer, j.namespace, j.Log)
			j.Log.Info("Remove cache in fuse pod", "pod", pod.Name, "cache", fuseCacheDirs)

			cacheDirsToBeDeleted := []string{}
			for _, cacheDir := range fuseCacheDirs {
				cacheDirsToBeDeleted = append(cacheDirsToBeDeleted, filepath.Join(cacheDir, uuid, "raw/chunks"))
			}
			err := fileUtils.DeleteCacheDirs(cacheDirsToBeDeleted)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (j *JuiceFSEngine) getCacheDirs(runtime *datav1alpha1.JuiceFSRuntime) (cacheDirs []string) {
	cacheDir := common.JuiceFSDefaultCacheDir
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		cacheDir = ""
		// if cache type hostpath, clean it
		if runtime.Spec.TieredStore.Levels[0].VolumeType == common.VolumeTypeHostPath {
			cacheDir = runtime.Spec.TieredStore.Levels[0].Path
		}
	}
	if cacheDir != "" {
		cacheDirs = strings.Split(cacheDir, ":")
	}

	// if cache-dir is set in worker option, it will override the cache-dir of worker in runtime
	workerOptions := runtime.Spec.Worker.Options
	for k, v := range workerOptions {
		if k == "cache-dir" {
			cacheDirs = append(cacheDirs, strings.Split(v, ":")...)
			break
		}
	}
	return
}

func (j *JuiceFSEngine) getFuseCacheDirs(runtime *datav1alpha1.JuiceFSRuntime) (cacheDirs []string) {
	cacheDir := common.JuiceFSDefaultCacheDir
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		cacheDir = ""
		// if cache type hostpath, clean it
		if runtime.Spec.TieredStore.Levels[0].VolumeType == common.VolumeTypeHostPath {
			cacheDir = runtime.Spec.TieredStore.Levels[0].Path
		}
	}
	if cacheDir != "" {
		cacheDirs = strings.Split(cacheDir, ":")
	}

	// if cache-dir is set in fuse option, it will override the cache-dir of fuse in runtime
	fuseOptions := runtime.Spec.Fuse.Options
	for k, v := range fuseOptions {
		if k == "cache-dir" {
			cacheDirs = append(cacheDirs, strings.Split(v, ":")...)
			break
		}
	}
	return
}

func (j *JuiceFSEngine) getUUID(pod corev1.Pod, containerName string) (uuid string, err error) {
	cm, err := j.GetValuesConfigMap()
	if err != nil {
		return
	}
	if cm == nil {
		j.Log.Info("value configMap not found")
		return
	}
	data := []byte(cm.Data["data"])
	fuseValues := make(map[string]interface{})
	err = yaml.Unmarshal(data, &fuseValues)
	if err != nil {
		return
	}

	edition := fuseValues["edition"].(string)
	source := fuseValues["source"].(string)
	if edition == EnterpriseEdition {
		uuid = source
		return
	}
	fileUtils := operations.NewJuiceFileUtils(pod.Name, containerName, j.namespace, j.Log)

	j.Log.Info("Get status in pod", "pod", pod.Name, "source", source)
	status, err := fileUtils.GetStatus(source)
	if err != nil {
		return
	}
	matchExp := regexp.MustCompile(`"UUID": "(.*)"`)
	idStr := matchExp.FindString(status)
	idStrs := strings.Split(idStr, "\"")
	if len(idStrs) < 4 {
		err = fmt.Errorf("parse uuid error, idStr: %s", idStr)
		return
	}

	uuid = idStrs[3]
	return
}

// destroyWorkers tears down all JuiceFS workers for the current runtime while holding SchedulerMutex.
// Worker and related label cleanup is delegated to Helper.TearDownWorkers.
func (j *JuiceFSEngine) destroyWorkers() (err error) {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return err
	}

	return j.Helper.TearDownWorkers(runtimeInfo)
}

func (j *JuiceFSEngine) cleanAll() (err error) {
	count, err := j.Helper.CleanUpFuse()
	if err != nil {
		j.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	j.Log.Info("clean up fuse count", "n", count)

	var (
		valueConfigmapName = j.getHelmValuesConfigMapName()
		configmapName      = j.name + "-config"
		namespace          = j.namespace
	)

	cms := []string{valueConfigmapName, configmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(j.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

func (j *JuiceFSEngine) cleanResidualResources() (err error) {
	// configmap
	var (
		workerConfigmapName = j.name + "-worker-script"
		fuseConfigmapName   = j.name + "-fuse-script"
		cms                 = []string{workerConfigmapName, fuseConfigmapName}
		namespace           = j.namespace
	)
	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(j.Client, cm, namespace)
		if err != nil {
			j.Log.Info("DeleteConfigMap", "err", err, "cm", cm)
			return
		}
	}

	// sa
	saName := j.name + "-loader"
	err = kubeclient.DeleteServiceAccount(j.Client, saName, namespace)
	if err != nil {
		j.Log.Info("DeleteServiceAccount", "err", err, "sa", saName)
		return
	}

	// role
	roleName := j.name + "-loader"
	err = kubeclient.DeleteRole(j.Client, roleName, namespace)
	if err != nil {
		j.Log.Info("DeleteRole", "err", err, "role", roleName)
		return
	}

	// roleBinding
	roleBindingName := j.name + "-loader"
	err = kubeclient.DeleteRoleBinding(j.Client, roleBindingName, namespace)
	if err != nil {
		j.Log.Info("DeleteRoleBinding", "err", err, "roleBinding", roleBindingName)
		return
	}
	return
}
