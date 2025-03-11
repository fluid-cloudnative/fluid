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

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/security"
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

func GetRuntimeRootMetadataPath(namespace, name, runtimeType string) string {
	return security.EscapeBashStr(fmt.Sprintf("%s/%s/%s/.meta", getMetadataRootPath(runtimeType), namespace, name))
}
func GetRuntimeFuseMetadataPath(namespace, name, runtimeType string) string {
	return security.EscapeBashStr(fmt.Sprintf("%s/fuse", GetRuntimeRootMetadataPath(namespace, name, runtimeType)))
}

func GetMetadataFuseLabelFileInMetadataPath(namespace, name, runtimeType string) string {
	return security.EscapeBashStr(fmt.Sprintf("%s/%s", GetRuntimeFuseMetadataPath(namespace, name, runtimeType), MetaDataFuseLabelFileName))
}

func LoadCurrentFuseGenerationFromMeta(namespace, name, runtimeType string) (string, error) {
	filePath := GetMetadataFuseLabelFileInMetadataPath(namespace, name, runtimeType)
	file, err := os.Open(filePath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open metadata file %s", filePath)
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

		if key == common.LabelRuntimeFuseGeneration {
			return value, nil
		}
	}
	return "", errors.New("failed to parse pod-template-generation in fuse.labels")
}
