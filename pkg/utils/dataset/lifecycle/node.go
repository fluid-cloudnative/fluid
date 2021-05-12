/*

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
	"github.com/pkg/errors"
	"strings"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

	label := runtimeInfo.GetCommonLabelname()
	log := rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())

	if len(node.Labels) > 0 {
		_, assigned = node.Labels[label]
	}

	exclusiveness := runtimeInfo.IsExclusive()
	if exclusiveness {
		log.Info("Placement Mode", "IsExclusive", exclusiveness)
		for _, nodeLabel := range node.Labels {
			if strings.Contains(nodeLabel, common.LabelAnnotationPrefix) {
				assigned = true
			}
		}
	}

	log.Info("Check alreadyAssigned", "node", node.Name, "label", label, "assigned", assigned)

	return

}

// CanbeAssigned checks if the node is already assigned the runtime engine
func CanbeAssigned(runtimeInfo base.RuntimeInfoInterface, node v1.Node) bool {
	// TODO(cheyang): the different dataset can be put in the same node, but it has to handle port conflict
	// Delete by (xieydd),  handle port conflict
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

func LabelCacheNode(nodeToLabel v1.Node, runtimeInfo base.RuntimeInfoInterface, client client.Client) (err error) {
	var (
		labelName       = runtimeInfo.GetRuntimeLabelname()
		labelCommonName = runtimeInfo.GetCommonLabelname()
		log             = rootLog.WithValues("runtime", runtimeInfo.GetName(), "namespace", runtimeInfo.GetNamespace())
	)

	exclusiveness := runtimeInfo.IsExclusive()
	log.Info("Placement Mode", "IsExclusive", exclusiveness)

	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	toUpdate := nodeToLabel.DeepCopy()
	if toUpdate.Labels == nil {
		toUpdate.Labels = make(map[string]string)
	}

	toUpdate.Labels[labelName] = "true"
	toUpdate.Labels[labelCommonName] = "true"
	if exclusiveness {
		toUpdate.Labels[utils.GetExclusiveKey()] = utils.GetExclusiveValue(runtimeInfo.GetNamespace(), runtimeInfo.GetName())
	}

	totalRequirement, err := resource.ParseQuantity("0Gi")
	if err != nil {
		return errors.Wrap(err, "Failed to parse the total requirement")
	}
	for key, requirement := range storageMap {
		value := utils.TranformQuantityToUnits(requirement)
		if key == common.MemoryCacheStore {
			toUpdate.Labels[runtimeInfo.GetLabelnameForMemory()] = value
		} else {
			toUpdate.Labels[runtimeInfo.GetLabelnameForDisk()] = value
		}
		totalRequirement.Add(*requirement)
	}
	totalValue := utils.TranformQuantityToUnits(&totalRequirement)
	toUpdate.Labels[runtimeInfo.GetLabelnameForTotal()] = totalValue

	// toUpdate.Labels[labelNameToAdd] = "true"
	err = client.Update(context.TODO(), toUpdate)
	if err != nil {
		log.Error(err, "LabelCacheNode")
		return err
	}
	return nil
}
