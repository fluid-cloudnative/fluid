/*
Copyright 2023 The Fluid Authors.

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

package jindo

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	jindoutils "github.com/fluid-cloudnative/fluid/pkg/utils/jindo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Use compiler to check if the struct implements all the interface
var _ controllers.RuntimeReconcilerInterface = (*RuntimeReconciler)(nil)

const controllerName string = "JindoRuntimeController"

// RuntimeReconciler reconciles a JindoRuntime object
type RuntimeReconciler struct {
	Scheme  *runtime.Scheme
	engines map[string]base.Engine
	mutex   *sync.Mutex
	*controllers.RuntimeReconciler
}

// NewRuntimeReconciler create controller for watching runtime custom resources created
func NewRuntimeReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *RuntimeReconciler {
	r := &RuntimeReconciler{
		Scheme:  scheme,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.RuntimeReconciler = controllers.NewRuntimeReconciler(r, client, log, recorder)
	return r
}

//Reconcile reconciles jindo runtime
// +kubebuilder:rbac:groups=data.fluid.io,resources=jindoruntimes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=jindoruntimes/status,verbs=get;update;patch

func (r *RuntimeReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	defer utils.TimeTrack(time.Now(), "Reconcile JindoRuntime", "request", req)
	ctx := cruntime.ReconcileRequestContext{
		Context:        context,
		Log:            r.Log.WithValues("jindoruntime", req.NamespacedName),
		NamespacedName: req.NamespacedName,
		Recorder:       r.Recorder,
		Category:       common.AccelerateCategory,
		RuntimeType:    common.JindoRuntime,
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
	ctx.EngineImpl = ddc.InferEngineImpl(runtime.Status, jindoutils.GetDefaultEngineImpl())
	ctx.Log.V(1).Info("process the runtime", "runtime", ctx.Runtime)

	// reconcile the implement
	return r.ReconcileInternal(ctx)
}

// SetupWithManager setups the manager with RuntimeReconciler
func (r *RuntimeReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options, eventDriven bool) (err error) {
	if eventDriven {
		err = watch.SetupWatcherWithReconciler(mgr, options, r, "")
	} else {
		err = ctrl.NewControllerManagedBy(mgr).
			WithOptions(options).
			For(&datav1alpha1.JindoRuntime{}).
			Complete(r)
	}
	return
}

func (r *RuntimeReconciler) ControllerName() string {
	return controllerName
}

func (r *RuntimeReconciler) ManagedResource() client.Object {
	return &datav1alpha1.JindoRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       datav1alpha1.JindoRuntimeKind,
			APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
		},
	}
}
