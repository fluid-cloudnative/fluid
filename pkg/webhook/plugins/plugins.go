/*
Copyright 2021 The Fluid Authors.

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
