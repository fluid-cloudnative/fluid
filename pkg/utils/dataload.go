package utils

import (
	"context"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDataLoad gets the DataLoad given its name and namespace
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

// GetDataLoadJob gets the DataLoad job given its name and namespace
func GetDataLoadJob(client client.Client, name, namespace string) (*batchv1.Job, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var job batchv1.Job
	if err := client.Get(context.TODO(), key, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// GetDataLoadReleaseName returns DataLoad helm release's name given the DataLoad's name
func GetDataLoadReleaseName(name string) string {
	return fmt.Sprintf("%s-loader", name)
}

// GetDataLoadJobName returns DataLoad job's name given the DataLoad helm release's name
func GetDataLoadJobName(releaseName string) string {
	return fmt.Sprintf("%s-job", releaseName)
}

// GetDataLoadRef returns the identity of the DataLoad by combining its namespace and name.
// The identity is used for identifying current lock holder on the target dataset.
func GetDataLoadRef(name, namespace string) string {
	// namespace may contain '-', use '/' as separator
	return fmt.Sprintf("%s/%s", namespace, name)
}
