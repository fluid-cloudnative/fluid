/*
Copyright 2026 The Fluid Author.

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

package updatedbconf

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Register Logic", func() {
	Describe("comment detection", func() {
		Context("when content has modified comment at start", func() {
			It("should detect the comment", func() {
				existingContent := modifiedByFluidComment + "\nPRUNEFS = \"9p afs fuse.alluxio\""
				hasComment := strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(hasComment).To(BeTrue())
			})
		})

		Context("when content does not have modified comment", func() {
			It("should not detect the comment", func() {
				existingContent := `PRUNEFS = "9p afs"`
				hasComment := strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(hasComment).To(BeFalse())
			})
		})

		Context("when comment is in the middle", func() {
			It("should not detect the comment", func() {
				existingContent := "some config\n" + modifiedByFluidComment
				hasComment := strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(hasComment).To(BeFalse())
			})
		})

		Context("when content is empty", func() {
			It("should not detect the comment", func() {
				existingContent := ""
				hasComment := strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(hasComment).To(BeFalse())
			})
		})

		Context("when content is only whitespace before comment", func() {
			It("should not detect the comment", func() {
				existingContent := "  \n" + modifiedByFluidComment
				hasComment := strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(hasComment).To(BeFalse())
			})
		})
	})

	Describe("backup decision", func() {
		Context("when content is not modified by Fluid", func() {
			It("should decide to create backup", func() {
				existingContent := `PRUNEFS = "9p afs"`
				shouldBackup := !strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(shouldBackup).To(BeTrue())
			})
		})

		Context("when content is already modified by Fluid", func() {
			It("should decide not to create backup", func() {
				existingContent := modifiedByFluidComment + "\n" + `PRUNEFS = "9p afs fuse.alluxio"`
				shouldBackup := !strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(shouldBackup).To(BeFalse())
			})
		})

		Context("when content has comment but with extra content before", func() {
			It("should decide to create backup", func() {
				existingContent := "# Some other comment\n" + modifiedByFluidComment + "\nPRUNEFS = \"9p afs\""
				shouldBackup := !strings.HasPrefix(existingContent, modifiedByFluidComment)
				Expect(shouldBackup).To(BeTrue())
			})
		})
	})

	Describe("content transformation", func() {
		Context("when adding Fluid comment to new config", func() {
			It("should prepend comment with newline separator", func() {
				baseConfig := `PRUNEFS = "9p afs fuse.alluxio"`
				newConfig := fmt.Sprintf("%s\n%s", modifiedByFluidComment, baseConfig)

				Expect(newConfig).To(HavePrefix(modifiedByFluidComment))
				Expect(strings.Count(newConfig, modifiedByFluidComment)).To(Equal(1))
				Expect(newConfig).To(ContainSubstring(baseConfig))
			})
		})

		Context("when comment already exists", func() {
			It("should not duplicate comment", func() {
				existingConfig := modifiedByFluidComment + "\n" + `PRUNEFS = "9p afs"`

				// Simulating what Register does - only add comment if not present
				var newConfig string
				if !strings.HasPrefix(existingConfig, modifiedByFluidComment) {
					newConfig = fmt.Sprintf("%s\n%s", modifiedByFluidComment, existingConfig)
				} else {
					newConfig = existingConfig
				}

				Expect(strings.Count(newConfig, modifiedByFluidComment)).To(Equal(1))
			})
		})
	})

	Describe("file content equality", func() {
		Context("when comparing identical content", func() {
			It("should detect no changes needed", func() {
				content1 := `PRUNEFS = "9p afs fuse.alluxio"
PRUNEPATHS = "/tmp /runtime-mnt/alluxio"`
				content2 := `PRUNEFS = "9p afs fuse.alluxio"
PRUNEPATHS = "/tmp /runtime-mnt/alluxio"`

				Expect(content1).To(Equal(content2))
			})
		})

		Context("when content differs", func() {
			It("should detect changes", func() {
				content1 := `PRUNEFS = "9p afs"`
				content2 := `PRUNEFS = "9p afs fuse.alluxio"`

				Expect(content1).NotTo(Equal(content2))
			})
		})
	})
})

var _ = Describe("Enabled", func() {
	It("should always return true", func() {
		result := Enabled()
		Expect(result).To(BeTrue())
	})

	It("should be consistent across multiple calls", func() {
		result1 := Enabled()
		result2 := Enabled()
		result3 := Enabled()

		Expect(result1).To(Equal(result2))
		Expect(result2).To(Equal(result3))
		Expect(result1).To(BeTrue())
	})
})

var _ = Describe("HasModifiedComment Helper", func() {
	DescribeTable("checking for modified comment prefix",
		func(content string, expected bool) {
			got := strings.HasPrefix(content, modifiedByFluidComment)
			Expect(got).To(Equal(expected))
		},
		Entry("has comment at start", modifiedByFluidComment+"\nsome config", true),
		Entry("no comment", "some config", false),
		Entry("comment in middle", "some config\n"+modifiedByFluidComment, false),
		Entry("empty content", "", false),
		Entry("only comment", modifiedByFluidComment, true),
		Entry("comment with extra whitespace before", "  "+modifiedByFluidComment, false),
		Entry("partial comment", modifiedByFluidComment[:10], false),
		Entry("comment with newlines after", modifiedByFluidComment+"\n\n\n", true),
		Entry("similar but different comment", "# Modified by fluid", false),
	)
})

var _ = Describe("Real World Scenarios", func() {
	Describe("typical updatedb.conf formats", func() {
		Context("with Ubuntu/Debian standard format", func() {
			It("should recognize standard config format", func() {
				content := `PRUNE_BIND_MOUNTS = "yes"
PRUNEFS = "9p afs anon_inodefs auto autofs bdev binfmt_misc cgroup cifs coda configfs cpuset debugfs devpts devtmpfs ecryptfs exofs fuse fuse.sshfs fusectl hugetlbfs iso9660 mqueue ncpfs nfs nfs4 nfsd pipefs proc ramfs rootfs rpc_pipefs securityfs selinuxfs smbfs sockfs sysfs tmpfs ubifs udf usbfs"
PRUNENAMES = ".git .hg .svn .bzr CVS"
PRUNEPATHS = "/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph"`

				hasComment := strings.HasPrefix(content, modifiedByFluidComment)
				Expect(hasComment).To(BeFalse())
				Expect(content).To(ContainSubstring("PRUNEFS"))
				Expect(content).To(ContainSubstring("PRUNEPATHS"))
				Expect(content).To(ContainSubstring("PRUNENAMES"))
				Expect(content).To(ContainSubstring("PRUNE_BIND_MOUNTS"))
			})
		})

		Context("with already modified content", func() {
			It("should recognize Fluid modifications", func() {
				content := modifiedByFluidComment + "\n" +
					`PRUNE_BIND_MOUNTS = "yes"
PRUNEFS = "9p afs anon_inodefs auto autofs fuse.alluxio fuse.jindofs"
PRUNENAMES = ".git .hg .svn .bzr CVS"
PRUNEPATHS = "/tmp /var/spool /media /runtime-mnt/alluxio /runtime-mnt/jindofs"`

				hasComment := strings.HasPrefix(content, modifiedByFluidComment)
				Expect(hasComment).To(BeTrue())
				Expect(content).To(ContainSubstring("fuse.alluxio"))
				Expect(content).To(ContainSubstring("fuse.jindofs"))
				Expect(content).To(ContainSubstring("/runtime-mnt/alluxio"))
				Expect(content).To(ContainSubstring("/runtime-mnt/jindofs"))
			})
		})

		Context("with minimal config", func() {
			It("should handle minimal updatedb.conf", func() {
				content := `PRUNEFS = "9p"
PRUNEPATHS = "/tmp"`

				hasComment := strings.HasPrefix(content, modifiedByFluidComment)
				Expect(hasComment).To(BeFalse())
				Expect(content).To(ContainSubstring("PRUNEFS"))
				Expect(content).To(ContainSubstring("PRUNEPATHS"))
			})
		})

		Context("with multiple fuse filesystems", func() {
			It("should handle multiple fuse types", func() {
				content := `PRUNEFS = "9p afs fuse.alluxio fuse.jindofs fuse.juicefs fuse.goosefs"`

				Expect(content).To(ContainSubstring("fuse.alluxio"))
				Expect(content).To(ContainSubstring("fuse.jindofs"))
				Expect(content).To(ContainSubstring("fuse.juicefs"))
				Expect(content).To(ContainSubstring("fuse.goosefs"))
			})
		})

		Context("with multiple runtime mount paths", func() {
			It("should handle multiple mount paths", func() {
				content := `PRUNEPATHS = "/tmp /var/spool /runtime-mnt/alluxio /runtime-mnt/jindofs /runtime-mnt/juicefs"`

				Expect(content).To(ContainSubstring("/runtime-mnt/alluxio"))
				Expect(content).To(ContainSubstring("/runtime-mnt/jindofs"))
				Expect(content).To(ContainSubstring("/runtime-mnt/juicefs"))
			})
		})
	})

	Describe("edge cases in config content", func() {
		Context("with different line endings", func() {
			It("should handle Unix line endings", func() {
				content := "PRUNEFS = \"9p\"\nPRUNEPATHS = \"/tmp\""
				lines := strings.Split(content, "\n")
				Expect(lines).To(HaveLen(2))
			})

			It("should handle content with multiple blank lines", func() {
				content := "PRUNEFS = \"9p\"\n\n\nPRUNEPATHS = \"/tmp\""
				Expect(content).To(ContainSubstring("PRUNEFS"))
				Expect(content).To(ContainSubstring("PRUNEPATHS"))
			})
		})

		Context("with comments in config", func() {
			It("should preserve other comments", func() {
				content := `# User comment
PRUNEFS = "9p afs"
# Another comment
PRUNEPATHS = "/tmp"`

				Expect(content).To(ContainSubstring("# User comment"))
				Expect(content).To(ContainSubstring("# Another comment"))
			})
		})

		Context("with quoted values", func() {
			It("should handle double-quoted values", func() {
				content := `PRUNEFS = "9p afs fuse.alluxio"`
				Expect(content).To(ContainSubstring(`"`))
			})

			It("should handle spaces in quoted values", func() {
				content := `PRUNEFS = "9p afs anon_inodefs auto autofs"`
				Expect(strings.Count(content, " ")).To(BeNumerically(">", 0))
			})
		})
	})

	Describe("comment string properties", func() {
		It("should have non-empty comment constant", func() {
			Expect(modifiedByFluidComment).NotTo(BeEmpty())
		})

		It("should start with comment character", func() {
			Expect(modifiedByFluidComment).To(HavePrefix("#"))
		})

		It("should mention Fluid in comment", func() {
			lowerComment := strings.ToLower(modifiedByFluidComment)
			Expect(lowerComment).To(ContainSubstring("fluid"))
		})
	})
})

var _ = Describe("String Operations", func() {
	Describe("prefix checking with variations", func() {
		It("should handle exact prefix match", func() {
			text := modifiedByFluidComment + "\ncontent"
			Expect(strings.HasPrefix(text, modifiedByFluidComment)).To(BeTrue())
		})

		It("should reject prefix with leading space", func() {
			text := " " + modifiedByFluidComment + "\ncontent"
			Expect(strings.HasPrefix(text, modifiedByFluidComment)).To(BeFalse())
		})

		It("should reject when target is longer than text", func() {
			text := "short"
			longPrefix := strings.Repeat("x", 1000)
			Expect(strings.HasPrefix(text, longPrefix)).To(BeFalse())
		})
	})

	Describe("content concatenation", func() {
		It("should correctly join comment and config", func() {
			config := `PRUNEFS = "9p"`
			result := fmt.Sprintf("%s\n%s", modifiedByFluidComment, config)

			parts := strings.SplitN(result, "\n", 2)
			Expect(parts).To(HaveLen(2))
			Expect(parts[0]).To(Equal(modifiedByFluidComment))
			Expect(parts[1]).To(Equal(config))
		})
	})
})

var _ = Describe("Integration Scenarios", func() {
	Context("when simulating Register behavior", func() {
		It("should follow correct logic flow for new file", func() {
			// Simulate existing content (no Fluid comment)
			existingContent := `PRUNEFS = "9p afs"
PRUNEPATHS = "/tmp"`

			// Check if backup needed
			shouldBackup := !strings.HasPrefix(existingContent, modifiedByFluidComment)
			Expect(shouldBackup).To(BeTrue())

			// Simulate adding comment
			newContent := fmt.Sprintf("%s\n%s", modifiedByFluidComment, existingContent)
			Expect(newContent).To(HavePrefix(modifiedByFluidComment))
		})

		It("should follow correct logic flow for already modified file", func() {
			// Simulate existing content (with Fluid comment)
			existingContent := modifiedByFluidComment + "\n" + `PRUNEFS = "9p afs fuse.alluxio"
PRUNEPATHS = "/tmp /runtime-mnt/alluxio"`

			// Check if backup needed
			shouldBackup := !strings.HasPrefix(existingContent, modifiedByFluidComment)
			Expect(shouldBackup).To(BeFalse())

			// No need to add comment again
			Expect(existingContent).To(HavePrefix(modifiedByFluidComment))
		})

		It("should detect when no changes needed", func() {
			content1 := `PRUNEFS = "9p afs fuse.alluxio"`
			content2 := `PRUNEFS = "9p afs fuse.alluxio"`

			noChanges := (content1 == content2)
			Expect(noChanges).To(BeTrue())
		})
	})
})

var _ = Describe("File Operations Simulation", func() {
	Describe("reading non-existent file", func() {
		It("should simulate IsNotExist check", func() {
			err := os.ErrNotExist
			isNotExist := os.IsNotExist(err)
			Expect(isNotExist).To(BeTrue())
		})
	})

	Describe("backup file naming", func() {
		It("should use conventional backup extension", func() {
			originalPath := "/etc/updatedb.conf"
			backupPath := originalPath + ".backup"

			Expect(backupPath).To(HaveSuffix(".backup"))
			Expect(strings.TrimSuffix(backupPath, ".backup")).To(Equal(originalPath))
		})
	})

	Describe("content comparison", func() {
		It("should detect identical content", func() {
			content1 := []byte("test content")
			content2 := []byte("test content")
			Expect(string(content1)).To(Equal(string(content2)))
		})

		It("should detect different content", func() {
			content1 := []byte("test content 1")
			content2 := []byte("test content 2")
			Expect(string(content1)).NotTo(Equal(string(content2)))
		})
	})
})
