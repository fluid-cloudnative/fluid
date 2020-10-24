package utils

import (
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDataLoad(client client.Client, name, namespace string) (*datav1alpha1.DataLoad, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var dataload datav1alpha1.DataLoad
	if err := client.Get(context.TODO(), key, &dataload); err != nil {
		return nil, err
	}
	return &dataload, nil
}
