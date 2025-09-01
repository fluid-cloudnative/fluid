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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// CheckMasterHealthy checks the sts healthy with role
func (e *Helper) CheckMasterHealthy(recorder record.EventRecorder, runtime datav1alpha1.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	sts *appsv1.StatefulSet) (err error) {
	var (
		healthy bool
	)
	if sts.Status.Replicas == sts.Status.ReadyReplicas {
		healthy = true
	}

	statusToUpdate := runtime.GetStatus()
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	if healthy {
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, "The master is ready.",
			"The master is ready.", corev1.ConditionTrue)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.MasterPhase = datav1alpha1.RuntimePhaseReady

	} else {
		// 1. Update the status
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, "The master is not ready.",
			fmt.Sprintf("The master %s in %s is not ready.", sts.Name, sts.Namespace), corev1.ConditionFalse)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.MasterPhase = datav1alpha1.RuntimePhaseNotReady

		// 2. Record the event
		err = fmt.Errorf("the master %s in %s is not ready. The expected number is %d, the actual number is %d",
			sts.Name,
			sts.Namespace,
			sts.Status.Replicas,
			sts.Status.ReadyReplicas)

		recorder.Eventf(runtime, corev1.EventTypeWarning, "MasterUnhealthy", err.Error())

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
