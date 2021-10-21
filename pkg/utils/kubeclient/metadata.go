package kubeclient

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// compareOwnerRefMatcheWithExpected checks if the ownerRefence belongs to the expected owner
func compareOwnerRefMatcheWithExpected(c client.Client,
	controllerRef *metav1.OwnerReference,
	namespace string,
	target runtime.Object) (matched bool, err error) {

	kind := target.GetObjectKind()
	controllerObject, err := resolveControllerRef(c, controllerRef, namespace, kind, target.DeepCopyObject())
	if err != nil || controllerObject == nil {
		return matched, err
	}

	// gv, _ := schema.ParseGroupVersion(runtime.GetObjectKind().GroupVersionKind().Group)

	// if gv.Group ==controllerRef.APIVersion

	// controllerRef.

	targetObject, err := meta.Accessor(target)
	if err != nil {
		return matched, err
	}

	matched = (controllerRef.UID == targetObject.GetUID())

	return matched, err
}

// resolveControllerRef resolves the parent object from the
func resolveControllerRef(c client.Client, controllerRef *metav1.OwnerReference, controllerNamespace string, objectKind schema.ObjectKind, obj runtime.Object) (result metav1.Object, err error) {
	if controllerRef == nil {
		log.Info("No controllerRef found")
		return nil, nil
	}

	controllerRefGV, err := schema.ParseGroupVersion(controllerRef.APIVersion)
	if err != nil {
		return nil, err
	}

	kind := objectKind.GroupVersionKind().Kind
	group := objectKind.GroupVersionKind().Group

	// We can't look up by UID, so look up by Name and then verify UID.
	// Don't even try to look up by Name if it's the wrong Kind.
	if controllerRef.Kind != kind ||
		controllerRefGV.Group != group {
		log.Info("Wrong Kind", "expected", kind, "actual", controllerRef.Kind)
		// return nil, fmt.Errorf("wrong kind to expect, expected %s but got %s",
		// 	expectedGroupVersionKind.Kind,
		// 	controllerRef.Kind)
		return nil, nil
	}

	err = c.Get(context.TODO(), types.NamespacedName{Name: controllerRef.Name, Namespace: controllerNamespace}, obj)
	if err != nil {
		return result, client.IgnoreNotFound(err)
	}

	return meta.Accessor(obj)
}
