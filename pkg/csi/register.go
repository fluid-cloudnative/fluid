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

package csi

import (
	"github.com/golang/glog"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/csi/plugins"
	"github.com/fluid-cloudnative/fluid/pkg/csi/recover"
	"github.com/fluid-cloudnative/fluid/pkg/csi/updatedbconf"
)

type registrationFuncs struct {
	enabled  func() bool
	register func(mgr manager.Manager, ctx config.RunningContext) error
}

var registraions map[string]registrationFuncs

func init() {
	registraions = map[string]registrationFuncs{}

	registraions["plugins"] = registrationFuncs{enabled: plugins.Enabled, register: plugins.Register}
	registraions["recover"] = registrationFuncs{enabled: recover.Enabled, register: recover.Register}
	registraions["updatedbconf"] = registrationFuncs{enabled: updatedbconf.Enabled, register: updatedbconf.Register}
}

// SetupWithManager registers all the enabled components defined in registrations to the controller manager.
func SetupWithManager(mgr manager.Manager, ctx config.RunningContext) error {
	for rName, r := range registraions {
		if r.enabled() {
			glog.Infof("Registering %s to controller manager", rName)
			if err := r.register(mgr, ctx); err != nil {
				glog.Errorf("Got error when registering %s, error: %v", rName, err)
				return err
			}
		} else {
			glog.Infof("%s is not enabled", rName)
		}
	}

	return nil
}
