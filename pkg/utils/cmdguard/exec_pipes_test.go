package cmdguard

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cmdguard pipes", func() {
	Describe("validateShellPipeString", func() {
		DescribeTable("pipe string validation",
			func(command string, wantErr bool) {
				err := validateShellPipeString(command)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("valid command with grep", "echo hello world | grep hello", true),
			Entry("valid command with wc -l", "ls file | wc -l", false),
			Entry("invalid command with xyz", "echo hello world | xyz", true),
			Entry("invalid command with kubectl", "kubectl hello world | xyz", true),
			Entry("illegal sequence in command with &", "echo hello world & echo y", true),
			Entry("illegal sequence in command with ;", "ls ; echo y", true),
			Entry("command with $", "ls $HOME", true),
			Entry("command with absolute path", "ls /etc", false),
		)
	})

	Describe("ShellCommand", func() {
		DescribeTable("shell command creation",
			func(name string, arg []string, wantCmd *exec.Cmd, wantErr bool) {
				gotCmd, err := ShellCommand(name, arg...)
				if wantErr {
					Expect(err).To(HaveOccurred())
					Expect(gotCmd).To(BeNil())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(gotCmd).NotTo(BeNil())
					Expect(gotCmd.Path).To(Equal(wantCmd.Path))
					Expect(gotCmd.Args).To(Equal(wantCmd.Args))
				}
			},
			Entry("valid simple command", "bash", []string{"-c", "ls"}, exec.Command("bash", "-c", "ls"), false),
			Entry("insufficient arguments", "bash", []string{"-c"}, nil, true),
			Entry("unknown shell command", "zsh", []string{"-c", "ls"}, nil, true),
			Entry("valid piped command", "bash", []string{"-c", "ls | grep something"}, exec.Command("bash", "-c", "ls | grep something"), false),
			Entry("invalid piped command", "bash", []string{"-c", "ls | random-command"}, nil, true),
		)
	})

	Describe("isValidCommand", func() {
		DescribeTable("command validity",
			func(cmd string, allowedCommands map[string]CommandValidater, want bool) {
				Expect(isValidCommand(cmd, allowedCommands)).To(Equal(want))
			},
			Entry("valid bash command", "bash", map[string]CommandValidater{"bash": ExactMatch}, true),
			Entry("valid sh command", "sh", map[string]CommandValidater{"bash": ExactMatch, "sh": ExactMatch}, true),
			Entry("invalid zsh command", "zsh", map[string]CommandValidater{"bash": ExactMatch}, false),
		)
	})

	Describe("splitShellCommand", func() {
		DescribeTable("splitting shell command",
			func(shellCommandSlice []string, wantShellCommand, wantPipedCommands string, wantErr bool) {
				gotShellCommand, gotPipedCommands, err := splitShellCommand(shellCommandSlice)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(gotShellCommand).To(Equal(wantShellCommand))
					Expect(gotPipedCommands).To(Equal(wantPipedCommands))
				}
			},
			Entry("valid shell command", []string{" bash ", "  -c", "echo foobar | grep foo"}, "bash -c", "echo foobar | grep foo", false),
			Entry("empty shell command", []string{}, "", "", true),
			Entry("invalid command without shell", []string{"echo foobar | grep foo"}, "", "", true),
			Entry("valid command without shell", []string{"test", "hello", "--help"}, "test hello", "--help", false),
		)
	})

	Describe("validateShellCommand", func() {
		DescribeTable("shell command validation",
			func(shellCommand string, wantErr bool) {
				err := validateShellCommand(shellCommand)
				if wantErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("bash shell", "bash -c", false),
			Entry("sh shell", "sh -c", false),
			Entry("zsh shell(invalid)", "zsh -c", true),
			Entry("bash command", "bash -s", true),
		)
	})
})
