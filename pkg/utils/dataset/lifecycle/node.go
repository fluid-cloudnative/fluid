package lifecycle

import (
	"context"
	"strings"

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

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("dataset.lifecycle")
}

// AlreadyAssigned checks if the node is already assigned the runtime engine
// If runtime engine cached dataset is exclusive, will check if any runtime engine already assigned the runtime engine
func AlreadyAssigned(runtimeInfo base.RuntimeInfoInterface, node v1.Node) (assigned bool) {
	// label := e.getCommonLabelname()

	label := runtimeInfo.GetCommonLabelname()

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
	label := common.Exclusive
	_, cannotBeAssigned := node.Labels[label]
	if cannotBeAssigned {
		log.Info("node ", node.Name, "is exclusiveness, can not be assigned")
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
		labelName          = runtimeInfo.GetRuntimeLabelname()
		labelCommonName    = runtimeInfo.GetCommonLabelname()
		labelExclusiveName string
	)

	exclusiveness := runtimeInfo.IsExclusive()
	log.Info("Placement Mode", "IsExclusive", exclusiveness)
	if exclusiveness {
		labelExclusiveName = common.Exclusive
	}

	storageMap := tieredstore.GetLevelStorageMap(runtimeInfo)

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		nodeName := nodeToLabel.Name
		node, err := kubeclient.GetNode(client, nodeName)
		if err != nil {
			log.Error(err, "GetNode In labelCacheNode")
			return err
		}
		toUpdate := node.DeepCopy()
		if toUpdate.Labels == nil {
			toUpdate.Labels = make(map[string]string)
		}

		toUpdate.Labels[labelName] = "true"
		toUpdate.Labels[labelCommonName] = "true"
		if exclusiveness {
			toUpdate.Labels[labelExclusiveName] = "true"
		}
		totalRequirement, err := resource.ParseQuantity("0Gi")
		if err != nil {
			log.Error(err, "Failed to parse the total requirement")
		}
		for key, requirement := range storageMap {
			value := utils.TranformQuantityToUnits(requirement)
			if key == common.MemoryCacheStore {
				toUpdate.Labels[runtimeInfo.GetStoragetLabelname(common.HumanReadType, common.MemoryStorageType)] = value
			} else {
				toUpdate.Labels[runtimeInfo.GetStoragetLabelname(common.HumanReadType, common.DiskStorageType)] = value
			}
			totalRequirement.Add(*requirement)
		}
		totalValue := utils.TranformQuantityToUnits(&totalRequirement)
		toUpdate.Labels[runtimeInfo.GetStoragetLabelname(common.HumanReadType, common.TotalStorageType)] = totalValue

		// toUpdate.Labels[labelNameToAdd] = "true"
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

	return nil
}
