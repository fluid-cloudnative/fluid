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
	result := line
	for s := range sensitiveKeys {
		// if the log line contains a secret key, redact its value
		if strings.Contains(result, s) {
			// Look for pattern "key=" and replace everything after = until space or end
			searchPattern := s + "="
			idx := strings.Index(result, searchPattern)

			for idx != -1 {
				// Find the end of the value (next space or end of string)
				startValue := idx + len(searchPattern)
				endValue := startValue

				// Find where the value ends (space, newline, or end of string)
				for endValue < len(result) && result[endValue] != ' ' && result[endValue] != '\n' && result[endValue] != '\t' {
					endValue++
				}

				// Replace the value with [ redacted ]
				result = result[:startValue] + "[ redacted ]" + result[endValue:]

				// Look for next occurrence
				idx = strings.Index(result[startValue+len("[ redacted ]"):], searchPattern)
				if idx != -1 {
					idx = idx + startValue + len("[ redacted ]")
				}
			}
		}
	}

	return result
}

func UpdateSensitiveKey(key string) {
	if _, found := sensitiveKeys[key]; !found {
		sensitiveKeys[key] = true
	}
}
