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

package recover

import (
	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Register initializes the fuse recover and registers it to the controller manager.
func Register(mgr manager.Manager, config config.Config) error {
	fuseRecover, err := NewFuseRecover(mgr.GetClient(), mgr.GetEventRecorderFor("FuseRecover"), config.RecoverFusePeriod)
	if err != nil {
		return err
	}

	if err = mgr.Add(fuseRecover); err != nil {
		return err
	}

	return nil
}

// Enabled checks if the fuse recover should be enabled.
func Enabled(cfg config.Config) bool {
	return cfg.RecoverFusePeriod > 0
}
