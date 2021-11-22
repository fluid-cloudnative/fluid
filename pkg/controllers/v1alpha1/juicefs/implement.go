package juicefs

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

const (
	runtimeType                  = common.JuiceFSRuntime
	runtimeResourceFinalizerName = "juicefs-runtime-controller-finalizer"
)

// getRuntime gets the runtime
func (r *JuiceFSRuntimeReconciler) getRuntime(ctx cruntime.ReconcileRequestContext) (*datav1alpha1.JuiceFSRuntime, error) {
	var runtime datav1alpha1.JuiceFSRuntime
	if err := r.Get(ctx, ctx.NamespacedName, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (r *JuiceFSRuntimeReconciler) GetOrCreateEngine(ctx cruntime.ReconcileRequestContext) (engine base.Engine, err error) {
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

func (r *JuiceFSRuntimeReconciler) RemoveEngine(ctx cruntime.ReconcileRequestContext) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	id := ddc.GenerateEngineID(ctx.NamespacedName)
	delete(r.engines, id)
}
