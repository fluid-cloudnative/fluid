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
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/datasetusageinjector"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/fileprefetcher"
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
	// cache handlers to avoid deserialization cost
	cacheHandlers = &Handlers{}
)

func RegisterMutatingHandlers(client client.Client) error {
	// ignore the register error
	_ = registry.Register(prefernodeswithoutcache.Name, prefernodeswithoutcache.NewPlugin)
	_ = registry.Register(mountpropagationinjector.Name, mountpropagationinjector.NewPlugin)
	_ = registry.Register(requirenodewithfuse.Name, requirenodewithfuse.NewPlugin)
	_ = registry.Register(nodeaffinitywithcache.Name, nodeaffinitywithcache.NewPlugin)
	_ = registry.Register(fusesidecar.Name, fusesidecar.NewPlugin)
	_ = registry.Register(datasetusageinjector.Name, datasetusageinjector.NewPlugin)
	_ = registry.Register(fileprefetcher.Name, fileprefetcher.NewPlugin)

	// get the handlers through the config file
	data, err := os.ReadFile(common.WebhookPluginFilePath)
	if err != nil {
		return err
	}

	profile := PluginsProfile{}
	err = yaml.Unmarshal(data, &profile)
	if err != nil {
		return err
	}

	cacheHandlers, err = newHandler(client, &profile)
	if err != nil {
		return err
	}

	return nil
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
	return cacheHandlers
}

func newHandler(client client.Client, profile *PluginsProfile) (handlers *Handlers, err error) {
	handlers = &Handlers{}
	pluginConfig := make(map[string]string, len(profile.PluginConfig))
	for i := range profile.PluginConfig {
		name := profile.PluginConfig[i].Name
		if _, ok := pluginConfig[name]; ok {
			log.Error(errors.New("repeated config for plugin, use the later"), "name", name)
		}
		pluginConfig[name] = profile.PluginConfig[i].Args
	}

	// new handler for serverful and serverless pod with/without dataset
	podWithDatasetHandler, err := newHandlerForType(client, profile.Plugins.Serverful.WithDataset, pluginConfig, "podWithDatasetHandler")
	if err != nil {
		return nil, err
	}
	handlers.podWithDatasetHandler = podWithDatasetHandler

	podWithoutDatasetHandler, err := newHandlerForType(client, profile.Plugins.Serverful.WithoutDataset, pluginConfig, "podWithoutDatasetHandler")
	if err != nil {
		return nil, err
	}
	handlers.podWithoutDatasetHandler = podWithoutDatasetHandler

	serverlessPodWithDatasetHandler, err := newHandlerForType(client, profile.Plugins.Serverless.WithDataset, pluginConfig, "serverlessPodWithDatasetHandler")
	if err != nil {
		return nil, err
	}
	handlers.serverlessPodWithDatasetHandler = serverlessPodWithDatasetHandler

	serverlessPodWithoutDatasetHandler, err := newHandlerForType(client, profile.Plugins.Serverless.WithoutDataset, pluginConfig, "serverlessPodWithoutDatasetHandler")
	if err != nil {
		return nil, err
	}
	handlers.serverlessPodWithoutDatasetHandler = serverlessPodWithoutDatasetHandler

	return handlers, nil
}

func newHandlerForType(client client.Client, pluginNames []string, pluginConfig map[string]string, pluginType string) ([]api.MutatingHandler, error) {
	var serverlessPodWithDatasetHandlerNames []string
	var serverlessPodWithDatasetHandler []api.MutatingHandler
	// failure as early as possible
	for _, name := range pluginNames {
		factory, ok := registry[name]
		if !ok {
			err := fmt.Errorf("unknown plugin name [%s]", name)
			log.Error(err, "plugin not exist", "pluginName", name)
			return nil, err
		}
		handler, err := factory(client, pluginConfig[name])
		if err != nil {
			log.Error(err, "new plugin occurs error", "pluginName", name)
			return nil, err
		}
		serverlessPodWithDatasetHandlerNames = append(serverlessPodWithDatasetHandlerNames, name)
		serverlessPodWithDatasetHandler = append(serverlessPodWithDatasetHandler, handler)
	}
	log.Info("register plugins", "plugin type", pluginType, "plugin names", serverlessPodWithDatasetHandlerNames)
	return serverlessPodWithDatasetHandler, nil
}
