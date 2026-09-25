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
	"strings"
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
		//
		// runtime.Replicas() > 0 excludes scaling to zero: graceful
		// decommission exists to redistribute cached blocks to surviving
		// workers, and there are none left to redistribute to once every
		// worker is being removed, so there's nothing for it to accomplish -
		// only a guaranteed wait for defaultWorkerDecommissionDeadline before
		// falling back to the same ungraceful scale-down anyway.
		if utilfeature.DefaultFeatureGate.Enabled(features.GracefulWorkerScaleDown) &&
			runtime.Replicas() > 0 && runtime.Replicas() < workers.Status.Replicas {

			decommissionStart, alreadyTracked := getDecommissionStart(runtime)
			if !alreadyTracked {
				decommissionStart = time.Now()
			}

			// A hard failure here (e.g. the master is on an Alluxio version
			// older than 2.9, where "fsadmin decommissionWorker" doesn't
			// exist) is deliberately NOT returned immediately: doing so would
			// bypass the deadline tracking below entirely, since
			// decommissionStart is only ever persisted inside the "not
			// drained" branch. Treating a drain error the same as "not
			// drained yet" ensures it is still bounded by
			// defaultWorkerDecommissionDeadline and eventually degrades to an
			// ungraceful scale-down, instead of stalling forever.
			drained, addresses, drainErr := e.drainScalingDownWorkers(ctx, runtime, runtime.Replicas(), workers.Status.Replicas)

			if !drained {
				elapsed := time.Since(decommissionStart)
				if elapsed > defaultWorkerDecommissionDeadline {
					// A worker that never finishes draining (unhealthy master,
					// unreplicable blocks, unsupported Alluxio version, ...)
					// would otherwise stall scale-down forever. Past the
					// deadline we fall through and proceed anyway so the
					// StatefulSet still converges; any data loss risk this
					// avoided is the same the cluster accepts today without
					// this feature.
					e.Log.Info("Worker decommission exceeded the deadline; forcing scale-down to proceed",
						"elapsed", elapsed, "deadline", defaultWorkerDecommissionDeadline, "lastError", drainErr)
				} else {
					// Refresh (not just initialize) the condition's Status,
					// Reason, and target-address message so
					// drainScalingDownWorkers can tell a confirmed-attempted
					// decommission apart from one that failed before ever
					// reaching the master, or one whose target set has since
					// gone stale (the user scaled down further before the
					// first attempt finished) - all need a retry, not a skip,
					// on the next reconcile. Status must be compared too: a
					// previous attempt that targeted the exact same addresses
					// and has since been cleared (Status=False, Reason/Message
					// unchanged) would otherwise look identical to this new,
					// still-active attempt and never get flipped back to
					// True. decommissionStart, not time.Now(), anchors
					// LastTransitionTime so the deadline clock isn't reset by
					// any of these fields changing.
					newCond := newDecommissioningCondition(decommissionStart, addresses, drainErr)
					_, existingCond := utils.GetRuntimeCondition(runtimeToUpdate.Status.Conditions, data.RuntimeWorkerDecommissioning)
					if existingCond == nil || existingCond.Status != newCond.Status ||
						existingCond.Reason != newCond.Reason || existingCond.Message != newCond.Message {
						runtimeToUpdate.Status.Conditions = utils.UpdateRuntimeCondition(
							runtimeToUpdate.Status.Conditions, newCond)
						if updateErr := e.Client.Status().Update(ctx, runtimeToUpdate); updateErr != nil {
							return updateErr
						}
					}
					if drainErr != nil {
						// Unlike errWorkersNotYetDrained, a real error is
						// logged at Error level by the caller below, so it's
						// visible on every reconcile instead of only once the
						// deadline is hit.
						return drainErr
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
		} else if _, tracked := getDecommissionStart(runtime); tracked {
			// A decommission attempt was tracked but we're no longer in an
			// active graceful scale-down this reconcile - scaled to zero,
			// scaled back up past it before the drain finished, or the
			// feature gate was disabled mid-flight. The condition above only
			// clears on a clean drain-to-completion inside the block, so
			// leaving via any of these other exits would otherwise strand it
			// at Status=True. A future, unrelated scale-down would then read
			// its stale LastTransitionTime as its own decommissionStart,
			// making the drain look like it's already run out the deadline
			// on its very first reconcile and skip straight to forcing scale-
			// down through - silently dropping the graceful guarantee for
			// that scale-down.
			runtimeToUpdate.Status.Conditions = clearDecommissioningCondition(runtimeToUpdate.Status.Conditions)
			if updateErr := e.Client.Status().Update(ctx, runtimeToUpdate); updateErr != nil {
				return updateErr
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
// worker count has already dropped to the desired level, along with the
// target addresses it computed - the caller reuses them to build the
// decommissioning condition instead of recomputing the same pod lookups.
func (e *AlluxioEngine) drainScalingDownWorkers(ctx context.Context, runtime *data.AlluxioRuntime, desiredReplicas, currentReplicas int32) (bool, []string, error) {
	masterPodName, masterContainerName := e.getMasterPodInfo()
	fileUtils := operations.NewAlluxioFileUtils(masterPodName, masterContainerName, e.namespace, e.Log)

	toDecommission, err := e.getDecommissionAddresses(ctx, runtime, desiredReplicas, currentReplicas)
	if err != nil {
		return false, nil, err
	}

	if len(toDecommission) == 0 {
		// All targeted pods are already gone from the cluster, or none of
		// them have been scheduled yet and so hold no data to protect.
		return true, toDecommission, nil
	}

	// alluxio fsadmin decommissionWorker execs into the master pod, which is
	// too heavy to reissue on every reconcile while we wait for a drain to
	// finish. Skip re-requesting it only once a decommission attempt has
	// actually reached the master for exactly this target set - not merely
	// been attempted: a failed attempt (transient network error, master pod
	// restarting, ...) never reached it, so skipping the retry would leave
	// the active worker count unchanged until the deadline forces an
	// ungraceful proceed, even though simply retrying next reconcile would
	// likely succeed. Comparing the target set (not just whether *something*
	// was attempted) also catches the case where the user scaled down
	// further before the first attempt finished, widening toDecommission to
	// include a worker that was never actually told to decommission.
	if !decommissionSucceeded(runtime, toDecommission) {
		if err := fileUtils.DecommissionWorkers(toDecommission); err != nil {
			return false, toDecommission, err
		}
	}

	activeCount, err := fileUtils.CountActiveWorkers()
	if err != nil {
		return false, toDecommission, err
	}

	if int32(activeCount) > desiredReplicas {
		e.Log.Info("Workers are still draining, will retry",
			"activeWorkers", activeCount, "desired", desiredReplicas)
		return false, toDecommission, nil
	}

	return true, toDecommission, nil
}

// getDecommissionAddresses returns the web-port addresses of the worker pods
// that need to be decommissioned to scale down from currentReplicas to
// desiredReplicas.
//
// A standard StatefulSet removes the highest-ordinal pods first, so the
// targets are ordinals [desiredReplicas, currentReplicas). "fsadmin
// decommissionWorker" addresses workers by their web port (not the RPC port
// used elsewhere) because it monitors the worker's workload as exposed on
// the web server. The worker registers with the master under its node's IP
// (see the ALLUXIO_WORKER_HOSTNAME wiring in charts/alluxio, which sources
// alluxio.worker.hostname from status.hostIP), not its pod IP, so that is
// the identity it must be addressed by. The port is read from each pod's own
// container spec (see workerWebPortFromPod) rather than assumed from
// runtime.Spec.Worker.Ports: in host network mode (the default, see
// data.IsHostNetwork) the operator allocates the worker web port dynamically
// whenever spec.worker.ports.web isn't pinned, and that allocated port is
// never written back to the runtime spec.
func (e *AlluxioEngine) getDecommissionAddresses(ctx context.Context, runtime *data.AlluxioRuntime, desiredReplicas, currentReplicas int32) ([]string, error) {
	workerStsName := e.getWorkerName()

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
			return nil, err
		}
		if pod.DeletionTimestamp != nil {
			// Already being torn down. Kubernetes has no distinct "Terminating"
			// Phase - a pod keeps reporting Running throughout its grace
			// period - so without this check a pod from an *earlier,
			// already-successful* drain could be re-targeted here: this
			// engine lowers the StatefulSet's Spec.Replicas only once that
			// earlier drain succeeds, and workers.Status.Replicas (used to
			// decide whether to enter this function at all, see SyncReplicas)
			// stays elevated until the terminating pod is actually gone,
			// which can briefly re-trigger this scan.
			continue
		}
		if pod.Status.HostIP == "" || pod.Status.Phase != corev1.PodRunning {
			// A pod can have a HostIP as soon as it's scheduled, well before
			// its containers actually start (Pending/ContainerCreating), so
			// HostIP alone doesn't mean the Alluxio worker process is up and
			// listening. Such a pod never registered with the master and
			// holds no cached blocks worth migrating; skip it rather than
			// aborting the whole batch so any sibling worker that IS ready
			// still gets decommissioned this reconcile.
			e.Log.Info("Worker pod has no host IP or isn't running yet, skipping decommission for this pod",
				"pod", podName, "phase", pod.Status.Phase)
			continue
		}
		workerWebPort, ok := workerWebPortFromPod(pod)
		if !ok {
			workerWebPort = e.getWorkerWebPort(runtime)
			e.Log.Info("Worker pod has no \"web\" container port, falling back to the configured/default port",
				"pod", podName, "port", workerWebPort)
		}
		hostIP := pod.Status.HostIP
		if strings.Contains(hostIP, ":") {
			// IPv6; needs bracketing in a host:port pair.
			hostIP = "[" + hostIP + "]"
		}
		addr := fmt.Sprintf("%s:%d", hostIP, workerWebPort)
		if _, dup := seen[addr]; dup {
			continue
		}
		seen[addr] = struct{}{}
		toDecommission = append(toDecommission, addr)
	}
	return toDecommission, nil
}

// getWorkerWebPort returns the runtime-spec-configured Alluxio worker web
// port, falling back to the Alluxio default when the runtime does not
// override it. This is only a fallback for getDecommissionAddresses, used
// when a worker pod's own container spec doesn't expose a "web" port:
// spec.worker.ports.web reflects the actual bound port only when the user
// pinned it, since in host network mode (the default) an unpinned port is
// allocated dynamically and never written back to the spec - see
// workerWebPortFromPod, which reads the port the worker container actually
// binds.
func (e *AlluxioEngine) getWorkerWebPort(runtime *data.AlluxioRuntime) int {
	if port, ok := runtime.Spec.Worker.Ports["web"]; ok && port > 0 {
		return port
	}
	return defaultWorkerWebPort
}

// workerWebPortFromPod returns the web port actually bound by the given
// worker pod's alluxio-worker container, and whether one was found. This is
// the source of truth for the port decommissionWorker must address: unlike
// runtime.Spec.Worker.Ports, it reflects the port in effect regardless of
// network mode or whether it was pinned or dynamically allocated (see
// getDecommissionAddresses). Matching must be scoped to the alluxio-worker
// container specifically - alluxio-job-worker also exposes a differently
// numbered port of the same name "web" (see
// charts/alluxio/templates/worker/statefulset.yaml).
func workerWebPortFromPod(pod *corev1.Pod) (int, bool) {
	for _, container := range pod.Spec.Containers {
		if container.Name != alluxioWorkerContainerName {
			continue
		}
		for _, port := range container.Ports {
			if port.Name == "web" {
				return int(port.ContainerPort), true
			}
		}
	}
	return 0, false
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

// decommissionSucceeded reports whether the most recently tracked
// decommission attempt actually reached the Alluxio master for exactly the
// given target addresses, as opposed to merely having been attempted, or
// having been attempted for a now-stale target set. A tracked-but-failed
// attempt (Reason == RuntimeWorkerDecommissionFailedReason) must still be
// retried, not skipped. So must an attempt tracked for a different address
// set: if the user scales down further before the first attempt finishes,
// drainScalingDownWorkers's target set widens, and a worker newly added to
// it was never actually told to decommission.
func decommissionSucceeded(runtime *data.AlluxioRuntime, addresses []string) bool {
	_, cond := utils.GetRuntimeCondition(runtime.Status.Conditions, data.RuntimeWorkerDecommissioning)
	if cond == nil || cond.Status != corev1.ConditionTrue || cond.Reason == data.RuntimeWorkerDecommissionFailedReason {
		return false
	}
	return cond.Message == decommissionTargetsMessage(addresses)
}

// decommissionTargetsMessage renders the addresses a decommission attempt
// targeted, in the canonical form newDecommissioningCondition and
// decommissionSucceeded both use to detect whether the target set has
// changed since the last recorded attempt.
func decommissionTargetsMessage(addresses []string) string {
	return fmt.Sprintf("Workers being decommissioned ahead of a scale-down: %s", strings.Join(addresses, ","))
}

// newDecommissioningCondition marks the state of a worker-drain attempt that
// didn't complete within one reconcile, so subsequent reconciles can measure
// elapsed time against defaultWorkerDecommissionDeadline and
// drainScalingDownWorkers can tell whether the last attempt needs retrying -
// either because it failed, or because the target address set has since
// changed.
func newDecommissioningCondition(start time.Time, addresses []string, drainErr error) data.RuntimeCondition {
	reason := data.RuntimeWorkerDecommissioningReason
	message := decommissionTargetsMessage(addresses)
	if drainErr != nil {
		reason = data.RuntimeWorkerDecommissionFailedReason
		message = fmt.Sprintf("Worker decommission attempt failed and will be retried: %v", drainErr)
	}
	cond := utils.NewRuntimeCondition(data.RuntimeWorkerDecommissioning, reason, message, corev1.ConditionTrue)
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
