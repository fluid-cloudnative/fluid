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

package jindocache

import (
	"fmt"
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubectl"
	"gopkg.in/yaml.v2"
)

func (e *JindoCacheEngine) setupMasterInernal() (err error) {
	var (
		chartName = utils.GetChartsDirectory() + "/jindocache"
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

func (e *JindoCacheEngine) generateJindoValueFile() (valueFileName string, err error) {
	// why need to delete configmap e.name+"-jindofs-config" ? Or it should be
	// err = kubeclient.DeleteConfigMap(e.Client, e.name+"-jindofs-config", e.namespace)
	err = kubeclient.DeleteConfigMap(e.Client, e.getConfigmapName(), e.namespace)
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

	err = kubectl.CreateConfigMapFromFile(e.getConfigmapName(), "data", valueFileName, e.namespace)
	if err != nil {
		return
	}
	return valueFileName, err
}

func (e *JindoCacheEngine) getConfigmapName() string {
	return e.name + "-" + e.runtimeType + "-values"
}
