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

package ctrl

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (e *Helper) CheckAndSyncFuseStatus(getRuntimeFn func(client.Client) (base.RuntimeInterface, error), fuseDsNamespacedName types.NamespacedName) (ready bool, err error) {
	fuseDs, err := kubeclient.GetDaemonset(e.client, fuseDsNamespacedName.Name, fuseDsNamespacedName.Namespace)
	if err != nil {
		return
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := getRuntimeFn(e.client)
		if err != nil {
			return err
		}

		oldStatus := runtime.GetStatus().DeepCopy()
		statusToUpdate := runtime.GetStatus()

		statusToUpdate.DesiredFuseNumberScheduled = fuseDs.Status.DesiredNumberScheduled
		statusToUpdate.CurrentFuseNumberScheduled = fuseDs.Status.CurrentNumberScheduled
		statusToUpdate.FuseNumberReady = fuseDs.Status.NumberReady
		statusToUpdate.FuseNumberAvailable = fuseDs.Status.NumberAvailable
		statusToUpdate.FuseNumberUnavailable = fuseDs.Status.NumberUnavailable

		// fluid assumes fuse components are always ready
		statusToUpdate.FusePhase = datav1alpha1.RuntimePhaseReady
		ready = true

		if len(statusToUpdate.Conditions) == 0 {
			statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesReady, datav1alpha1.RuntimeFusesReadyReason, "The fuses are ready.", corev1.ConditionTrue)
		statusToUpdate.FuseReason = cond.Reason
		statusToUpdate.Conditions = utils.UpdateRuntimeCondition(statusToUpdate.Conditions, cond)

		if !reflect.DeepEqual(oldStatus, statusToUpdate) {
			return e.client.Status().Update(context.TODO(), runtime)
		}

		return nil
	})

	if err != nil {
		return false, errors.Wrapf(err, "failed to update fuse ready status in runtime status")
	}

	return ready, nil
}

// CleanUpFuse will cleanup node label for Fuse.
func (e *Helper) CleanUpFuse() (count int, err error) {
	var (
		nodeList     = &corev1.NodeList{}
		fuseLabelKey = utils.GetFuseLabelName(e.runtimeInfo.GetNamespace(), e.runtimeInfo.GetName(), e.runtimeInfo.GetOwnerDatasetUID())
	)

	labelNames := []string{fuseLabelKey}
	e.log.Info("check node labels", "labelNames", labelNames)
	fuseLabelSelector, err := labels.Parse(fmt.Sprintf("%s=true", fuseLabelKey))
	if err != nil {
		return
	}

	err = e.client.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: fuseLabelSelector,
	})
	if err != nil {
		return count, err
	}

	nodes := nodeList.Items
	if len(nodes) == 0 {
		e.log.Info("No node with fuse label need to be delete")
		return
	} else {
		e.log.Info("Try to clean the fuse label for nodes", "len", len(nodes))
	}

	var labelsToModify common.LabelsToModify
	labelsToModify.Delete(fuseLabelKey)

	for _, node := range nodes {
		_, err = utils.ChangeNodeLabelWithPatchMode(e.client, &node, labelsToModify)
		if err != nil {
			e.log.Error(err, "Error when patching labels on node", "nodeName", node.Name)
			return count, errors.Wrapf(err, "error when patching labels on node %s", node.Name)
		}
		count++
	}

	return
}

// GetFuseNodes gets the node of fuses
func (e *Helper) GetFuseNodes() (nodes []corev1.Node, err error) {
	var (
		nodeList     = &corev1.NodeList{}
		fuseLabelKey = utils.GetFuseLabelName(e.runtimeInfo.GetNamespace(), e.runtimeInfo.GetName(), e.runtimeInfo.GetOwnerDatasetUID())
	)

	labelNames := []string{fuseLabelKey}
	e.log.V(1).Info("check node labels", "labelNames", labelNames)
	fuseLabelSelector, err := labels.Parse(fmt.Sprintf("%s=true", fuseLabelKey))
	if err != nil {
		return
	}

	err = e.client.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: fuseLabelSelector,
	})
	if err != nil {
		return nodes, err
	}

	nodes = nodeList.Items
	if len(nodes) == 0 {
		e.log.Info("No node with fuse label is found")
		return
	} else {
		e.log.Info("Find the fuse label for nodes", "len", len(nodes))
	}

	return
}

// GetIpAddressesOfFuse gets Ipaddresses from the Fuse Node
func (e *Helper) GetIpAddressesOfFuse() (ipAddresses []string, err error) {
	nodes, err := e.GetFuseNodes()
	if err != nil {
		return
	}
	ipAddresses = kubeclient.GetIpAddressesOfNodes(nodes)
	return
}
