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

package juicefs

import (
	"context"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (j JuiceFSEngine) CheckMasterReady() (ready bool, err error) {
	// JuiceFS Runtime has no master role
	return true, nil
}

func (j JuiceFSEngine) ShouldSetupMaster() (should bool, err error) {
	runtime, err := j.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.WorkerPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}
	return
}

func (j JuiceFSEngine) SetupMaster() (err error) {
	workerName := j.getWorkerName()

	// 1. Setup
	_, err = kubeclient.GetStatefulSet(j.Client, workerName, j.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		//1. Is not found error
		j.Log.V(1).Info("SetupMaster", "worker", workerName)
		return j.setupMasterInternal()
	} else if err != nil {
		//2. Other errors
		return
	} else {
		//3.The fuse has been set up
		j.Log.V(1).Info("The worker has been set.")
	}

	// 2. Update the status of the runtime
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := j.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()

		runtimeToUpdate.Status.WorkerPhase = datav1alpha1.RuntimePhaseNotReady
		replicas := runtimeToUpdate.Spec.Worker.Replicas
		if replicas == 0 {
			replicas = 1
		}

		// Init selector for worker
		runtimeToUpdate.Status.Selector = j.getWorkerSelectors()
		runtimeToUpdate.Status.DesiredWorkerNumberScheduled = replicas

		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason,
			"The worker is initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return j.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	if err != nil {
		j.Log.Error(err, "Update runtime status")
		return err
	}

	return
}
