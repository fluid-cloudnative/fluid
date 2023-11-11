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

package alluxio

import (
	"fmt"
	"os"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
	"sigs.k8s.io/yaml"
)

// setup the cache master
func (e *AlluxioEngine) setupMasterInternal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/" + common.AlluxioChart
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
	valueFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("%s-%s-values.yaml", e.name, e.runtimeType))
	if err != nil {
		e.Log.Error(err, "failed to create value file", "valueFile", valueFile.Name())
		return valueFileName, err
	}

	valueFileName = valueFile.Name()
	e.Log.V(1).Info("Save the values file", "valueFile", valueFileName)

	err = os.WriteFile(valueFileName, data, 0400)
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

func (e *AlluxioEngine) getMountConfigmapName() string {
	return e.name + "-mount-config"
}
