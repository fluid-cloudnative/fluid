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

package plugins

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/fusesidecar"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/mountpropagationinjector"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/nodeaffinitywithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/requirenodewithfuse"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MutatingHandler interface {
	// Mutate injects affinity info into pod
	// if a plugin return true, it means that no need to call other plugins
	// map[string]base.RuntimeInfoInterface's key is pvcName
	Mutate(*corev1.Pod, map[string]base.RuntimeInfoInterface) (shouldStop bool, err error)
	// GetName returns the name of plugin
	GetName() string
}

// Plugins record the active plugins
// including two kinds: plugins for pod with no dataset mounted and with dataset mounted
type plugins struct {
	podWithoutDatasetHandler []MutatingHandler
	podWithDatasetHandler    []MutatingHandler
	serverlessPodHandler     []MutatingHandler
}

func (p *plugins) GetPodWithoutDatasetHandler() []MutatingHandler {
	return p.podWithoutDatasetHandler
}

func (p *plugins) GetPodWithDatasetHandler() []MutatingHandler {
	return p.podWithDatasetHandler
}

func (p *plugins) GetServerlessPodHandler() []MutatingHandler {
	return p.serverlessPodHandler
}

// Registry return active plugins in a defined order
func Registry(client client.Client) plugins {
	return plugins{
		podWithoutDatasetHandler: []MutatingHandler{
			prefernodeswithoutcache.NewPlugin(client),
		},
		podWithDatasetHandler: []MutatingHandler{
			requirenodewithfuse.NewPlugin(client),
			nodeaffinitywithcache.NewPlugin(client),
			mountpropagationinjector.NewPlugin(client),
		}, serverlessPodHandler: []MutatingHandler{
			fusesidecar.NewPlugin(client),
		},
	}
}
