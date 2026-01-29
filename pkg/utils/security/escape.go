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

import (
	"fmt"
	"strings"
)

// According to https://www.gnu.org/software/bash/manual/html_node/ANSI_002dC-Quoting.html#ANSI_002dC-Quoting
// a -> a
// a b -> a b
// $a -> $'$a'
// $'a' -> $'$\'a\”
func EscapeBashStr(s string) string {
	if !containsOne(s, []rune{'$', '`', '&', ';', '>', '|', '(', ')'}) {
		return s
	}

	// Escape backslashes first
	s = strings.ReplaceAll(s, `\`, `\\`)

	// Then escape single quotes
	s = strings.ReplaceAll(s, `'`, `\'`)

	return fmt.Sprintf(`$'%s'`, s)
}

func containsOne(target string, chars []rune) bool {
	charMap := make(map[rune]bool, len(chars))
	for _, c := range chars {
		charMap[c] = true
	}
	for _, s := range target {
		if charMap[s] {
			return true
		}
	}
	return false
}
