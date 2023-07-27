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

package jindofsx

import (
	"fmt"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataprocess"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func (e *JindoFSxEngine) generateDataProcessValueFile(ctx cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		err = fmt.Errorf("object is not of type DataProcess", object)
		return "", err
	}

	targetDataset, err := utils.GetDataset(e.Client, dataProcess.Spec.Dataset.Name, dataProcess.Spec.Dataset.Namespace)
	if err != nil {
		return "", errors.Wrap(err, "failed to get dataset")
	}

	dataProcessValue := dataprocess.GenDataProcessValue(targetDataset, dataProcess)

	data, err := yaml.Marshal(dataProcessValue)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal dataProcessValue of DataProcess %s/%s", object.GetNamespace(), object.GetName())
	}

	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-process-values.yaml", dataProcess.Namespace, dataProcess.Name))
	if err != nil {
		return "", errors.Wrapf(err, "failed to create temp file to store values for DataProcess %s/%s", dataProcess.Namespace, dataProcess.Name)
	}

	err = os.WriteFile(valueFile.Name(), data, 0o400)
	if err != nil {
		return "", errors.Wrapf(err, "failed to write temp file %s", valueFile.Name())
	}

	return valueFile.Name(), nil
}
