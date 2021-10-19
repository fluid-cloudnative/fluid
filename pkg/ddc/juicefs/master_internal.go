/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// setup fuse
func (j *JuiceFSEngine) setupMasterInternal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/" + common.JuiceFSChart
	)

	runtime, err := j.getRuntime()
	if err != nil {
		return
	}

	valuefileName, err := j.generateJuicefsValueFile(runtime)
	if err != nil {
		return
	}

	found, err := helm.CheckRelease(j.name, j.namespace)
	if err != nil {
		return
	}

	if found {
		j.Log.Info("The release is already installed", "name", j.name, "namespace", j.namespace)
		return
	}

	return helm.InstallRelease(j.name, j.namespace, valuefileName, chartName)
}

// generate juicefs struct
func (j *JuiceFSEngine) generateJuicefsValueFile(runtime *datav1alpha1.JuiceFSRuntime) (valueFileName string, err error) {
	// labelName := common.LabelAnnotationStorageCapacityPrefix + e.runtimeType + "-" + e.name
	// configmapName := e.name + "-" + e.runtimeType + "-values"
	//1. Transform the runtime to value
	value, err := j.transform(runtime)
	if err != nil {
		return
	}

	j.Log.Info("Generate values", "value", value)

	data, err := yaml.Marshal(value)
	if err != nil {
		return
	}

	//2. Get the template value file
	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-values.yaml", j.name, j.runtimeType))
	if err != nil {
		j.Log.Error(err, "failed to create value file", "valueFile", valueFile.Name())
		return valueFileName, err
	}

	valueFileName = valueFile.Name()
	j.Log.V(1).Info("Save the values file", "valueFile", valueFileName)

	err = ioutil.WriteFile(valueFileName, data, 0400)
	if err != nil {
		return
	}

	return valueFileName, err
}
