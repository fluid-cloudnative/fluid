package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaceKey(metaObj metav1.Object) string {
	name := metaObj.GetName()
	namespace := metaObj.GetNamespace()
	if len(namespace) == 0 {
		namespace = corev1.NamespaceDefault
	}
	return fmt.Sprintf("%s/%s", namespace, name)
}
