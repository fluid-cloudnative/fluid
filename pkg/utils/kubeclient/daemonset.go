package kubeclient

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDaemonSet gets daemonSet with given name and namespace. If not found, no error will be returned.
func GetDaemonSet(client client.Client, name, namespace string) (*appsv1.DaemonSet, error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	daemonSet := &appsv1.DaemonSet{}

	if err := client.Get(context.TODO(), key, daemonSet); err != nil {
		return nil, err
	}

	return daemonSet, nil
}
