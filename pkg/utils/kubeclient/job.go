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

package kubeclient

import (
	"context"
	"k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetJob gets the job given its name and namespace
func GetJob(client client.Client, name, namespace string) (*v1.Job, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var job v1.Job
	if err := client.Get(context.TODO(), key, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
