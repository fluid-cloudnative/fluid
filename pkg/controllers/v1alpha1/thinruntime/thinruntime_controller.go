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

package thinruntime

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const controllerName string = "ThinRuntimeController"

var (
	_ controllers.RuntimeReconcilerInterface = (*ThinRuntimeReconciler)(nil)
)

// ThinRuntimeReconciler reconciles a ThinRuntime object
type ThinRuntimeReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	engines map[string]base.Engine
	mutex   *sync.Mutex
	*controllers.RuntimeReconciler
}

func (r *ThinRuntimeReconciler) ControllerName() string {
	return controllerName
}

func (r *ThinRuntimeReconciler) ManagedResource() client.Object {
	return &datav1alpha1.ThinRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       datav1alpha1.ThinRuntimeKind,
			APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
		},
	}
}

//+kubebuilder:rbac:groups=data.fluid.io,resources=thinruntimes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.fluid.io,resources=thinruntimes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=data.fluid.io,resources=thinruntimes/finalizers,verbs=update

func (r *ThinRuntimeReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	defer utils.TimeTrack(time.Now(), "Reconcile", "request", req)
	ctx := cruntime.ReconcileRequestContext{
		Context:        context,
		Log:            r.Log.WithValues("thinruntime", req.NamespacedName),
		NamespacedName: req.NamespacedName,
		Recorder:       r.Recorder,
		Category:       common.AccelerateCategory,
		RuntimeType:    runtimeType,
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
	ctx.Log.V(1).Info("process the runtime", "runtime", ctx.Runtime)

	// reconcile the implement
	return r.ReconcileInternal(ctx)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ThinRuntimeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		// For().
		Complete(r)
}
