package kubeclient

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkIfOwnerRefMatcheExpected checks if the ownerRefence belongs to the expected owner
func checkIfOwnerRefMatcheExpected(c client.Client,
	controllerRef *metav1.OwnerReference,
	runtime runtime.Object) (matched bool, err error) {
	owner := runtime.(metav1.Object)
	parentObject, err := resolveControllerRef(c, controllerRef, owner.GetNamespace(), runtime.GetObjectKind().GroupVersionKind())
	if err != nil {
		return matched, err
	}

	matched = (parentObject.GetUID() == controllerRef.UID)

	return matched, err
}

// resolveControllerRef resolves the parent object from the
func resolveControllerRef(c client.Client, controllerRef *metav1.OwnerReference, namespace string, expectedGroupVersionKind schema.GroupVersionKind) (metav1.Object, error) {

	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != expectedGroupVersionKind.Kind {
		log.Info("Wrong Kind", "expected", expectedGroupVersionKind.Kind, "actual", controllerRef.Kind)
		return nil, fmt.Errorf("wrong kind to expect, expected %s but got %s",
			expectedGroupVersionKind.Kind,
			controllerRef.Kind)
	}

	set := &appsv1.StatefulSet{}

	err := c.Get(context.TODO(), types.NamespacedName{Name: controllerRef.Name, Namespace: namespace}, set)
	if err != nil {
		return set, err
	}

	if set.UID != controllerRef.UID {
		// The controller we found with this Name is not the same one that the
		// ControllerRef points to.
		return set, nil
	}
	return set, nil
}
