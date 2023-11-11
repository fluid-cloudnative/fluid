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
package mutating

import (
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func buildRuntimeInfoInternal(client client.Client,
	pvc *corev1.PersistentVolumeClaim,
	log logr.Logger) (runtimeInfo base.RuntimeInfoInterface, err error) {
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "mutating.buildRuntimeInfoInternalByPVC",
			"pvc.name", pvc.GetName(), "pvc.namespace", pvc.GetNamespace())
	}
	namespace := pvc.GetNamespace()
	if len(namespace) == 0 {
		namespace = corev1.NamespaceDefault
	}
	pvcName := pvc.GetName()
	if datasetName, exists := common.GetManagerDatasetFromLabels(pvc.Labels); exists {
		pvcName = datasetName
	}

	dataset, err := utils.GetDataset(client, pvcName, namespace)
	if err != nil {
		return
	}

	if dataset.Status.Phase == v1alpha1.NotBoundDatasetPhase || dataset.Status.Phase == v1alpha1.NoneDatasetPhase {
		_, cond := utils.GetDatasetCondition(dataset.Status.Conditions, v1alpha1.DatasetNotReady)
		if cond != nil {
			err = fmt.Errorf("dataset \"%s/%s\" not ready because %s", dataset.Namespace, dataset.Name, cond.Message)
			return
		}
		err = fmt.Errorf("dataset \"%s/%s\" not bound", dataset.Namespace, dataset.Name)
		return
	}

	runtimeInfo, err = base.GetRuntimeInfo(client, pvcName, namespace)
	if err != nil {
		log.Error(err, "unable to get runtimeInfo, get failure", "runtime", pvc.GetName(), "namespace", namespace)
		return
	}
	runtimeInfo.SetDeprecatedNodeLabel(false)
	return
}
