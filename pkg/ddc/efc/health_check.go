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
	"fmt"
	"reflect"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// CheckRuntimeHealthy checks the healthy of the runtime
func (e *EFCEngine) CheckRuntimeHealthy() (err error) {
	// 1. Check the healthy of the master
	err = e.checkMasterHealthy()
	if err != nil {
		e.Log.Error(err, "The master is not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	// 2. Check the healthy of the workers
	err = e.checkWorkersHealthy()
	if err != nil {
		e.Log.Error(err, "The workers are not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	// 3. Check the healthy of the fuse
	err = e.checkFuseHealthy()
	if err != nil {
		e.Log.Error(err, "The fuse is not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	_, err = e.syncWorkersEndpoints()
	if err != nil {
		e.Log.Error(err, "The worker endpoints is not healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "Failed to update dataset")
		}
		return
	}

	updateErr := e.UpdateDatasetStatus(data.BoundDatasetPhase)
	if updateErr != nil {
		e.Log.Error(updateErr, "Failed to update dataset")
	}

	return
}

// checkMasterHealthy checks the master healthy
func (e *EFCEngine) checkMasterHealthy() (err error) {
	healthy := false
	master, err := kubeclient.GetStatefulSet(e.Client, e.getMasterName(), e.namespace)
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
		}

		var cond data.RuntimeCondition
		if master.Status.Replicas != master.Status.ReadyReplicas {
			cond = utils.NewRuntimeCondition(data.RuntimeMasterReady, "The master is not ready.",
				fmt.Sprintf("The master %s in %s is not ready.", master.Name, master.Namespace), v1.ConditionFalse)
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady
			e.Log.Error(err, "the master are not ready")
		} else {
			healthy = true
			cond = utils.NewRuntimeCondition(data.RuntimeMasterReady, "The master is ready.",
				"The master is ready.", v1.ConditionTrue)
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseReady
		}

		runtimeToUpdate.Status.Conditions = utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if err != nil {
				e.Log.Error(err, "Failed to update the runtime")
				return err
			}
		}

		return nil
	})

	if err != nil {
		e.Log.Error(err, "Failed update runtime")
		return err
	}

	if !healthy {
		err = fmt.Errorf("the master %s in %s is not ready. The expected number is %d, the actual number is %d",
			master.Name,
			master.Namespace,
			master.Status.Replicas,
			master.Status.ReadyReplicas)
	}

	return err
}

// checkWorkersHealthy check workers number changed
func (e *EFCEngine) checkWorkersHealthy() (err error) {
	healthy := false
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
		}

		var cond data.RuntimeCondition
		if workers.Status.ReadyReplicas == 0 && *workers.Spec.Replicas > 0 {
			cond = utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are not ready.",
				fmt.Sprintf("The statefulset %s in %s are not ready, the Unavailable number is %d, please fix it.",
					workers.Name,
					workers.Namespace,
					*workers.Spec.Replicas-workers.Status.ReadyReplicas), v1.ConditionFalse)
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
			e.Log.Error(err, "the workers are not ready")
		} else {
			healthy = true
			if workers.Status.ReadyReplicas == *workers.Spec.Replicas {
				cond = utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are ready.",
					"The workers are ready", v1.ConditionTrue)
				runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
			} else {
				cond = utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are partial ready.",
					"The workers are partial ready", v1.ConditionTrue)
				runtimeToUpdate.Status.WorkerPhase = data.RuntimePhasePartialReady
			}
		}

		runtimeToUpdate.Status.Conditions = utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

		runtimeToUpdate.Status.WorkerNumberReady = int32(workers.Status.ReadyReplicas)
		runtimeToUpdate.Status.WorkerNumberAvailable = int32(workers.Status.CurrentReplicas)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			updateErr := e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if updateErr != nil {
				return updateErr
			}
		}

		return err
	})

	if err != nil {
		e.Log.Error(err, "Failed update runtime")
		return err
	}

	if !healthy {
		err = fmt.Errorf("the workers %s in %s are not ready, the unhealthy number %d",
			workers.Name,
			workers.Namespace,
			*workers.Spec.Replicas-workers.Status.ReadyReplicas)
	}

	return err
}

// checkFuseHealthy check fuses number changed
func (e *EFCEngine) checkFuseHealthy() (err error) {
	healthy := false
	fuses, err := e.getDaemonset(e.getFuseName(), e.namespace)
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
		}

		var cond data.RuntimeCondition
		if fuses.Status.NumberUnavailable > 0 ||
			(fuses.Status.DesiredNumberScheduled > 0 && fuses.Status.NumberAvailable == 0) {
			cond = utils.NewRuntimeCondition(data.RuntimeFusesReady, "The Fuses are not ready.",
				fmt.Sprintf("The daemonset %s in %s are not ready, the unhealthy number %d",
					fuses.Name,
					fuses.Namespace,
					fuses.Status.UpdatedNumberScheduled), v1.ConditionFalse)
			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseNotReady
			e.Log.Error(err, "Failed to check the fuse healthy")
		} else {
			healthy = true
			cond = utils.NewRuntimeCondition(data.RuntimeFusesReady, "The Fuses are ready.",
				"The fuses are ready", v1.ConditionFalse)
			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseReady
		}

		runtimeToUpdate.Status.Conditions = utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

		runtimeToUpdate.Status.DesiredFuseNumberScheduled = int32(fuses.Status.DesiredNumberScheduled)

		runtimeToUpdate.Status.FuseNumberReady = int32(fuses.Status.NumberReady)
		runtimeToUpdate.Status.FuseNumberAvailable = int32(fuses.Status.NumberAvailable)
		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			updateErr := e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			if updateErr != nil {
				return updateErr
			}
		}

		return err
	})

	if err != nil {
		e.Log.Error(err, "Failed update runtime")
		return err
	}

	if !healthy {
		err = fmt.Errorf("the daemonset %s in %s are not ready, the unhealthy number %d",
			fuses.Name,
			fuses.Namespace,
			fuses.Status.UpdatedNumberScheduled)
	}
	return err
}
