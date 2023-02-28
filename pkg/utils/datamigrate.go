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

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

// GetDataMigrateJob gets the DataMigrate job given its name and namespace
func GetDataMigrateJob(client client.Client, name, namespace string) (*batchv1.Job, error) {
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

// GetDataMigrateReleaseName returns DataMigrate helm release's name given the DataMigrate's name
func GetDataMigrateReleaseName(name string) string {
	return fmt.Sprintf("%s-migrate", name)
}

// GetDataMigrateJobName returns DataMigrate job's name given the DataMigrate helm release's name
func GetDataMigrateJobName(releaseName string) string {
	return fmt.Sprintf("%s-migrate", releaseName)
}

// GetDataMigrateRef returns the identity of the DataMigrate by combining its namespace and name.
// The identity is used for identifying current lock holder on the target dataset.
func GetDataMigrateRef(name, namespace string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func GetTargetDatasetOfMigrate(client client.Client, dataMigrate datav1alpha1.DataMigrate) (targetDataset *datav1alpha1.Dataset, err error) {
	var fromDataset, toDataset *datav1alpha1.Dataset
	if dataMigrate.Spec.To.DataSet != nil && dataMigrate.Spec.To.DataSet.Name != "" {
		toDataset, err = GetDataset(client, dataMigrate.Spec.To.DataSet.Name, dataMigrate.Spec.To.DataSet.Namespace)
		if err != nil {
			return
		}

		// if runtimeType is not specified, we will use the toDataset as the targetDataset
		if dataMigrate.Spec.RuntimeType == "" {
			targetDataset = toDataset
			return
		}

		// if runtimeType is specified, check if toDataset's accelerate runtime type is the same as the runtimeType
		index, boundedRuntime := GetRuntimeByCategory(toDataset.Status.Runtimes, common.AccelerateCategory)
		if index == -1 {
			err = fmt.Errorf("bounded accelerate runtime not ready")
			return
		}
		if boundedRuntime.Type == dataMigrate.Spec.RuntimeType {
			targetDataset = toDataset
			return
		}
	}
	if dataMigrate.Spec.From.DataSet != nil && dataMigrate.Spec.From.DataSet.Name != "" {
		fromDataset, err = GetDataset(client, dataMigrate.Spec.From.DataSet.Name, dataMigrate.Spec.From.DataSet.Namespace)
		if err != nil {
			return
		}
		// if runtimeType is not specified, we will use the fromDataset as the targetDataset
		if dataMigrate.Spec.RuntimeType == "" {
			targetDataset = fromDataset
			return
		}

		// if runtimeType is specified, check if fromDataset's accelerate runtime type is the same as the runtimeType
		index, boundedRuntime := GetRuntimeByCategory(fromDataset.Status.Runtimes, common.AccelerateCategory)
		if index == -1 {
			err = fmt.Errorf("bounded accelerate runtime not ready")
			return
		}
		if boundedRuntime.Type == dataMigrate.Spec.RuntimeType {
			targetDataset = fromDataset
			return
		}
	}

	return nil, apierrors.NewNotFound(schema.GroupResource{
		Group:    datav1alpha1.Group,
		Resource: datav1alpha1.Version,
	}, "dataset")
}
