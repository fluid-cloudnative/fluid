/*
Copyright 2026 The Fluid Authors.

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

package plugins

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolveMountSource", func() {
	var (
		tempDir   string
		fluidPath string
		outside   string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "resolve-mount-source-*")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		fluidPath = filepath.Join(tempDir, "runtime-mnt", "fuse")
		outside = filepath.Join(tempDir, "outside")
		Expect(os.MkdirAll(filepath.Join(fluidPath, "sub", "nested"), 0750)).To(Succeed())
		Expect(os.MkdirAll(outside, 0750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(outside, "passwd"), []byte("secret"), 0600)).To(Succeed())
	})

	It("returns the mount point itself for an empty subPath", func() {
		source, closer, err := resolveMountSource(fluidPath, "")
		Expect(err).NotTo(HaveOccurred())
		defer closer()
		Expect(source).To(Equal(fluidPath))
	})

	It("resolves a nested subPath", func() {
		source, closer, err := resolveMountSource(fluidPath, "sub/nested")
		Expect(err).NotTo(HaveOccurred())
		defer closer()
		Expect(source).NotTo(BeEmpty())
	})

	It("rejects a subPath escaping the mount point", func() {
		_, _, err := resolveMountSource(fluidPath, "../../outside")
		Expect(err).To(HaveOccurred())
	})

	It("rejects a subPath whose last component is a symlink", func() {
		Expect(os.Symlink(outside, filepath.Join(fluidPath, "evil"))).To(Succeed())

		_, _, err := resolveMountSource(fluidPath, "evil")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("symlink"))
	})

	// Regression test: an Lstat of the joined path reports "passwd" as a regular file and misses
	// that the "link" component already led out of the mount point.
	It("rejects a subPath whose intermediate component is a symlink", func() {
		Expect(os.Symlink(outside, filepath.Join(fluidPath, "link"))).To(Succeed())

		_, _, err := resolveMountSource(fluidPath, "link/passwd")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("symlink"))
	})
})
