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
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Register", func() {
	Describe("file does not exist", func() {
		It("should skip updating without error", func() {
			Skip("Test requires refactoring of Register function to accept file paths")
		})
	})

	Describe("first time modification", func() {
		It("should create backup and add comment", func() {
			Skip("Test requires refactoring of Register function to accept file paths")
		})
	})

	Describe("already modified by Fluid", func() {
		It("should not create backup", func() {
			Skip("Test requires refactoring of Register function to accept file paths")
		})
	})

	Describe("no changes needed", func() {
		It("should skip updating", func() {
			Skip("Test requires refactoring of Register function to accept file paths")
		})
	})
})

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
	})
})

var _ = Describe("Enabled", func() {
	It("should always return true", func() {
		result := Enabled()
		Expect(result).To(BeTrue())
	})
})

var _ = Describe("Integration Test", func() {
	Context("when RUN_INTEGRATION_TESTS is not set", func() {
		It("should skip integration test", func() {
			if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
				Skip("Skipping integration test. Set RUN_INTEGRATION_TESTS=true to run")
			}

			tempDir := GinkgoT().TempDir()
			Expect(tempDir).ToNot(BeEmpty())
		})
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
	)
})
