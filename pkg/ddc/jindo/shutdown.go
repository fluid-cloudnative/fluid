package jindo

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// shut down the Jindo engine
func (e *JindoEngine) Shutdown() (err error) {
	if e.retryShutdown < e.gracefulShutdownLimits {
		//err = e.cleanupCache()
		err = nil
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed",
				// "engine", e,
				"retry times", e.retryShutdown)
			return
		}
	}

	/* TODO metadata sync
	if e.MetadataSyncDoneCh != nil {
		close(e.MetadataSyncDoneCh)
	}*/

	err = e.destroyWorkers(-1)
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
func (e *JindoEngine) destroyWorkers(workers int32) (err error) {
	var (
		nodeList *corev1.NodeList = &corev1.NodeList{}

		labelName          = e.getRuntimeLabelname()
		labelCommonName    = e.getCommonLabelname()
		labelMemoryName    = e.getStoragetLabelname(humanReadType, memoryStorageType)
		labelDiskName      = e.getStoragetLabelname(humanReadType, diskStorageType)
		labelTotalname     = e.getStoragetLabelname(humanReadType, totalStorageType)
		labelExclusiveName = common.Exclusive
	)

	err = e.List(context.TODO(), nodeList, &client.ListOptions{})
	if err != nil {
		return
	}

	labelNames := []string{labelName, labelTotalname, labelDiskName, labelMemoryName, labelCommonName}

	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return
	}

	if runtimeInfo.IsExclusive() {
		labelNames = append(labelNames, labelExclusiveName)
	}

	// 1.select the nodes
	// TODO(cheyang) Need consider node selector
	var i int32 = 0
	for _, node := range nodeList.Items {
		if workers >= 0 {
			if i > workers {
				e.Log.Info("destroy workers", "count", i)
				break
			}
		}

		// nodes = append(nodes, &node)
		toUpdate := node.DeepCopy()
		if len(toUpdate.Labels) == 0 {
			continue
		}

		for _, label := range labelNames {
			delete(toUpdate.Labels, label)
		}

		if len(toUpdate.Labels) < len(node.Labels) {
			err := e.Client.Update(context.TODO(), toUpdate)
			if err != nil {
				return err
			}
		}

		i++
	}

	return
}
