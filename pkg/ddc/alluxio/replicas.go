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
	stderrors "errors"
	"fmt"
	"reflect"
	"time"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/alluxio/operations"
	"github.com/fluid-cloudnative/fluid/pkg/features"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// errWorkersNotYetDrained marks the normal, transient state during scale-in
// where the targeted workers have not finished migrating their cached blocks
// to the surviving workers yet. It lets the caller log this at Info level
// instead of Error, while still propagating a non-nil error so the existing
// fixed-interval reconcile requeue (see runtime_controller.go) kicks in.
var errWorkersNotYetDrained = stderrors.New("workers not yet drained")

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

		// When the GracefulWorkerScaleDown feature is enabled and we detect a
		// scale-in, decommission the targeted workers before the StatefulSet
		// controller terminates them. This gives the Alluxio master a chance to
		// migrate their cached blocks to the surviving workers. The reconciler
		// requeues until the active worker count has dropped to the desired
		// level.
		//
		// workers.Status.Replicas (the number of Pods the StatefulSet controller
		// has actually created) is used rather than workers.Spec.Replicas: the
		// spec is the target this engine itself lowers once a drain succeeds, so
		// relying on it could under-count pods that still exist but whose spec
		// update already landed.
		if utilfeature.DefaultFeatureGate.Enabled(features.GracefulWorkerScaleDown) &&
			runtime.Replicas() < workers.Status.Replicas {

			decommissionStart, alreadyTracked := getDecommissionStart(runtime)
			if !alreadyTracked {
				decommissionStart = time.Now()
			}

			drained, drainErr := e.drainScalingDownWorkers(ctx, runtime, runtime.Replicas(), workers.Status.Replicas)
			if drainErr != nil {
				return drainErr
			}

			if !drained {
				elapsed := time.Since(decommissionStart)
				if elapsed > defaultWorkerDecommissionDeadline {
					// A worker that never finishes draining (unhealthy master,
					// unreplicable blocks, ...) would otherwise stall scale-down
					// forever. Past the deadline we fall through and proceed
					// anyway so the StatefulSet still converges; any data loss
					// risk this avoided is the same the cluster accepts today
					// without this feature.
					e.Log.Info("Worker decommission exceeded the deadline; forcing scale-down to proceed",
						"elapsed", elapsed, "deadline", defaultWorkerDecommissionDeadline)
				} else {
					if !alreadyTracked {
						runtimeToUpdate.Status.Conditions = utils.UpdateRuntimeCondition(
							runtimeToUpdate.Status.Conditions, newDecommissioningCondition(decommissionStart))
						if updateErr := e.Client.Status().Update(ctx, runtimeToUpdate); updateErr != nil {
							return updateErr
						}
					}
					return fmt.Errorf("%w: scale-in to %d replicas will resume on next reconcile",
						errWorkersNotYetDrained, runtime.Replicas())
				}
			}

			if alreadyTracked {
				runtimeToUpdate.Status.Conditions = clearDecommissioningCondition(runtimeToUpdate.Status.Conditions)
				if updateErr := e.Client.Status().Update(ctx, runtimeToUpdate); updateErr != nil {
					return updateErr
				}
			}
		}

		err = e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
		return err
	})
	if err != nil {
		if stderrors.Is(err, errWorkersNotYetDrained) {
			e.Log.Info(err.Error(), "name", e.name, "namespace", e.namespace)
		} else {
			_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to sync replicas", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		}
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
func (e *AlluxioEngine) drainScalingDownWorkers(ctx context.Context, runtime *data.AlluxioRuntime, desiredReplicas, currentReplicas int32) (bool, error) {
	masterPodName, masterContainerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(masterPodName, masterContainerName, e.namespace, e.Log)

	workerRPCPort := e.getWorkerRPCPort(runtime)
	workerStsName := e.getWorkerName()

	// Collect RPC addresses of the pods that will be terminated on scale-down.
	// The worker registers with the master under its node's IP (see the
	// ALLUXIO_WORKER_HOSTNAME wiring in charts/alluxio, which sources
	// alluxio.worker.hostname from status.hostIP), not its pod IP, so that is
	// the identity "fsadmin decommissionWorker" must be addressed by.
	//
	// Pods sharing a node produce the same HostIP; seen tracks addresses
	// already added so the request doesn't list the same worker twice.
	var toDecommission []string
	seen := make(map[string]struct{})
	for ord := desiredReplicas; ord < currentReplicas; ord++ {
		podName := fmt.Sprintf("%s-%d", workerStsName, ord)
		pod := &corev1.Pod{}
		if err := e.Client.Get(ctx,
			types.NamespacedName{Name: podName, Namespace: e.namespace}, pod); err != nil {
			if errors.IsNotFound(err) {
				// Pod is already gone; nothing to decommission here.
				continue
			}
			return false, err
		}
		if pod.Status.HostIP == "" {
			e.Log.Info("Worker pod has no host IP yet, will retry", "pod", podName)
			return false, nil
		}
		addr := fmt.Sprintf("%s:%d", pod.Status.HostIP, workerRPCPort)
		if _, dup := seen[addr]; dup {
			continue
		}
		seen[addr] = struct{}{}
		toDecommission = append(toDecommission, addr)
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
func (e *AlluxioEngine) getWorkerRPCPort(runtime *data.AlluxioRuntime) int {
	if port, ok := runtime.Spec.Worker.Ports["rpc"]; ok && port > 0 {
		return port
	}
	return defaultWorkerRPCPort
}

// getDecommissionStart returns when the current worker-drain attempt began,
// based on the RuntimeWorkerDecommissioning condition set the first time a
// scale-down's drain didn't finish within one reconcile. The bool reports
// whether such an in-progress attempt is already being tracked.
func getDecommissionStart(runtime *data.AlluxioRuntime) (time.Time, bool) {
	_, cond := utils.GetRuntimeCondition(runtime.Status.Conditions, data.RuntimeWorkerDecommissioning)
	if cond == nil || cond.Status != corev1.ConditionTrue {
		return time.Time{}, false
	}
	return cond.LastTransitionTime.Time, true
}

// newDecommissioningCondition marks the start of a worker-drain attempt that
// didn't complete within one reconcile, so subsequent reconciles can measure
// elapsed time against defaultWorkerDecommissionDeadline.
func newDecommissioningCondition(start time.Time) data.RuntimeCondition {
	cond := utils.NewRuntimeCondition(data.RuntimeWorkerDecommissioning, data.RuntimeWorkerDecommissioningReason,
		"Workers are being decommissioned ahead of a scale-down.", corev1.ConditionTrue)
	cond.LastTransitionTime = metav1.NewTime(start)
	return cond
}

// clearDecommissioningCondition marks a tracked drain attempt as finished,
// whether because it succeeded or because defaultWorkerDecommissionDeadline
// forced the scale-down to proceed anyway.
func clearDecommissioningCondition(conditions []data.RuntimeCondition) []data.RuntimeCondition {
	idx, cond := utils.GetRuntimeCondition(conditions, data.RuntimeWorkerDecommissioning)
	if cond == nil {
		return conditions
	}
	cleared := *cond
	cleared.Status = corev1.ConditionFalse
	cleared.LastTransitionTime = metav1.Now()
	conditions[idx] = cleared
	return conditions
}
