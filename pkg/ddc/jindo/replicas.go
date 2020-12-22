package jindo

import cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"

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
		// scale in
		e.Log.V(1).Info("Scale in to be implemented")
	} else {
		e.Log.V(1).Info("Nothing to do")
	}

	return
}
