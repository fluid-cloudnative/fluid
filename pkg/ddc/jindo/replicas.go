package jindo

import (
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e JindoEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkertName()})
		if err != nil {
			if fluiderrs.IsDeprecated(err) {
				e.Log.Info("Warning: Deprecated mode is not support, so skip handling", "details", err)
				return nil
			}
			return err
		}

		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		// err = e.Helper.SetupWorkers(runtimeToUpdate, runtimeToUpdate.Status, workers)
		err = e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
		if err != nil {
			e.Log.Error(err, "Failed to sync the replicas")
		}
		return nil
	})
	if err != nil {
		e.Log.Error(err, "Failed to sync the replicas")
	}

	return
}
