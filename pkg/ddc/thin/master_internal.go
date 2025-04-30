/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"fmt"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (t *ThinEngine) setupMasterInternal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/" + common.ThinChart
	)

	runtime, err := t.getRuntime()
	if err != nil {
		return
	}

	profile, err := t.getThinRuntimeProfile()
	if err != nil && !errors.IsNotFound(err) {
		return
	}

	valuefileName, err := t.generateThinValueFile(runtime, profile)
	if err != nil {
		return
	}

	found, err := helm.CheckRelease(t.name, t.namespace)
	if err != nil {
		return
	}

	if found {
		t.Log.Info("The release is already installed", "name", t.name, "namespace", t.namespace)
		return
	}

	return helm.InstallRelease(t.name, t.namespace, valuefileName, chartName)
}

func (t *ThinEngine) generateThinValueFile(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile) (valueFileName string, err error) {
	enableRuntimeHelmValueConfig := t.ifRuntimeHelmValueEnable()

	if enableRuntimeHelmValueConfig {
		//0. Check if the configmap exists
		err = kubeclient.DeleteConfigMap(t.Client, t.getHelmValuesConfigMapName(), t.namespace)
		if err != nil {
			t.Log.Error(err, "Failed to clean value files")
			return
		}
	}

	// labelName := common.LabelAnnotationStorageCapacityPrefix + e.runtimeType + "-" + e.name
	// configmapName := e.name + "-" + e.runtimeType + "-values"
	//1. Transform the runtime to value
	value, err := t.transform(runtime, profile)
	if err != nil {
		return
	}

	t.Log.Info("Generate values", "value", value)

	data, err := yaml.Marshal(value)
	if err != nil {
		return
	}

	//2. Get the template value file
	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-values.yaml", t.name, t.engineImpl))
	if err != nil {
		t.Log.Error(err, "failed to create value file", "valueFile", valueFile.Name())
		return valueFileName, err
	}

	valueFileName = valueFile.Name()
	t.Log.Info("Save the values file", "valueFile", valueFileName)

	err = os.WriteFile(valueFileName, data, 0400)
	if err != nil {
		return
	}

	if enableRuntimeHelmValueConfig {
		//3. Save the configfile into configmap
		runtimeInfo := t.runtimeInfo
		ownerDatasetId := utils.GetDatasetId(runtimeInfo.GetNamespace(), runtimeInfo.GetName(), runtimeInfo.GetOwnerDatasetUID())
		err = kubeclient.CreateConfigMap(t.Client, t.getHelmValuesConfigMapName(), t.namespace, "data", data, ownerDatasetId)
		if err != nil {
			return
		}
	}

	return valueFileName, err
}

func (t *ThinEngine) ifRuntimeHelmValueEnable() bool {
	runtime := t.runtime
	if runtime == nil {
		return false
	}
	value, exist := runtime.Annotations[common.AnnotationEnableRuntimeHelmValueConfig]
	if !exist {
		return false
	}
	enableRuntimeHelmValueConfig, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return enableRuntimeHelmValueConfig
}

func (t *ThinEngine) getHelmValuesConfigMapName() string {
	return fmt.Sprintf("%s-%s-values", t.name, t.engineImpl)
}
