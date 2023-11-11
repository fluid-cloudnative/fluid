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

package mountpropagationinjector

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

const Name = "MountPropagationInjector"

var (
	log = ctrl.Log.WithName(Name)
)

type MountPropagationInjector struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *MountPropagationInjector {
	return &MountPropagationInjector{
		client: c,
		name:   Name,
	}
}

func (p *MountPropagationInjector) GetName() string {
	return p.name
}

func (p *MountPropagationInjector) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if len(runtimeInfos) == 0 {
		return
	}
	datasetNames := make([]string, len(runtimeInfos))
	for name, runtimeInfo := range runtimeInfos {
		if runtimeInfo == nil {
			err = fmt.Errorf("RuntimeInfo is nil")
			shouldStop = true
			return
		}
		// do not use the runtime name, as the pvc may be the dataset mounting another dataset
		datasetNames = append(datasetNames, name)
	}
	log.V(1).Info("InjectMountPropagation", "datasetNames", datasetNames)
	utils.InjectMountPropagation(datasetNames, pod)

	return
}
