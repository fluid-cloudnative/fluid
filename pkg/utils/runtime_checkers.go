package utils

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CheckObject(client client.Client, key types.NamespacedName, obj client.Object) (found bool, err error) {
	if err = client.Get(context.TODO(), key, obj); err != nil {
		if IgnoreNotFound(err) == nil {
			log.V(1).Info("Failed in finding the object, skip.", "runtime", obj)
			err = nil
		}
		return
	}
	found = true
	log.Info("Succeed in finding the object", "runtime", obj)
	return
}
