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

package datatable

import (
	"context"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/datatable"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	// finalizer for datable
	DataTableFinalizerName = "datatable-controller-finalizer"
	// common name
	CommonName   = "datatable-common"
	ResyncPeriod = time.Duration(5 * time.Second)
)

// DataTableReconciler reconciles a DataTable object
type DataTableReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

type reconcileRequestContext struct {
	context.Context
	Log       logr.Logger
	Datatable datav1alpha1.DataTable
	types.NamespacedName
	FinalizerName string
}

//+kubebuilder:rbac:groups=data.fluid.io,resources=datatables,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=data.fluid.io,resources=datatables/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=data.fluid.io,resources=datatables/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DataTable object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *DataTableReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {

	// init the reconcileRequestContext struct
	ctx := reconcileRequestContext{
		Context:        context,
		Log:            r.Log.WithValues("datatable", req.NamespacedName),
		NamespacedName: req.NamespacedName,
		FinalizerName:  DataTableFinalizerName,
	}

	ctx.Log.V(1).Info("Process the request", "request", req)

	if err := r.Get(ctx, req.NamespacedName, &ctx.Datatable); err != nil {
		// can not get the datatable obj
		// TODO: err handle
		ctx.Log.Error(err, "Fail to get the datatable", req)
		return utils.RequeueIfError(err)
	} else {
		return r.reconcileDataTable(ctx)
	}
}

func (r *DataTableReconciler) reconcileDataTable(ctx reconcileRequestContext) (result ctrl.Result, err error) {

	ip, err := datatable.GetOrCreateAlluxio(r.Client, CommonName, ctx.Datatable.Namespace)
	if err != nil || len(ip) == 0 {
		if err != nil {
			ctx.Log.Error(err, "GetOrCreateAlluxio error")
		} else {
			ctx.Log.V(1).Info("The alluxio is starting...")
		}
		return utils.RequeueAfterInterval(10 * time.Second)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataTableReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.DataTable{}).
		Complete(r)
}
