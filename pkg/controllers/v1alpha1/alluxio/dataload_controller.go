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

//Reconcile reconciles the AlluxioDataLoad Object
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=data.fluid.io,resources=alluxiodataloads/status,verbs=get;update;patch

func (r *DataLoadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("alluxiodataload", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

//SetupWithManager setups the manager with AlluxioDataLoad
func (r *DataLoadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datav1alpha1.AlluxioDataLoad{}).
		Complete(r)
}
