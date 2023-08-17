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
	"github.com/dazheng/gohive"
	"github.com/fluid-cloudnative/fluid/pkg/utils/hive"
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
	DataTableFinalizerName = "datatable-controller-finalizer"       // finalizer for datable
	CommonName             = "datatable-common"                     // the name of dataset and runtime
	ResyncPeriod           = time.Duration(5 * time.Second)         // requeue period
	FinalizerName          = "fluid-datatable-controller-finalizer" // finalizer name
)

var MasterIP string // alluxio master IP

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
//+kubebuilder:rbac:groups=data.fluid.io,resources=datases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxioruntimes,verbs=get;list;watch;create;update;patch;delete

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
		if utils.IgnoreNotFound(err) != nil {
			ctx.Log.Error(err, "Fail to get the datatable", req)
			return utils.RequeueIfError(err)
		} else {
			ctx.Log.V(1).Info("Fail to get the datatable")
			return ctrl.Result{}, nil
		}
	} else {
		return r.reconcileDataTable(ctx)
	}
}

func (r *DataTableReconciler) reconcileDataTable(ctx reconcileRequestContext) (result ctrl.Result, err error) {

	// 0. create dataset and alluxio runtime if they do not exist, or get the master ip
	MasterIP, err = datatable.GetOrCreateAlluxio(r.Client, CommonName, ctx.Datatable.Namespace)
	if err != nil || len(MasterIP) == 0 {
		if err != nil {
			ctx.Log.Error(err, "GetOrCreateAlluxio error")
		} else {
			ctx.Log.V(1).Info("The alluxio is starting...")
		}
		return utils.RequeueAfterInterval(10 * time.Second)
	}
	ctx.Datatable.Status.CacheMasterIP = MasterIP // TODO: Update the status
	r.Log.V(1).Info("STEP 0: Get or Create alluxio successfully and get the master IP")

	// 1. create the hive client
	host := ctx.Datatable.Spec.Url
	conn, err := hive.CreateHiveClient(host)
	if err != nil {
		ctx.Log.Error(err, "Fail to create the hive client")
		return utils.RequeueIfError(err)
	}
	defer conn.Close()
	r.Log.V(1).Info("STEP 1: Create the hive client successfully")

	// 2. delete this datatable if it is mark the deletiontimestamp
	if utils.HasDeletionTimestamp(ctx.Datatable.ObjectMeta) {
		return r.ReconcileDatatableDeletion(ctx, conn)
	}

	// 3. add finalizer if it does not have
	if !utils.ContainsString(ctx.Datatable.ObjectMeta.GetFinalizers(), FinalizerName) {
		return r.AddFinalizerAndRequeue(ctx)
	}

	return ctrl.Result{}, nil
}

// ReconcileDatatableDeletion recover the table location and remove the finalizer
func (r *DataTableReconciler) ReconcileDatatableDeletion(ctx reconcileRequestContext, conn *gohive.Connection) (ctrl.Result, error) {
	datatable := ctx.Datatable
	if err := hive.ChangeSchemaURLForRecover(r.Client, datatable, conn); err != nil {
		r.Log.Error(err, "Fail to recover the table location")
		return utils.RequeueIfError(err)
	}
	if !datatable.ObjectMeta.GetDeletionTimestamp().IsZero() {
		datatable.ObjectMeta.Finalizers = utils.RemoveString(datatable.ObjectMeta.GetFinalizers(), FinalizerName)

		if err := r.Update(ctx, &datatable); err != nil {
			r.Log.Error(err, "Fail to update the datatable")
			utils.RequeueIfError(err)
		}

		r.Log.V(1).Info("Clean the finalizers", datatable)
	}

	r.Log.V(1).Info("STEP 3: Delete the datatable successfully")

	return ctrl.Result{}, nil
}

// AddFinalizerAndRequeue add the finalizer and requeue
func (r *DataTableReconciler) AddFinalizerAndRequeue(ctx reconcileRequestContext) (ctrl.Result, error) {

	prevGeneration := ctx.Datatable.ObjectMeta.GetGeneration()
	ctx.Datatable.ObjectMeta.Finalizers = append(ctx.Datatable.ObjectMeta.Finalizers, FinalizerName)
	if err := r.Update(ctx, &ctx.Datatable); err != nil {
		ctx.Log.Error(err, "Fail to add the finalizer")
		return utils.RequeueIfError(err)
	}
	newGeneration := ctx.Datatable.ObjectMeta.GetGeneration()

	r.Log.V(1).Info("STEP 3: Add the finalizer successfully")

	return utils.RequeueImmediatelyUnlessGenerationChanged(prevGeneration, newGeneration)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DataTableReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&datav1alpha1.DataTable{}).
		Complete(r)
}
