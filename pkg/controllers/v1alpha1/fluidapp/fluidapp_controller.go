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

package fluidapp

import (
	"context"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const controllerName string = "FluidAppController"

type FluidAppReconciler struct {
	client.Client
	Recorder record.EventRecorder
	*FluidAppReconcilerImplement
}

func (f *FluidAppReconciler) ControllerName() string {
	return controllerName
}

func (f *FluidAppReconciler) ManagedResource() client.Object {
	return &corev1.Pod{}
}

type reconcileRequestContext struct {
	context.Context
	Log logr.Logger
	pod *corev1.Pod
	types.NamespacedName
}

func NewFluidAppReconciler(client client.Client,
	log logr.Logger,
	recorder record.EventRecorder) *FluidAppReconciler {
	return &FluidAppReconciler{
		Client:                      client,
		Recorder:                    recorder,
		FluidAppReconcilerImplement: NewFluidAppReconcilerImplement(client, log, recorder),
	}
}

// Reconcile reconciles Pod
// +kubebuilder:rbac:groups=v1,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=v1,resources=pods/status,verbs=get;update;patch
func (f *FluidAppReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	requestCtx := reconcileRequestContext{
		Context:        ctx,
		Log:            f.Log.WithValues("fluidapp", request.NamespacedName),
		NamespacedName: request.NamespacedName,
	}
	pod, err := f.fetchFluidApp(request.NamespacedName)
	if err != nil {
		requestCtx.Log.Error(err, "fetch pod error")
		return reconcile.Result{}, err
	}
	if pod == nil {
		requestCtx.Log.Info("pod not found", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}
	requestCtx.pod = pod

	return f.internalReconcile(requestCtx)
}

func (f *FluidAppReconciler) internalReconcile(ctx reconcileRequestContext) (ctrl.Result, error) {
	pod := ctx.pod

	// umount fuse sidecar
	err := f.umountFuseSidecar(pod)
	if err != nil {
		ctx.Log.Error(err, "umount fuse sidecar error", "podName", pod.Name, "podNamespace", pod.Namespace)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (f *FluidAppReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return watch.SetupAppWatcherWithReconciler(mgr, options, f)
}

func (f *FluidAppReconciler) fetchFluidApp(name types.NamespacedName) (*corev1.Pod, error) {
	pod := corev1.Pod{}
	if err := f.Get(context.TODO(), name, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &pod, nil
}
