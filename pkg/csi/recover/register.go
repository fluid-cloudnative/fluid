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
	"os"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/csi/features"
	utilfeature "github.com/fluid-cloudnative/fluid/pkg/utils/feature"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Register initializes the fuse recover and registers it to the controller manager.
func Register(mgr manager.Manager, ctx config.RunningContext) error {
	fuseRecover, err := NewFuseRecover(mgr.GetClient(), mgr.GetEventRecorderFor("FuseRecover"), mgr.GetAPIReader(), ctx.VolumeLocks)
	if err != nil {
		return err
	}

	if err = mgr.Add(fuseRecover); err != nil {
		return err
	}

	return nil
}

// Enabled checks if the fuse recover should be enabled.
func Enabled() bool {
	if os.Getenv("NODEPUBLISH_METHOD") == common.NodePublishMethodSymlink {
		// not support auto recovery for nodePublishMethod symlink
		return false
	}
	return utilfeature.DefaultFeatureGate.Enabled(features.FuseRecovery)
}
