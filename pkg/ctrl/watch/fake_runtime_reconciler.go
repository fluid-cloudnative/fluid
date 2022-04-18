package watch

import (
	"context"
	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FakeRuntimeReconciler struct {
}

// Reconciler performs a full reconciliation for the object referred to by the Request.
// The Controller will requeue the Request to be processed again if an error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *FakeRuntimeReconciler) Reconcile(context context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	return
}

// ControllerName gets the name of controller
func (r *FakeRuntimeReconciler) ControllerName() string {
	return ""
}

// ManagedResource which is handled by controller
func (r *FakeRuntimeReconciler) ManagedResource() (c client.Object) {
	return &datav1alpha1.JindoRuntime{
		TypeMeta: metav1.TypeMeta{
			Kind:       datav1alpha1.JindoRuntimeKind,
			APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
		},
	}
}

type FakePodReconciler struct {
}

// Reconcile Reconciler performs a full reconciliation for the object referred to by the Request.
// The Controller will requeue the Request to be processed again if an error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *FakePodReconciler) Reconcile(context context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	return
}

// ControllerName gets the name of controller
func (r *FakePodReconciler) ControllerName() string {
	return ""
}

// ManagedResource which is handled by controller
func (r *FakePodReconciler) ManagedResource() (c client.Object) {
	return &corev1.Pod{}
}
