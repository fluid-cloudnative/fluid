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
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"reflect"
	"strings"
	"time"

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
	// Label to be added
	var (
		// runtimeLabel indicates the specific runtime pod is on the node
		// e.g. fluid.io/s-alluxio-default-hbase=true
		runtimeLabel = runtimeInfo.GetRuntimeLabelname()

		// commonLabel indicates that any of fluid supported runtime is on the node
		// e.g. fluid.io/s-default-hbase=true
		commonLabel = runtimeInfo.GetCommonLabelname()

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

		toUpdate.Labels[runtimeLabel] = "true"
		toUpdate.Labels[commonLabel] = "true"
		if exclusiveness {
			toUpdate.Labels[exclusiveLabel] = utils.GetExclusiveValue(runtimeInfo.GetNamespace(), runtimeInfo.GetName())
		}

		labelNodeWithCapacityInfo(toUpdate, runtimeInfo)

		err = client.Update(context.TODO(), toUpdate)
		if err != nil {
			log.Error(err, "LabelCachedNodes")
			return err
		}
		return nil
	})

	if err != nil {
		log.Error(err, "LabelCacheNode")
		return err
	}

	// Wait at most 30s for cache in controller-runtime successfully catching up with api-server
	// This is to ensure the controller can get up-to-date cluster status during the following scheduling
	// loop.
	if err := wait.Poll(1*time.Second, 30*time.Second, func() (done bool, err error) {
		node := &v1.Node{}
		namespacedName := types.NamespacedName{
			Name: nodeName,
		}
		if err := client.Get(context.TODO(), namespacedName, node); err != nil {
			return false, fmt.Errorf("failed to get node: %w", err)
		}
		return reflect.DeepEqual(toUpdate, node), nil
	}); err != nil {
		log.Error(err, "wait polling in LabelCacheNode")
	}

	return nil
}

func labelNodeWithCapacityInfo(toUpdate *v1.Node, runtimeInfo base.RuntimeInfoInterface) {
	var (
		// memCapacityLabel indicates in-memory cache capacity assigned on the node
		// e.g. fluid.io/s-h-alluxio-m-default-hbase=1GiB
		memCapacityLabel = runtimeInfo.GetLabelnameForMemory()

		// diskCapacityLabel indicates on-disk cache capacity assigned on the node
		// e.g. fluid.io/s-h-alluxio-d-default-hbase=2GiB
		diskCapacityLabel = runtimeInfo.GetLabelnameForDisk()

		// totalCapacityLabel indicates total cache capacity assigned on the node
		// e.g. fluid.io/s-h-alluxio-t-default-hbase=3GiB
		totalCapacityLabel = runtimeInfo.GetLabelnameForTotal()
	)

	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	totalRequirement := resource.MustParse("0Gi")
	for key, requirement := range storageMap {
		value := utils.TranformQuantityToUnits(requirement)
		if key == common.MemoryCacheStore {
			toUpdate.Labels[memCapacityLabel] = value
		} else {
			toUpdate.Labels[diskCapacityLabel] = value
		}
		totalRequirement.Add(*requirement)
	}
	totalValue := utils.TranformQuantityToUnits(&totalRequirement)
	toUpdate.Labels[totalCapacityLabel] = totalValue
}
