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

type PluginsProfile struct {
	Plugins      Plugins        `yaml:"plugins"`
	PluginConfig []PluginConfig `yaml:"pluginConfig"`
}

type Plugins struct {
	Serverful  Serverful  `yaml:"serverful"`
	Serverless Serverless `yaml:"serverless"`
}

type Serverful struct {
	WithDataset    []string `yaml:"withDataset"`
	WithoutDataset []string `yaml:"withoutDataset"`
}

type Serverless struct {
	WithDataset    []string `yaml:"withDataset"`
	WithoutDataset []string `yaml:"withoutDataset"`
}

type PluginConfig struct {
	// Name defines the name of plugin being configured
	Name string `yaml:"name"`
	// Args defines the arguments passed to the plugins at the time of initialization. Args can have arbitrary structure.
	Args interface{} `yaml:"args,omitempty"`
}
