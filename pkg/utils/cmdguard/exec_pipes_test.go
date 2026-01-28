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
		// Test for disallowed first commands
		It("should reject command with disallowed first command (echo)", func() {
			err := validateShellPipeString("echo hello world | grep hello")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with kubectl as first command", func() {
			err := validateShellPipeString("kubectl get pods | grep Running")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with curl as first command", func() {
			err := validateShellPipeString("curl example.com | grep title")
			Expect(err).To(HaveOccurred())
		})

		// Test for allowed first commands
		It("should accept command with ls as first command", func() {
			err := validateShellPipeString("ls /var | grep log")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept command with df as first command", func() {
			err := validateShellPipeString("df -h | grep /dev")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept command with mount as first command", func() {
			err := validateShellPipeString("mount | grep nfs")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept command with alluxio as first command", func() {
			err := validateShellPipeString("alluxio fs ls / | grep data")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept command with goosefs as first command", func() {
			err := validateShellPipeString("goosefs fs ls / | grep data")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for allowed piped commands
		It("should accept command with wc -l", func() {
			err := validateShellPipeString("ls file | wc -l")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept command with grep as piped command", func() {
			err := validateShellPipeString("ls /var | grep log")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for disallowed piped commands
		It("should reject command with unknown piped command", func() {
			err := validateShellPipeString("ls /var | xyz")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with sed as piped command", func() {
			err := validateShellPipeString("ls /var | sed 's/a/b/'")
			Expect(err).To(HaveOccurred())
		})

		// Test for command sequences
		It("should reject command with & sequence", func() {
			err := validateShellPipeString("echo hello world & echo y")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with ; sequence", func() {
			err := validateShellPipeString("ls ; echo y")
			Expect(err).To(HaveOccurred())
		})

		// Test for variable expansion and injection attempts
		It("should reject command with $ variable", func() {
			err := validateShellPipeString("ls $HOME")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with backticks", func() {
			err := validateShellPipeString("ls `whoami`")
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with parentheses", func() {
			err := validateShellPipeString("ls (echo test)")
			Expect(err).To(HaveOccurred())
		})

		// Test for allowed expressions
		It("should accept allowed expression ${METAURL}", func() {
			err := validateShellPipeString("ls ${METAURL}")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for safe paths
		It("should accept command with absolute path", func() {
			err := validateShellPipeString("ls /etc")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for multiple piped commands
		It("should accept multiple valid piped commands", func() {
			err := validateShellPipeString("ls /var | grep log | wc -l")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject multiple piped commands with invalid piped command", func() {
			err := validateShellPipeString("ls /var | grep log | sed 's/a/b/'")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("ShellCommand", func() {
	Context("when creating shell commands", func() {
		// Test for basic command creation
		It("should create valid simple command", func() {
			cmd, err := ShellCommand("bash", "-c", "ls")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
			Expect(cmd.Path).To(Equal(exec.Command("bash").Path))
			Expect(cmd.Args).To(Equal([]string{"bash", "-c", "ls"}))
		})

		It("should create valid command with absolute path", func() {
			cmd, err := ShellCommand("bash", "-c", "ls /etc")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		// Test for insufficient arguments
		It("should reject insufficient arguments", func() {
			cmd, err := ShellCommand("bash", "-c")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject command without arguments", func() {
			cmd, err := ShellCommand("bash")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		// Test for shell validation
		It("should reject unknown shell command", func() {
			cmd, err := ShellCommand("zsh", "-c", "ls")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should accept bash shell", func() {
			cmd, err := ShellCommand("bash", "-c", "ls")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		It("should accept sh shell", func() {
			cmd, err := ShellCommand("sh", "-c", "ls")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		// Test for valid piped commands (FIXED)
		It("should create valid piped command with ls and grep", func() {
			cmd, err := ShellCommand("bash", "-c", "ls | grep something")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		It("should create valid piped command with df and wc", func() {
			cmd, err := ShellCommand("bash", "-c", "df -h | wc -l")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		// Test for invalid piped commands
		It("should reject invalid piped command with disallowed first command", func() {
			cmd, err := ShellCommand("bash", "-c", "echo test | grep test")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject invalid piped command with unknown piped command", func() {
			cmd, err := ShellCommand("bash", "-c", "ls | random-command")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		// Test for dangerous sequences
		It("should reject command with semicolon sequence", func() {
			cmd, err := ShellCommand("bash", "-c", "ls; rm -rf /")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject command with ampersand sequence", func() {
			cmd, err := ShellCommand("bash", "-c", "ls & whoami")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject command with dollar variable", func() {
			cmd, err := ShellCommand("bash", "-c", "ls $HOME")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		It("should reject command with backticks", func() {
			cmd, err := ShellCommand("bash", "-c", "ls `whoami`")
			Expect(err).To(HaveOccurred())
			Expect(cmd).To(BeNil())
		})

		// Test for allowed commands
		It("should accept allowed commands like alluxio", func() {
			cmd, err := ShellCommand("bash", "-c", "alluxio fs ls /")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		It("should accept allowed commands like goosefs", func() {
			cmd, err := ShellCommand("bash", "-c", "goosefs fs ls /")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		It("should accept allowed commands like ddc-helm", func() {
			cmd, err := ShellCommand("bash", "-c", "ddc-helm version")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})

		// Test for allowed expressions
		It("should accept command with allowed expression ${METAURL}", func() {
			cmd, err := ShellCommand("bash", "-c", "ls ${METAURL}")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd).NotTo(BeNil())
		})
	})
})

var _ = Describe("isValidCommand", func() {
	Context("when checking command validity", func() {
		allowedBash := map[string]CommandValidater{"bash": ExactMatch}
		allowedMultiple := map[string]CommandValidater{"bash": ExactMatch, "sh": ExactMatch}

		// Test for exact match validation
		It("should accept valid bash command with exact match", func() {
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

		It("should reject empty command", func() {
			result := isValidCommand("", allowedBash)
			Expect(result).To(BeFalse())
		})

		// Test for prefix match validation
		It("should work with PrefixMatch for matching prefix", func() {
			allowedPrefix := map[string]CommandValidater{"ls": PrefixMatch}
			result := isValidCommand("ls -la", allowedPrefix)
			Expect(result).To(BeTrue())
		})

		It("should work with PrefixMatch for exact match", func() {
			allowedPrefix := map[string]CommandValidater{"ls": PrefixMatch}
			result := isValidCommand("ls", allowedPrefix)
			Expect(result).To(BeTrue())
		})

		It("should reject non-matching prefix", func() {
			allowedPrefix := map[string]CommandValidater{"ls": PrefixMatch}
			result := isValidCommand("cat file", allowedPrefix)
			Expect(result).To(BeFalse())
		})

		It("should reject partial prefix match", func() {
			allowedPrefix := map[string]CommandValidater{"ls": PrefixMatch}
			result := isValidCommand("lsof", allowedPrefix)
			Expect(result).To(BeFalse())
		})

		// Test for multiple commands with different validators
		It("should work with multiple commands and different validators", func() {
			allowedMixed := map[string]CommandValidater{
				"bash":    ExactMatch,
				"ls":      PrefixMatch,
				"alluxio": PrefixMatch,
			}
			Expect(isValidCommand("bash", allowedMixed)).To(BeTrue())
			Expect(isValidCommand("bash -c", allowedMixed)).To(BeFalse())      // ExactMatch
			Expect(isValidCommand("ls -la", allowedMixed)).To(BeTrue())        // PrefixMatch
			Expect(isValidCommand("alluxio fs ls", allowedMixed)).To(BeTrue()) // PrefixMatch
		})
	})
})

var _ = Describe("splitShellCommand", func() {
	Context("when splitting shell commands", func() {
		// Test for valid splits
		It("should split valid shell command", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{" bash ", "  -c", "echo foobar | grep foo"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("bash -c"))
			Expect(pipedCmd).To(Equal("echo foobar | grep foo"))
		})

		It("should split valid command with sh", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{"sh", "-c", "ls /var"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("sh -c"))
			Expect(pipedCmd).To(Equal("ls /var"))
		})

		It("should handle extra whitespace", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{"  sh  ", " -c  ", "ls"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("sh -c"))
			Expect(pipedCmd).To(Equal("ls"))
		})

		It("should split command with complex pipe string", func() {
			shellCmd, pipedCmd, err := splitShellCommand([]string{"bash", "-c", "ls /var | grep log | wc -l"})
			Expect(err).NotTo(HaveOccurred())
			Expect(shellCmd).To(Equal("bash -c"))
			Expect(pipedCmd).To(Equal("ls /var | grep log | wc -l"))
		})

		// Test for invalid splits
		It("should reject empty shell command", func() {
			_, _, err := splitShellCommand([]string{})
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with only one argument", func() {
			_, _, err := splitShellCommand([]string{"bash"})
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with only two arguments", func() {
			_, _, err := splitShellCommand([]string{"bash", "-c"})
			Expect(err).To(HaveOccurred())
		})

		It("should reject too many arguments", func() {
			_, _, err := splitShellCommand([]string{"bash", "-c", "ls", "extra"})
			Expect(err).To(HaveOccurred())
		})

		It("should reject command with five arguments", func() {
			_, _, err := splitShellCommand([]string{"bash", "-c", "ls", "extra1", "extra2"})
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("validateShellCommand", func() {
	Context("when validating shell commands", func() {
		// Test for allowed shells
		It("should accept bash shell with -c flag", func() {
			err := validateShellCommand("bash -c")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept sh shell with -c flag", func() {
			err := validateShellCommand("sh -c")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for disallowed shells
		It("should reject zsh shell", func() {
			err := validateShellCommand("zsh -c")
			Expect(err).To(HaveOccurred())
		})

		It("should reject fish shell", func() {
			err := validateShellCommand("fish -c")
			Expect(err).To(HaveOccurred())
		})

		It("should reject ksh shell", func() {
			err := validateShellCommand("ksh -c")
			Expect(err).To(HaveOccurred())
		})

		// Test for wrong flags
		It("should reject bash with wrong flag (-s)", func() {
			err := validateShellCommand("bash -s")
			Expect(err).To(HaveOccurred())
		})

		It("should reject bash with wrong flag (-x)", func() {
			err := validateShellCommand("bash -x")
			Expect(err).To(HaveOccurred())
		})

		It("should reject sh with wrong flag (-e)", func() {
			err := validateShellCommand("sh -e")
			Expect(err).To(HaveOccurred())
		})

		// Test for missing flags
		It("should reject bash without flag", func() {
			err := validateShellCommand("bash")
			Expect(err).To(HaveOccurred())
		})

		It("should reject sh without flag", func() {
			err := validateShellCommand("sh")
			Expect(err).To(HaveOccurred())
		})

		// Test for empty or malformed input
		It("should reject empty string", func() {
			err := validateShellCommand("")
			Expect(err).To(HaveOccurred())
		})

		It("should reject only flag", func() {
			err := validateShellCommand("-c")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("checkIllegalSequence", func() {
	Context("when checking for illegal sequences", func() {
		// Test for command sequences
		It("should reject script with semicolon", func() {
			err := checkIllegalSequence("ls; rm -rf /")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with ampersand", func() {
			err := checkIllegalSequence("ls & whoami")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with double ampersand", func() {
			err := checkIllegalSequence("ls && whoami")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with double pipe", func() {
			err := checkIllegalSequence("ls || echo fail")
			Expect(err).To(HaveOccurred())
		})

		// Test for variable expansion
		It("should reject script with dollar sign", func() {
			err := checkIllegalSequence("echo $USER")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with dollar and braces", func() {
			err := checkIllegalSequence("echo ${USER}")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with dollar and parentheses", func() {
			err := checkIllegalSequence("echo $(whoami)")
			Expect(err).To(HaveOccurred())
		})

		// Test for command substitution
		It("should reject script with backticks", func() {
			err := checkIllegalSequence("ls `pwd`")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with nested backticks", func() {
			err := checkIllegalSequence("echo `ls `pwd``")
			Expect(err).To(HaveOccurred())
		})

		// Test for quotes (preventing quote injection)
		It("should reject script with single quotes", func() {
			err := checkIllegalSequence("echo 'test'")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with multiple single quotes", func() {
			err := checkIllegalSequence("echo 'hello' 'world'")
			Expect(err).To(HaveOccurred())
		})

		// Test for parentheses (subshell execution)
		It("should reject script with parentheses", func() {
			err := checkIllegalSequence("ls (test)")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with opening parenthesis", func() {
			err := checkIllegalSequence("(ls)")
			Expect(err).To(HaveOccurred())
		})

		It("should reject script with closing parenthesis", func() {
			err := checkIllegalSequence("ls)")
			Expect(err).To(HaveOccurred())
		})

		// Test for redirection
		It("should reject script with redirect append", func() {
			err := checkIllegalSequence("echo test >> file")
			Expect(err).To(HaveOccurred())
		})

		// Test for allowed expressions
		It("should accept allowed expression ${METAURL}", func() {
			err := checkIllegalSequence("ls ${METAURL}")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept command with ${METAURL} in middle", func() {
			err := checkIllegalSequence("alluxio fs ls ${METAURL}/data")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept multiple ${METAURL} references", func() {
			err := checkIllegalSequence("cp ${METAURL}/src ${METAURL}/dst")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for safe commands
		It("should accept safe command", func() {
			err := checkIllegalSequence("ls /var/log")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept safe command with flags", func() {
			err := checkIllegalSequence("ls -la /home")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept safe command with multiple arguments", func() {
			err := checkIllegalSequence("mount -t nfs server:/path /mnt")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept empty string", func() {
			err := checkIllegalSequence("")
			Expect(err).NotTo(HaveOccurred())
		})

		// Test for combinations that should still be blocked
		It("should reject dollar even after allowed expression", func() {
			err := checkIllegalSequence("ls ${METAURL} && echo $HOME")
			Expect(err).To(HaveOccurred())
		})

		It("should reject semicolon with allowed expression", func() {
			err := checkIllegalSequence("ls ${METAURL}; rm -rf /")
			Expect(err).To(HaveOccurred())
		})
	})
})
