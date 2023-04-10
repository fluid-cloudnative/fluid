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
package updatedbconf

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
)

// Register update the host /etc/updatedb.conf
func Register(_ manager.Manager, cfg config.Config) error {
	content, err := os.ReadFile(updatedbConfPath)
	if os.IsNotExist(err) {
		glog.Info("/etc/updatedb.conf not exist, skip updating")
		return nil
	}
	if err != nil {
		return err
	}
	newconfig, err := updateConfig(string(content), cfg.PruneFs, []string{cfg.PrunePath})
	if err != nil {
		glog.Warningf("failed to update updatedb.conf %s ", err)
		return nil
	}
	if newconfig == string(content) {
		glog.Info("/etc/updatedb.conf has no changes, skip updating")
		return nil
	}
	// if the old file does not have the `modifiedByFluidComment` comments
	// we consider this an original config file that has never been modified
	// by Fluid before and should do a backup for that.
	if !strings.HasPrefix(string(content), modifiedByFluidComment) {
		glog.Info("backup old /etc/updatedb.conf to /etc/updatedb.conf.backup")
		err = os.WriteFile(updatedbConfBackupPath, content, 0644)
		if err != nil {
			return err
		}
		newconfig = fmt.Sprintf("%s\n%s", modifiedByFluidComment, newconfig)
		glog.Info("backup complete, now update /etc/updatedb.conf")
	}
	return os.WriteFile(updatedbConfPath, []byte(newconfig), 0644)
}

// Enabled checks if the updatedb config modifier should be enabled
func Enabled() bool {
	return true
}
