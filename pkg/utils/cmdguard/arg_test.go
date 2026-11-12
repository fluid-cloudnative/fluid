/*
Copyright 2024 The Fluid Authors.

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

package cmdguard

import "testing"

func TestValidateArg(t *testing.T) {
	var tests = []struct {
		name    string
		arg     string
		wantErr bool
	}{
		// Ordinary paths must keep working.
		{"plain path", "/var/lib/fluid/backup", false},
		{"dash, underscore and dot", "/pvc/sub-dir_1/a.b", false},
		{"scheme prefix", "pvc://pvc1/erf", false},
		{"space", "/tmp/a b", false},
		{"empty", "", false},

		// Covered by illegalChars, shared with checkCommandArgs.
		{"semicolon", "/tmp/a;b", true},
		{"ampersand", "/tmp/a&b", true},
		{"pipe", "/tmp/a|b", true},
		{"dollar", "/tmp/a$b", true},
		{"backquote", "/tmp/a`b", true},
		{"single quote", "/tmp/a'b", true},
		{"parentheses", "/tmp/a(b)c", true},
		{"output redirection", "/tmp/a>b", true},

		// Covered by extraArgIllegalChars, which checkCommandArgs does not reject.
		{"input redirection", "/tmp/a<b", true},
		{"double quote", "/tmp/a\"b", true},
		{"backslash", "/tmp/a\\b", true},
		{"history expansion", "/tmp/a!b", true},
		{"glob star", "/tmp/a*b", true},
		{"glob question mark", "/tmp/a?b", true},
		{"brace expansion", "/tmp/{a,b}", true},
		{"bracket expansion", "/tmp/a[0-9]b", true},
		{"home expansion", "~/backup", true},
		{"comment", "/tmp/a#b", true},

		// Control characters, which neither character list covers.
		{"newline", "/tmp/a\nb", true},
		{"carriage return", "/tmp/a\rb", true},
		{"tab", "/tmp/a\tb", true},
		{"NUL", "/tmp/a\x00b", true},
		{"DEL", "/tmp/a\x7fb", true},
	}

	for _, test := range tests {
		err := ValidateArg(test.arg)
		if test.wantErr && err == nil {
			t.Errorf("%s: %q should be rejected, but err is nil", test.name, test.arg)
		}
		if !test.wantErr && err != nil {
			t.Errorf("%s: %q should be accepted, but got err: %v", test.name, test.arg, err)
		}
	}
}
