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

package watch

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	webhookReconcile "github.com/fluid-cloudnative/fluid/pkg/controllers/v1alpha1/webhook"
	"github.com/fluid-cloudnative/fluid/pkg/webhook"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	log = ctrl.Log.WithName("watch")
)

// This controller will be moved to RuntimeReconcilerInterface later
type Controller interface {
	// Reconciler performs a full reconciliation for the object referred to by the Request.
	// The Controller will requeue the Request to be processed again if an error is non-nil or
	// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
	Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error)

	// ControllerName gets the name of controller
	ControllerName() string

	// ManagedResource which is handled by controller
	ManagedResource() client.Object
}

func SetupWatcherWithReconciler(mgr ctrl.Manager, options controller.Options, r Controller) (err error) {
	options.Reconciler = r
	c, err := controller.New(r.ControllerName(), mgr, options)
	if err != nil {
		return err
	}

	runtimeEventHandler := &runtimeEventHandler{}
	err = c.Watch(&source.Kind{Type: r.ManagedResource()}, &handler.EnqueueRequestForObject{}, predicate.Funcs{
		CreateFunc: runtimeEventHandler.onCreateFunc(r),
		UpdateFunc: runtimeEventHandler.onUpdateFunc(r),
		DeleteFunc: runtimeEventHandler.onDeleteFunc(r),
	})
	if err != nil {
		log.Error(err, "Failed to watch JindoRuntime")
		return err
	}

	statefulsetEventHandler := &statefulsetEventHandler{}
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    r.ManagedResource(),
	}, predicate.Funcs{
		CreateFunc: statefulsetEventHandler.onCreateFunc(r),
		UpdateFunc: statefulsetEventHandler.onUpdateFunc(r),
		DeleteFunc: statefulsetEventHandler.onDeleteFunc(r),
	})
	if err != nil {
		return err
	}

	daemonsetEventHandler := &daemonsetEventHandler{}
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    r.ManagedResource(),
	}, predicate.Funcs{
		CreateFunc: daemonsetEventHandler.onCreateFunc(r),
		UpdateFunc: daemonsetEventHandler.onUpdateFunc(r),
		DeleteFunc: daemonsetEventHandler.onDeleteFunc(r),
	})
	if err != nil {
		return err
	}

	return
}

// isObjectInManaged checks if the object is managed by Fluid runtime controller
func isObjectInManaged(obj metav1.Object, r Controller) (managed bool) {
	if controllerRef := metav1.GetControllerOf(obj); controllerRef != nil && isOwnerMatched(controllerRef, r) {
		log.V(1).Info("Controller will handle the object due to owner reference is matched with runtime", "name", obj.GetName(), "namespace", obj.GetNamespace())
		managed = true
	} else {
		log.V(1).Info("Skip the object due to the  owner reference is not matched with fluid runtime", "name", obj.GetName(), "namespace", obj.GetNamespace())
	}
	return managed
}

// isOwnerMatched checks if controllerRef matches with the controller
func isOwnerMatched(controllerRef *metav1.OwnerReference, c Controller) bool {
	target := c.ManagedResource()
	// kind := target.GetObjectKind().GroupVersionKind().Kind
	// apiVersion := target.GetObjectKind().GroupVersionKind().Group + "/" + target.GetObjectKind().GroupVersionKind().Version
	apiVersion, kind := target.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

	return kind == controllerRef.Kind && apiVersion == controllerRef.APIVersion
}

func SetupWatcherForWebhook(mgr ctrl.Manager, certBuilder *webhook.CertificateBuilder, caCert []byte) (err error) {
	options := controller.Options{}
	webhookName := common.WebhookName
	options.Reconciler = &webhookReconcile.WebhookReconciler{
		CertBuilder: certBuilder,
		WebhookName: webhookName,
		CaCert:      caCert,
	}
	webhookController, err := controller.New("webhook-controller", mgr, options)
	if err != nil {
		return err
	}

	mutatingWebhookConfigurationEventHandler := &mutatingWebhookConfigurationEventHandler{}
	err = webhookController.Watch(&source.Kind{
		Type: &admissionregistrationv1.MutatingWebhookConfiguration{},
	}, &handler.EnqueueRequestForObject{},
		predicate.Funcs{
			CreateFunc: mutatingWebhookConfigurationEventHandler.onCreateFunc(webhookName),
			UpdateFunc: mutatingWebhookConfigurationEventHandler.onUpdateFunc(webhookName),
			DeleteFunc: mutatingWebhookConfigurationEventHandler.onDeleteFunc(webhookName),
		})
	if err != nil {
		log.Error(err, "Failed to watch mutatingWebhookConfiguration")
		return err
	}

	return
}
