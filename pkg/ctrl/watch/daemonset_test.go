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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestDaemonSetEventHandler_OnCreateFunc(t *testing.T) {

	// 1. the Object is RuntimeInterface
	createEvent := event.CreateEvent{
		Object: &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       datav1alpha1.JindoRuntimeKind,
						APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						Controller: utilpointer.Bool(true),
					},
				},
			},
		},
	}
	daemonsetEventHandler := &daemonsetEventHandler{}

	f := daemonsetEventHandler.onCreateFunc(&FakeRuntimeReconciler{})
	predicate := f(createEvent)

	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", createEvent)
	}

	// 2. the Object is not RuntimeInterface
	createEvent.Object = &corev1.Pod{}
	predicate = f(createEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", createEvent)
	}

	// 3. Skip the runtime which is deleting
	createEvent = event.CreateEvent{
		Object: &appsv1.DaemonSet{},
	}
	createEvent.Object.SetDeletionTimestamp(&metav1.Time{})
	predicate = f(createEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", createEvent)
	}

}

func TestDaemonSetEventHandler_OnUpdateFunc(t *testing.T) {

	updateRuntimeEvent := event.UpdateEvent{
		ObjectOld: &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       datav1alpha1.JindoRuntimeKind,
						APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						Controller: utilpointer.Bool(true),
					},
				},
				ResourceVersion: "123",
			},
		},
		ObjectNew: &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       datav1alpha1.JindoRuntimeKind,
						APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						Controller: utilpointer.Bool(true),
					},
				},
				ResourceVersion: "456",
			},
		},
	}
	daemonsetEventHandler := &daemonsetEventHandler{}

	f := daemonsetEventHandler.onUpdateFunc(&FakeRuntimeReconciler{})
	predicate := f(updateRuntimeEvent)

	// 1. expect the updateEvent is validated
	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", updateRuntimeEvent)
	}

	// 2. expect the updateEvent is not validated due to the resource version is equal
	updateRuntimeEvent.ObjectOld.SetResourceVersion("456")
	predicate = f(updateRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updateRuntimeEvent)
	}

	// 3. expect the updateEvent is not validated due to the object is not kind of runtimeInterface
	updateRuntimeEvent.ObjectOld = &corev1.Pod{}
	updateRuntimeEvent.ObjectNew = &corev1.Pod{}
	predicate = f(updateRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updateRuntimeEvent)
	}

	// 4. expect the updateEvent is not validate due the old Object  is not kind of the runtimeInterface
	updateRuntimeEvent.ObjectNew = &appsv1.DaemonSet{}
	predicate = f(updateRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updateRuntimeEvent)
	}
}

func TestDaemonSetEventHandler_OnDeleteFunc(t *testing.T) {

	// 1. the Object is RuntimeInterface
	delRuntimeEvent := event.DeleteEvent{
		Object: &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       datav1alpha1.JindoRuntimeKind,
					APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
					Controller: utilpointer.Bool(true),
				},
			},
		}},
	}
	daemonsetEventHandler := &daemonsetEventHandler{}

	f := daemonsetEventHandler.onDeleteFunc(&FakeRuntimeReconciler{})
	predicate := f(delRuntimeEvent)

	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", delRuntimeEvent)
	}

	// 2. the Object is not RuntimeInterface
	delRuntimeEvent.Object = &corev1.Pod{}
	predicate = f(delRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", delRuntimeEvent)
	}
}
