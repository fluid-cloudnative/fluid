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

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
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
	masterReady, err := e.CheckMasterReady()
	if err != nil {
		e.Log.Error(err, "failed  to check if master is ready")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "failed  to update dataset status to \"Failed\"")
		}
		return
	}

	if !masterReady {
		return fmt.Errorf("the master \"%s\" is not healthy, expect at least one replica is ready", e.getMasterName())
	}

	// 2. Check the healthy of the workers
	workerReady, err := e.CheckWorkersReady()
	if err != nil {
		e.Log.Error(err, "failed  to check if workers are ready")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "failed  to update dataset status to \"Failed\"")
		}
		return
	}

	if !workerReady {
		return fmt.Errorf("the worker \"%s\" is not healthy, expect at least one replica is ready", e.getWorkerName())
	}

	// 3. Check the healthy of the fuse
	fuseReady, err := e.checkFuseHealthy()
	if err != nil {
		e.Log.Error(err, "failed  to check fuse is healthy")
		updateErr := e.UpdateDatasetStatus(data.FailedDatasetPhase)
		if updateErr != nil {
			e.Log.Error(updateErr, "failed  to update dataset status to \"Failed\"")
		}
		return
	}

	if !fuseReady {
		// fluid assumes fuse is always ready, so it's a protective branch.
		return fmt.Errorf("the fuse \"%s\" is not healthy", e.getFuseName())
	}

	err = e.UpdateDatasetStatus(data.BoundDatasetPhase)
	if err != nil {
		e.Log.Error(err, "failed  to update dataset status to \"Bound\"")
		return
	}

	return
}

// checkFuseHealthy check fuses number changed
func (e *AlluxioEngine) checkFuseHealthy() (ready bool, err error) {
	getRuntimeFn := func(client client.Client) (base.RuntimeInterface, error) {
		return utils.GetAlluxioRuntime(client, e.name, e.namespace)
	}

	ready, err = e.Helper.CheckAndSyncFuseStatus(getRuntimeFn, types.NamespacedName{Namespace: e.namespace, Name: e.getFuseName()})
	if err != nil {
		e.Log.Error(err, "failed  to check and update fuse status")
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
