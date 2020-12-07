package jindo

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func (e *JindoEngine) setupMasterInernal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/jindo"
	)
	valuefileName, err := e.generateJindoValueFile()
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

func (e *JindoEngine) generateJindoValueFile() (valueFileName string, err error) {
	err = kubeclient.DeleteConfigMap(e.Client, e.name+"-jindofs-config", e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to clean value files")
	}
	value, err := e.transform(e.runtime)
	if err != nil {
		return
	}
	data, err := yaml.Marshal(value)
	if err != nil {
		return
	}
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
	return valueFileName, err
}
