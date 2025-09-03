/*
Copyright 2020 The Fluid Authors.

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

package alluxio

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// CheckRuntimeHealthy checks the healthy of the runtime
func (e *AlluxioEngine) CheckRuntimeHealthy() (err error) {

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
	_, err = e.checkFuseHealthy()
	if err != nil {
		e.Log.Error(err, "The fuse is not healthy")
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
func (e *AlluxioEngine) checkMasterHealthy() (err error) {
	masterName := e.getMasterName()

	healthy := false
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, e.namespace)

	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if master.Status.Replicas != master.Status.ReadyReplicas {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(data.RuntimeMasterReady, "The master is not ready.",
				fmt.Sprintf("The master %s in %s is not ready.", master.Name, master.Namespace), corev1.ConditionFalse)
			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady

			return err
		} else {
			cond := utils.NewRuntimeCondition(data.RuntimeMasterReady, "The master is ready.",
				"The master is ready.", corev1.ConditionTrue)
			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseReady
			healthy = true
		}

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
func (e *AlluxioEngine) checkWorkersHealthy() (err error) {
	// Check the status of workers
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
	if err != nil {
		if fluiderrs.IsDeprecated(err) {
			e.Log.Info("Warning: the current runtime is created by runtime controller before v0.7.0, checking worker health state is not supported. To support these features, please create a new dataset", "details", err)
			e.Recorder.Event(e.runtime, corev1.EventTypeWarning, common.RuntimeDeprecated, "The runtime is created by controllers before v0.7.0, to fully enable latest capabilities, please delete the runtime and create a new one")
			return nil
		}
		return err
	}

	healthy := false
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {

		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if workers.Status.ReadyReplicas == 0 && *workers.Spec.Replicas > 0 {
			// if workers.Status.NumberReady != workers.Status.DesiredNumberScheduled {
			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []data.RuntimeCondition{}
			}
			cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are not ready.",
				fmt.Sprintf("The statefulset %s in %s are not ready, the Unavailable number is %d, please fix it.",
					workers.Name,
					workers.Namespace,
					*workers.Spec.Replicas-workers.Status.ReadyReplicas), corev1.ConditionFalse)

			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}

			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady

			// runtimeToUpdate.Status.DesiredWorkerNumberScheduled
			// runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady

			e.Log.Error(err, "the workers are not ready")
		} else {
			healthy = true
			cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are ready.",
				"The workers are ready", corev1.ConditionTrue)

			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}
			// runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
		}
		// runtimeToUpdate.Status.DesiredWorkerNumberScheduled = int32(workers.Status.DesiredNumberScheduled)
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
func (e *AlluxioEngine) checkFuseHealthy() (ready bool, err error) {
	getRuntimeFn := func(client client.Client) (base.RuntimeInterface, error) {
		return utils.GetAlluxioRuntime(client, e.name, e.namespace)
	}

	ready, err = e.Helper.CheckAndUpdateFuseStatus(getRuntimeFn, types.NamespacedName{Namespace: e.namespace, Name: e.getFuseName()})
	if err != nil {
		e.Log.Error(err, "fail to check and update fuse status")
		return
	}

	if !ready {
		e.Log.Info("fuses are not ready")
	}

	return
}

// checkExistenceOfMaster check engine existed
func (e *AlluxioEngine) checkExistenceOfMaster() (err error) {

	master, masterErr := kubeclient.GetStatefulSet(e.Client, e.getMasterName(), e.namespace)

	if (masterErr != nil && errors.IsNotFound(masterErr)) || *master.Spec.Replicas <= 0 {
		cond := utils.NewRuntimeCondition(data.RuntimeMasterReady, "The master are not ready.",
			fmt.Sprintf("The statefulset %s in %s is not found, or the replicas is <= 0 ,please fix it.",
				e.getMasterName(),
				e.namespace), corev1.ConditionFalse)

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {

			runtime, err := e.getRuntime()
			if err != nil {
				return err
			}

			runtimeToUpdate := runtime.DeepCopy()

			_, oldCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, cond.Type)

			if oldCond == nil || oldCond.Type != cond.Type {
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
						cond)
			}

			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady
			//when master is not ready, the worker and fuse should be not ready.
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
			runtimeToUpdate.Status.FusePhase = data.RuntimePhaseNotReady
			e.Log.Error(err, "the master are not ready")

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				updateErr := e.Client.Status().Update(context.TODO(), runtimeToUpdate)
				if updateErr != nil {
					return updateErr
				}

				updateErr = e.UpdateDatasetStatus(data.FailedDatasetPhase)
				if updateErr != nil {
					e.Log.Error(updateErr, "Failed to update dataset")
					return updateErr
				}
			}

			return err
		})

		//the totalErr promise the sync will return and Requeue
		totalErr := fmt.Errorf("the master engine is not existed %v", err)
		return totalErr
	} else {
		return nil
	}
}
