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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDataOperationJobByCronjob gets the DataOperation(i.e. DataMigrate, DataLoad) job by cronjob given its name and namespace
func ListDataOperationJobByCronjob(c client.Client, cronjobNamespacedName types.NamespacedName) ([]batchv1.Job, error) {
	jobLabelSelector, err := labels.Parse(fmt.Sprintf("cronjob=%s", cronjobNamespacedName.Name))
	if err != nil {
		return nil, err
	}
	var jobList batchv1.JobList
	if err := c.List(context.TODO(), &jobList, &client.ListOptions{
		LabelSelector: jobLabelSelector,
		Namespace:     cronjobNamespacedName.Namespace,
	}); err != nil {
		return nil, err
	}
	return jobList.Items, nil
}
