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

	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/utils/ptr"
)

// GetWorkersAsStatefulset gets workers as statefulset object. if it returns deprecated errors, it indicates that
// not support anymore.
func GetWorkersAsStatefulset(client client.Client, key types.NamespacedName) (workers *appsv1.StatefulSet, err error) {
	workers, err = kubeclient.GetStatefulSet(client, key.Name, key.Namespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			_, dsErr := kubeclient.GetDaemonset(client, key.Name, key.Namespace)
			// return workers, fluiderr.NewDeprecated()
			// find the daemonset successfully
			if dsErr == nil {
				return workers, fluiderrs.NewDeprecated(schema.GroupResource{
					Group:    appsv1.SchemeGroupVersion.Group,
					Resource: "daemonsets",
				}, key)
			}
		}
	}

	return
}

// CheckworkersHealthy checks the sts healthy with role
func (e *Helper) CheckWorkersHealthy(recorder record.EventRecorder, runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	sts *appsv1.StatefulSet) (err error) {
	var (
		healthy bool
	)

	if sts.Spec.Replicas == ptr.To[int32](0) || sts.Status.ReadyReplicas > 0 {
		healthy = true
	}

	statusToUpdate := runtime.GetStatus()
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	if healthy {
		var cond datav1alpha1.RuntimeCondition
		if sts.Status.Replicas == sts.Status.ReadyReplicas {
			statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhaseReady
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, "The worker are ready.",
				"The worker are ready.", corev1.ConditionTrue)
		} else {
			statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhasePartialReady
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, "The worker are partial ready.",
				"The worker are partial ready.", corev1.ConditionTrue)
		}
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)
		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
	} else {
		// 1. Update the status
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, "The workers are not ready.",
			fmt.Sprintf("The workers %s in %s are not ready.", sts.Name, sts.Namespace), corev1.ConditionFalse)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhaseNotReady

		// 2. Record the event
		err = fmt.Errorf("the workers %s in namespace %s are not ready. The expected number is %d, the actual number is %d",
			sts.Name,
			sts.Namespace,
			sts.Status.Replicas,
			sts.Status.ReadyReplicas)

		recorder.Eventf(runtime, corev1.EventTypeWarning, "WorkersUnhealthy", err.Error())
	}

	status := *statusToUpdate
	if !reflect.DeepEqual(status, currentStatus) {
		updateErr := e.client.Status().Update(context.TODO(), runtime)
		if updateErr != nil {
			return updateErr
		}
	}

	if err != nil {
		return
	}

	return

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

	desireReplicas := runtime.Replicas()
	if *workers.Spec.Replicas != desireReplicas {
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
		statusToUpdate := runtime.GetStatus()

		if workers.Status.ReadyReplicas > 0 {
			if runtime.Replicas() == workers.Status.ReadyReplicas {
				statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhaseReady
			} else {
				statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhasePartialReady
			}
		} else {
			statusToUpdate.WorkerPhase = datav1alpha1.RuntimePhaseNotReady
		}

		statusToUpdate.DesiredWorkerNumberScheduled = runtime.Replicas()
		statusToUpdate.CurrentWorkerNumberScheduled = statusToUpdate.DesiredWorkerNumberScheduled

		if len(statusToUpdate.Conditions) == 0 {
			statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
		}
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

// CheckAndUpdateWorkerStatus checks the worker statefulset's status and update it to runtime's status accordingly.
// It returns readyOrPartialReady to indicate if the worker statefulset is (partial) ready or not ready.
func (e *Helper) CheckAndUpdateWorkerStatus(getRuntimeFn func(client.Client) (base.RuntimeInterface, error), workerStsNamespacedName types.NamespacedName) (readyOrPartialReady bool, err error) {
	workers, err := GetWorkersAsStatefulset(e.client,
		workerStsNamespacedName)
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			e.log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
			readyOrPartialReady = true
			return readyOrPartialReady, nil
		}
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
