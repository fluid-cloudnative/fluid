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

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// The common part of the engine which can be reused
type Helper struct {
	runtimeInfo base.RuntimeInfoInterface

	client client.Client

	log logr.Logger
}

func BuildHelper(runtimeInfo base.RuntimeInfoInterface, client client.Client, log logr.Logger) *Helper {
	return &Helper{
		runtimeInfo: runtimeInfo,
		client:      client,
		log:         log,
	}
}
func (e *Helper) SyncMasterHealthStateToStatus(runtime base.RuntimeInterface, expectReplicas int32, masterSts *appsv1.StatefulSet) (healthy bool) {
	var (
		phase datav1alpha1.RuntimePhase     = kubeclient.GetPhaseFromStatefulset(expectReplicas, *masterSts)
		cond  datav1alpha1.RuntimeCondition = datav1alpha1.RuntimeCondition{}
	)

	switch phase {
	case datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady:
		healthy = true
	default:
		healthy = false
	}

	// must not DeepCopy() here because we'll update runtime's status later
	statusToUpdate := runtime.GetStatus()
	statusToUpdate.MasterPhase = phase
	statusToUpdate.MasterNumberReady = masterSts.Status.ReadyReplicas
	statusToUpdate.CurrentMasterNumberScheduled = masterSts.Status.Replicas
	statusToUpdate.DesiredMasterNumberScheduled = expectReplicas
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	switch phase {
	case datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady:
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
			"The master is ready.", corev1.ConditionTrue)
	case datav1alpha1.RuntimePhaseNotReady:
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
			"The master is not ready.", corev1.ConditionFalse)
	}

	if cond.Type != "" {
		statusToUpdate.Conditions = utils.UpdateRuntimeCondition(statusToUpdate.Conditions, cond)
	}

	return healthy
}

func (e *Helper) SyncWorkerHealthStateToStatus(runtime base.RuntimeInterface, expectReplicas int32, workerSts *appsv1.StatefulSet) (readyOrPartialReady bool) {

	var (
		phase datav1alpha1.RuntimePhase     = kubeclient.GetPhaseFromStatefulset(expectReplicas, *workerSts)
		cond  datav1alpha1.RuntimeCondition = datav1alpha1.RuntimeCondition{}
	)

	switch phase {
	case datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady:
		readyOrPartialReady = true
	default:
		readyOrPartialReady = false
	}

	// update the status as the workers are ready

	statusToUpdate := runtime.GetStatus()
	statusToUpdate.WorkerPhase = phase
	statusToUpdate.WorkerNumberReady = workerSts.Status.ReadyReplicas
	statusToUpdate.WorkerNumberAvailable = workerSts.Status.AvailableReplicas
	statusToUpdate.CurrentWorkerNumberScheduled = workerSts.Status.Replicas
	statusToUpdate.DesiredWorkerNumberScheduled = expectReplicas
	statusToUpdate.WorkerNumberUnavailable = expectReplicas - workerSts.Status.AvailableReplicas
	if statusToUpdate.WorkerNumberUnavailable <= 0 {
		statusToUpdate.WorkerNumberUnavailable = 0
	}
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	switch phase {
	case datav1alpha1.RuntimePhaseReady:
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
			"The workers are ready.", corev1.ConditionTrue)
	case datav1alpha1.RuntimePhasePartialReady:
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
			"The workers are partially ready.", corev1.ConditionTrue)
	case datav1alpha1.RuntimePhaseNotReady:
		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
			"The workers are not ready.", corev1.ConditionFalse)
	}

	if cond.Type != "" {
		statusToUpdate.Conditions =
			utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
				cond)
	}

	return readyOrPartialReady
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

// CheckAndUpdateWorkerStatus checks if workers are ready
// func (e *Helper) CheckAndUpdateWorkerStatus(runtime base.RuntimeInterface, workers *appsv1.StatefulSet) (readyOrPartialReady bool, err error) {

// 	var (
// 		phase datav1alpha1.RuntimePhase     = kubeclient.GetPhaseFromStatefulset(runtime.Replicas(), *workers)
// 		cond  datav1alpha1.RuntimeCondition = datav1alpha1.RuntimeCondition{}
// 	)

// 	switch phase {
// 	case datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady:
// 		readyOrPartialReady = true
// 	default:
// 		e.log.Info("workers are not ready", "phase", phase)
// 	}

// 	// update the status as the workers are ready
// 	oldStatus := runtime.GetStatus().DeepCopy()

// 	statusToUpdate := runtime.GetStatus()
// 	statusToUpdate.WorkerPhase = phase
// 	statusToUpdate.WorkerNumberReady = workers.Status.ReadyReplicas
// 	statusToUpdate.WorkerNumberAvailable = workers.Status.AvailableReplicas
// 	statusToUpdate.CurrentWorkerNumberScheduled = workers.Status.Replicas
// 	statusToUpdate.DesiredWorkerNumberScheduled = runtime.Replicas()
// 	statusToUpdate.WorkerNumberUnavailable = runtime.Replicas() - workers.Status.AvailableReplicas
// 	if statusToUpdate.WorkerNumberUnavailable <= 0 {
// 		statusToUpdate.WorkerNumberUnavailable = 0
// 	}
// 	if len(statusToUpdate.Conditions) == 0 {
// 		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
// 	}

// 	switch phase {
// 	case datav1alpha1.RuntimePhaseReady:
// 		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
// 			"The workers are ready.", corev1.ConditionTrue)
// 	case datav1alpha1.RuntimePhasePartialReady:
// 		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
// 			"The workers are partially ready.", corev1.ConditionTrue)
// 	case datav1alpha1.RuntimePhaseNotReady:
// 		cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
// 			"The workers are not ready.", corev1.ConditionFalse)
// 	}

// 	if cond.Type != "" {
// 		statusToUpdate.Conditions =
// 			utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
// 				cond)
// 	}

// 	if !reflect.DeepEqual(oldStatus, statusToUpdate) {
// 		err = e.client.Status().Update(context.TODO(), runtime)
// 	} else {
// 		e.log.V(1).Info("No need to update runtime status for checking healthy")
// 	}

// 	return
// }
