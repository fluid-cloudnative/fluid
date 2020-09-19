/*

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

package alluxio

import (
	"fmt"
	"io/ioutil"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
	yaml "gopkg.in/yaml.v2"
)

// setup the cache master
func (e *AlluxioEngine) setupMasterInternal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/" + common.ALLUXIO_CHART
	)

	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	valuefileName, err := e.generateAlluxioValueFile(runtime)
	if err != nil {
		return
	}

	found, err := helm.CheckRelease(e.name, e.namespace)
	if err != nil {
		return
	}

	if found {
		e.Log.Info("The release is already installed", "name", e.name, "namespace", e.namespace)
		return
	}

	return helm.InstallRelease(e.name, e.namespace, valuefileName, chartName)
}

// generate alluxio struct
func (e *AlluxioEngine) generateAlluxioValueFile(runtime *datav1alpha1.AlluxioRuntime) (valueFileName string, err error) {

	//0. Check if the configmap exists
	err = kubeclient.DeleteConfigMap(e.Client, e.getConfigmapName(), e.namespace)

	if err != nil {
		e.Log.Error(err, "Failed to clean value files")
		return
	}

	// labelName := common.LabelAnnotationStorageCapacityPrefix + e.runtimeType + "-" + e.name
	// configmapName := e.name + "-" + e.runtimeType + "-values"
	//1. Transform the runtime to value
	value, err := e.transform(runtime)
	if err != nil {
		return
	}

	e.Log.Info("Generate values", "value", value)

	data, err := yaml.Marshal(value)
	if err != nil {
		return
	}

	//2. Get the template value file
	valueFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s-%s-values.yaml", e.name, e.runtimeType))
	if err != nil {
		e.Log.Error(err, "failed to create value file", "valueFile", valueFile.Name())
		return valueFileName, err
	}

	valueFileName = valueFile.Name()
	e.Log.V(1).Info("Save the values file", "valueFile", valueFileName)

	err = ioutil.WriteFile(valueFileName, data, 0400)
	if err != nil {
		return
	}

	//3. Save the configfile into configmap
	err = kubectl.CreateConfigMapFromFile(e.getConfigmapName(), "data", valueFileName, e.namespace)
	if err != nil {
		return
	}

	return valueFileName, err
}

func (e *AlluxioEngine) getConfigmapName() string {
	return e.name + "-" + e.runtimeType + "-values"
}
