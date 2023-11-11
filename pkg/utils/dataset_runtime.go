/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
					Controller: utilpointer.BoolPtr(true),
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
							Controller: utilpointer.BoolPtr(true),
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
