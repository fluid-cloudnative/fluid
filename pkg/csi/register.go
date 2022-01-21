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

package csi

import (
	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/csi/plugin"
	"github.com/fluid-cloudnative/fluid/pkg/csi/recover"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var csiRegisterFuncs []func(manager manager.Manager, config config.Config) error

func init() {
	csiRegisterFuncs = append(csiRegisterFuncs, plugin.Register)
	csiRegisterFuncs = append(csiRegisterFuncs, recover.Register)
}

func SetupWithManager(m manager.Manager, config config.Config) error {
	for _, f := range csiRegisterFuncs {
		if err := f(m, config); err != nil {
			return err
		}
	}

	return nil
}
