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
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// CheckFuseHealthy checks the ds healthy with role
func (e *Helper) CheckFuseHealthy(recorder record.EventRecorder, runtime base.RuntimeInterface, fuseName string) error {
	currentStatus := runtime.GetStatus()

	runtimeInfo := e.runtimeInfo
	ds, err := kubeclient.GetDaemonset(e.client, fuseName, runtimeInfo.GetNamespace())
	if err != nil {
		e.log.Error(err, "Failed to get Fuse Daemonset", "fuseDaemonsetName", ds.Name, "fuseDaemonsetNamespace", ds.Namespace)
		return err
	}

	healthy := true
	if ds.Status.NumberUnavailable > 0 ||
		(ds.Status.DesiredNumberScheduled > 0 && ds.Status.NumberAvailable == 0) {
		healthy = false
	}

	updateRuntimeStatus := func(fusePhase datav1alpha1.RuntimePhase,
		reason, message string) error {
		statusToUpdate := runtime.GetStatus()

		conditionStatus := corev1.ConditionFalse
		if fusePhase == datav1alpha1.RuntimePhaseReady {
			conditionStatus = corev1.ConditionTrue
		}

		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesReady, reason, message, conditionStatus)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.FusePhase = fusePhase
		statusToUpdate.FuseNumberReady = int32(ds.Status.NumberReady)
		statusToUpdate.FuseNumberAvailable = int32(ds.Status.NumberAvailable)
		statusToUpdate.FuseNumberUnavailable = int32(ds.Status.NumberUnavailable)
		if !reflect.DeepEqual(*statusToUpdate, currentStatus) {
			e.log.V(1).Info("Update runtime status", "runtime", fmt.Sprintf("%s/%s", runtime.GetNamespace(), runtime.GetName()))
			return e.client.Status().Update(context.TODO(), runtime)
		}
		return nil
	}

	if healthy {
		if err := updateRuntimeStatus(
			datav1alpha1.RuntimePhaseReady,
			"The fuse is ready.",
			"The fuse is ready."); err != nil {
			return err
		}
		return nil
	}

	// handle unhealthy case
	// 1. Update the status
	if err := updateRuntimeStatus(
		datav1alpha1.RuntimePhaseNotReady,
		"The fuses are not ready.",
		fmt.Sprintf("The fuses %s in %s are not ready.", ds.Name, ds.Namespace)); err != nil {
		return err
	}

	// 2. Record the event
	unavailablePodNames, err := kubeclient.GetUnavailableDaemonPodNames(e.client, ds)
	if err != nil {
		e.log.Error(err, "Failed to get UnavailableDaemonPodNames", "fuseDaemonsetName", ds.Name, "fuseDaemonsetNamespace", ds.Namespace)
		return err
	}
	recorder.Eventf(runtime, corev1.EventTypeWarning, "FuseUnhealthy",
		fmt.Errorf("the fuse %s in %s are not ready. The expected number is %d, the actual number is %d, the unhealthy pods are %v",
			ds.Name,
			ds.Namespace,
			ds.Status.DesiredNumberScheduled,
			ds.Status.NumberReady,
			unavailablePodNames).Error())

	return nil
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
