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

// GetDataLoad gets the DataLoad given its name and namespace
func GetDataLoad(client client.Client, name, namespace string) (*datav1alpha1.DataLoad, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var dataload datav1alpha1.DataLoad
	if err := client.Get(context.TODO(), key, &dataload); err != nil {
		return nil, err
	}
	return &dataload, nil
}

// GetDataLoadReleaseName returns DataLoad helm release's name given the DataLoad's name
func GetDataLoadReleaseName(name string) string {
	return fmt.Sprintf("%s-loader", name)
}

// GetDataLoadJobName returns DataLoad job's name given the DataLoad helm release's name
func GetDataLoadJobName(releaseName string) string {
	return fmt.Sprintf("%s-job", releaseName)
}
