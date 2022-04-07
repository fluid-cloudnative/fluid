/*
Copyright 2021 The Fluid Authors.

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
