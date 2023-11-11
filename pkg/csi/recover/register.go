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
