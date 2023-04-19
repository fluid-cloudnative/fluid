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

package efc

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"

	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
)

// getRuntime gets the runtime
func (r *RuntimeReconciler) getRuntime(ctx cruntime.ReconcileRequestContext) (*datav1alpha1.EFCRuntime, error) {
	var runtime datav1alpha1.EFCRuntime
	if err := r.Get(ctx, ctx.NamespacedName, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetOrCreateEngine gets or creates engine
func (r *RuntimeReconciler) GetOrCreateEngine(
	ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error) {
	var found bool
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if engine, found = r.engines[id]; !found {
		engine, err = ddc.CreateEngine(id,
			ctx)
		if err != nil {
			return nil, err
		}
		r.engines[id] = engine
		r.Log.V(1).Info("Put Engine to engine map")
	} else {
		r.Log.V(1).Info("Get Engine from engine map")
	}

	return engine, err
}

// RemoveEngine removes the engine
func (r *RuntimeReconciler) RemoveEngine(ctx cruntime.ReconcileRequestContext) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	delete(r.engines, id)
}
