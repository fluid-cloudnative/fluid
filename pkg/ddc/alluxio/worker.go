/*

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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

// SetupWorkers checks the desired and current replicas of workers and makes an update
// over the status by setting phases and conditions. The function
// calls for a status update and finally returns error if anything unexpected happens.
func (e *AlluxioEngine) SetupWorkers() (err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		e.Log.Error(err, "setupWorker")
		return err
	}

	replicas := runtime.Replicas()

	currentReplicas, err := e.AssignNodesToCache(replicas)
	if err != nil {
		return err
	}

	e.Log.Info("check the desired and current replicas",
		"desriedReplicas", replicas,
		"currentReplicas", currentReplicas)

	if currentReplicas == 0 {
		return fmt.Errorf("the number of the current workers which can be scheduled is 0")
	}

	// 2. Update the status
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			e.Log.Error(err, "setupWorker")
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()

		runtimeToUpdate.Status.WorkerPhase = datav1alpha1.RuntimePhaseNotReady
		runtimeToUpdate.Status.DesiredWorkerNumberScheduled = replicas
		runtimeToUpdate.Status.CurrentWorkerNumberScheduled = currentReplicas
		runtimeToUpdate.Status.FusePhase = datav1alpha1.RuntimePhaseNotReady

		if runtimeToUpdate.Spec.Fuse.Global {
			fuseName := e.getFuseDaemonsetName()
			fuses, err := e.getDaemonset(fuseName, e.namespace)
			if err != nil {
				e.Log.Error(err, "setupWorker")
				return err
			}

			// Clean the label to start the daemonset deployment
			fusesToUpdate := fuses.DeepCopy()
			e.Log.Info("check node labels of fuse before cleaning balloon key", "labels", fusesToUpdate.Spec.Template.Spec.NodeSelector)
			delete(fusesToUpdate.Spec.Template.Spec.NodeSelector, common.FLUID_FUSE_BALLOON_KEY)
			e.Log.Info("check node labels of fuse after cleaning balloon key", "labels", fusesToUpdate.Spec.Template.Spec.NodeSelector)
			err = e.Client.Update(context.TODO(), fusesToUpdate)
			if err != nil {
				e.Log.Error(err, "setupWorker")
				return err
			}
			runtimeToUpdate.Status.DesiredFuseNumberScheduled = fuses.Status.DesiredNumberScheduled
			runtimeToUpdate.Status.CurrentFuseNumberScheduled = fuses.Status.CurrentNumberScheduled
		} else {
			runtimeToUpdate.Status.DesiredFuseNumberScheduled = replicas
			runtimeToUpdate.Status.CurrentFuseNumberScheduled = currentReplicas
		}
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersInitialized, datav1alpha1.RuntimeWorkersInitializedReason,
			"The workers are initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)
		fuseCond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesInitialized, datav1alpha1.RuntimeFusesInitializedReason,
			"The fuses are initialized.", corev1.ConditionTrue)
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				fuseCond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	return
}

// ShouldSetupWorkers checks if we need setup the workers
func (e *AlluxioEngine) ShouldSetupWorkers() (should bool, err error) {
	runtime, err := e.getRuntime()
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

// are the workers ready
func (e *AlluxioEngine) CheckWorkersReady() (ready bool, err error) {
	var (
		workerReady, fuseReady, workerPartialReady, fusePartialReady bool
		workerName                                                   string = e.getWorkerDaemonsetName()
		fuseName                                                     string = e.getFuseDaemonsetName()
		namespace                                                    string = e.namespace
	)

	runtime, err := e.getRuntime()
	if err != nil {
		return ready, err
	}

	// runtime, err := e.getRuntime()
	// if err != nil {
	// 	return
	// }

	workers, err := e.getDaemonset(workerName, namespace)
	if err != nil {
		return ready, err
	}

	// replicas := runtime.Replicas()

	if workers.Status.NumberReady > 0 {
		if runtime.Replicas() == workers.Status.NumberReady {
			workerReady = true
		} else if workers.Status.NumberReady >= 1 {
			workerPartialReady = true
		}
	}

	e.Log.Info("Fuse deploy mode", "global", runtime.Spec.Fuse.Global)
	fuses, err := e.getDaemonset(fuseName, namespace)
	if fuses.Status.NumberAvailable > 0 {
		if runtime.Spec.Fuse.Global {
			if fuses.Status.DesiredNumberScheduled == fuses.Status.CurrentNumberScheduled {
				fuseReady = true
			} else {
				fusePartialReady = true
			}
		} else {
			if runtime.Spec.Replicas == fuses.Status.NumberReady {
				fuseReady = true
			} else if fuses.Status.NumberReady >= 1 {
				fusePartialReady = true
			}
		}
	}

	if (workerReady || workerPartialReady) && (fuseReady || fusePartialReady) {
		ready = true
	} else {
		e.Log.Info("workers are not ready", "workerReady", workerReady,
			"workerPartialReady", workerPartialReady,
			"fuseReady", fuseReady,
			"fusePartialReady", fusePartialReady)
		return
	}

	// update the status as the workers are ready
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		if len(runtimeToUpdate.Status.Conditions) == 0 {
			runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
		}
		cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
			"The workers are ready.", corev1.ConditionTrue)
		if workerPartialReady {
			cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkersReady, datav1alpha1.RuntimeWorkersReadyReason,
				"The workers are partially ready.", corev1.ConditionTrue)

			runtimeToUpdate.Status.WorkerPhase = datav1alpha1.RuntimePhasePartialReady
		}
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				cond)
		fuseCond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesReady, datav1alpha1.RuntimeFusesReadyReason,
			"The fuses are ready.", corev1.ConditionTrue)

		if fusePartialReady {
			fuseCond = utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesReady, datav1alpha1.RuntimeFusesReadyReason,
				"The fuses are partially ready.", corev1.ConditionTrue)

			runtimeToUpdate.Status.FusePhase = datav1alpha1.RuntimePhasePartialReady
		}
		runtimeToUpdate.Status.Conditions =
			utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions,
				fuseCond)

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		}

		return nil
	})

	return
}
