/*
Copyright 2023 The Fluid Authors.

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

package lifecycle

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	fluidctrl "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

var rootLog logr.Logger

func init() {
	rootLog = ctrl.Log.WithName("dataset.lifecycle")
}

func SyncScheduleInfoToCacheNodes(runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
	defer utils.TimeTrack(time.Now(), "SyncScheduleInfoToCacheNodes", "name", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())

	var (
		currentCacheNodeNames  []string
		previousCacheNodeNames []string
	)

	workers, err := fluidctrl.GetWorkersAsStatefulset(client,
		types.NamespacedName{Namespace: runtimeInfo.GetNamespace(), Name: runtimeInfo.GetWorkerStatefulsetName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			rootLog.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			return nil
		}
		return err
	}

	workerSelector, err := metav1.LabelSelectorAsSelector(workers.Spec.Selector)

	workerPods, err := kubeclient.GetPodsForStatefulSet(client, workers, workerSelector)
	if err != nil {
		return err
	}

	for _, pod := range workerPods {
		if !podutil.IsPodReady(&pod) {
			rootLog.V(1).Info("Skip the pod because it's not ready", "pod", pod.Name)
			continue
		}

		if len(pod.Spec.NodeName) != 0 {
			currentCacheNodeNames = append(currentCacheNodeNames, pod.Spec.NodeName)
		}
	}

	// find the nodes which already have the runtime labels
	previousCacheNodeNames, err = getAssignedNodes(runtimeInfo, client)

	currentCacheNodeNames = utils.RemoveDuplicateStr(currentCacheNodeNames)
	previousCacheNodeNames = utils.RemoveDuplicateStr(previousCacheNodeNames)

	addedCacheNodenames := utils.SubtractString(currentCacheNodeNames, previousCacheNodeNames)
	removedCacheNodenames := utils.SubtractString(previousCacheNodeNames, currentCacheNodeNames)

	if len(addedCacheNodenames) > 0 {
		for _, nodeName := range addedCacheNodenames {
			node := corev1.Node{}
			err = client.Get(context.TODO(), types.NamespacedName{
				Name: nodeName,
			}, &node)
			if err != nil {
				rootLog.Error(err, "Failed to find new cache node", "node", nodeName)
				return err
			}
			if !CheckIfRuntimeInNode(node, runtimeInfo) {
				err = LabelCacheNode(node, runtimeInfo, client)
				if err != nil {
					rootLog.Error(err, "Failed to label new cache node", "node", nodeName)
					return err
				}
			} else {
				rootLog.Info("The node is already added to cache", "node", nodeName)
			}
		}
	}

	if len(removedCacheNodenames) > 0 {
		for _, nodeName := range removedCacheNodenames {
			node := corev1.Node{}
			err = client.Get(context.TODO(), types.NamespacedName{
				Name: nodeName,
			}, &node)
			if utils.IgnoreNotFound(err) != nil {
				rootLog.Error(err, "Failed to find new cache node", "node", nodeName)
				return err
			}
			if CheckIfRuntimeInNode(node, runtimeInfo) {
				err = UnlabelCacheNode(node, runtimeInfo, client)
				if err != nil {
					rootLog.Error(err, "Failed to unlabel cache node", "node", nodeName)
					return err
				}
			} else {
				rootLog.Info("The node is already removed from cache", "node", nodeName)
			}

		}
	}
	return nil
}

func getAssignedNodes(runtimeInfo base.RuntimeInfoInterface, cli client.Client) (nodeNames []string, err error) {
	var (
		nodeList     = &corev1.NodeList{}
		runtimeLabel = runtimeInfo.GetRuntimeLabelName()
	)

	nodeNames = []string{}
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", runtimeLabel))
	if err != nil {
		return
	}

	err = cli.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})
	if err != nil {
		return
	}

	for _, node := range nodeList.Items {
		nodeNames = append(nodeNames, node.Name)
	}

	return
}

// CheckIfRuntimeInNode checks if the the runtime on this node
func CheckIfRuntimeInNode(node corev1.Node, runtimeInfo base.RuntimeInfoInterface) (found bool) {
	key := runtimeInfo.GetRuntimeLabelName()
	return findLabelNameOnNode(node, key)
}

// findLabelNameOnNode checks if the label exist
func findLabelNameOnNode(node corev1.Node, key string) (found bool) {
	labels := node.Labels
	if len(labels) == 0 {
		return
	}
	_, found = labels[key]
	return
}

// LabelCacheNode adds labels on a selected node to indicate the node is scheduled with corresponding runtime
func LabelCacheNode(nodeToLabel corev1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
	defer utils.TimeTrack(time.Now(), "LabelCacheNode", "runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace(), "node", nodeToLabel.Name)
	// Label to be added
	var (
		// runtimeLabel indicates the specific runtime pod is on the node
		// e.g. fluid.io/s-alluxio-default-hbase=true
		runtimeLabel = runtimeInfo.GetRuntimeLabelName()

		// commonLabel indicates that any of fluid supported runtime is on the node
		// e.g. fluid.io/s-default-hbase=true
		commonLabel = runtimeInfo.GetCommonLabelName()

		// exclusiveLabel is the label key indicates the node is exclusively assigned
		// e.g. fluid_exclusive=default_hbase
		exclusiveLabel string
	)

	log := rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())

	exclusiveness := runtimeInfo.IsExclusive()
	log.Info("Placement Mode", "IsExclusive", exclusiveness)
	if exclusiveness {
		exclusiveLabel = utils.GetExclusiveKey()
	}

	nodeName := nodeToLabel.Name
	var toUpdate *corev1.Node
	var modifiedLabels []string
	var labelsToModify common.LabelsToModify
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		node, err := kubeclient.GetNode(client, nodeName)
		if err != nil {
			log.Error(err, "GetNode In labelCacheNode")
			return err
		}

		toUpdate = node.DeepCopy()
		if toUpdate.Labels == nil {
			toUpdate.Labels = make(map[string]string)
		}

		labelsToModify.Add(runtimeLabel, "true")
		labelsToModify.Add(commonLabel, "true")

		if exclusiveness {
			exclusiveLabelValue := utils.GetExclusiveValue(runtimeInfo.GetNamespace(), runtimeInfo.GetName())
			labelsToModify.Add(exclusiveLabel, exclusiveLabelValue)
		}

		err = increaseDatasetNum(toUpdate, runtimeInfo, &labelsToModify)
		if err != nil {
			log.Error(err, "fail to update datasetNum label")
			return err
		}

		labelNodeWithCapacityInfo(toUpdate, runtimeInfo, &labelsToModify)

		// Update the toUpdate in UPDATE mode
		// modifiedLabels, err = utils.ChangeNodeLabelWithUpdateMode(client, toUpdate, labelToModify)
		// Update the toUpdate in PATCH mode
		modifiedLabels, err = utils.ChangeNodeLabelWithPatchMode(client, toUpdate, labelsToModify)

		if err != nil {
			log.Error(err, fmt.Sprintf("update node labels failed, node name: %s, labels:%v", node.Name, node.Labels))
			return err
		}
		return nil
	})

	if err != nil {
		log.Error(err, fmt.Sprintf("fail to update the labels, node name: %s", nodeName))
		return err
	}

	// Wait infinitely with 30-second loops for cache in controller-runtime successfully catching up with api-server
	// This is to ensure the controller can get up-to-date cluster status during the following scheduling
	// loop.
	pollStartTime := time.Now()
	for i := 1; ; i++ {
		if err := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 30*time.Second, true, func(ctx context.Context) (done bool, err error) {
			node, err := kubeclient.GetNode(client, nodeName)
			if err != nil {
				return false, fmt.Errorf("failed to get node: %w", err)
			}
			return utils.ContainsAll(node.Labels, modifiedLabels), nil
		}); err == nil {
			break
		}
		// if timeout, retry infinitely
		if wait.Interrupted(err) {
			log.Error(err, fmt.Sprintf("client cache can't catch up with api-server after %v secs", i*30), "nodeName", nodeName)
			continue
		}
		log.Error(err, "wait polling in LabelCacheNode")
		return err
	}
	utils.TimeTrack(pollStartTime, "polling up-to-date cache status when scheduling", "nodeToLabel", nodeToLabel.Name)

	return nil
}

func labelNodeWithCapacityInfo(toUpdate *corev1.Node, runtimeInfo base.RuntimeInfoInterface, labelsToModify *common.LabelsToModify) {
	var (
		// memCapacityLabel indicates in-memory cache capacity assigned on the node
		// e.g. fluid.io/s-h-alluxio-m-default-hbase=1GiB
		memCapacityLabel = runtimeInfo.GetLabelNameForMemory()

		// diskCapacityLabel indicates on-disk cache capacity assigned on the node
		// e.g. fluid.io/s-h-alluxio-d-default-hbase=2GiB
		diskCapacityLabel = runtimeInfo.GetLabelNameForDisk()

		// totalCapacityLabel indicates total cache capacity assigned on the node
		// e.g. fluid.io/s-h-alluxio-t-default-hbase=3GiB
		totalCapacityLabel = runtimeInfo.GetLabelNameForTotal()
	)

	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	totalRequirement := resource.MustParse("0Gi")
	for key, requirement := range storageMap {
		value := utils.TranformQuantityToUnits(requirement)
		if key == common.MemoryCacheStore {
			labelsToModify.Add(memCapacityLabel, value)
		} else {
			labelsToModify.Add(diskCapacityLabel, value)
		}
		totalRequirement.Add(*requirement)
	}
	totalValue := utils.TranformQuantityToUnits(&totalRequirement)
	labelsToModify.Add(totalCapacityLabel, totalValue)
}

// UnlabelCacheNode remove labels on a selected node to indicate the node doesn't have the cache for
func UnlabelCacheNode(node corev1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {

	var (
		labelExclusiveName = utils.GetExclusiveKey()
		labelName          = runtimeInfo.GetRuntimeLabelName()
		labelCommonName    = runtimeInfo.GetCommonLabelName()
		labelMemoryName    = runtimeInfo.GetLabelNameForMemory()
		labelDiskName      = runtimeInfo.GetLabelNameForDisk()
		labelTotalName     = runtimeInfo.GetLabelNameForTotal()
	)

	labelNames := []string{labelName, labelTotalName, labelDiskName, labelMemoryName, labelCommonName}
	log := rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())
	log.Info("check node labels", "labelNames", labelNames)

	nodeName := node.Name
	nodeToUpdate := &corev1.Node{}
	var labelsToModify common.LabelsToModify
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err = client.Get(context.TODO(), types.NamespacedName{Name: nodeName}, nodeToUpdate)
		if err != nil {
			log.Error(err, "Fail to get node", "nodename", nodeName)
			return err
		}

		nodeToUpdate = nodeToUpdate.DeepCopy()
		for _, label := range labelNames {
			labelsToModify.Delete(label)
		}

		exclusiveLabelValue := utils.GetExclusiveValue(runtimeInfo.GetNamespace(), runtimeInfo.GetName())
		if val, exist := nodeToUpdate.Labels[labelExclusiveName]; exist && val == exclusiveLabelValue {
			labelsToModify.Delete(labelExclusiveName)
		}

		err = DecreaseDatasetNum(nodeToUpdate, runtimeInfo, &labelsToModify)
		if err != nil {
			return err
		}

		// Update the toUpdate in UPDATE mode
		// modifiedLabels, err := utils.ChangeNodeLabelWithUpdateMode(e.Client, toUpdate, labelToModify)
		// Update the toUpdate in PATCH mode
		modifiedLabels, err := utils.ChangeNodeLabelWithPatchMode(client, nodeToUpdate, labelsToModify)
		if err != nil {
			log.Error(err, "Failed to change node label with patch mode")
		}
		log.Info("Destroy worker", "Dataset", runtimeInfo.GetName(), "deleted worker node", node.Name, "removed or updated labels", modifiedLabels)
		return err

	})

	return
}

// DecreaseDatasetNum deletes the datasetNum label or updates the number of the dataset in the specific node.
func DecreaseDatasetNum(toUpdate *corev1.Node, runtimeInfo base.RuntimeInfoInterface, labelsToModify *common.LabelsToModify) error {
	var labelDatasetNum = runtimeInfo.GetDatasetNumLabelName()
	if val, exist := toUpdate.Labels[labelDatasetNum]; exist {
		currentDataset, err := strconv.Atoi(val)
		if err != nil {
			rootLog.Error(err, "The dataset number format error")
			return err
		}
		if currentDataset < 2 {
			labelsToModify.Delete(labelDatasetNum)
		} else {
			labelDatasetNumValue := strconv.Itoa(currentDataset - 1)
			labelsToModify.Update(labelDatasetNum, labelDatasetNumValue)
		}
	}
	return nil
}

// increaseDatasetNum adds the datasetNum label or updates the number of the dataset in the specific node.
func increaseDatasetNum(toUpdate *corev1.Node, runtimeInfo base.RuntimeInfoInterface, labelsToModify *common.LabelsToModify) error {
	var labelDatasetNum = runtimeInfo.GetDatasetNumLabelName()
	if currentDatasetNum, ok := toUpdate.Labels[labelDatasetNum]; ok {
		currentData, err := strconv.Atoi(currentDatasetNum)
		if err != nil {
			return err
		}
		datasetLabelValue := strconv.Itoa(currentData + 1)
		labelsToModify.Update(labelDatasetNum, datasetLabelValue)
	} else {
		labelsToModify.Add(labelDatasetNum, "1")
	}
	return nil
}
