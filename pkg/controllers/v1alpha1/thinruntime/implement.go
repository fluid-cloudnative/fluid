/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package thinruntime

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

const (
	runtimeType                  = common.ThinRuntime
	runtimeResourceFinalizerName = "thin-runtime-controller-finalizer"
)

// getRuntime gets the runtime
func (r *ThinRuntimeReconciler) getRuntime(ctx cruntime.ReconcileRequestContext) (*datav1alpha1.ThinRuntime, error) {
	var runtime datav1alpha1.ThinRuntime
	if err := r.Get(ctx, ctx.NamespacedName, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (r *ThinRuntimeReconciler) GetOrCreateEngine(ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error) {
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

func (r *ThinRuntimeReconciler) RemoveEngine(ctx cruntime.ReconcileRequestContext) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	delete(r.engines, id)
}
