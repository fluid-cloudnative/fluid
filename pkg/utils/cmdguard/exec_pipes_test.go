package cmdguard

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

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

var _ = Describe("ValidateShellPipeString", func() {
	Context("when validating pipe commands", func() {
		It("should reject command with grep", func() {
			err := validateShellPipeString("echo hello world | grep hello")
			Expect(err).To(HaveOccurred())
		})

		It("should accept command with wc -l", func() {
			err := validateShellPipeString("ls file | wc -l")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject command with unknown piped command", func() {
			err := validateShellPipeString("echo hello world | xyz")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with kubectl", func() {
			err := validateShellPipeString("kubectl hello world | xyz")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with & sequence", func() {
			err := validateShellPipeString("echo hello world & echo y")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with ; sequence", func() {
			err := validateShellPipeString("ls ; echo y")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with $ variable", func() {
			err := validateShellPipeString("ls $HOME")
			Expect(err).To(HaveOccurred())
		})

		It("should accept command with absolute path", func() {
			err := validateShellPipeString("ls /etc")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject command with backticks", func() {
			err := validateShellPipeString("ls `whoami`")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with parentheses", func() {
			err := validateShellPipeString("ls (echo test)")
			Expect(err).To(HaveOccurred())
		})

		It("should accept allowed expression ${METAURL}", func() {
			err := validateShellPipeString("ls ${METAURL}")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept multiple valid piped commands", func() {
			err := validateShellPipeString("ls /var | grep log | wc -l")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("ShellCommand", func() {
	Context("when creating shell commands", func() {
		It("should create valid simple command", func() {
			cmd, err := ShellCommand("bash", "-c", "ls")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Path).To(Equal(exec.Command("bash").Path))
			Expect(cmd.Args).To(Equal([]string{"bash", "-c", "ls"}))
		})

		It("should reject insufficient arguments", func() {
			cmd, err := ShellCommand("bash", "-c")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject unknown shell command", func() {
			cmd, err := ShellCommand("zsh", "-c", "ls")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should create valid piped command", func() {
			cmd, err := ShellCommand("bash", "-c", "ls | grep something")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject invalid piped command", func() {
			cmd, err := ShellCommand("bash", "-c", "ls | random-command")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should accept sh shell", func() {
			cmd, err := ShellCommand("sh", "-c", "ls")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		It("should reject command with dangerous sequences", func() {
			cmd, err := ShellCommand("bash", "-c", "ls; rm -rf /")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should accept allowed commands like alluxio", func() {
			cmd, err := ShellCommand("bash", "-c", "alluxio fs ls /")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})
	})
})

var _ = Describe("IsValidCommand", func() {
	Context("when checking command validity", func() {
		allowedBash := map[string]CommandValidater{"bash": ExactMatch}
		allowedMultiple := map[string]CommandValidater{"bash": ExactMatch, "sh": ExactMatch}

		It("should accept valid bash command", func() {
			result := isValidCommand("bash", allowedBash)
			Expect(result).To(BeTrue())
		})

		It("should accept valid sh command from multiple allowed", func() {
			result := isValidCommand("sh", allowedMultiple)
			Expect(result).To(BeTrue())
		})

		It("should reject invalid zsh command", func() {
			result := isValidCommand("zsh", allowedBash)
			Expect(result).To(BeFalse())
		})

		It("should work with PrefixMatch", func() {
			allowedPrefix := map[string]CommandValidater{"ls": PrefixMatch}
			result := isValidCommand("ls -la", allowedPrefix)
			Expect(result).To(BeTrue())
		})

		It("should reject non-matching prefix", func() {
			allowedPrefix := map[string]CommandValidater{"ls": PrefixMatch}
			result := isValidCommand("cat file", allowedPrefix)
			Expect(result).To(BeFalse())
		})
	})
})

var _ = Describe("splitShellCommand", func() {
	Context("when splitting shell commands", func() {
		It("should split valid shell command", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{" bash ", "  -c", "echo foobar | grep foo"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("bash -c"))
			Expect(pipedCmd).To(Equal("echo foobar | grep foo"))
		})

		It("should reject empty shell command", func() {
			_, _, err := splitShellCommand([]string{})
			Expect(err).To(HaveOccurred())
		})

		It("should reject invalid command without shell", func() {
			_, _, err := splitShellCommand([]string{"echo foobar | grep foo"})
			Expect(err).To(HaveOccurred())
		})

		It("should split valid command without shell", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{"test", "hello", "--help"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("test hello"))
			Expect(pipedCmd).To(Equal("--help"))
		})

		It("should handle extra whitespace", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{"  sh  ", " -c  ", "ls"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("sh -c"))
			Expect(pipedCmd).To(Equal("ls"))
		})

		It("should reject too many arguments", func() {
			_, _, err := splitShellCommand([]string{"bash", "-c", "ls", "extra"})
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("validateShellCommand", func() {
	Context("when validating shell commands", func() {
		It("should accept bash shell", func() {
			err := validateShellCommand("bash -c")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept sh shell", func() {
			err := validateShellCommand("sh -c")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject zsh shell", func() {
			err := validateShellCommand("zsh -c")
			Expect(err).To(HaveOccurred())
		})

		It("should reject bash with wrong flag", func() {
			err := validateShellCommand("bash -s")
			Expect(err).To(HaveOccurred())
		})

		It("should reject bash without flag", func() {
			err := validateShellCommand("bash")
			Expect(err).To(HaveOccurred())
		})

		It("should reject unknown shell", func() {
			err := validateShellCommand("fish -c")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("checkIllegalSequence", func() {
	Context("when checking for illegal sequences", func() {
		It("should reject script with semicolon", func() {
			err := checkIllegalSequence("ls; rm -rf /")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with ampersand", func() {
			err := checkIllegalSequence("ls & whoami")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with dollar sign", func() {
			err := checkIllegalSequence("echo $USER")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with backticks", func() {
			err := checkIllegalSequence("ls `pwd`")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with single quotes", func() {
			err := checkIllegalSequence("echo 'test'")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with double pipe", func() {
			err := checkIllegalSequence("ls || echo fail")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with redirect append", func() {
			err := checkIllegalSequence("echo test >> file")
			Expect(err).To(HaveOccurred())
		})

		It("should accept allowed expression ${METAURL}", func() {
			err := checkIllegalSequence("ls ${METAURL}")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept safe command", func() {
			err := checkIllegalSequence("ls /var/log")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject parentheses", func() {
			err := checkIllegalSequence("ls (test)")
			Expect(err).To(HaveOccurred())
		})
	})
})
