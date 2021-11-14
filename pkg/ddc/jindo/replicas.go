package jindo

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

	var (
		cond datav1alpha1.RuntimeCondition
	)

	if runtime.Replicas() != runtime.Status.DesiredWorkerNumberScheduled {
		// 1. Update scale condtion
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

			if runtime.Replicas() < runtime.Status.DesiredFuseNumberScheduled {
				cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkerScaledIn, datav1alpha1.RuntimeWorkersScaledInReason,
					"The workers scaled in.", corev1.ConditionTrue)
			} else {
				cond = utils.NewRuntimeCondition(datav1alpha1.RuntimeWorkerScaledOut, datav1alpha1.RuntimeWorkersScaledOutReason,
					"The workers scaled out.", corev1.ConditionTrue)
			}
			runtimeToUpdate.Status.Conditions =
				utils.UpdateRuntimeCondition(runtimeToUpdate.Status.Conditions, cond)

			if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
				return e.Client.Status().Update(context.TODO(), runtimeToUpdate)
			}

			return nil
		})

		if err != nil {
			return err
		}

		// 2. setup the workers for scaling
		err = e.SetupWorkers()
		if err != nil {
			return err
		}
		_, err = e.CheckWorkersReady()
		if err != nil {
			e.Log.Error(err, "Check if the workers are ready")
			return err
		}
	} else {
		e.Log.V(1).Info("Nothing to do")
		return
	}

	return
}
