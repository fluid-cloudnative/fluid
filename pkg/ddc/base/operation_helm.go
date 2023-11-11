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
			chartName = operation.GetChartsDirectory() + "/" + ctx.RuntimeType
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
