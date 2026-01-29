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

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestSecurity(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Security Suite")
}

var _ = Describe("EscapeBashStr", func() {
	Context("when string contains no special characters", func() {
		It("should return the original string for simple alphanumeric", func() {
			result := EscapeBashStr("abc")
			Expect(result).To(Equal("abc"))
		})

		It("should return the original string with hyphens", func() {
			result := EscapeBashStr("test-volume")
			Expect(result).To(Equal("test-volume"))
		})

		It("should return the original string for URLs without special chars", func() {
			result := EscapeBashStr("http://minio.kube-system:9000/minio/dynamic-ce")
			Expect(result).To(Equal("http://minio.kube-system:9000/minio/dynamic-ce"))
		})

		It("should return the original string with underscores and dots", func() {
			result := EscapeBashStr("test_file.txt")
			Expect(result).To(Equal("test_file.txt"))
		})

		It("should return the original string with slashes", func() {
			result := EscapeBashStr("/path/to/file")
			Expect(result).To(Equal("/path/to/file"))
		})
	})

	Context("when string contains command substitution", func() {
		It("should escape dollar sign with parentheses", func() {
			result := EscapeBashStr("$(cat /proc/self/status | grep CapEff > /test.txt)")
			Expect(result).To(Equal("$'$(cat /proc/self/status | grep CapEff > /test.txt)'"))
		})

		It("should escape backticks", func() {
			result := EscapeBashStr("hel`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'hel`cat /proc/self/status`lo'"))
		})

		It("should escape multiple command substitution attempts", func() {
			result := EscapeBashStr("`whoami` and `date`")
			Expect(result).To(Equal("$'`whoami` and `date`'"))
		})
	})

	Context("when string contains quotes", func() {
		It("should escape single quotes with backticks", func() {
			result := EscapeBashStr("'h'el`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'\\'h\\'el`cat /proc/self/status`lo'"))
		})

		It("should handle already escaped single quotes", func() {
			result := EscapeBashStr("\\'h\\'el`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'\\\\\\'h\\\\\\'el`cat /proc/self/status`lo'"))
		})

		It("should handle ANSI-C quoted strings with single quotes", func() {
			result := EscapeBashStr("$'h'el`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'$\\'h\\'el`cat /proc/self/status`lo'"))
		})
	})

	Context("when string contains backslashes", func() {
		It("should escape backslash before backtick", func() {
			result := EscapeBashStr("hel\\`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'hel\\\\`cat /proc/self/status`lo'"))
		})

		It("should handle double backslash before backtick", func() {
			result := EscapeBashStr("hel\\\\`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'hel\\\\\\\\`cat /proc/self/status`lo'"))
		})

		It("should handle backslash with single quote and backtick", func() {
			result := EscapeBashStr("hel\\'`cat /proc/self/status`lo")
			Expect(result).To(Equal("$'hel\\\\\\'`cat /proc/self/status`lo'"))
		})

		It("should handle multiple backslashes", func() {
			result := EscapeBashStr("test\\\\\\\\value")
			Expect(result).To(Equal("$'test\\\\\\\\\\\\\\\\value'"))
		})
	})

	Context("when string contains shell operators", func() {
		It("should escape ampersand", func() {
			result := EscapeBashStr("command1 & command2")
			Expect(result).To(Equal("$'command1 & command2'"))
		})

		It("should escape semicolon", func() {
			result := EscapeBashStr("command1; command2")
			Expect(result).To(Equal("$'command1; command2'"))
		})

		It("should escape pipe", func() {
			result := EscapeBashStr("command1 | command2")
			Expect(result).To(Equal("$'command1 | command2'"))
		})

		It("should escape greater than", func() {
			result := EscapeBashStr("echo test > file.txt")
			Expect(result).To(Equal("$'echo test > file.txt'"))
		})

		It("should escape parentheses", func() {
			result := EscapeBashStr("(command1)")
			Expect(result).To(Equal("$'(command1)'"))
		})
	})

	Context("when string contains multiple special characters", func() {
		It("should handle complex injection attempts", func() {
			result := EscapeBashStr("'; rm -rf /; echo '")
			Expect(result).To(Equal("$'\\'; rm -rf /; echo \\''"))
		})

		It("should handle dollar sign with ampersand", func() {
			result := EscapeBashStr("$VAR && malicious")
			Expect(result).To(Equal("$'$VAR && malicious'"))
		})

		It("should handle nested quotes and operators", func() {
			result := EscapeBashStr("'test' | grep 'pattern' > output.txt")
			Expect(result).To(Equal("$'\\'test\\' | grep \\'pattern\\' > output.txt'"))
		})
	})

	Context("edge cases", func() {
		It("should handle empty string", func() {
			result := EscapeBashStr("")
			Expect(result).To(Equal(""))
		})

		It("should handle string with only spaces", func() {
			result := EscapeBashStr("   ")
			Expect(result).To(Equal("$'   '"))
		})

		It("should handle string with newlines (security critical)", func() {
			result := EscapeBashStr("line1\nline2")
			Expect(result).To(Equal("$'line1\nline2'"))
		})

		It("should handle string starting with special character", func() {
			result := EscapeBashStr("$test")
			Expect(result).To(Equal("$'$test'"))
		})

		It("should handle string ending with special character", func() {
			result := EscapeBashStr("test$")
			Expect(result).To(Equal("$'test$'"))
		})

		It("should handle newline character injection attempt", func() {
			result := EscapeBashStr("line1\nrm -rf /")
			Expect(result).To(Equal("$'line1\nrm -rf /'"))
		})

		It("should handle carriage return injection", func() {
			result := EscapeBashStr("test\rmalicious")
			Expect(result).To(Equal("$'test\rmalicious'"))
		})

		It("should handle tab character", func() {
			result := EscapeBashStr("test\tvalue")
			Expect(result).To(Equal("$'test\tvalue'"))
		})
	})
})

