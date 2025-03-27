/*
Copyright 2025 The Fluid Authors.

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

package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/validation"
)

const (
	MetaDataFuseImageVersionFileName = "fuse.ImageVersion"
)

// getMetadataRootPathFromEnv gets the value of the env variable named MOUNT_ROOT
func getMetadataRootPathFromEnv() (string, error) {
	metaRoot := os.Getenv(MountRoot)

	if err := validation.IsValidMountRoot(metaRoot); err != nil {
		return metaRoot, err
	}
	return metaRoot, nil
}

// getMountRoot returns the default path, if it's not set
func getMetadataRootPath(runtimeType string) (path string) {
	path, err := getMetadataRootPathFromEnv()
	if err != nil {
		path = "/" + runtimeType
	} else {
		path = path + "/" + runtimeType
	}
	return
}

func GetRuntimeMetadataPath(namespace, name, runtimeType string) (mountPath string) {
	mountRoot := getMetadataRootPath(runtimeType)
	return fmt.Sprintf("%s/%s/%s/.meta", mountRoot, namespace, name)
}

func GetMetadataFuseImageVersion(namespace, name, runtimeType string) (mountPath string) {
	return fmt.Sprintf("%s/%s", GetRuntimeMetadataPath(namespace, name, runtimeType), MetaDataFuseImageVersionFileName)
}

func GetGenerateFuseImageVersionMetadataPostStartCmd(namespace, name, runtimeType string) string {
	return fmt.Sprintf("if [ ! -z $%s ]; then echo $%s > %s; fi",
		common.EnvRuntimeFuseImageVersion,
		common.EnvRuntimeFuseImageVersion,
		GetMetadataFuseImageVersion(namespace, name, runtimeType),
	)
}

func GetFuseImageVersionFromMetadata(namespace, name, runtimeType string) (string, bool, error) {
	fuseImageVersionFile := GetMetadataFuseImageVersion(namespace, name, runtimeType)
	if _, err := os.Stat(fuseImageVersionFile); err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}

	content, err := os.ReadFile(fuseImageVersionFile)
	if err != nil {
		return "", false, err
	}

	return strings.ReplaceAll(string(content), "\n", ""), true, nil
}
