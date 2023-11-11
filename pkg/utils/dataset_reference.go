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
