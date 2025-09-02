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
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
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