var _ = Describe("containsOne", func() {
	Context("when checking for character existence", func() {
		It("should return true when target contains one of the characters", func() {
			backtick := '`'
			result := containsOne("hello$world", []rune{'$', backtick, '&'})
			Expect(result).To(BeTrue())
		})

		It("should return true when target contains multiple matching characters", func() {
			result := containsOne("$test&value", []rune{'$', '&', ';'})
			Expect(result).To(BeTrue())
		})

		It("should return false when target contains none of the characters", func() {
			backtick := '`'
			result := containsOne("hello-world", []rune{'$', backtick, '&'})
			Expect(result).To(BeFalse())
		})

		It("should return false for empty string", func() {
			backtick := '`'
			result := containsOne("", []rune{'$', backtick, '&'})
			Expect(result).To(BeFalse())
		})

		It("should return false when checking empty character list", func() {
			result := containsOne("hello$world", []rune{})
			Expect(result).To(BeFalse())
		})

		It("should handle single character target", func() {
			backtick := '`'
			result := containsOne("$", []rune{'$', backtick})
			Expect(result).To(BeTrue())
		})

		It("should handle unicode characters", func() {
			result := containsOne("hello世界", []rune{'世', '$'})
			Expect(result).To(BeTrue())
		})

		It("should be case sensitive", func() {
			result := containsOne("Hello", []rune{'h'})
			Expect(result).To(BeFalse())
		})

		It("should detect backslash character", func() {
			result := containsOne("test\\value", []rune{'\\', '$'})
			Expect(result).To(BeTrue())
		})

		It("should detect single quote character", func() {
			result := containsOne("test'value", []rune{'\'', '$'})
			Expect(result).To(BeTrue())
		})

		It("should detect newline character", func() {
			result := containsOne("test\nvalue", []rune{'\n', '$'})
			Expect(result).To(BeTrue())
		})
	})
})
