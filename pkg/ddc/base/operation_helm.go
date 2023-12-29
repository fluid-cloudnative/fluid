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

package base

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
)

func InstallDataOperationHelmIfNotExist(ctx cruntime.ReconcileRequestContext, operation dataoperation.OperationInterface,
	yamlGenerator DataOperatorYamlGenerator) (err error) {
	log := ctx.Log.WithName("InstallDataOperationHelmIfNotExist")

	operationTypeName := string(operation.GetOperationType())
	releaseNamespacedName := operation.GetReleaseNameSpacedName()
	var existed bool
	existed, err = helm.CheckRelease(releaseNamespacedName.Name, releaseNamespacedName.Namespace)
	if err != nil {
		log.Error(err, "failed to check if release exists", "releaseName", releaseNamespacedName.Name,
			"namespace", releaseNamespacedName.Namespace)
		return err
	}

	// 2. install the helm chart if not exists
	if !existed {
		log.Info(fmt.Sprintf("%s job helm chart not installed yet, will install", operationTypeName))
		var valueFileName string
		valueFileName, err = yamlGenerator.GetDataOperationValueFile(ctx, operation)
		if err != nil {
			log.Error(err, "failed to generate chart's value file")
			return err
		}

		var chartName string
		if operation.GetOperationType() == datav1alpha1.DataProcessType {
			// for DataProcess, all engine share the same chart
			chartName = operation.GetChartsDirectory() + "/" + "common"
		} else {
			chartName = operation.GetChartsDirectory() + "/" + ctx.EngineImpl
		}

		err = helm.InstallRelease(releaseNamespacedName.Name, releaseNamespacedName.Namespace, valueFileName, chartName)
		if err != nil {
			log.Error(err, "failed to install chart")
			return err
		}
		log.Info(fmt.Sprintf("%s job helm chart successfully installed", operationTypeName),
			"namespace", releaseNamespacedName.Namespace, "releaseName", releaseNamespacedName.Name)

	}

	return nil
}
