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
