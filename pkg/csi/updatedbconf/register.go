package updatedbconf

import (
	"os"

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
	return os.WriteFile(updatedbConfPath, []byte(newconfig), 0644)
}

// Enabled checks if the updatedb config modifier should be enabled
func Enabled() bool {
	return true
}
