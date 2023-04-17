package kubeclient

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDeployment gets the deployment by name and namespace
func GetDeployment(c client.Client, key types.NamespacedName) (deploy *appsv1.Deployment, err error) {
	deploy = &appsv1.Deployment{}
	err = c.Get(context.TODO(), key, deploy)
	return deploy, err
}
