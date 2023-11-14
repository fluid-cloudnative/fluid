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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetDataProcess(client client.Client, name, namespace string) (*datav1alpha1.DataProcess, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	var dataprocess datav1alpha1.DataProcess
	if err := client.Get(context.TODO(), key, &dataprocess); err != nil {
		return nil, err
	}

	return &dataprocess, nil
}

// GetDataProcessReleaseName returns the helm release name given the DataProcess's name.
func GetDataProcessReleaseName(name string) string {
	return fmt.Sprintf("%s-processor", name)
}

func GetDataProcessJobName(releaseName string) string {
	return fmt.Sprintf("%s-job", releaseName)
}
