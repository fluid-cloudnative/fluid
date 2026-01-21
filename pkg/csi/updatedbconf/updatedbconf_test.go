/*
Copyright 2024 The Fluid Author.

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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUpdatedbconf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Updatedbconf Suite")
}

var _ = Describe("updateLine", func() {
	Context("when adding new values to a configuration line", func() {
		It("should add new filesystem types to PRUNEFS", func() {
			line := `PRUNEFS="foo bar"`
			newValues := []string{"fuse.alluxio-fuse", "JuiceFS"}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar fuse.alluxio-fuse JuiceFS"`))
		})

		It("should add new paths to PRUNEPATHS", func() {
			line := `PRUNEPATHS="/tmp /var/spool"`
			newValues := []string{"/runtime-mnt"}
			
			result, changed := updateLine(line, configKeyPrunePaths, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEPATHS="/tmp /var/spool /runtime-mnt"`))
		})

		It("should handle line with extra spaces", func() {
			line := `PRUNEFS  =  "foo bar"  `
			newValues := []string{"JuiceFS"}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar JuiceFS"`))
		})
	})

	Context("when values already exist", func() {
		It("should not add duplicate values and return false", func() {
			line := `PRUNEFS="foo bar JuiceFS"`
			newValues := []string{"foo", "bar"}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeFalse())
			Expect(result).To(Equal(line))
		})

		It("should add only new values when some already exist", func() {
			line := `PRUNEFS="foo bar"`
			newValues := []string{"foo", "JuiceFS", "bar", "fuse.alluxio-fuse"}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar JuiceFS fuse.alluxio-fuse"`))
		})
	})

	Context("when handling edge cases", func() {
		It("should handle empty existing values", func() {
			line := `PRUNEFS=""`
			newValues := []string{"foo", "bar"}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS=" foo bar"`))
		})

		It("should handle empty new values slice", func() {
			line := `PRUNEFS="foo bar"`
			newValues := []string{}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeFalse())
			Expect(result).To(Equal(line))
		})

		It("should filter out empty string values", func() {
			line := `PRUNEFS="foo"`
			newValues := []string{"", "bar", ""}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar"`))
		})

		It("should return false when only empty strings provided", func() {
			line := `PRUNEFS="foo"`
			newValues := []string{"", ""}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeFalse())
			Expect(result).To(Equal(line))
		})

		It("should handle line with leading space in value", func() {
			line := `PRUNEFS=" foo bar"`
			newValues := []string{"baz"}
			
			result, changed := updateLine(line, configKeyPruneFs, newValues)
			
			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar baz"`))
		})
	})
})

var _ = Describe("updateConfig", func() {
	var (
		newFs    []string
		newPaths []string
	)

	BeforeEach(func() {
		newFs = []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"}
		newPaths = []string{"/runtime-mnt"}
	})

	Context("when adding new filesystems and paths", func() {
		It("should add new paths and filesystems to existing configuration", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot"
PRUNEFS="foo bar"`

			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should not duplicate paths when path already exists", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar"`

			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should create new PRUNEFS and PRUNEPATHS entries when they don't exist", func() {
			content := `PRUNE_BIND_MOUNTS="yes"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring(`PRUNEFS="fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`))
			Expect(result).To(ContainSubstring(`PRUNEPATHS="/runtime-mnt"`))
		})
	})

	Context("when handling edge cases", func() {
		It("should handle empty content", func() {
			content := ""

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring(`PRUNEFS="fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`))
			Expect(result).To(ContainSubstring(`PRUNEPATHS="/runtime-mnt"`))
		})

		It("should handle empty newFs slice", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool"
PRUNEFS="foo bar"`

			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /runtime-mnt"
PRUNEFS="foo bar"`

			result, err := updateConfig(content, []string{}, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should handle empty newPaths slice", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool"
PRUNEFS="foo bar"`

			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should return original content when both newFs and newPaths are empty", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool"
PRUNEFS="foo bar"`

			result, err := updateConfig(content, []string{}, []string{})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(content))
		})

		It("should not duplicate filesystems when they already exist", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp"
PRUNEFS="foo bar fuse.alluxio-fuse JuiceFS"`

			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse JuiceFS fuse.jindofs-fuse fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should return original content when all values already exist", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/runtime-mnt"
PRUNEFS="fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(content))
		})
	})

	Context("when handling multiple paths", func() {
		It("should add multiple new paths correctly", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp"
PRUNEFS="foo"`

			multiplePaths := []string{"/runtime-mnt", "/data-mnt", "/cache-mnt"}
			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /runtime-mnt /data-mnt /cache-mnt"
PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, multiplePaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should skip already existing paths in multiple paths", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /data-mnt"
PRUNEFS="foo"`

			multiplePaths := []string{"/runtime-mnt", "/data-mnt", "/cache-mnt"}
			expected := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /data-mnt /runtime-mnt /cache-mnt"
PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, multiplePaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})
	})

	Context("when preserving existing configuration", func() {
		It("should preserve other configuration lines", func() {
			content := `# Comment line
PRUNE_BIND_MOUNTS="yes"
PRUNENAMES=".git .bzr .hg .svn"
PRUNEPATHS="/tmp"
PRUNEFS="foo"
ANOTHER_CONFIG="value"`

			expected := `# Comment line
PRUNE_BIND_MOUNTS="yes"
PRUNENAMES=".git .bzr .hg .svn"
PRUNEPATHS="/tmp /runtime-mnt"
PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"
ANOTHER_CONFIG="value"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should preserve blank lines", func() {
			content := `PRUNE_BIND_MOUNTS="yes"

PRUNEPATHS="/tmp"

PRUNEFS="foo"`

			expected := `PRUNE_BIND_MOUNTS="yes"

PRUNEPATHS="/tmp /runtime-mnt"

PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})

		It("should handle lines with only whitespace", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
   
PRUNEPATHS="/tmp"
		
PRUNEFS="foo"`

			expected := `PRUNE_BIND_MOUNTS="yes"
   
PRUNEPATHS="/tmp /runtime-mnt"
		
PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})
	})

	Context("when PRUNEFS appears multiple times", func() {
		It("should update all PRUNEFS lines", func() {
			content := `PRUNEFS="foo"
PRUNEPATHS="/tmp"
PRUNEFS="bar"`

			expected := `PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"
PRUNEPATHS="/tmp /runtime-mnt"
PRUNEFS="bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})
	})

	Context("when creating config from scratch", func() {
		It("should append PRUNEFS before PRUNEPATHS when both are missing", func() {
			content := `PRUNE_BIND_MOUNTS="yes"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring(`PRUNEFS="`))
			Expect(result).To(ContainSubstring(`PRUNEPATHS="`))
		})

		It("should only append PRUNEFS when PRUNEPATHS exists", func() {
			content := `PRUNEPATHS="/tmp"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(ContainSubstring(`PRUNEPATHS="/tmp /runtime-mnt"`))
			Expect(result).To(ContainSubstring(`PRUNEFS="fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`))
		})

		It("should only append PRUNEPATHS when PRUNEFS exists", func() {
			content := `PRUNEFS="foo"`

			expected := `PRUNEFS="foo fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"
PRUNEPATHS="/runtime-mnt"`

			result, err := updateConfig(content, newFs, newPaths)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})
	})
})