package jindo

import (
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/client-go/util/retry"
)

func (e JindoEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {

	var (
		workerName string = e.getWorkertName()
		namespace  string = e.namespace
	)

	workers, err := kubeclient.GetStatefulSet(e.Client, workerName, namespace)
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		runtimeToUpdate := runtime.DeepCopy()
		// err = e.Helper.SetupWorkers(runtimeToUpdate, runtimeToUpdate.Status, workers)
		e.Helper.SyncReplicas(ctx, runtimeToUpdate, runtimeToUpdate.Status, workers)
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
