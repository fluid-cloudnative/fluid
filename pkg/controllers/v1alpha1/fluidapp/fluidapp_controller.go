/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package fluidapp

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl/watch"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
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
	pod, err := kubeclient.GetPodByName(f.Client, request.Name, request.Namespace)
	if err != nil {
		requestCtx.Log.Error(err, "fetch pod error")
		return reconcile.Result{}, err
	}
	if pod == nil {
		requestCtx.Log.Info("pod not found", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}
	requestCtx.pod = pod

	if !watch.ShouldInQueue(pod) {
		requestCtx.Log.Info("pod should not in queue", "name", request.Name, "namespace", request.Namespace)
		return reconcile.Result{}, nil
	}
	return f.internalReconcile(requestCtx)
}

func (f *FluidAppReconciler) internalReconcile(ctx reconcileRequestContext) (ctrl.Result, error) {
	pod := ctx.pod

	// umount fuse sidecars
	err := f.umountFuseSidecars(pod)
	if err != nil {
		ctx.Log.Error(err, "umount fuse sidecar error", "podName", pod.Name, "podNamespace", pod.Namespace)
		return utils.RequeueIfError(err)
	}
	return utils.NoRequeue()
}

func (f *FluidAppReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return watch.SetupAppWatcherWithReconciler(mgr, options, f)
}

func NewCache(scheme *runtime.Scheme) cache.NewCacheFunc {
	return cache.BuilderWithOptions(cache.Options{
		Scheme: scheme,
		SelectorsByObject: cache.SelectorsByObject{
			&corev1.Pod{}: {Label: labels.SelectorFromSet(labels.Set{
				common.InjectSidecarDone: common.True,
			})},
		},
	})
}
