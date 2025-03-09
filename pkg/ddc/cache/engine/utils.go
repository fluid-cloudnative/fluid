package engine

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

func (e *CacheEngine) getRuntimeConfigDir() string {
	return fmt.Sprintf("/etc/fluid/config")
}

func (e *CacheEngine) getRuntimeConfigPath() string {
	return fmt.Sprintf("%s/config.json", e.getRuntimeConfigDir())
}

func (e *CacheEngine) getRuntimeEncryptOptionPath(secretName string) string {
	return fmt.Sprintf("/etc/fluid/secrets/%s", secretName)
}

func (e *CacheEngine) getRuntimeEncryptVolumeName(secretName string) string {
	return fmt.Sprintf("runtime-mount-secret-%s", secretName)
}

func (e *CacheEngine) getRuntimeConfigCmName() (targetPath string) {
	return fmt.Sprintf("fluid-cache-runtime-config-%s", e.name)
}

func (e *CacheEngine) getRuntimeConfigVolumeName() (targetPath string) {
	return fmt.Sprintf("fluid-cache-runtime-%s-config", e.name)
}

func (e *CacheEngine) getServiceName(runtimeName, componentName string) (targetPath string) {
	return fmt.Sprintf("cacheruntime-%s-%s", runtimeName, componentName)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.CacheRuntime
	} else {
		path = path + "/" + common.CacheRuntime
	}
	return
}

func (e *CacheEngine) getTargetPath() (targetPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/cache-fuse", mountRoot, e.namespace, e.name)
}

func (e *CacheEngine) getRuntimeTargetPathVolumeName() string {
	return fmt.Sprintf("fluid-cache-runtime-shared-path")
}

func (e *CacheEngine) getComponentName(componentType common.ComponentType) string {
	return fmt.Sprintf("%s-%s", e.name, componentType)
}
