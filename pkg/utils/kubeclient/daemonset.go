/*
Copyright 2023 The Fluid Author.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeclient

import (
	"context"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDaemonset gets the daemonset by name and namespace
func GetDaemonset(c client.Reader, name string, namespace string) (ds *appsv1.DaemonSet, err error) {
	ds = &appsv1.DaemonSet{}
	err = c.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, ds)

	return ds, err
}

func UpdateDaemonSetUpdateStrategy(client client.Client, name, namespace string, strategy appsv1.DaemonSetUpdateStrategy) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		ds, err := GetDaemonset(client, name, namespace)
		if err != nil {
			return err
		}
		dsToUpdate := ds.DeepCopy()
		dsToUpdate.Spec.UpdateStrategy = strategy
		if !reflect.DeepEqual(ds.Spec.UpdateStrategy, dsToUpdate.Spec.UpdateStrategy) {
			err = client.Update(context.TODO(), dsToUpdate)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
