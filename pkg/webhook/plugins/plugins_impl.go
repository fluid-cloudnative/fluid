/*
Copyright 2023 The Fluid Authors.

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
	"errors"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/fusesidecar"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/mountpropagationinjector"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/nodeaffinitywithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/requirenodewithfuse"
	"gopkg.in/yaml.v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	log = ctrl.Log.WithName("plugins")
	// registry stores all plugins, mapping by name.
	registry = api.Registry{}
	cclient  client.Client

	handlers = &Handlers{}

	fluidNameSpace = common.NamespaceFluidSystem
)

func init() {
	nameSpace, err := utils.GetEnvByKey(common.MyPodNamespace)
	if err != nil || nameSpace == "" {
		log.Info(fmt.Sprintf("can not get non-empty fluid system namespace from env, use %s", common.NamespaceFluidSystem))
	} else {
		fluidNameSpace = nameSpace
	}
}

func RegisterMutatingHandlers(client client.Client) {
	// ignore the register error
	_ = registry.Register(prefernodeswithoutcache.Name, prefernodeswithoutcache.NewPlugin)
	_ = registry.Register(mountpropagationinjector.Name, mountpropagationinjector.NewPlugin)
	_ = registry.Register(requirenodewithfuse.Name, requirenodewithfuse.NewPlugin)
	_ = registry.Register(nodeaffinitywithcache.Name, nodeaffinitywithcache.NewPlugin)
	_ = registry.Register(fusesidecar.Name, fusesidecar.NewPlugin)

	cclient = client
}

type Handlers struct {
	podWithDatasetHandler              []api.MutatingHandler
	podWithoutDatasetHandler           []api.MutatingHandler
	serverlessPodWithDatasetHandler    []api.MutatingHandler
	serverlessPodWithoutDatasetHandler []api.MutatingHandler
}

func (h *Handlers) GetPodWithoutDatasetHandler() []api.MutatingHandler {
	return h.podWithoutDatasetHandler
}

func (h *Handlers) GetPodWithDatasetHandler() []api.MutatingHandler {
	return h.podWithDatasetHandler
}

func (h *Handlers) GetServerlessPodWithDatasetHandler() []api.MutatingHandler {
	return h.serverlessPodWithDatasetHandler
}

func (h *Handlers) GetServerlessPodWithoutDatasetHandler() []api.MutatingHandler {
	return h.serverlessPodWithoutDatasetHandler
}

func GetRegistryHandler() api.RegistryHandler {
	cm, err := kubeclient.GetConfigmapByName(cclient, common.PluginProfileConfigMapName, fluidNameSpace)
	if err != nil {
		log.Error(err, "get plugins config map error, skip mutating")
		return &Handlers{}
	}
	if cm == nil {
		log.Error(errors.New("plugins config map not exist"), "skip mutating")
		return &Handlers{}
	}
	profile := PluginsProfile{}
	err = yaml.Unmarshal([]byte(cm.Data[common.PluginProfileKeyName]), &profile)
	if err != nil {
		log.Error(err, "plugins config map format is wrong, skip mutating")
		return &Handlers{}
	}

	plgs := &Handlers{}

	pluginConfig := make(map[string]interface{}, len(profile.PluginConfig))
	for i := range profile.PluginConfig {
		name := profile.PluginConfig[i].Name
		if _, ok := pluginConfig[name]; ok {
			log.Error(errors.New("repeated config for plugin, use the later"), "name", name)
		}
		pluginConfig[name] = profile.PluginConfig[i].Args
	}

	for _, name := range profile.Plugins.Serverful.WithDataset {
		if factory, ok := registry[name]; ok {
			plgs.podWithDatasetHandler = append(plgs.podWithDatasetHandler, factory(cclient, pluginConfig[name]))
		}
	}
	for _, name := range profile.Plugins.Serverful.WithoutDataset {
		if factory, ok := registry[name]; ok {
			plgs.podWithoutDatasetHandler = append(plgs.podWithoutDatasetHandler, factory(cclient, pluginConfig[name]))
		}
	}
	for _, name := range profile.Plugins.Serverless.WithDataset {
		if factory, ok := registry[name]; ok {
			plgs.serverlessPodWithDatasetHandler = append(plgs.serverlessPodWithDatasetHandler, factory(cclient, pluginConfig[name]))
		}
	}
	for _, name := range profile.Plugins.Serverless.WithoutDataset {
		if factory, ok := registry[name]; ok {
			plgs.serverlessPodWithoutDatasetHandler = append(plgs.serverlessPodWithoutDatasetHandler, factory(cclient, pluginConfig[name]))
		}
	}

	return plgs
}
