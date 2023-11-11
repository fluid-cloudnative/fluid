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

package fusesidecar

import (
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/application/inject"
	"github.com/fluid-cloudnative/fluid/pkg/application/inject/fuse"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin is for pods without a mounted dataset.
   They should prefer nods without cache workers on them.

*/

const Name string = "FuseSidecar"

type FuseSidecar struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) *FuseSidecar {
	return &FuseSidecar{
		client: c,
		name:   Name,
	}
}

func (p *FuseSidecar) GetName() string {
	return p.name
}

func (p *FuseSidecar) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	// if the pod has no mounted datasets, should exit and call other plugins
	if utils.IsTimeTrackerDebugEnabled() {
		defer utils.TimeTrack(time.Now(), "FuseSidecar.Mutate",
			"pod.name", pod.GetName(), "pvc.namespace", pod.GetNamespace())
	}
	if len(runtimeInfos) == 0 {
		return
	}

	var injector inject.Injector = fuse.NewInjector(p.client)
	out, err := injector.InjectPod(pod, runtimeInfos)
	if err != nil {
		return shouldStop, err
	}
	out.DeepCopyInto(pod)
	return
}
