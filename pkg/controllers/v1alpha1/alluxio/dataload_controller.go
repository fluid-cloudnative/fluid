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

package alluxio

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
)

// DataLoadReconciler reconciles a AlluxioDataLoad object
type DataLoadReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type ReconcileRequestContext struct {
	// necessary info used in reconcile process
}

func (r *DataLoadReconciler) ReconcileDataload(ctx ReconcileRequestContext) (ctrl.Result, error) {
	/*
		//1. Get Alluxio Dataload Object `dataload`
	*/

	/*
		// 2. Check dataload's Phase
			// == `DataloadPhaseNone` -> change to `DataloadPhasePending`, return RequeueImmediately
			// == `DataloadPhaseComplete` -> return NoRequeue TODO: observe alluxio worker dynamic scaling
			// == `DataloadPhaseFailed` -> return NoRequeue   TODO: backoff restart with limit
			// == `DataloadPhaseLoading` -> To 3, 5
			// == `DataloadPhasePending` -> To 3, 4
	*/

	/*
		// 3. Get dataset object using `dataload.spec.Dataset`
	*/

	/*
		// ----------- DataloadPhasePending Logic --------------- //
		// 4. Check if dataset exists && dataset.phase
			// == Existed && Bound ->
				// If runtime.status.currentWorkerNumberScheduled existed &&
				//		runtime.status.workerNumberAvailable existed &&
				// 		runtime.status.currentWorkerNumberScheduled != 0 &&
				//		runtime.status.currentWorkerNumberScheduled == runtime.status.workerNumberAvailable (DATASET BOUND AND ALL WORKERS AVAILABLE) ->
					// i. Get runtime.status.workerNumberAvailable;
					// ii. Helm install dataloader Job with release name `<dataset>-load` and workerNumberAvailable;
					// iii. change phase to `DataloadPhaseLoading`;
					// iiii. return Requeue(10 seconds);
				// otherwise (DATASET BOUND BUT SOME WORKERS NOT AVAILABLE) -> return Requeue(5s)
			// == otherwise (DATASET NOT EXISTED OR NOT BOUND) -> return Requeue(10s)
	*/

	/*
		// ----------- DataloadPhaseLoading Logic --------------- //
		// 5. Check if dataset exists && dataset.phase
			// == Existed && Bound ->
				// Check `<dataset>-loader` job status
					// job.status.succeeded == job.spec.completions (JOB SUCCEEDED) ->
						// i. change phase to `DataloadPhaseComplete`;
						// ii. update dataload.conditions with job conditions;
						// iii. return NoRequeue;
					// job.status.failed existed && job.status.failed != 0 (JOB FAILED) ->
						// i. change phase to `DataloadPhaseFailed`;
						// ii. update dataload.conditions with job conditions;
						// iii. return NoRequque;
					// otherwise (JOB STILL RUNNING) -> return Requeue(10s)
			// == otherwise (DATASET OR RUNTIME STATUS CHANGED DURING LOADING) ->
				// i. check if release exists && helm del `<dataset>-load`;
				// ii. change phase to `DataloadPhasePending`;
				// iii. return Requque(10s);
	*/
}

func (r *DataLoadReconciler) ReconcileDataloadDeletion(ctx ReconcileRequestContext) (ctrl.Result, error) {
	/*
		// 1. Get `dataload.dataset`
	*/

	/*
		// 2. Check if helm release `<dataset>-load` exists
			// == existed -> helm del `<dataset>-load`
			// == otherwise -> do nothing
	*/

	/*
		// 3. Remove finalizer
	*/

}

//Reconcile reconciles the AlluxioDataLoad Object
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads/status,verbs=get;update;patch

func (r *DataLoadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("alluxiodataload", req.NamespacedName)

	/*
		// 1. Get necessary info
	*/
	ctx := ReconcileRequestContext{}

	/*
		// 2. Reconcile delete the runtime
			// 2.1 Check if deletionTimestamp exists
			// 2.2 Do deletion: r.reconcileDataloadDeletion(ctx)
	*/

	/*
		// 3. Add finalizer if necessary
			// 3.1 Check if finalizer already exists
			// 3.2 AddFinalizerAndRequeue
	*/

	/*
		// 4. Do reconcile dataload
	*/
	return r.ReconcileDataload(ctx)
	//return ctrl.Result{}, nil
}

//SetupWithManager setups the manager with AlluxioDataLoad
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.AlluxioDataLoad{}).
		Complete(r)
}
