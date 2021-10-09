package jindo

import (
	"context"

	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/client-go/util/retry"
)

func (e JindoEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {

	var (
		workerName string = e.getWorkertName()
		namespace  string = e.namespace
	)
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		workers, err := e.getStatefulset(workerName, namespace)
		if err != nil {
			return err
		}

		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		desireReplicas := runtime.Replicas()
		if *workers.Spec.Replicas != desireReplicas {
			workerToUpdate := workers.DeepCopy()
			workerToUpdate.Spec.Replicas = &desireReplicas
			err = e.Client.Update(context.TODO(), workerToUpdate)
			if err != nil {
				return err
			}
		} else {
			e.Log.V(1).Info("Nothing to do for syncing")
		}

		return nil
	})

	return
}
