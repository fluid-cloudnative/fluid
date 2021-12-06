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
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// CheckFuseHealthy checks the ds healthy with role
func (e *Helper) CheckFuseHealthy(runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	ds *appsv1.DaemonSet, recorder record.EventRecorder) (err error) {
	var healthy bool
	if ds.Status.NumberUnavailable == 0 {
		healthy = true
	}

	statusToUpdate := runtime.GetStatus()
	if len(statusToUpdate.Conditions) == 0 {
		statusToUpdate.Conditions = []datav1alpha1.RuntimeCondition{}
	}

	if healthy {
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesReady, "The fuse is ready.",
			"The fuse is ready.", corev1.ConditionTrue)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.FusePhase = datav1alpha1.RuntimePhaseReady
	} else {
		// 1. Update the status
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesReady, "The fuses are not ready.",
			fmt.Sprintf("The fuses %s in %s are not ready.", ds.Name, ds.Namespace), corev1.ConditionFalse)
		_, oldCond := utils.GetRuntimeCondition(statusToUpdate.Conditions, cond.Type)

		if oldCond == nil || oldCond.Type != cond.Type {
			statusToUpdate.Conditions =
				utils.UpdateRuntimeCondition(statusToUpdate.Conditions,
					cond)
		}
		statusToUpdate.FusePhase = datav1alpha1.RuntimePhaseNotReady

		// 2. Record the event
		unavailablePodNames, err := kubeclient.GetUnavailableDaemonPods(e.client, ds)
		if err != nil {
			return err
		}

		// 3. Set error
		msg := fmt.Sprintf("the fuse %s in %s are not ready. The expected number is %d, the actual number is %d, the unhealthy pods are %v",
			ds.Name,
			ds.Namespace,
			ds.Status.DesiredNumberScheduled,
			ds.Status.NumberReady,
			unavailablePodNames)

		recorder.Eventf(runtime, corev1.EventTypeWarning, "FuseUnhealthy", msg)
	}

	if err != nil {
		return
	}

	status := *statusToUpdate
	if !reflect.DeepEqual(status, currentStatus) {
		return e.client.Status().Update(context.TODO(), runtime)
	}

	return
}
