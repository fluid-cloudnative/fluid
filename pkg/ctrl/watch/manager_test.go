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
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestIsObjectInManaged(t *testing.T) {
	reconciler := &FakeRuntimeReconciler{}

	_ = reconciler.ControllerName()
	_ = reconciler.ManagedResource()

	matchedPod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:       datav1alpha1.JindoRuntimeKind,
				APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
				Controller: utilpointer.BoolPtr(true),
			},
		},
	}}

	isManaged := isObjectInManaged(&matchedPod, reconciler)

	if !isManaged {
		t.Errorf("The object %v is not managed by %v which is not expected.", matchedPod, reconciler)
	}

	notmatchedPod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:       datav1alpha1.JindoRuntimeKind,
				APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
			},
		},
	}}

	isManaged = isObjectInManaged(&notmatchedPod, reconciler)

	if isManaged {
		t.Errorf("The object %v is  managed by %v which is not expected.", matchedPod, reconciler)
	}

	notmatchedPod = corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		OwnerReferences: []metav1.OwnerReference{
			{
				Kind:       "Test",
				APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
			},
		},
	}}

	isManaged = isObjectInManaged(&notmatchedPod, reconciler)

	if isManaged {
		t.Errorf("The object %v is  managed by %v which is not expected.", matchedPod, reconciler)
	}

}
