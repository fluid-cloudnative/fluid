/*
Copyright 2022 The Fluid Authors.

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

package jindofsx

import (
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (e JindoFSxEngine) SyncReplicas(ctx cruntime.ReconcileRequestContext) (err error) {

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}
		if runtime.Spec.Worker.Disabled {
			e.Log.Info("Skip syncing replicas for worker when it's disabled")
			return nil
		}

		workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
			types.NamespacedName{Namespace: e.namespace, Name: e.getWorkerName()})
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
