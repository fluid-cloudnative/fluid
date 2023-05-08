/*
Copyright 2022 The Fluid Authors.

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
	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubelet"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Register initializes the csi driver and registers it to the controller manager.
func Register(mgr manager.Manager, cfg config.Config) error {
	client, err := kubelet.InitNodeAuthorizedClient(cfg.KubeletConfigPath)
	if err != nil {
		return err
	}

	csiDriver := NewDriver(cfg.NodeId, cfg.Endpoint, mgr.GetClient(), mgr.GetAPIReader(), client)

	if err := mgr.Add(csiDriver); err != nil {
		return err
	}

	return nil
}

// Enabled checks if the csi driver should be enabled.
func Enabled() bool {
	return true
}
