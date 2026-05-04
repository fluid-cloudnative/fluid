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
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Precomputed max lengths for the 63-char DNS limit
const (
	secretVolumeNamePrefix   = "cache-mnt-secret-"
	secretMaxTotalLength     = validation.DNS1035LabelMaxLength
	prefixSecretVolumeLength = len(secretVolumeNamePrefix)
	hashSuffixLength         = 8
	truncatedSecretMaxLength = secretMaxTotalLength - prefixSecretVolumeLength - hashSuffixLength
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
func getSecretVolumeName(name string) string {
	fullName := fmt.Sprintf("%s%s", secretVolumeNamePrefix, name)
	// check volume name length
	if len(fullName) <= validation.DNS1035LabelMaxLength {
		return fullName
	}

	// Case 2: Long name - truncate + hash (fallback)
	// Step 1: Truncate secret to 36 chars
	truncatedName := name
	if len(truncatedName) > truncatedSecretMaxLength {
		truncatedName = truncatedName[:truncatedSecretMaxLength]
	}

	// Step 2: Generate 8-char SHA-256 hash of the ORIGINAL secret name (prevents collisions)
	hash := sha256.Sum256([]byte(name))
	shortHash := hex.EncodeToString(hash[:])[:hashSuffixLength]

	// Step 3: Combine to exact 63 chars
	volumeName := secretVolumeNamePrefix + truncatedName + shortHash

	return volumeName
}

// getSecretMountPath generates the base mount path for a secret in the container
func getSecretMountPath(secretName string) string {
	return fmt.Sprintf("/etc/fluid/secrets/%s", secretName)
}

// getSecretFilePath generates the full file path for a secret key in the container
func getSecretFilePath(secretName, secretKey string) string {
	return fmt.Sprintf("%s/%s", getSecretMountPath(secretName), secretKey)
}
