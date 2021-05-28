package jindo

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/retry"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// shut down the Jindo engine
func (e *JindoEngine) Shutdown() (err error) {

	err = e.invokeCleanCache()
	if err != nil {
		return
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

// destroyMaster Destroies the master
func (e *JindoEngine) destroyMaster() (err error) {
	found := false
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
	var (
		valueConfigmapName = e.name + "-" + e.runtimeType + "-values"
		configmapName      = e.name + "-config"
		namespace          = e.namespace
	)

	cms := []string{valueConfigmapName, configmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(e.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// destroyWorkers will delete the workers by number of the workers, if workers is -1, it means all the workers are deleted
func (e *JindoEngine) destroyWorkers(expectedWorkers int32) (currentWorkers int32, err error) {
	runtimeInfo, err := e.getRuntimeInfo()
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
		labelTotalname     = runtimeInfo.GetLabelNameForTotal()
	    labelDatasetNum    = runtimeInfo.GetDatasetNumLabelName()
	)

	labelNames := []string{labelName, labelTotalname, labelDiskName, labelMemoryName, labelCommonName}
	e.Log.Info("check node labels", "labelNames", labelNames)
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", labelCommonName))
	if err != nil {
		return
	}

	err = e.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})
	if err != nil {
		return currentWorkers, err
	}

	currentWorkers = int32(len(nodeList.Items))
	if expectedWorkers >= currentWorkers {
		e.Log.Info("No need to scale in. Skip.")
		return currentWorkers, nil
	}

	var nodes []corev1.Node
	if expectedWorkers >= 0 {
		e.Log.Info("Scale in Jindo workers", "expectedWorkers", expectedWorkers)
		// This is a scale in operation
		runtimeInfo, err := e.getRuntimeInfo()
		if err != nil {
			e.Log.Error(err, "getRuntimeInfo when scaling in")
			return currentWorkers, err
		}

		fuseGlobal, _ := runtimeInfo.GetFuseDeployMode()
		nodes, err = e.sortNodesToShutdown(nodeList.Items, fuseGlobal)
		if err != nil {
			return currentWorkers, err
		}

	} else {
		// Destroy all workers. This is a subprocess during deletion of JindoRuntime
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
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			node, err := kubeclient.GetNode(e.Client, nodeName)
			if err != nil {
				e.Log.Error(err, "Fail to get node", "nodename", nodeName)
				return err
			}

			toUpdate := node.DeepCopy()
			for _, label := range labelNames {
				delete(toUpdate.Labels, label)
			}

			exclusiveLabelValue := utils.GetExclusiveValue(e.namespace, e.name)
			if val, exist := toUpdate.Labels[labelExclusiveName]; exist && val == exclusiveLabelValue {
				delete(toUpdate.Labels, labelExclusiveName)
				labelNames = append(labelNames, labelExclusiveName)
			}

			isDeleted, err := lifecycle.DecreaseDatasetNum(toUpdate, runtimeInfo)
			if err != nil {
				return err
			}
			if isDeleted {
				labelNames = append(labelNames, labelDatasetNum)
			}

			err = e.Client.Update(context.TODO(), toUpdate)
			if err != nil {
				return err
			}
			e.Log.Info("Destroy worker", "Dataset", e.name, "deleted worker node", node.Name, "removed labels", labelNames)
			return nil

		})

		if err != nil {
			return currentWorkers, err
		}
		currentWorkers--
	}

	return currentWorkers, nil
}

func (e *JindoEngine) sortNodesToShutdown(candidateNodes []corev1.Node, fuseGlobal bool) (nodes []corev1.Node, err error) {
	if !fuseGlobal {
		// If fuses are deployed in non-global mode, workers and fuses will be scaled in together.
		// It can be dangerous if we scale in nodes where there are pods using the related pvc.
		// So firstly we filter out such nodes
		pvcMountNodes, err := kubeclient.GetPvcMountNodes(e.Client, e.name, e.namespace)
		if err != nil {
			e.Log.Error(err, "GetPvcMountNodes when scaling in")
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
	// TODO support jindo calculate node usedCapacity
	return nodes, nil
}
