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

package efc

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
)

func (e *EFCEngine) ShouldSetupMaster() (should bool, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.MasterPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}

	return
}

// SetupMaster setups the master and updates the status
// It will print the information in the Debug window according to the Master status
// It return any cache error encountered
func (e *EFCEngine) SetupMaster() (err error) {
	// 1. Setup the efc cluster
	masterName := e.getMasterName()
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		// Is not found error
		e.Log.V(1).Info("SetupMaster", "master", e.getMasterName())
		return e.setupMasterInternal()
	} else if err != nil {
		// Other errors
		return
	} else {
		// The master has been set up
		e.Log.V(1).Info("The master has been set.", "replicas", master.Status.ReadyReplicas)
	}

	// 2. Update the status of the runtime
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()

		runtimeToUpdate.Status.MasterPhase = datav1alpha1.RuntimePhaseNotReady
		replicas := runtimeToUpdate.MasterReplicas()

		// Init selector for worker
		runtimeToUpdate.Status.Selector = e.getWorkerSelectors()

		runtimeToUpdate.Status.DesiredMasterNumberScheduled = replicas
		runtimeToUpdate.Status.ValueFileConfigmap = e.getConfigmapName()

		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterInitialized, datav1alpha1.RuntimeMasterInitializedReason,
			"The master is initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	if err != nil {
		e.Log.Error(err, "Update runtime status")
		return err
	}

	return
}

func (e *EFCEngine) CheckMasterReady() (ready bool, err error) {
	// 1. Check the status
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	master, err := kubeclient.GetStatefulSet(e.Client, e.getMasterName(), e.namespace)
	if err != nil {
		return
	}

	masterReplicas := runtime.MasterReplicas()
	if masterReplicas == master.Status.ReadyReplicas {
		ready = true
	} else {
		e.Log.Info("The master is not ready.", "replicas", masterReplicas,
			"readyReplicas", master.Status.ReadyReplicas)
	}

	// 2. Update the phase
	if ready {
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			runtime, err := e.getRuntime()
			if err != nil {
				return err
			}
			runtimeToUpdate := runtime.DeepCopy()

			runtimeToUpdate.Status.CurrentMasterNumberScheduled = int32(master.Status.ReadyReplicas)
			runtimeToUpdate.Status.MasterPhase = datav1alpha1.RuntimePhaseReady

			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeMasterReady, datav1alpha1.RuntimeMasterReadyReason,
				"The master is ready.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
					cond)

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			}

			return nil
		})

		if err != nil {
			e.Log.Error(err, "Update runtime status")
			return
		}
	}

	if err != nil {
		return
	}

	return
}
