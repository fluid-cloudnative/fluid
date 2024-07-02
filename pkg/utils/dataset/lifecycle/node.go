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

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/tieredstore"
)

var rootLog logr.Logger

func init() {
	rootLog = ctrl.Log.WithName("dataset.lifecycle")
}

// AlreadyAssigned checks if the node is already assigned the runtime engine
// If runtime engine cached dataset is exclusive, will check if any runtime engine already assigned the runtime engine
func AlreadyAssigned(runtimeInfo base.RuntimeInfoInterface, node v1.Node) (assigned bool) {
	// label := e.getCommonLabelname()

	label := runtimeInfo.GetCommonLabelName()
	log := rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())

	if len(node.Labels) > 0 {
		_, assigned = node.Labels[label]
	}

	log.Info("Check alreadyAssigned", "node", node.Name, "label", label, "assigned", assigned)

	return

}

// CanbeAssigned checks if the node is already assigned the runtime engine
func CanbeAssigned(runtimeInfo base.RuntimeInfoInterface, node v1.Node) bool {
	// TODO(xieydd): Resource consumption of multi dataset same node
	// if e.alreadyAssignedByFluid(node) {
	// 	return false
	// }
	label := utils.GetExclusiveKey()
	log := rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())
	value, cannotBeAssigned := node.Labels[label]
	if cannotBeAssigned {
		log.Info("node ", node.Name, "is exclusive and already be assigned, can not be assigned",
			"key", label,
			"value", value)
		return false
	}

	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	for key, requirement := range storageMap {
		if key == common.MemoryCacheStore {
			nodeMemoryCapacity := *node.Status.Allocatable.Memory()
			if requirement.Cmp(nodeMemoryCapacity) <= 0 {
				log.Info("requirement is less than node memory capacity", "requirement", requirement,
					"nodeMemoryCapacity", nodeMemoryCapacity)
			} else {
				log.Info("requirement is more than node memory capacity", "requirement", requirement,
					"nodeMemoryCapacity", nodeMemoryCapacity)
				return false
			}
		}

		// } else {
		// 	nodeDiskCapacity := *node.Status.Allocatable.StorageEphemeral()
		// 	if requirement.Cmp(nodeDiskCapacity) <= 0 {
		// 		log.Info("requirement is less than node disk capacity", "requirement", requirement,
		// 			"nodeDiskCapacity", nodeDiskCapacity)
		// 	} else {
		// 		log.Info("requirement is more than node disk capacity", "requirement", requirement,
		// 			"nodeDiskCapacity", nodeDiskCapacity)
		// 		return false
		// 	}
		// }
	}

	return true

}

// CheckIfRuntimeInNode checks if the the runtime on this node
func CheckIfRuntimeInNode(node v1.Node, runtimeInfo base.RuntimeInfoInterface) (found bool) {
	key := runtimeInfo.GetRuntimeLabelName()
	return findLabelNameOnNode(node, key)
}

// findLabelNameOnNode checks if the label exist
func findLabelNameOnNode(node v1.Node, key string) (found bool) {
	labels := node.Labels
	if len(labels) == 0 {
		return
	}
	_, found = labels[key]
	return
}

// LabelCacheNode adds labels on a selected node to indicate the node is scheduled with corresponding runtime
func LabelCacheNode(nodeToLabel v1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
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
	var toUpdate *v1.Node
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

func labelNodeWithCapacityInfo(toUpdate *v1.Node, runtimeInfo base.RuntimeInfoInterface, labelsToModify *common.LabelsToModify) {
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
func UnlabelCacheNode(node v1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {

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
	nodeToUpdate := &v1.Node{}
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
func DecreaseDatasetNum(toUpdate *v1.Node, runtimeInfo base.RuntimeInfoInterface, labelsToModify *common.LabelsToModify) error {
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
func increaseDatasetNum(toUpdate *v1.Node, runtimeInfo base.RuntimeInfoInterface, labelsToModify *common.LabelsToModify) error {
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
