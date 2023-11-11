/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package ctrl

import (
	"context"
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
			return e.client.Status().Update(context.TODO(), runtime)
		}
	}

	return

}

// CheckWorkersReady checks if workers are ready
func (e *Helper) CheckWorkersReady(runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	workers *appsv1.StatefulSet) (ready bool, err error) {

	var (
		phase datav1alpha1.RuntimePhase     = kubeclient.GetPhaseFromStatefulset(runtime.Replicas(), *workers)
		cond  datav1alpha1.RuntimeCondition = datav1alpha1.RuntimeCondition{}
	)

	switch phase {
	case datav1alpha1.RuntimePhaseReady, datav1alpha1.RuntimePhasePartialReady:
		ready = true
	default:
		e.log.Info("workers are not ready", "phase", phase)
	}

	// update the status as the workers are ready
	if phase != currentStatus.WorkerPhase {
		statusToUpdate := runtime.GetStatus()
		statusToUpdate.WorkerPhase = phase

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

		if !reflect.DeepEqual(currentStatus, statusToUpdate) {
			err = e.client.Status().Update(context.TODO(), runtime)
		}

	} else {
		e.log.V(1).Info("No need to update runtime status for checking healthy")
	}

	return
}
