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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("configUpdate", func() {
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
})
