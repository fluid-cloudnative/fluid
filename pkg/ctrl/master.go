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
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)

// CheckMasterHealthy checks the sts healthy with role
func (e *Helper) CheckMasterHealthy(recorder record.EventRecorder, runtime base.RuntimeInterface,
	currentStatus datav1alpha1.RuntimeStatus,
	sts *appsv1.StatefulSet) (err error) {
	var (
		healthy             bool
		selector            labels.Selector
		unavailablePodNames []types.NamespacedName
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

		selector, err = metav1.LabelSelectorAsSelector(sts.Spec.Selector)
		if err != nil {
			return fmt.Errorf("error converting StatefulSet %s in namespace %s selector: %v", sts.Name, sts.Namespace, err)
		}

		unavailablePodNames, err = kubeclient.GetUnavailablePodNamesForStatefulSet(e.client, sts, selector)
		if err != nil {
			return err
		}

		// 3. Set event
		err = fmt.Errorf("the master %s in %s is not ready. The expected number is %d, the actual number is %d, the unhealthy pods are %v",
			sts.Name,
			sts.Namespace,
			sts.Status.Replicas,
			sts.Status.ReadyReplicas,
			unavailablePodNames)

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
