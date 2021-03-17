package jindo

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
)

func (e JindoEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {

	runtime, err := e.getRuntime()
	if err != nil {
		return err
	}

	if runtime.Replicas() > runtime.Status.CurrentWorkerNumberScheduled {
		err = e.SetupWorkers()
		if err != nil {
			return err
		}
		_, err = e.CheckWorkersReady()
		if err != nil {
			e.Log.Error(err, "Check if the workers are ready")
			return err
		}

		// _, err := e.CheckAndUpdateRuntimeStatus()
		// if err != nil {
		// 	e.Log.Error(err, "Check if the runtime is ready")
		// 	return err
		// }
	} else if runtime.Replicas() < runtime.Status.CurrentWorkerNumberScheduled {
		replicas := runtime.Replicas()
		e.Log.Info("Scaling in Jindo workers", "expectedReplicas", replicas)
		curReplicas, err := e.destroyWorkers(replicas)
		if err != nil {
			return err
		}

		if curReplicas > replicas {
			ctx.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeScaleInFailed,
				"Alluxio workers are being used by some pods, can't scale in (expected replicas: %v, current replicas: %v)",
				replicas, curReplicas)
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			runtime, err := e.getRuntime()
			if err != nil {
				e.Log.Error(err, "scale in when sync replicas")
				return err
			}

			runtimeToUpdate := runtime.DeepCopy()

			if len(runtimeToUpdate.Status.Conditions) == 0 {
				runtimeToUpdate.Status.Conditions = []datav1alpha1.RuntimeCondition{}
			}

			runtimeToUpdate.Status.DesiredWorkerNumberScheduled = replicas
			runtimeToUpdate.Status.CurrentWorkerNumberScheduled = curReplicas
			cond := utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkerScaledIn, datav1alpha1.RuntimeWorkersScaledInReason,
				"The workers scaled in.", corev1.ConditionTrue)
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

			if !runtimeToUpdate.Spec.Fuse.Global {
				runtimeToUpdate.Status.DesiredFuseNumberScheduled = replicas
				runtimeToUpdate.Status.CurrentWorkerNumberScheduled = curReplicas
				fuseCond := utils.NewRuntimeCondition(datav1alpha1.RuntimeFusesScaledIn, datav1alpha1.RuntimeFusesScaledInReason,
					"The fuses scaled in.", corev1.ConditionTrue)
				runtimeToUpdate.Status.Conditions =
					utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, fuseCond)
			}

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			}

			return nil
		})

		if err != nil {
			return err
		}
	} else {
		e.Log.V(1).Info("Nothing to do")
	}

	return
}
