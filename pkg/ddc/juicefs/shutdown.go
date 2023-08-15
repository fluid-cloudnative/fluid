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
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	_, err = j.destroyWorkers(-1)
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
	var valueConfigMapName = j.getConfigmapName()

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
		if utils.IgnoreNotFound(err) == nil {
			j.Log.Info("worker of runtime %s namespace %s has been shutdown.", runtime.Name, runtime.Namespace)
			return nil
		} else {
			return err
		}
	}

	if len(pods) == 0 {
		j.Log.Info("no worker pod of runtime %s namespace %s", runtime.Name, runtime.Namespace)
		return
	}
	uuid, err := j.getUUID(pods[0], common.JuiceFSWorkerContainer)
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

// destroyWorkers attempts to delete the workers until worker num reaches the given expectedWorkers, if expectedWorkers is -1, it means all the workers should be deleted
// This func returns currentWorkers representing how many workers are left after this process.
func (j *JuiceFSEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	runtimeInfo, err := j.getRuntimeInfo()
	if err != nil {
		return currentWorkers, err
	}

	var (
		nodeList           = &corev1.NodeList{}
		labelExclusiveName = utils.GetExclusiveKey()
		labelName          = runtimeInfo.GetRuntimeLabelName()
		labelCommonName    = runtimeInfo.GetCommonLabelName()
		labelMemoryName    = runtimeInfo.GetLabelNameForMemory()
		labelDiskName      = runtimeInfo.GetLabelNameForDisk()
		labelTotalName     = runtimeInfo.GetLabelNameForTotal()
	)

	labelNames := []string{labelName, labelTotalName, labelDiskName, labelMemoryName, labelCommonName}
	j.Log.Info("check node labels", "labelNames", labelNames)

	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", labelCommonName))
	if err != nil {
		return currentWorkers, err
	}

	err = j.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})

	if err != nil {
		return currentWorkers, err
	}

	currentWorkers = int32(len(nodeList.Items))
	if expectedWorkers >= currentWorkers {
		j.Log.Info("No need to scale in. Skip.")
		return currentWorkers, nil
	}

	var nodes []corev1.Node
	if expectedWorkers >= 0 {
		j.Log.Info("Scale in juicefs workers", "expectedWorkers", expectedWorkers)
		// This is a scale in operation
		runtimeInfo, err := j.getRuntimeInfo()
		if err != nil {
			j.Log.Error(err, "getRuntimeInfo when scaling in")
			return currentWorkers, err
		}

		fuseGlobal, _ := runtimeInfo.GetFuseDeployMode()
		nodes, err = j.sortNodesToShutdown(nodeList.Items, fuseGlobal)
		if err != nil {
			return currentWorkers, err
		}

	} else {
		// Destroy all workers. This is a subprocess during deletion of JuiceFSRuntime
		nodes = nodeList.Items
	}

	// 1.select the nodes
	for _, node := range nodes {
		if expectedWorkers == currentWorkers {
			break
		}

		if len(node.Labels) == 0 {
			continue
		}

		nodeName := node.Name
		var labelsToModify common.LabelsToModify
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			node, err := kubeclient.GetNode(j.Client, nodeName)
			if err != nil {
				j.Log.Error(err, "Fail to get node", "nodename", nodeName)
				return err
			}

			toUpdate := node.DeepCopy()
			for _, label := range labelNames {
				labelsToModify.Delete(label)
			}

			exclusiveLabelValue := utils.GetExclusiveValue(j.namespace, j.name)
			if val, exist := toUpdate.Labels[labelExclusiveName]; exist && val == exclusiveLabelValue {
				labelsToModify.Delete(labelExclusiveName)
			}

			err = lifecycle.DecreaseDatasetNum(toUpdate, runtimeInfo, &labelsToModify)
			if err != nil {
				return err
			}
			// Update the toUpdate in UPDATE mode
			// modifiedLabels, err := utils.ChangeNodeLabelWithUpdateMode(e.Client, toUpdate, labelToModify)
			// Update the toUpdate in PATCH mode
			modifiedLabels, err := utils.ChangeNodeLabelWithPatchMode(j.Client, toUpdate, labelsToModify)
			if err != nil {
				return err
			}
			j.Log.Info("Destroy worker", "Dataset", j.name, "deleted worker node", node.Name, "removed or updated labels", modifiedLabels)
			return nil
		})

		if err != nil {
			return currentWorkers, err
		}

		currentWorkers--
	}

	return currentWorkers, nil
}

func (j *JuiceFSEngine) sortNodesToShutdown(candidateNodes []corev1.Node, fuseGlobal bool) (nodes []corev1.Node, err error) {
	if !fuseGlobal {
		// If fuses are deployed in non-global mode, workers and fuses will be scaled in together.
		// It can be dangerous if we scale in nodes where there are pods using the related pvc.
		// So firstly we filter out such nodes
		pvcMountNodes, err := kubeclient.GetPvcMountNodes(j.Client, j.name, j.namespace)
		if err != nil {
			j.Log.Error(err, "GetPvcMountNodes when scaling in")
			return nil, err
		}

		for _, node := range candidateNodes {
			if _, found := pvcMountNodes[node.Name]; !found {
				nodes = append(nodes, node)
			}
		}
	} else {
		// If fuses are deployed in global mode. Scaling in workers has nothing to do with fuses.
		// All nodes with related label can be candidate nodes.
		nodes = candidateNodes
	}

	// Prefer to choose nodes with less data cache
	//Todo

	return nodes, nil
}

func (j *JuiceFSEngine) cleanAll() (err error) {
	count, err := j.Helper.CleanUpFuse()
	if err != nil {
		j.Log.Error(err, "Err in cleaning fuse")
		return err
	}
	j.Log.Info("clean up fuse count", "n", count)

	var (
		valueConfigmapName = j.name + "-" + j.runtimeType + "-values"
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
