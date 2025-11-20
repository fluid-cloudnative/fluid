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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// GetWorkersAsStatefulset gets workers as statefulset object. if it returns deprecated errors, it indicates that
// not support anymore.
func GetWorkersAsStatefulset(client client.Client, key types.NamespacedName) (workers *appsv1.StatefulSet, err error) {
	return kubeclient.GetStatefulSet(client, key.Name, key.Namespace)
}

func (e *Helper) GetWorkerNodes() (nodes []corev1.Node, err error) {
	var (
		nodeList = &corev1.NodeList{}
		// runtimeLabel indicates the specific runtime pod is on the node
		// e.g. fluid.io/s-alluxio-default-hbase=true
		runtimeLabelKey = e.runtimeInfo.GetRuntimeLabelName()
	)

	labelNames := []string{runtimeLabelKey}
	e.log.Info("check node labels", "labelNames", labelNames)
	runtimeLabelSelector, err := labels.Parse(fmt.Sprintf("%s=true", runtimeLabelKey))
	if err != nil {
		return
	}

	err = e.client.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: runtimeLabelSelector,
	})
	if err != nil {
		return nodes, err
	}

	nodes = nodeList.Items
	if len(nodes) == 0 {
		e.log.Info("No node with runtime label is found")
		return
	} else {
		e.log.Info("Find the runtime label for nodes", "len", len(nodes))
	}

	return
}

// GetIpAddressesOfWorker gets Ipaddresses from the Worker Node
func (e *Helper) GetIpAddressesOfWorker() (ipAddresses []string, err error) {
	nodes, err := e.GetWorkerNodes()
	if err != nil {
		return
	}
	ipAddresses = kubeclient.GetIpAddressesOfNodes(nodes)
	return
}

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *Helper) SetupWorkers(runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	workers *appsv1.StatefulSet) (err error) {

	var (
		desireReplicas int32 = runtime.Replicas()
		actualReplicas int32 = 0
	)

	if workers.Spec.Replicas != nil {
		actualReplicas = *workers.Spec.Replicas
	}

	if actualReplicas != desireReplicas {
		// workerToUpdate, err := e.buildWorkersAffinity(workers)

		workerToUpdate, err := e.BuildWorkersAffinity(workers)
		if err != nil {
			return err
		}

		workerToUpdate.Spec.Replicas = &desireReplicas
		err = e.client.Update(context.TODO(), workerToUpdate)
		if err != nil {
			return err
		}

		workers = workerToUpdate
	} else {
		e.log.V(1).Info("Nothing to do for syncing")
	}

	if *workers.Spec.Replicas != runtime.GetStatus().DesiredWorkerNumberScheduled {
		// DO NOT DeepCopy here because the status might be updated somewhere else
		statusToUpdate := runtime.GetStatus()

		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason,
			"The workers are initialized.", corev1.ConditionTrue)
		statusToUpdate.Conditions =
			utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
				cond)

		status := *statusToUpdate
		if !reflect.DeepEqual(status, currentStatus) {
			e.log.V(1).Info("Update runtime status", "runtime", fmt.Sprintf("%s/%s", runtime.GetNamespace(), runtime.GetName()))
			return e.client.Status().Update(context.TODO(), runtime)
		}
	}

	return

}

// CheckAndSyncWorkerStatus checks the worker statefulset's status and update it to runtime's status accordingly.
// It returns readyOrPartialReady to indicate if the worker statefulset is (partial) ready or not ready.
func (e *Helper) CheckAndSyncWorkerStatus(getRuntimeFn func(client.Client) (base.RuntimeInterface, error), workerStsNamespacedName types.NamespacedName) (readyOrPartialReady bool, err error) {
	workers, err := GetWorkersAsStatefulset(e.client,
		workerStsNamespacedName)
	if err != nil {
		return readyOrPartialReady, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := getRuntimeFn(e.client)
		if err != nil {
			return err
		}

		oldStatus := runtime.GetStatus().DeepCopy()
		statusToUpdate := runtime.GetStatus()
		var expectReplicas int32
		if workers.Spec.Replicas != nil {
			expectReplicas = *workers.Spec.Replicas
		} else {
			expectReplicas = 0
		}

		statusToUpdate.DesiredWorkerNumberScheduled = expectReplicas
		statusToUpdate.CurrentWorkerNumberScheduled = workers.Status.Replicas
		statusToUpdate.WorkerNumberReady = workers.Status.ReadyReplicas
		statusToUpdate.WorkerNumberAvailable = workers.Status.AvailableReplicas
		statusToUpdate.WorkerNumberUnavailable = workers.Status.Replicas - workers.Status.AvailableReplicas
		if statusToUpdate.WorkerNumberUnavailable < 0 {
			statusToUpdate.WorkerNumberUnavailable = 0
		}

		phase := kubeclient.GetPhaseFromStatefulset(expectReplicas, *workers)
		statusToUpdate.WorkerPhase = phase

		var cond datav1alpha1.RuntimeCondition
		if len(statusToUpdate.Conditions) == 0 {
			statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		switch phase {
		case datav1alpha1.RuntimePhaseReady:
			readyOrPartialReady = true
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
				"The workers are ready.", corev1.ConditionTrue)
		case datav1alpha1.RuntimePhasePartialReady:
			readyOrPartialReady = true
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
				"The workers are partially ready.", corev1.ConditionTrue)
		case datav1alpha1.RuntimePhaseNotReady:
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
				"The workers are not ready.", corev1.ConditionFalse)
		}

		if len(cond.Type) != 0 {
			statusToUpdate.Conditions = utils.UpdateRuntimeCondition(statusToUpdate.Conditions, cond)
		}

		if !reflect.DeepEqual(oldStatus, statusToUpdate) {
			return e.client.Status().Update(context.TODO(), runtime)
		}

		return nil
	})

	if err != nil {
		return false, errors.Wrap(err, "failed to update worker ready status in runtime status")
	}

	return readyOrPartialReady, nil
}
