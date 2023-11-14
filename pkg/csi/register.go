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
