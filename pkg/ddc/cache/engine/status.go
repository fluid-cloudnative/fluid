/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"context"
	"fmt"
	"reflect"
	"time"

	fluidapi "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/cache/component"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e *CacheEngine) setMasterComponentStatus(componentInfo *common.ComponentStatusInfo, status *fluidapi.CacheRuntimeStatus) (ready bool, err error) {
	manager := component.NewComponentHelper(common.ComponentTypeMaster, e.Client)

	masterStatus, err := manager.ConstructComponentStatus(context.TODO(), &componentInfo.ComponentIdentity)
	if err != nil {
		return false, err
	}
	if masterStatus.ReadyReplicas == masterStatus.DesiredReplicas {
		masterStatus.Phase = fluidapi.RuntimePhaseReady
		ready = true
	} else {
		masterStatus.Phase = fluidapi.RuntimePhaseNotReady
	}
	status.Master = masterStatus

	return ready, err
}
func (e *CacheEngine) setWorkerComponentStatus(componentInfo *common.ComponentStatusInfo, status *fluidapi.CacheRuntimeStatus) (ready bool, err error) {
	manager := component.NewComponentHelper(common.ComponentTypeWorker, e.Client)

	workerStatus, affinity, err := manager.ConstructComponentStatusAndAffinity(context.TODO(), &componentInfo.ComponentIdentity)
	if err != nil {
		return false, err
	}

	if workerStatus.DesiredReplicas == 0 {
		workerStatus.Phase = fluidapi.RuntimePhaseReady
		ready = true
	} else if workerStatus.ReadyReplicas > 0 {
		if workerStatus.DesiredReplicas == workerStatus.ReadyReplicas {
			workerStatus.Phase = fluidapi.RuntimePhaseReady
			ready = true
		} else if workerStatus.ReadyReplicas >= 1 {
			workerStatus.Phase = fluidapi.RuntimePhasePartialReady
			ready = true
		}
	} else {
		workerStatus.Phase = fluidapi.RuntimePhaseNotReady
	}
	status.Worker = workerStatus
	status.CacheAffinity = affinity

	return ready, nil
}
func (e *CacheEngine) setClientComponentStatus(componentInfo *common.ComponentStatusInfo, status *fluidapi.CacheRuntimeStatus) (fullyReady bool, err error) {
	manager := component.NewComponentHelper(common.ComponentTypeClient, e.Client)

	clientStatus, err := manager.ConstructComponentStatus(context.TODO(), &componentInfo.ComponentIdentity)
	if err != nil {
		return false, err
	}
	if clientStatus.DesiredReplicas > 0 && clientStatus.ReadyReplicas >= clientStatus.DesiredReplicas {
		clientStatus.Phase = fluidapi.RuntimePhaseReady
		fullyReady = true
	} else if clientStatus.ReadyReplicas > 0 {
		clientStatus.Phase = fluidapi.RuntimePhasePartialReady
	} else {
		clientStatus.Phase = fluidapi.RuntimePhaseNotReady
	}
	status.Client = clientStatus

	return fullyReady, nil
}

func (e *CacheEngine) CheckAndUpdateRuntimeStatus(value *common.CacheRuntimeStatusValue) (bool, error) {
	return e.checkAndUpdateRuntimeStatus(context.TODO(), value)
}

func (e *CacheEngine) checkAndUpdateRuntimeStatus(ctx context.Context, value *common.CacheRuntimeStatusValue,
	conditions ...*fluidapi.RuntimeCondition) (bool, error) {
	runtimeReady := false
	var latestRuntime *fluidapi.CacheRuntime
	var pendingEvents []fluidapi.RuntimeCondition

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// Reset readiness and transition events on each retry to avoid stale state after conflicts.
		masterReady, workerReady, clientFullyReady := true, true, false
		runtimeReady = false
		latestRuntime = nil
		pendingEvents = nil

		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if value.Master.Enabled {
			masterReady, err = e.setMasterComponentStatus(value.Master, &runtimeToUpdate.Status)
			if err != nil {
				return err
			}
		}

		if value.Worker.Enabled {
			workerReady, err = e.setWorkerComponentStatus(value.Worker, &runtimeToUpdate.Status)
			if err != nil {
				return err
			}
		}

		if value.Client.Enabled {
			clientFullyReady, err = e.setClientComponentStatus(value.Client, &runtimeToUpdate.Status)
			if err != nil {
				return err
			}
		}

		runtimeReady = masterReady && workerReady
		if !runtimeReady {
			e.Log.Info(fmt.Sprintf(
				"MasterReady: %v, workerReady: %v, clientFullyReady: %v, clientPhase: %s",
				masterReady,
				workerReady,
				clientFullyReady,
				runtimeToUpdate.Status.Client.Phase,
			))
		}

		// Update the setup time
		if runtimeReady && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		selector, err := e.getWorkerSelectors()
		if err != nil {
			return err
		}
		runtimeToUpdate.Status.Selector = selector

		runtimeToUpdate.Status.ValueFile = common.GetCacheRuntimeConfigConfigMapName(e.name)
		pendingEvents = applyRuntimeConditions(&runtimeToUpdate.Status, conditions...)
		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(ctx, runtimeToUpdate)
			if err == nil {
				latestRuntime = runtimeToUpdate
			}
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to update runtime status", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return false, err
	}
	e.recordConditionEvents(latestRuntime, pendingEvents)

	return runtimeReady, nil
}

// applyRuntimeConditions updates each non-nil condition in status and returns
// tieredStore conflict conditions that transitioned or changed message.
func applyRuntimeConditions(status *fluidapi.CacheRuntimeStatus, conditions ...*fluidapi.RuntimeCondition) []fluidapi.RuntimeCondition {
	var events []fluidapi.RuntimeCondition
	for _, cond := range conditions {
		if cond == nil {
			continue
		}
		_, old := utils.GetRuntimeCondition(status.Conditions, cond.Type)
		if !applyRuntimeCondition(status, *cond) {
			continue
		}
		if cond.Type == fluidapi.RuntimeTieredStoreSynced && cond.Status == corev1.ConditionFalse &&
			(old == nil || old.Status != corev1.ConditionFalse || old.Reason != cond.Reason || old.Message != cond.Message) {
			events = append(events, *cond)
		}
	}
	return events
}

func (e *CacheEngine) recordConditionEvents(runtime *fluidapi.CacheRuntime, conditions []fluidapi.RuntimeCondition) {
	if e.Recorder == nil || runtime == nil {
		return
	}
	for _, cond := range conditions {
		e.Recorder.Eventf(runtime, corev1.EventTypeWarning, cond.Reason, "%s", cond.Message)
	}
}

func applyRuntimeCondition(status *fluidapi.CacheRuntimeStatus, cond fluidapi.RuntimeCondition) bool {
	_, old := utils.GetRuntimeCondition(status.Conditions, cond.Type)
	if old != nil && old.Status == cond.Status {
		if old.Reason == cond.Reason && old.Message == cond.Message {
			return false
		}
		cond.LastTransitionTime = old.LastTransitionTime
	}
	status.Conditions = utils.UpdateRuntimeCondition(status.Conditions, cond)
	return true
}
