package engine

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/dataset/lifecycle"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Shutdown and clean up the engine
func (e *CacheEngine) Shutdown() (err error) {
	//1. clean cache
	if e.retryShutdown < e.gracefulShutdownLimits {
		err = e.invokeCleanCache()
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed", "retry times", e.retryShutdown)
			return
		}
	}

	//3. clean components
	err = e.destroyComponents()
	if err != nil {
		return
	}

	//4. clean all related resources
	err = e.cleanAll()
	return err
}

func (e *CacheEngine) invokeCleanCache() error {
	return nil
}

func (e *CacheEngine) releasePorts() error {
	return nil
}

func (e *CacheEngine) destroyComponents() error {
	//  SchedulerMutex only for patch mode
	lifecycle.SchedulerMutex.Lock()
	defer lifecycle.SchedulerMutex.Unlock()

	var (
		nodeList           = &corev1.NodeList{}
		labelExclusiveName = utils.GetExclusiveKey()
		labelName          = utils.GetRuntimeLabelName(false, e.runtimeType, e.namespace, e.name, "")
		labelCommonName    = utils.GetCommonLabelName(false, e.namespace, e.name, "")
		labelFuseName      = utils.GetFuseLabelName(e.namespace, e.name, "")
	)

	labelNames := []string{labelName, labelCommonName, labelFuseName}
	e.Log.Info("check node labels", "labelNames", labelNames)

	fuseLabels, err := labels.Parse(fmt.Sprintf("%s=true", labelFuseName))
	if err != nil {
		return err
	}

	if err := e.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: fuseLabels,
	}); err != nil {
		return err
	}
	nodes := nodeList.Items

	// 1.select the nodes
	for _, node := range nodes {
		if len(node.Labels) == 0 {
			continue
		}

		nodeName := node.Name
		var labelsToModify common.LabelsToModify
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			node, err := kubeclient.GetNode(e.Client, nodeName)
			if err != nil {
				e.Log.Error(err, "Fail to get node", "nodename", nodeName)
				return err
			}

			toUpdate := node.DeepCopy()
			for _, label := range labelNames {
				labelsToModify.Delete(label)
			}

			if val, exist := toUpdate.Labels[labelExclusiveName]; exist &&
				val == fmt.Sprintf("%s_%s", e.namespace, e.name) {
				labelsToModify.Delete(labelExclusiveName)
			}

			modifiedLabels, err := utils.ChangeNodeLabelWithPatchMode(e.Client, toUpdate, labelsToModify)
			if err != nil {
				return err
			}
			e.Log.Info("Destroy worker", "Dataset", e.name, "deleted worker node", node.Name, "removed or updated labels", modifiedLabels)
			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}

func (e *CacheEngine) cleanAll() error {
	return nil
}
