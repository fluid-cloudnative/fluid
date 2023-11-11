/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
