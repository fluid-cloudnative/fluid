package engine

import (
	"context"
	"fmt"
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"reflect"
	"time"
)

func (e *CacheEngine) CheckAndUpdateRuntimeStatus(value *common.CacheRuntimeValue) (bool, error) {

	var (
		masterReady, workerReady, runtimeReady bool
	)

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()
		if value.Master.Enabled {
			masterStatus, err := e.masterHelper.ConstructComponentStatus(context.TODO(), value.Master)
			if err != nil {
				return err
			}
			if masterStatus.ReadyReplicas == masterStatus.DesiredReplicas {
				masterStatus.Phase = data.RuntimePhaseReady
				masterReady = true
			} else {
				masterStatus.Phase = data.RuntimePhaseNotReady
			}
			runtimeToUpdate.Status.Master = masterStatus
		} else {
			masterReady = true
		}

		if value.Worker.Enabled {
			workerStatus, err := e.workerHelper.ConstructComponentStatus(context.TODO(), value.Worker)
			if err != nil {
				return err
			}
			if runtime.Replicas() == 0 {
				workerStatus.Phase = data.RuntimePhaseReady
				workerReady = true
			} else if workerStatus.ReadyReplicas > 0 {
				if runtime.Replicas() == workerStatus.ReadyReplicas {
					workerStatus.Phase = data.RuntimePhaseReady
					workerReady = true
				} else if workerStatus.ReadyReplicas >= 1 {
					workerStatus.Phase = data.RuntimePhasePartialReady
					workerReady = true
				}
			} else {
				workerStatus.Phase = data.RuntimePhaseNotReady
			}
			runtimeToUpdate.Status.Worker = workerStatus
		} else {
			workerReady = true
		}

		if value.Client.Enabled {
			clientStatus, err := e.clientHelper.ConstructComponentStatus(context.TODO(), value.Client)
			if err != nil {
				return err
			}
			if clientStatus.DesiredReplicas > 0 {
				if clientStatus.DesiredReplicas == clientStatus.ReadyReplicas {
					clientStatus.Phase = data.RuntimePhaseReady
				} else if clientStatus.ReadyReplicas >= 1 {
					clientStatus.Phase = data.RuntimePhasePartialReady
				}
			}
			runtimeToUpdate.Status.Client = clientStatus
		}

		if masterReady && workerReady {
			runtimeReady = true
		} else {
			e.Log.Info(fmt.Sprintf("MasterReady: %v, workerReady: %v", masterReady, workerReady))

		}
		// Update the setup time
		if runtimeReady && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to update runtime status", types.NamespacedName{Namespace: e.namespace, Name: e.name})
		return false, err
	}

	return runtimeReady, nil
}
