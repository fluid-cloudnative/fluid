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
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RemoveNotFoundDatasetRef checks datasetRef existence and remove if reference dataset is not found
func RemoveNotFoundDatasetRef(client client.Client, dataset datav1alpha1.Dataset, log logr.Logger) ([]string, error) {
	// check reference dataset existence
	var datasetRefToUpdate []string
	for _, datasetRefName := range dataset.Status.DatasetRef {
		namespacedName := strings.Split(datasetRefName, "/")
		if len(namespacedName) < 2 {
			log.Info("invalid datasetRef", "datasetRef", datasetRefName)
			continue
		}
		dataset, err := GetDataset(client, namespacedName[1], namespacedName[0])
		if err != nil && IgnoreNotFound(err) != nil {
			log.Error(err, "Failed to get reference dataset", "DatasetDeleteError", datasetRefName)
			return dataset.Status.DatasetRef, err
		}
		if dataset != nil {
			datasetRefToUpdate = append(datasetRefToUpdate, datasetRefName)
		} else {
			log.V(1).Info("dataset not found, remove it from datasetRef", "remove dataset namespacedName", namespacedName)
		}
	}
	return datasetRefToUpdate, nil
}
