package utils

import (
	"context"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CheckAlluxioRuntime checks Alluxio Runtime object with the given name and namespace
func CheckAlluxioRuntime(client client.Client, key types.NamespacedName) (found bool, err error) {
	var runtime data.AlluxioRuntime
	return checkObject(client, key, &runtime)
}

// CheckJindoRuntime checks Jindo Runtime object with the given name and namespace
func CheckJindoRuntime(client client.Client, key types.NamespacedName) (bool, error) {
	var runtime data.JindoRuntime
	return checkObject(client, key, &runtime)
}

// CheckGooseFSRuntime checks GooseFS Runtime object with the given name and namespace
func CheckGooseFSRuntime(client client.Client, key types.NamespacedName) (bool, error) {
	var runtime data.GooseFSRuntime
	return checkObject(client, key, &runtime)
}

// CheckJuiceFSRuntime checks JuiceFS Runtime object with the given name and namespace
func CheckJuiceFSRuntime(client client.Client, key types.NamespacedName) (bool, error) {
	var runtime data.JuiceFSRuntime
	return checkObject(client, key, &runtime)
}

func checkObject(client client.Client, key types.NamespacedName, obj client.Object) (found bool, err error) {
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
