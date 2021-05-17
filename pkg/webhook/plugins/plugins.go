/*

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
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AffinityInterface interface {
	// InjectAffinity injects affinity info into pod
	// if a plugin return true, it means that no need to call other plugins
	InjectAffinity(*corev1.Pod, []base.RuntimeInfoInterface) (shouldStop bool)
	// GetName returns the name of plugin
	GetName() string
}

// Registry return a slice of active plugins in a defined order
func Registry(client client.Client) []AffinityInterface {
	return []AffinityInterface{
		prefernodeswithoutcache.NewPlugin(client),
		prefernodeswithcache.NewPlugin(client),
	}
}
