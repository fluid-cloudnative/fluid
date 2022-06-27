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
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

// SyncRuntime syncs the runtime spec
func (e *JindoFSxEngine) SyncRuntime(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {

	masterChanged, err := e.syncMasterSpec(ctx)
	if err != nil {
		return
	}
	if masterChanged {
		e.Log.Info("Master Spec is changed", "name", ctx.Name, "namespace", ctx.Namespace)
		return masterChanged, err
	}

	workerChanged, err := e.syncWorkerSpec(ctx)
	if err != nil {
		return
	}
	if workerChanged {
		e.Log.Info("Worker Spec is changed", "name", ctx.Name, "namespace", ctx.Namespace)
		return workerChanged, err
	}

	fuseChanged, err := e.syncFuseSpec(ctx)
	if err != nil {
		return
	}
	if fuseChanged {
		e.Log.Info("Fuse Spec is changed", "name", ctx.Name, "namespace", ctx.Namespace)
		return fuseChanged, err
	}

	return
}

func (e *JindoFSxEngine) syncMasterSpec(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	return
}

func (e *JindoFSxEngine) syncWorkerSpec(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	return
}

func (e *JindoFSxEngine) syncFuseSpec(ctx cruntime.ReconcileRequestContext) (changed bool, err error) {
	return
}
