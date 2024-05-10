/*
Copyright 2023 The Fluid Authors.

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

package utils

import (
	"context"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilpointer "k8s.io/utils/pointer"
)

func GetRuntimeByCategory(runtimes []datav1alpha1.Runtime, category common.Category) (index int, runtime *datav1alpha1.Runtime) {
	if runtimes == nil {
		return -1, nil
	}
	for i := range runtimes {
		if runtimes[i].Category == category {
			return i, &runtimes[i]
		}
	}
	return -1, nil
}

// CreateRuntimeForReferenceDatasetIfNotExist creates runtime for ReferenceDataset
func CreateRuntimeForReferenceDatasetIfNotExist(client client.Client, dataset *datav1alpha1.Dataset) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := GetThinRuntime(client,
			dataset.GetName(),
			dataset.GetNamespace())
		// 1. if err is null, which indicates that the runtime exists, then return
		if err == nil {
			runtimeToUpdate := runtime.DeepCopy()
			runtimeToUpdate.SetOwnerReferences([]metav1.OwnerReference{
				{
					Kind:       dataset.GetObjectKind().GroupVersionKind().Kind,
					APIVersion: dataset.APIVersion,
					Name:       dataset.GetName(),
					UID:        dataset.GetUID(),
					Controller: utilpointer.Bool(true),
				}})
			if !reflect.DeepEqual(runtimeToUpdate, runtime) {
				err = client.Update(context.TODO(), runtimeToUpdate)
				return err
			}
			return nil
		}

		// 2. If the runtime doesn't exist
		if IgnoreNotFound(err) == nil {
			var runtime datav1alpha1.ThinRuntime = datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      dataset.Name,
					Namespace: dataset.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind:       dataset.GetObjectKind().GroupVersionKind().Kind,
							APIVersion: dataset.APIVersion,
							Name:       dataset.GetName(),
							UID:        dataset.GetUID(),
							Controller: utilpointer.Bool(true),
						},
					},
				},
			}
			err = client.Create(context.TODO(), &runtime)
		}
		return err

	})

	return
}
