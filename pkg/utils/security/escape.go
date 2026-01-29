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
// $'a' -> $'$\'a\''
func EscapeBashStr(s string) string {
	// Check if string contains any shell-sensitive characters that require escaping
	// Added '\', '\'', '\n', '\r', and '\t' to the list as identified by security review
	if !containsOne(s, []rune{'$', '`', '&', ';', '>', '|', '(', ')', '\'', '\\', '\n', '\r', '\t', ' '}) {
		return s
	}

	// Build the escaped string manually to handle all special cases correctly
	var result strings.Builder

	for _, ch := range s {
		switch ch {
		case '\\':
			// Escape backslashes by doubling them
			result.WriteString(`\\`)
		case '\'':
			// Escape single quotes
			result.WriteString(`\'`)
		case '\n':
			// Preserve newline as literal \n in ANSI-C quoted string
			result.WriteString(`\n`)
		case '\r':
			// Preserve carriage return as literal \r in ANSI-C quoted string
			result.WriteString(`\r`)
		case '\t':
			// Preserve tab as literal \t in ANSI-C quoted string
			result.WriteString(`\t`)
		default:
			// All other characters (including $, `, &, etc.) are safe within $'...'
			result.WriteRune(ch)
		}
	}

	// Wrap in ANSI-C quoting format
	return fmt.Sprintf(`$'%s'`, result.String())
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
