package kubeclient

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func resolveControllerRef(c client.Client, expectedGroupVersionKind schema.GroupVersionKind, namespace string, controllerRef *metav1.OwnerReference) (obj metav1.Object, err error) {

	if controllerRef.Kind != expectedGroupVersionKind.Kind {
		return nil, nil
	}
}
