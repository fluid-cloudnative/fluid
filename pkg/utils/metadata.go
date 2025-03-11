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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	MetaDataFuseLabelFileName = "labels"
)

// getMountRoot returns the default path, if it's not set
func getMetadataRootPath(runtimeType string) (path string) {
	path, err := GetMountRoot()
	if err != nil {
		path = "/" + runtimeType
	} else {
		path = path + "/" + runtimeType
	}
	return
}

func GetRuntimeRootMetadataPath(namespace, name, runtimeType string) (mountPath string) {
	mountPath = fmt.Sprintf("%s/%s/%s/.meta", getMetadataRootPath(runtimeType), namespace, name)
	log.Info("DEBUG", "RootMetadata:", mountPath)
	return
}
func GetRuntimeFuseMetadataPath(namespace, name, runtimeType string) (mountPath string) {
	mountPath = fmt.Sprintf("%s/fuse", GetRuntimeRootMetadataPath(namespace, name, runtimeType))
	log.Info("DEBUG", "FuseMetadata:", mountPath)
	return
}

func GetMetadataFuseLabelFileInMetadataPath(namespace, name, runtimeType string) (mountPath string) {
	mountPath = fmt.Sprintf("%s/%s", GetRuntimeFuseMetadataPath(namespace, name, runtimeType), MetaDataFuseLabelFileName)
	log.Info("DEBUG", "FuseMetadata:", mountPath)
	return
}

func LoadCurrentFuseGenerationFromMeta(namespace, name, runtimeType string) (string, error) {
	file, err := os.Open(GetMetadataFuseLabelFileInMetadataPath(namespace, name, runtimeType))
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], "\"")

		if key == "pod-template-generation" {
			return value, nil
		}
	}
	return "", errors.New("pod-template-generation in fuse.labels")
}
