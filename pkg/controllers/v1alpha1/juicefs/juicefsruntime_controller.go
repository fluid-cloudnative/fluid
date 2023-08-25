/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"context"
	"sync"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

const controllerName string = "JuiceFSRuntimeController"

var (
	_ controllers.RuntimeReconcilerInterface = (*JuiceFSRuntimeReconciler)(nil)
)

// JuiceFSRuntimeReconciler reconciles a JuiceFSRuntime object
type JuiceFSRuntimeReconciler struct {
	Scheme  *runtime.Scheme
	engines map[string]base.Engine
	mutex   *sync.Mutex
	*controllers.RuntimeReconciler
}

func (r *JuiceFSRuntimeReconciler) ControllerName() string {
	return controllerName
}

func (r *JuiceFSRuntimeReconciler) ManagedResource() client.Object {
	return &datav1alpha1.JuiceFSRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       datav1alpha1.JuiceFSRuntimeKind,
			APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
		},
	}
}

// NewRuntimeReconciler create controller for watching runtime custom resources created
func NewRuntimeReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *JuiceFSRuntimeReconciler {
	r := &JuiceFSRuntimeReconciler{
		Scheme:  scheme,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.RuntimeReconciler = controllers.NewRuntimeReconciler(r, client, log, recorder)
	return r
}

//+kubebuilder:rbac:groups=data.fluid.io,resources=juicefsruntimes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.fluid.io,resources=juicefsruntimes/status,verbs=get;update;patch

func (r *JuiceFSRuntimeReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	defer utils.TimeTrack(time.Now(), "Reconcile", "request", req)
	ctx := cruntime.ReconcileRequestContext{
		Context:        context,
		Log:            r.Log.WithValues("juicefsruntime", req.NamespacedName),
		NamespacedName: req.NamespacedName,
		Recorder:       r.Recorder,
		Category:       common.AccelerateCategory,
		RuntimeType:    common.JuiceFSRuntime,
		Client:         r.Client,
		FinalizerName:  runtimeResourceFinalizerName,
	}

	ctx.Log.V(1).Info("process the request", "request", req)

	//	1.Load the Runtime
	runtime, err := r.getRuntime(ctx)
	if err != nil {
		if utils.IgnoreNotFound(err) == nil {
			ctx.Log.V(1).Info("The runtime is not found", "runtime", ctx.NamespacedName)
			return ctrl.Result{}, nil
		} else {
			ctx.Log.Error(err, "Failed to get the ddc runtime")
			return utils.RequeueIfError(errors.Wrap(err, "Unable to get ddc runtime"))
		}
	}
	ctx.Runtime = runtime
	ctx.EngineImpl = ddc.InferEngineImpl(runtime.Status, common.JuiceFSEngineImpl)
	ctx.Log.V(1).Info("process the runtime", "runtime", ctx.Runtime)

	// reconcile the implement
	return r.ReconcileInternal(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *JuiceFSRuntimeReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options, eventDriven bool) error {
	if eventDriven {
		return watch.SetupWatcherWithReconciler(mgr, options, r, "")
	} else {
		return ctrl.NewControllerManagedBy(mgr).
			WithOptions(options).
			For(&datav1alpha1.JuiceFSRuntime{}).
			Complete(r)
	}
}

func NewCache(scheme *runtime.Scheme) cache.NewCacheFunc {
	return cache.BuilderWithOptions(cache.Options{
		Scheme: scheme,
		SelectorsByObject: cache.SelectorsByObject{
			&appsv1.StatefulSet{}: {Label: labels.SelectorFromSet(labels.Set{
				common.App: common.JuiceFSRuntime,
			})},
			&appsv1.DaemonSet{}: {Label: labels.SelectorFromSet(labels.Set{
				common.App: common.JuiceFSRuntime,
			})},
		},
	})
}
