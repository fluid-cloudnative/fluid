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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// CheckWorkersReady checks if workers are ready
func (e *Helper) SyncReplicas(ctx cruntime.ReconcileRequestContext,
	runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	workers *appsv1.StatefulSet) (err error) {

	var cond datav1alpha1.RuntimeCondition

	if runtime.Replicas() != currentStatus.DesiredWorkerNumberScheduled {
		// 1. Update scale condtion
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
