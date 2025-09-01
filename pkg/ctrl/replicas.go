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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// CheckWorkersReady checks if workers are ready
func (e *Helper) SyncReplicas(ctx cruntime.ReconcileRequestContext,
	runtime datav1alpha1.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	workers *appsv1.StatefulSet) (err error) {

	var cond datav1alpha1.RuntimeCondition

	if runtime.Replicas() != currentStatus.DesiredWorkerNumberScheduled {
		// 1. Update scale condition
		statusToUpdate := runtime.GetStatus()
		if len(statusToUpdate.Conditions) == 0 {
			statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
		}

		if runtime.Replicas() < currentStatus.DesiredWorkerNumberScheduled {
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkerScaledIn, datav1alpha1.RuntimeWorkersScaledInReason,
				"The workers scaled in.", corev1.ConditionTrue)
		} else {
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkerScaledOut, datav1alpha1.RuntimeWorkersScaledOutReason,
				"The workers scaled out.", corev1.ConditionTrue)
		}
		statusToUpdate.Conditions =
			utils.UpdateRuntimeCondition(statusToUpdate.Conditions, cond)

		// 2. Record events
		if runtime.Replicas() < currentStatus.DesiredWorkerNumberScheduled {
			ctx.Recorder.Eventf(runtime, corev1.EventTypeNormal, common.Succeed, "Runtime scaled in. current replicas: %d, desired replicas: %d.",
				runtime.Replicas(),
				currentStatus.DesiredWorkerNumberScheduled)
		} else {
			ctx.Recorder.Eventf(runtime, corev1.EventTypeNormal, common.Succeed, "Runtime scaled out. current replicas: %d, desired replicas: %d.",
				runtime.Replicas(),
				currentStatus.DesiredWorkerNumberScheduled)
		}

		// 3. setup the workers for scaling
		err = e.SetupWorkers(runtime, currentStatus, workers)
		if err != nil {
			return
		}

	} else {
		e.log.V(1).Info("Nothing to do")
		return
	}

	return
}
