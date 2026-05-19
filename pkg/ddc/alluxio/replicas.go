/*
Copyright 2023 The Fluid Authors.

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

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/features"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// SyncReplicas syncs the replicas
func (e *AlluxioEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
		if err != nil {
			if errors.IsNotFound(err) {
				cond := utils.NewRuntimeCondition(data.RuntimeWorkersReady, "The workers are not ready.",
					fmt.Sprintf("The statefulset %s in %s is not found, please fix it.",
						e.getWorkerName(),
						e.namespace), corev1.ConditionFalse)

				updateErr := retry.RetryOnConflict(retry.DefaultBackoff, func() error {

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

					runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
					e.Log.Error(err, "the worker are not ready")

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
				totalErr := fmt.Errorf("the master engine does not exist: %v", updateErr)
				return totalErr
			}
			return err
		}
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()

		// When the AdvancedStatefulSet feature is enabled and we detect a scale-in,
		// decommission the targeted workers before the StatefulSet controller
		// terminates them. This gives the Alluxio master a chance to migrate their
		// cached blocks to the surviving workers. The reconciler requeues until the
		// active worker count has dropped to the desired level.
		if utilfeature.DefaultFeatureGate.Enabled(features.AdvancedStatefulSet) &&
			workers.Spec.Replicas != nil &&
			runtime.Replicas() < *workers.Spec.Replicas {

			drained, drainErr := e.drainScalingDownWorkers(runtime.Replicas(), *workers.Spec.Replicas)
			if drainErr != nil {
				return drainErr
			}
			if !drained {
				return fmt.Errorf("workers not yet drained; scale-in to %d replicas will resume on next reconcile",
					runtime.Replicas())
			}
		}

		err = e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
		return err
	})
	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to sync replicas", types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}

// drainScalingDownWorkers decommissions the Alluxio workers that are about to be
// removed when scaling from currentReplicas down to desiredReplicas.
//
// A standard StatefulSet removes the highest-ordinal pods first, so the targets
// are ordinals [desiredReplicas, currentReplicas). The function issues a
// decommission request via the master and returns whether Alluxio's active
// worker count has already dropped to the desired level.
func (e *AlluxioEngine) drainScalingDownWorkers(desiredReplicas, currentReplicas int32) (bool, error) {
	masterPodName, masterContainerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(masterPodName, masterContainerName, e.namespace, e.Log)

	workerRPCPort := e.getWorkerRPCPort()
	workerStsName := e.getWorkerName()

	// Collect RPC addresses of the pods that will be terminated on scale-down.
	var toDecommission []string
	for ord := desiredReplicas; ord < currentReplicas; ord++ {
		podName := fmt.Sprintf("%s-%d", workerStsName, ord)
		pod := &corev1.Pod{}
		if err := e.Client.Get(context.TODO(),
			types.NamespacedName{Name: podName, Namespace: e.namespace}, pod); err != nil {
			if errors.IsNotFound(err) {
				// Pod is already gone; nothing to decommission here.
				continue
			}
			return false, err
		}
		if pod.Status.PodIP == "" {
			e.Log.Info("Worker pod has no IP yet, will retry", "pod", podName)
			return false, nil
		}
		toDecommission = append(toDecommission,
			fmt.Sprintf("%s:%d", pod.Status.PodIP, workerRPCPort))
	}

	if len(toDecommission) == 0 {
		// All targeted pods are already gone from the cluster.
		return true, nil
	}

	if err := fileUtils.DecommissionWorkers(toDecommission); err != nil {
		return false, err
	}

	activeCount, err := fileUtils.CountActiveWorkers()
	if err != nil {
		return false, err
	}

	if int32(activeCount) > desiredReplicas {
		e.Log.Info("Workers are still draining, will retry",
			"activeWorkers", activeCount, "desired", desiredReplicas)
		return false, nil
	}

	return true, nil
}

// getWorkerRPCPort returns the configured Alluxio worker RPC port, falling back
// to the Alluxio default when the runtime does not override it.
func (e *AlluxioEngine) getWorkerRPCPort() int {
	runtime, err := e.getRuntime()
	if err != nil {
		return defaultWorkerRPCPort
	}
	if port, ok := runtime.Spec.Worker.Ports["rpc"]; ok && port > 0 {
		return port
	}
	return defaultWorkerRPCPort
}
