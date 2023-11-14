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

package testutil

import (
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// SortEnvVarByName sorts the value of an environment variable by its name
func SortEnvVarByName(envs []corev1.EnvVar, name string) []corev1.EnvVar {
	// Search for the environment variable with the given name
	for i := range envs {
		if envs[i].Name == name {
			// Sort the value of the environment variable
			envs[i].Value = sortEnvVarValue(envs[i].Value)
			return envs
		}
	}
	return envs
}

// sortEnvVarValue sorts the value of an environment variable
func sortEnvVarValue(value string) string {
	// Split the value into key=value pairs
	pairs := strings.Split(value, ",")
	kvMap := make(map[string]string, len(pairs))

	// Add the key=value pairs to a map
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			kvMap[kv[0]] = kv[1]
		}
	}

	// Sort the keys and build a new value
	var sortedPairs []string
	for key, value := range kvMap {
		sortedPairs = append(sortedPairs, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(sortedPairs)
	return strings.Join(sortedPairs, ",")
}
