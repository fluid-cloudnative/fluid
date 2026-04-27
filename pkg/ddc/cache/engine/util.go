/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// GetComponentName gets the component name using runtime name and component type.
func GetComponentName(runtimeName string, componentType common.ComponentType) string {
	return fmt.Sprintf("%s-%s", runtimeName, componentType)
}

// GetComponentServiceName gets the component service name.
func GetComponentServiceName(runtimeName string, componentType common.ComponentType) string {
	return fmt.Sprintf("svc-%s-%s", runtimeName, componentType)
}

// getFuseMountPointVolumeName get the volume name of mount path in fuse pod.
func (e *CacheEngine) getFuseMountPointVolumeName() string {
	return "fluid-cache-runtime-shared-path"
}

func (e *CacheEngine) getFuseMountPoint() string {
	mountRoot, err := utils.GetMountRoot()
	if err != nil {
		mountRoot = "/" + common.CacheRuntime
	} else {
		mountRoot = mountRoot + "/" + common.CacheRuntime
	}

	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/cache-fuse", mountRoot, e.namespace, e.name)
}

// getRuntimeConfigConfigMapName get the configmap name of the runtime config.
func (e *CacheEngine) getRuntimeConfigConfigMapName() string {
	return fmt.Sprintf("fluid-runtime-config-%s", e.name)
}
func (e *CacheEngine) getRuntimeConfigVolumeName() (targetPath string) {
	return fmt.Sprintf("fluid-runtime-%s-config", e.name)
}

// getRuntimeConfigDir defines the mount directory of runtime config in the pod.
func (e *CacheEngine) getRuntimeConfigDir() string {
	return "/etc/fluid/config"
}

// getRuntimeConfigPath defines the mount path of runtime config in the pod.
func (e *CacheEngine) getRuntimeConfigPath() string {
	return fmt.Sprintf("%s/%s", e.getRuntimeConfigDir(), e.getRuntimeConfigFileName())
}
func (e *CacheEngine) getRuntimeConfigFileName() string {
	return "runtime.json"
}

func (e *CacheEngine) getRuntimeClassExtraConfigMapVolumeName(name string) string {
	return fmt.Sprintf("fluid-extra-%s-%s", e.name, name)
}

// getSecretVolumeName generates the volume name for a secret mount
func getSecretVolumeName(secretName string) string {
	return fmt.Sprintf("cache-mount-secret-%s", secretName)
}

// getSecretMountPath generates the base mount path for a secret in the container
func getSecretMountPath(secretName string) string {
	return fmt.Sprintf("/etc/fluid/secrets/%s", secretName)
}

// getSecretFilePath generates the full file path for a secret key in the container
func getSecretFilePath(secretName, secretKey string) string {
	return fmt.Sprintf("%s/%s", getSecretMountPath(secretName), secretKey)
}
