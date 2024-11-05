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
func Register(_ manager.Manager, ctx config.RunningContext) error {
	content, err := os.ReadFile(updatedbConfPath)
	if os.IsNotExist(err) {
		glog.Info("/etc/updatedb.conf not exist, skip updating")
		return nil
	}
	if err != nil {
		return err
	}
	newConfig, err := updateConfig(string(content), ctx.PruneFs, []string{ctx.PrunePath})
	if err != nil {
		glog.Warningf("failed to update updatedb.conf %s ", err)
		return nil
	}
	if newConfig == string(content) {
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
		newConfig = fmt.Sprintf("%s\n%s", modifiedByFluidComment, newConfig)
		glog.Info("backup complete, now update /etc/updatedb.conf")
	}
	return os.WriteFile(updatedbConfPath, []byte(newConfig), 0644)
}

// Enabled checks if the updatedb config modifier should be enabled
func Enabled() bool {
	return true
}
