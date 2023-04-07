/*

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

package cachefs

import (
	"context"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const controllerName string = "CacheFSRuntimeController"

// Use compiler to check if the struct implements all the interface
var _ controllers.RuntimeReconcilerInterface = (*CacheFSRuntimeReconciler)(nil)

// CacheFSRuntimeReconciler reconciles a CacheFSRuntime object
type CacheFSRuntimeReconciler struct {
	engines map[string]base.Engine
	mutex   *sync.Mutex
	client.Client
	Scheme *runtime.Scheme
	*controllers.RuntimeReconciler
}

// NewRuntimeReconciler create controller for watching runtime custom resources created
func NewRuntimeReconciler(client client.Client,
	log logr.Logger,
	scheme *runtime.Scheme,
	recorder record.EventRecorder) *CacheFSRuntimeReconciler {
	r := &CacheFSRuntimeReconciler{
		Scheme:  scheme,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.RuntimeReconciler = controllers.NewRuntimeReconciler(r, client, log, recorder)
	return r
}

//+kubebuilder:rbac:groups=data.fluid.io,resources=cachefsruntimes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.fluid.io,resources=cachefsruntimes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=data.fluid.io,resources=cachefsruntimes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CacheFSRuntime object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *CacheFSRuntimeReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {

	defer utils.TimeTrack(time.Now(), "Reconcile", "request", req)
	ctx := cruntime.ReconcileRequestContext{
		Context:        context,
		Log:            r.Log.WithValues("cachefsruntime", req.NamespacedName),
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
func (r *CacheFSRuntimeReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options, eventDriven bool) error {
	if eventDriven {
		return watch.SetupWatcherWithReconciler(mgr, options, r)
	} else {
		return ctrl.NewControllerManagedBy(mgr).
			WithOptions(options).
			For(&datav1alpha1.JuiceFSRuntime{}).
			Complete(r)
	}
}

func (r *CacheFSRuntimeReconciler) ControllerName() string {
	return controllerName
}

func (r *CacheFSRuntimeReconciler) ManagedResource() client.Object {
	return &datav1alpha1.CacheFSRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       datav1alpha1.CacheFSRuntimeKind,
			APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
		},
	}
}

func NewCache(scheme *runtime.Scheme) cache.NewCacheFunc {
	// For reference dataset, controller cares about fuse daemonsets of other runtime types
	daemonSetSelector := labels.NewSelector()
	req, err := labels.NewRequirement(common.App, selection.In, []string{
		common.ThinRuntime,
		common.AlluxioRuntime,
		common.JindoRuntime,
		common.JuiceFSRuntime,
		common.GooseFSRuntime,
		common.EACRuntime,
		common.CacheFSRuntime,
	})
	if err != nil {
		panic(err)
	}
	daemonSetSelector.Add(*req)

	return cache.BuilderWithOptions(cache.Options{
		Scheme: scheme,
		SelectorsByObject: cache.SelectorsByObject{
			&appsv1.StatefulSet{}: {Label: labels.SelectorFromSet(labels.Set{
				common.App: common.CacheFSRuntime,
			})},
			&appsv1.DaemonSet{}: {Label: daemonSetSelector},
		},
	})
}
