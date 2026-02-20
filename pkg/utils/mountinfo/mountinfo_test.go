/*
Copyright 2021 The Fluid Authors.

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

package mountinfo

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func loadMountInfoFromString(str string) (map[string]*Mount, error) {
	return readMountInfo(strings.NewReader(str))
}

var _ = Describe("MountInfo", func() {
	Describe("loadMountInfoFromString", func() {
		Context("when loading basic mount information", func() {
			It("should successfully parse a single mountpoint", func() {
				mountinfo := `
15 0 259:3 / / rw,relatime shared:1 - ext4 /dev/root rw,data=ordered
`
				mountMap, err := loadMountInfoFromString(mountinfo)
				Expect(err).NotTo(HaveOccurred())

				mnt, ok := mountMap["/"]
				Expect(ok).To(BeTrue(), "mount path should exist in map")
				Expect(mnt.MountPath).To(Equal("/"))
				Expect(mnt.FilesystemType).To(Equal("ext4"))
				Expect(mnt.PeerGroups).To(HaveLen(1))
				Expect(mnt.PeerGroups[1]).To(BeTrue())
				Expect(mnt.Subtree).To(Equal("/"))
				Expect(mnt.ReadOnly).To(BeFalse())
				Expect(mnt.Count).To(Equal(1))
			})
		})

		Context("when mount has no peer group", func() {
			It("should return empty map for mounts without peer groups", func() {
				mountinfo := `
15 0 259:3 / / rw,relatime - ext4 /dev/root rw,data=ordered
`
				mountMap, err := loadMountInfoFromString(mountinfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(mountMap).To(BeEmpty(), "mounts without peer groups should not be included")
			})
		})

		Context("when multiple mounts exist for same path", func() {
			It("should keep the last mount and increment count", func() {
				mountinfo := `
15 0 259:3 / / rw,relatime shared:1 - ext4 /dev/root rw,data=ordered
15 0 259:3 / / rw,relatime shared:2 - ext4 /dev/root rw,data=ordered
`
				mountMap, err := loadMountInfoFromString(mountinfo)
				Expect(err).NotTo(HaveOccurred())

				mnt, ok := mountMap["/"]
				Expect(ok).To(BeTrue())
				Expect(mnt.MountPath).To(Equal("/"))
				Expect(mnt.FilesystemType).To(Equal("ext4"))
				Expect(mnt.PeerGroups).To(HaveLen(1))
				Expect(mnt.PeerGroups[2]).To(BeTrue())
				Expect(mnt.Subtree).To(Equal("/"))
				Expect(mnt.ReadOnly).To(BeFalse())
				Expect(mnt.Count).To(Equal(2))
			})
		})

		Context("when mount info is empty or malformed", func() {
			It("should handle empty string", func() {
				mountMap, err := loadMountInfoFromString("")
				Expect(err).NotTo(HaveOccurred())
				Expect(mountMap).To(BeEmpty())
			})

			It("should handle whitespace only", func() {
				mountMap, err := loadMountInfoFromString("   \n\t  \n")
				Expect(err).NotTo(HaveOccurred())
				Expect(mountMap).To(BeEmpty())
			})

			It("should skip invalid lines", func() {
				mountinfo := `
invalid line here
15 0 259:3 / / rw,relatime shared:1 - ext4 /dev/root rw,data=ordered
another invalid line
`
				mountMap, err := loadMountInfoFromString(mountinfo)
				Expect(err).NotTo(HaveOccurred())
				Expect(mountMap).To(HaveLen(1))
			})
		})

		Context("when mount has read-only flag", func() {
			It("should correctly parse ro flag", func() {
				mountinfo := `
1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0
`
				mountMap, err := loadMountInfoFromString(mountinfo)
				Expect(err).NotTo(HaveOccurred())

				mnt, ok := mountMap["/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"]
				Expect(ok).To(BeTrue())
				Expect(mnt.ReadOnly).To(BeTrue())
			})
		})

		Context("when mount has multiple peer groups", func() {
			It("should parse all peer groups correctly", func() {
				mountinfo := `
1764 1620 0:388 / /runtime-mnt/juicefs ro,relatime shared:475 master:478 - fuse.juicefs JuiceFS:minio ro
`
				mountMap, err := loadMountInfoFromString(mountinfo)
				Expect(err).NotTo(HaveOccurred())

				mnt, ok := mountMap["/runtime-mnt/juicefs"]
				Expect(ok).To(BeTrue())
				Expect(mnt.PeerGroups).To(HaveLen(2))
				Expect(mnt.PeerGroups[475]).To(BeTrue())
				Expect(mnt.PeerGroups[478]).To(BeTrue())
			})
		})
	})

	Describe("parseMountInfoLine", func() {
		Context("with valid mount info lines", func() {
			It("should parse basic fuse.juicefs mount", func() {
				line := "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other"

				mnt := parseMountInfoLine(line)
				Expect(mnt).NotTo(BeNil())
				Expect(mnt.Subtree).To(Equal("/"))
				Expect(mnt.MountPath).To(Equal("/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"))
				Expect(mnt.FilesystemType).To(Equal("fuse.juicefs"))
				Expect(mnt.PeerGroups).To(HaveLen(1))
				Expect(mnt.PeerGroups[475]).To(BeTrue())
				Expect(mnt.ReadOnly).To(BeTrue())
				Expect(mnt.Count).To(Equal(1))
			})

			It("should parse mount with no peer group", func() {
				line := "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other"

				mnt := parseMountInfoLine(line)
				Expect(mnt).NotTo(BeNil())
				Expect(mnt.PeerGroups).To(BeEmpty())
				Expect(mnt.ReadOnly).To(BeTrue())
			})

			It("should parse mount with multiple peer groups", func() {
				line := "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 master:478 - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other"

				mnt := parseMountInfoLine(line)
				Expect(mnt).NotTo(BeNil())
				Expect(mnt.PeerGroups).To(HaveLen(2))
				Expect(mnt.PeerGroups[475]).To(BeTrue())
				Expect(mnt.PeerGroups[478]).To(BeTrue())
			})

			It("should parse mount with rw flag as not readonly", func() {
				line := "15 0 259:3 / / rw,relatime shared:1 - ext4 /dev/root rw,data=ordered"

				mnt := parseMountInfoLine(line)
				Expect(mnt).NotTo(BeNil())
				Expect(mnt.ReadOnly).To(BeFalse())
			})
		})

		Context("with invalid mount info lines", func() {
			It("should return nil for line with invalid peer group number", func() {
				line := "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:abc - fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other"

				mnt := parseMountInfoLine(line)
				Expect(mnt).To(BeNil())
			})

			It("should return nil for line missing separator", func() {
				line := "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other"

				mnt := parseMountInfoLine(line)
				Expect(mnt).To(BeNil())
			})

			It("should return nil for line with trailing separator", func() {
				line := "1764 1620 0:388 / /runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse ro,relatime shared:475 fuse.juicefs JuiceFS:minio ro,user_id=0,group_id=0,default_permissions,allow_other -"

				mnt := parseMountInfoLine(line)
				Expect(mnt).To(BeNil())
			})

			It("should return nil for empty line", func() {
				mnt := parseMountInfoLine("")
				Expect(mnt).To(BeNil())
			})

			It("should return nil for line with too few fields", func() {
				line := "1764 1620 0:388"

				mnt := parseMountInfoLine(line)
				Expect(mnt).To(BeNil())
			})

			It("should return nil for line with invalid format before separator", func() {
				line := "1764 1620 - fuse.juicefs device rw"

				mnt := parseMountInfoLine(line)
				Expect(mnt).To(BeNil())
			})
		})

		Context("with escaped characters in paths", func() {
			It("should unescape mount paths with octal sequences", func() {
				line := "15 0 259:3 / /path\\040with\\040spaces rw,relatime shared:1 - ext4 /dev/root rw"

				mnt := parseMountInfoLine(line)
				Expect(mnt).NotTo(BeNil())
				Expect(mnt.MountPath).To(Equal("/path with spaces"))
			})

			It("should unescape subtree paths", func() {
				line := "15 0 259:3 /sub\\040tree /mount rw,relatime shared:1 - ext4 /dev/root rw"

				mnt := parseMountInfoLine(line)
				Expect(mnt).NotTo(BeNil())
				Expect(mnt.Subtree).To(Equal("/sub tree"))
			})
		})
	})

	Describe("peerGroupFromString", func() {
		Context("with valid peer group strings", func() {
			It("should parse shared peer group", func() {
				pgTag, pg, err := peerGroupFromString("shared:475")

				Expect(err).NotTo(HaveOccurred())
				Expect(pgTag).To(Equal("shared"))
				Expect(pg).To(Equal(475))
			})

			It("should parse master peer group", func() {
				pgTag, pg, err := peerGroupFromString("master:475")

				Expect(err).NotTo(HaveOccurred())
				Expect(pgTag).To(Equal("master"))
				Expect(pg).To(Equal(475))
			})
		})

		Context("with invalid peer group strings", func() {
			It("should return error for unbindable", func() {
				pgTag, pg, err := peerGroupFromString("unbindable")

				Expect(err).To(HaveOccurred())
				Expect(pgTag).To(Equal(""))
				Expect(pg).To(Equal(-1))
			})

			It("should return error for private", func() {
				pgTag, pg, err := peerGroupFromString("private")

				Expect(err).To(HaveOccurred())
				Expect(pgTag).To(Equal(""))
				Expect(pg).To(Equal(-1))
			})

			It("should return error for invalid format", func() {
				pgTag, pg, err := peerGroupFromString("invalid")

				Expect(err).To(HaveOccurred())
				Expect(pgTag).To(Equal(""))
				Expect(pg).To(Equal(-1))
			})

			It("should return error for colon only", func() {
				pgTag, pg, err := peerGroupFromString(":")

				Expect(err).To(HaveOccurred())
				Expect(pgTag).To(Equal(""))
				Expect(pg).To(Equal(-1))
			})
		})
	})

	Describe("unescapeString", func() {
		Context("with unescaped strings", func() {
			It("should return the same string", func() {
				str := "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"

				result := unescapeString(str)
				Expect(result).To(Equal(str))
			})

			It("should handle empty string", func() {
				result := unescapeString("")
				Expect(result).To(Equal(""))
			})

			It("should handle string without backslashes", func() {
				str := "simple/path/without/escapes"

				result := unescapeString(str)
				Expect(result).To(Equal(str))
			})
		})

		Context("with escaped strings", func() {
			It("should unescape octal sequence \\040 to space", func() {
				str := "\\040"

				result := unescapeString(str)
				Expect(result).To(Equal(" "))
			})

			It("should unescape octal sequence \\123 to 'S'", func() {
				str := "\\123abc"

				result := unescapeString(str)
				Expect(result).To(Equal("Sabc"))
			})

			It("should unescape multiple octal sequences", func() {
				str := "path\\040with\\040spaces"

				result := unescapeString(str)
				Expect(result).To(Equal("path with spaces"))
			})

			It("should unescape tab character \\011", func() {
				str := "path\\011with\\011tabs"

				result := unescapeString(str)
				Expect(result).To(Equal("path\twith\ttabs"))
			})

			It("should unescape newline character \\012", func() {
				str := "line1\\012line2"

				result := unescapeString(str)
				Expect(result).To(Equal("line1\nline2"))
			})

			It("should handle backslash at end of string", func() {
				str := "path\\"

				result := unescapeString(str)
				Expect(result).To(Equal("path\\"))
			})

			It("should handle incomplete octal sequence", func() {
				str := "path\\04"

				result := unescapeString(str)
				Expect(result).To(Equal("path\\04"))
			})

			It("should handle mixed escaped and unescaped content", func() {
				str := "/mnt/path\\040with\\040spaces/and/normal"

				result := unescapeString(str)
				Expect(result).To(Equal("/mnt/path with spaces/and/normal"))
			})

			It("should handle consecutive backslashes", func() {
				str := "path\\\\test"

				result := unescapeString(str)
				Expect(result).To(Equal("path\\\\test"))
			})

			It("should unescape \\000 to null byte", func() {
				str := "test\\000data"

				result := unescapeString(str)
				Expect(result).To(Equal("test\x00data"))
			})

			It("should unescape \\177 to DEL character", func() {
				str := "test\\177data"

				result := unescapeString(str)
				Expect(result).To(Equal("test\x7fdata"))
			})
		})
	})
})
