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
	stdlog "log"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	fluidctrl "github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

var rootLog logr.Logger
var nodeExcludeSelector labels.Selector

func init() {
	rootLog = ctrl.Log.WithName("dataset.lifecycle")
	if err := parseNodeExcludeSelectorFromEnv(); err != nil {
		stdlog.Fatal(err)
	}
	if nodeExcludeSelector != nil {
		stdlog.Printf("Found non-empty nodeExcludeSelector \"%s\"", nodeExcludeSelector.String())
	} else {
		stdlog.Print("nodeExcludeSelector is empty, no node would be excluded when syncing schedule info of runtimes")
	}
}

func SyncScheduleInfoToCacheNodes(runtimeInfo base.RuntimeInfoInterface, client client.Client) error {
	defer utils.TimeTrack(time.Now(), "SyncScheduleInfoToCacheNodes", "name", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())

	// Get current cache nodes from worker pods
	desiredNodeNames, err := getDesiredNodesWithScheduleInfo(runtimeInfo, client)
	if err != nil {
		return err
	}

	// Get previously assigned cache nodes
	actualNodeNames, err := getActualNodesWithScheduleInfo(runtimeInfo, client)
	if err != nil {
		return err
	}

	// Calculate node differences
	nodesToAddScheduleInfo, nodesToRemoveScheduleInfo := calculateNodeDifferences(desiredNodeNames, actualNodeNames)

	// Apply changes to nodes
	applyNodeChanges(nodesToAddScheduleInfo, nodesToRemoveScheduleInfo, runtimeInfo, client)

	return nil
}

// getDesiredNodesWithScheduleInfo retrieves the desired cache nodes with schedule info
func getDesiredNodesWithScheduleInfo(runtimeInfo base.RuntimeInfoInterface, client client.Client) ([]string, error) {
	workers, err := fluidctrl.GetWorkersAsStatefulset(client,
		types.NamespacedName{Namespace: runtimeInfo.GetNamespace(), Name: runtimeInfo.GetWorkerStatefulsetName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			rootLog.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			return nil, nil
		}
		return nil, err
	}

	workerSelector, err := metav1.LabelSelectorAsSelector(workers.Spec.Selector)
	if err != nil {
		return nil, err
	}

	workerPods, err := kubeclient.GetPodsForStatefulSet(client, workers, workerSelector)
	if err != nil {
		return nil, err
	}

	var nodeNames []string
	for _, pod := range workerPods {
		if pod.Spec.NodeName != "" {
			nodeNames = append(nodeNames, pod.Spec.NodeName)
		}
	}

	return utils.RemoveDuplicateStr(nodeNames), nil
}

// getActualNodesWithScheduleInfo retrieves the actual cache nodes with schedule info
func getActualNodesWithScheduleInfo(runtimeInfo base.RuntimeInfoInterface, cli client.Client) (nodeNames []string, err error) {
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

	return utils.RemoveDuplicateStr(nodeNames), nil
}

// calculateNodeDifferences calculates which nodes need to be added or removed
func calculateNodeDifferences(currentNodes, previousNodes []string) (nodesToAddLabels, nodesToRemovedLabels []string) {
	nodesToAddLabels = utils.SubtractString(currentNodes, previousNodes)
	nodesToRemovedLabels = utils.SubtractString(previousNodes, currentNodes)
	return
}

// applyNodeChanges applies the calculated changes to the nodes
func applyNodeChanges(nodesToAddScheduleInfo, nodesToRemoveScheduleInfo []string, runtimeInfo base.RuntimeInfoInterface, client client.Client) {
	// Add schedule info to new nodes
	for _, nodeName := range nodesToAddScheduleInfo {
		if err := addScheduleInfoToNode(nodeName, runtimeInfo, client); err != nil {
			rootLog.Info("Failed to add schedule info to node, continuing", "node", nodeName, "error", err)
		}
	}

	// Remove schedule info from old nodes
	for _, nodeName := range nodesToRemoveScheduleInfo {
		if err := removeScheduleInfoFromNode(nodeName, runtimeInfo, client); err != nil {
			rootLog.Info("Failed to remove schedule info from node, continuing", "node", nodeName, "error", err)
		}
	}
}

func addScheduleInfoToNode(nodeName string, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
	node := corev1.Node{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Name: nodeName,
	}, &node)
	if err != nil {
		return err
	}

	if nodeExcludeSelector != nil && nodeExcludeSelector.Matches(labels.Set(node.Labels)) {
		rootLog.V(1).Info("Skip addScheduleInfoToNode to node that matches nodeExcludeSelector", "nodeExcludeSelector", nodeExcludeSelector.String(), "node", node.Name)
		return nil
	}

	if hasRuntimeLabel(node, runtimeInfo) {
		rootLog.Info("Node already added schedule info, skip.", "node", nodeName)
		return
	}

	err = labelCacheNode(node, runtimeInfo, client)
	if err != nil {
		return err
	}

	return nil
}

func removeScheduleInfoFromNode(nodeName string, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
	node := corev1.Node{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Name: nodeName,
	}, &node)
	if err != nil {
		return err
	}

	if !hasRuntimeLabel(node, runtimeInfo) {
		rootLog.Info("Node doesn't have existing schedule info, skip.", "node", nodeName)
		return
	}

	err = unlabelCacheNode(node, runtimeInfo, client)
	if err != nil {
		return err
	}

	return nil
}

// hasRuntimeLabel checks if the node has the runtime label
func hasRuntimeLabel(node corev1.Node, runtimeInfo base.RuntimeInfoInterface) bool {
	key := runtimeInfo.GetRuntimeLabelName()
	if len(node.Labels) == 0 {
		return false
	}
	_, found := node.Labels[key]
	return found
}

// labelCacheNode adds labels on a selected node to indicate the node is scheduled with corresponding runtime
func labelCacheNode(nodeToLabel corev1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
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

	placementMode := runtimeInfo.GetPlacementModeWithDefault(datav1alpha1.ExclusiveMode)
	log.Info("Placement Mode", "mode", placementMode)
	if placementMode == datav1alpha1.ExclusiveMode {
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

		if placementMode == datav1alpha1.ExclusiveMode {
			exclusiveLabelValue := runtimeInfo.GetExclusiveLabelValue()
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

// unlabelCacheNode remove labels on a selected node to indicate the node doesn't have the cache for
func unlabelCacheNode(node corev1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {

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

		exclusiveLabelValue := runtimeInfo.GetExclusiveLabelValue()
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

func parseNodeExcludeSelectorFromEnv() error {
	excludeSelectorStr := utils.GetStringValueFromEnv(common.EnvScheduleInfoExcludeNodeSelector, "")
	if len(excludeSelectorStr) != 0 {
		tmpSelector, err := metav1.ParseToLabelSelector(excludeSelectorStr)
		if err != nil {
			return fmt.Errorf("failed to parse node exclude selector \"%v\" from env %s: %v", excludeSelectorStr, common.EnvScheduleInfoExcludeNodeSelector, err)
		}

		nodeExcludeSelector, err = metav1.LabelSelectorAsSelector(tmpSelector)
		if err != nil {
			return fmt.Errorf("failed to parse node exclude selector \"%v\" from %s: %v", tmpSelector.String(), common.EnvScheduleInfoExcludeNodeSelector, err)
		}
	}

	return nil
}
