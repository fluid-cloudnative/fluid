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

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// GetDataMigrate gets the DataMigrate given its name and namespace
func GetDataMigrate(client client.Client, name, namespace string) (*datav1alpha1.DataMigrate, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var dataMigrate datav1alpha1.DataMigrate
	if err := client.Get(context.TODO(), key, &dataMigrate); err != nil {
		return nil, err
	}
	return &dataMigrate, nil
}

// GetDataMigrateReleaseName returns DataMigrate helm release's name given the DataMigrate's name
func GetDataMigrateReleaseName(name string) string {
	return fmt.Sprintf("%s-migrate", name)
}

// GetDataMigrateJobName returns DataMigrate job(or cronjob)'s name given the DataMigrate helm release's name
func GetDataMigrateJobName(releaseName string) string {
	return fmt.Sprintf("%s-migrate", releaseName)
}

func GetTargetDatasetNamespacedNameOfMigrate(client client.Client, dataMigrate *datav1alpha1.DataMigrate) (namespacedName types.NamespacedName, err error) {
	if (dataMigrate.Spec.To.DataSet == nil || dataMigrate.Spec.To.DataSet.Name == "") && (dataMigrate.Spec.From.DataSet == nil || dataMigrate.Spec.From.Dataset.Name == "") {
		err = fmt.Errorf("invalid spec: either %v or %v must be set", field.NewPath("spec").Child("to").Child("dataset"), field.NewPath("spec").Child("from").Child("dataset"))
		return
	}

	// if runtimeType is not specified, we will simply use the toDataset as the first choice, and the fromDataset as the second.
	if dataMigrate.Spec.RuntimeType == "" {
		var targetDataset *datav1alpha1.DatasetToMigrate
		if dataMigrate.Spec.To.DataSet != nil && len(dataMigrate.Spec.To.DataSet.Name) > 0 {
			targetDataset = dataMigrate.Spec.To.DataSet
		} else {
			targetDataset = dataMigrate.Spec.From.DataSet
		}

		var namespace string = targetDataset.Namespace
		if len(namespace) == 0 {
			namespace = dataMigrate.Namespace
		}
		namespacedName = types.NamespacedName{
			Namespace: namespace,
			Name:      targetDataset.Name,
		}
		return
	}

	// when runtimeType is explicitly set, check whether toDataset or fromDataset has the runtimeType
	datasetsToCheck := []*datav1alpha1.DatasetToMigrate{dataMigrate.Spec.To.DataSet, dataMigrate.Spec.From.DataSet}
	checkedRuntimeTypes := []string{}
	for _, toCheck := range datasetsToCheck {
		if toCheck != nil && len(toCheck.Name) > 0 {
			var namespace string = toCheck.Namespace
			if len(namespace) == 0 {
				namespace = dataMigrate.Namespace
			}

			dataset, innerErr := GetDataset(client, toCheck.Name, namespace)
			if innerErr != nil {
				err = errors.Wrapf(innerErr, "failed to get dataset \"%s/%s\" from DataMigrate \"%s/%s\"'s spec (%v)", namespace, toCheck.Name, dataMigrate.Namespace, dataMigrate.Name)
				return
			}

			index, boundedRuntime := GetRuntimeByCategory(dataset.Status.Runtimes, common.AccelerateCategory)
			if index == -1 {
				err = fmt.Errorf("bounded accelerate runtime not ready for dataset \"%s/%s\"", namespace, toCheck.Name)
				return
			}

			if boundedRuntime.Type == dataMigrate.Spec.RuntimeType {
				namespacedName = types.NamespacedName{
					Namespace: namespace,
					Name:      toCheck.Name,
				}
				return
			}

			// boundedRuntime.Type is not matching with dataMigrate.Spec.RuntimeType
			checkedRuntimeTypes = append(checkedRuntimeTypes, boundedRuntime.Type)
		}
	}

	// no bounded dataset with the runtimeType can be found
	err = fmt.Errorf("invalid spec: the valid runtime type of the dataset is %v, but the specified runtime type in dataMigrate is %s",
		checkedRuntimeTypes, dataMigrate.Spec.RuntimeType)
	return
}

func GetTargetDatasetOfMigrate(client client.Client, dataMigrate *datav1alpha1.DataMigrate) (targetDataset *datav1alpha1.Dataset, err error) {
	namespacedName, err := GetTargetDatasetNamespacedNameOfMigrate(client, dataMigrate)
	if err != nil {
		return
	}

	return GetDataset(client, namespacedName.Name, namespacedName.Namespace)
}
