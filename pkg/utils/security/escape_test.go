/*
Copyright 2023 The Fluid Author.

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

import "testing"

func TestEscapeBashStr(t *testing.T) {
	cases := [][]string{
		{"abc", "abc"},
		{"test-volume", "test-volume"},
		{"http://minio.kube-system:9000/minio/dynamic-ce", "http://minio.kube-system:9000/minio/dynamic-ce"},
		{"$(cat /proc/self/status | grep CapEff > /test.txt)", "$'$(cat /proc/self/status | grep CapEff > /test.txt)'"},
		{"hel`cat /proc/self/status`lo", "$'hel`cat /proc/self/status`lo'"},
		{"'h'el`cat /proc/self/status`lo", "$'\\'h\\'el`cat /proc/self/status`lo'"},
		{"\\'h\\'el`cat /proc/self/status`lo", "$'\\'h\\'el`cat /proc/self/status`lo'"},
		{"$'h'el`cat /proc/self/status`lo", "$'$\\'h\\'el`cat /proc/self/status`lo'"},
		{"hel\\`cat /proc/self/status`lo", "$'hel\\\\`cat /proc/self/status`lo'"},
		{"hel\\\\`cat /proc/self/status`lo", "$'hel\\\\`cat /proc/self/status`lo'"},
		{"hel\\'`cat /proc/self/status`lo", "$'hel\\'`cat /proc/self/status`lo'"},
	}
	for _, c := range cases {
		escaped := EscapeBashStr(c[0])
		if escaped != c[1] {
			t.Errorf("escapeBashVar(%s) = %s, want %s", c[0], escaped, c[1])
		}
	}
}
