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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("updateLine", func() {
	Context("when adding new values to existing line", func() {
		It("should append new values and return true", func() {
			line := `PRUNEFS="foo bar"`
			newValues := []string{"baz", "qux"}

			result, changed := updateLine(line, "PRUNEFS", newValues)

			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar baz qux"`))
		})
	})

	Context("when all values already exist", func() {
		It("should return original line and false", func() {
			line := `PRUNEFS="foo bar baz"`
			newValues := []string{"foo", "bar"}

			result, changed := updateLine(line, "PRUNEFS", newValues)

			Expect(changed).To(BeFalse())
			Expect(result).To(Equal(line))
		})
	})

	Context("when new values contain empty strings", func() {
		It("should skip empty values", func() {
			line := `PRUNEFS="foo"`
			newValues := []string{"", "bar", ""}

			result, changed := updateLine(line, "PRUNEFS", newValues)

			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar"`))
		})
	})

	Context("when line has various spacing formats", func() {
		It("should handle line with spaces around equals", func() {
			line := `PRUNEFS = "foo"`
			newValues := []string{"bar"}

			result, changed := updateLine(line, "PRUNEFS", newValues)

			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEFS="foo bar"`))
		})

		It("should handle line with extra spaces in value", func() {
			line := `PRUNEFS="  foo  bar  "`
			newValues := []string{"baz"}

			result, changed := updateLine(line, "PRUNEFS", newValues)

			Expect(changed).To(BeTrue())
			Expect(result).To(ContainSubstring("baz"))
		})
	})

	Context("when no new values provided", func() {
		It("should return false for empty slice", func() {
			line := `PRUNEFS="foo bar"`
			newValues := []string{}

			result, changed := updateLine(line, "PRUNEFS", newValues)

			Expect(changed).To(BeFalse())
			Expect(result).To(Equal(line))
		})
	})

	Context("when dealing with PRUNEPATHS", func() {
		It("should handle path values correctly", func() {
			line := `PRUNEPATHS="/tmp /var"`
			newValues := []string{"/runtime-mnt", "/new-path"}

			result, changed := updateLine(line, "PRUNEPATHS", newValues)

			Expect(changed).To(BeTrue())
			Expect(result).To(Equal(`PRUNEPATHS="/tmp /var /runtime-mnt /new-path"`))
		})
	})
})

var _ = Describe("updateConfig", func() {
	Context("when adding new path and fs", func() {
		It("should update configuration correctly", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot"
PRUNEFS="foo bar"`
			newFs := []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"}
			newPaths := []string{"/runtime-mnt"}
			want := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(want))
		})
	})

	Context("when no new path is needed", func() {
		It("should only update filesystem types", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar"`
			newFs := []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"}
			newPaths := []string{"/runtime-mnt"}
			want := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /var/spool /media /var/lib/os-prober /var/lib/ceph /home/.ecryptfs /var/lib/schroot /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"`

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(want))
		})
	})

	Context("when path or fs config is empty", func() {
		It("should add new configuration lines", func() {
			content := `PRUNE_BIND_MOUNTS="yes"`
			newFs := []string{"fuse.alluxio-fuse", "fuse.jindofs-fuse", "JuiceFS", "fuse.goosefs-fuse"}
			newPaths := []string{"/runtime-mnt"}
			want := `PRUNE_BIND_MOUNTS="yes"
PRUNEFS="fuse.alluxio-fuse fuse.jindofs-fuse JuiceFS fuse.goosefs-fuse"
PRUNEPATHS="/runtime-mnt"`

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(want))
		})
	})

	Context("when no changes are needed", func() {
		It("should return original content", func() {
			content := `PRUNE_BIND_MOUNTS="yes"
PRUNEPATHS="/tmp /runtime-mnt"
PRUNEFS="foo bar fuse.alluxio-fuse"`
			newFs := []string{"fuse.alluxio-fuse"}
			newPaths := []string{"/runtime-mnt"}

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(content))
		})
	})

	Context("when only PRUNEFS config exists", func() {
		It("should add PRUNEPATHS line", func() {
			content := `PRUNEFS="foo bar"`
			newFs := []string{}
			newPaths := []string{"/runtime-mnt"}
			want := `PRUNEFS="foo bar"
PRUNEPATHS="/runtime-mnt"`

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(want))
		})
	})

	Context("when only PRUNEPATHS config exists", func() {
		It("should add PRUNEFS line", func() {
			content := `PRUNEPATHS="/tmp"`
			newFs := []string{"fuse.alluxio-fuse"}
			newPaths := []string{}
			want := `PRUNEPATHS="/tmp"
PRUNEFS="fuse.alluxio-fuse"`

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(want))
		})
	})

	Context("when content has leading/trailing whitespace on lines", func() {
		It("should handle whitespace correctly", func() {
			content := `  PRUNEFS="foo"  
  PRUNEPATHS="/tmp"  `
			newFs := []string{"bar"}
			newPaths := []string{"/new"}

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(ContainSubstring("bar"))
			Expect(got).To(ContainSubstring("/new"))
		})
	})

	Context("when both newFs and newPaths are empty", func() {
		It("should return original content if configs exist", func() {
			content := `PRUNEFS="foo"
PRUNEPATHS="/tmp"`
			newFs := []string{}
			newPaths := []string{}

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(content))
		})

		It("should not add lines if configs don't exist", func() {
			content := `PRUNE_BIND_MOUNTS="yes"`
			newFs := []string{}
			newPaths := []string{}

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(content))
		})
	})

	Context("when content is empty", func() {
		It("should add both config lines", func() {
			content := ``
			newFs := []string{"fuse.alluxio-fuse"}
			newPaths := []string{"/runtime-mnt"}

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(ContainSubstring(`PRUNEFS="fuse.alluxio-fuse"`))
			Expect(got).To(ContainSubstring(`PRUNEPATHS="/runtime-mnt"`))
		})
	})

	Context("when content has multiple lines with comments", func() {
		It("should preserve comments and update configs", func() {
			content := `# This is a comment
PRUNEFS="foo"
# Another comment
PRUNEPATHS="/tmp"`
			newFs := []string{"bar"}
			newPaths := []string{"/new"}

			got, err := updateConfig(content, newFs, newPaths)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(ContainSubstring("# This is a comment"))
			Expect(got).To(ContainSubstring("# Another comment"))
			Expect(got).To(ContainSubstring("bar"))
			Expect(got).To(ContainSubstring("/new"))
		})
	})
})
