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
	"fmt"
	"reflect"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/utils/ptr"
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

// datasetControllerOwnerReference builds the controller ownerReference which points to the given dataset.
// Kind and APIVersion are resolved via the shared transformer helper (backed by the scheme), falling back to the
// well-known values of the Dataset CRD as a last resort if unavailable. The owner based watch of the dataset
// controller resolves a dependent through those two fields, so they must be set.
func datasetControllerOwnerReference(dataset *datav1alpha1.Dataset) metav1.OwnerReference {
	ref := transformer.GenerateOwnerReferenceFromObject(dataset)
	kind := ref.Kind
	if len(kind) == 0 {
		kind = datav1alpha1.Datasetkind
	}
	apiVersion := ref.APIVersion
	if len(apiVersion) == 0 {
		apiVersion = datav1alpha1.GroupVersion.String()
	}

	return metav1.OwnerReference{
		Kind:       kind,
		APIVersion: apiVersion,
		Name:       dataset.GetName(),
		UID:        dataset.GetUID(),
		Controller: ptr.To(true),
	}
}

// CreateRuntimeForReferenceDatasetIfNotExist creates runtime for ReferenceDataset
func CreateRuntimeForReferenceDatasetIfNotExist(client client.Client, dataset *datav1alpha1.Dataset) (err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := GetThinRuntime(client,
			dataset.GetName(),
			dataset.GetNamespace())
		// 1. if err is null, which indicates that the runtime exists, then return
		if err == nil {
			// 1.1 The runtime is being deleted, it can neither be adopted nor be re-created for now.
			// Return an error (not a conflict error, so retry.RetryOnConflict won't swallow it) to
			// let the caller requeue until the terminating runtime is really gone.
			if HasDeletionTimestamp(runtime.ObjectMeta) {
				return fmt.Errorf("the ThinRuntime %s/%s is terminating, wait for it to be deleted before creating a new one for the reference dataset",
					runtime.GetNamespace(), runtime.GetName())
			}

			runtimeToUpdate := runtime.DeepCopy()
			runtimeToUpdate.SetOwnerReferences([]metav1.OwnerReference{
				datasetControllerOwnerReference(dataset)})
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
						datasetControllerOwnerReference(dataset),
					},
					Labels: map[string]string{
						common.LabelAnnotationDatasetId: GetDatasetId(dataset.GetNamespace(), dataset.GetName(), string(dataset.GetUID())),
					},
				},
			}
			err = client.Create(context.TODO(), &runtime)
		}
		return err

	})

	return
}
