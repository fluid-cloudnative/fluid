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

package security

import "strings"

var sensitiveKeys map[string]bool = map[string]bool{
	"aws.secretKey":          true,
	"aws.accessKeyId":        true,
	"fs.oss.accessKeyId":     true,
	"fs.oss.accessKeySecret": true,
}

func FilterCommand(command []string) (filteredCommand []string) {
	for _, str := range command {
		filteredCommand = append(filteredCommand, FilterString(str))
	}

	return
}

func FilterString(line string) string {
	for s := range sensitiveKeys {
		// if the log line contains a secret value redact it
		if strings.Contains(line, s) {
			line = s + "=[ redacted ]"
		}
	}

	return line
}

func UpdateSensitiveKey(key string) {
	if _, found := sensitiveKeys[key]; !found {
		sensitiveKeys[key] = true
	}
}
