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
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MountInfo", func() {
	var (
		peerGroup1           map[int]bool
		peerGroup2           map[int]bool
		mockGlobalMount      *Mount
		mockBindMount        *Mount
		mockBindSubPathMount *Mount
		mockMountPoints      map[string]*Mount
	)

	BeforeEach(func() {
		peerGroup1 = map[int]bool{475: true}
		peerGroup2 = map[int]bool{476: true}

		mockGlobalMount = &Mount{
			Subtree:        "/",
			MountPath:      "/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse",
			FilesystemType: "fuse.juicefs",
			PeerGroups:     peerGroup2,
			ReadOnly:       false,
		}

		mockBindMount = &Mount{
			Subtree:        "/",
			MountPath:      "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
			FilesystemType: "fuse.juicefs",
			PeerGroups:     peerGroup1,
			ReadOnly:       false,
		}

		mockBindSubPathMount = &Mount{
			Subtree:        "/",
			MountPath:      "/var/lib/kubelet/pods/6fe8418f-3f78-4adb-9e02-416d8601c1b6/volume-subpaths/default-jfsdemo/demo/0",
			FilesystemType: "fuse.juicefs",
			PeerGroups:     peerGroup1,
			ReadOnly:       false,
		}

		mockMountPoints = map[string]*Mount{
			"/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse":                                                          mockGlobalMount,
			"/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount": mockBindMount,
			"/var/lib/kubelet/pods/6fe8418f-3f78-4adb-9e02-416d8601c1b6/volume-subpaths/default-jfsdemo/demo/0":          mockBindSubPathMount,
		}
	})

	Describe("getBindMounts", func() {
		Context("when processing valid mount points", func() {
			It("should correctly identify bind mounts and subpath mounts", func() {
				bindMountByName := getBindMounts(mockMountPoints)

				Expect(bindMountByName).To(HaveKey("default-jfsdemo"))
				Expect(bindMountByName["default-jfsdemo"]).To(HaveLen(2))
				Expect(bindMountByName["default-jfsdemo"]).To(ContainElement(mockBindMount))
				Expect(bindMountByName["default-jfsdemo"]).To(ContainElement(mockBindSubPathMount))
			})

			It("should handle multiple datasets with different namespaces", func() {
				anotherBindMount := &Mount{
					Subtree:        "/",
					MountPath:      "/var/lib/kubelet/pods/test-uid/volumes/kubernetes.io~csi/prod-dataset/mount",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     peerGroup1,
					ReadOnly:       false,
				}

				mountPoints := map[string]*Mount{
					"/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount": mockBindMount,
					"/var/lib/kubelet/pods/test-uid/volumes/kubernetes.io~csi/prod-dataset/mount":                                anotherBindMount,
				}

				bindMountByName := getBindMounts(mountPoints)

				Expect(bindMountByName).To(HaveLen(2))
				Expect(bindMountByName).To(HaveKey("default-jfsdemo"))
				Expect(bindMountByName).To(HaveKey("prod-dataset"))
			})
		})

		Context("when mount path doesn't contain kubernetes.io~csi or volume-subpaths", func() {
			It("should return empty map for bind mounts", func() {
				mountPoints := map[string]*Mount{
					"kubernetes.io~csi/mount": {
						Subtree:        "/",
						MountPath:      "kubernetes.io~csi/mount",
						FilesystemType: "ext4",
						PeerGroups:     nil,
						ReadOnly:       false,
						Count:          0,
					},
				}

				bindMountByName := getBindMounts(mountPoints)

				Expect(bindMountByName).To(BeEmpty())
			})
		})

		Context("when mount path has insufficient fields", func() {
			It("should skip mounts with less than 3 fields for csi mounts", func() {
				mountPoints := map[string]*Mount{
					"kubernetes.io~csi/mount": {
						Subtree:        "/",
						MountPath:      "kubernetes.io~csi/mount",
						FilesystemType: "ext4",
						PeerGroups:     peerGroup1,
						ReadOnly:       false,
					},
				}

				bindMountByName := getBindMounts(mountPoints)

				Expect(bindMountByName).To(BeEmpty())
			})

			It("should skip subpath mounts with less than 4 fields", func() {
				mountPoints := map[string]*Mount{
					"/volume-subpaths/test": {
						Subtree:        "/",
						MountPath:      "/volume-subpaths/test",
						FilesystemType: "ext4",
						PeerGroups:     peerGroup1,
						ReadOnly:       false,
					},
				}

				bindMountByName := getBindMounts(mountPoints)

				Expect(bindMountByName).To(BeEmpty())
			})
		})

		Context("when processing empty mount map", func() {
			It("should return empty bind mount map", func() {
				bindMountByName := getBindMounts(map[string]*Mount{})

				Expect(bindMountByName).To(BeEmpty())
			})
		})

		Context("when processing subpath mounts", func() {
			It("should correctly extract namespace-dataset from subpath", func() {
				subPathMount := &Mount{
					Subtree:        "/data",
					MountPath:      "/var/lib/kubelet/pods/pod-123/volume-subpaths/staging-mydata/app/0",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     peerGroup1,
					ReadOnly:       false,
				}

				mountPoints := map[string]*Mount{
					subPathMount.MountPath: subPathMount,
				}

				bindMountByName := getBindMounts(mountPoints)

				Expect(bindMountByName).To(HaveKey("staging-mydata"))
				Expect(bindMountByName["staging-mydata"]).To(ContainElement(subPathMount))
			})
		})
	})

	Describe("getBrokenBindMounts", func() {
		Context("when bind mounts have different peer groups than global mount", func() {
			It("should identify them as broken mounts", func() {
				globalMountByName := map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				}
				bindMountByName := map[string][]*Mount{
					"default-jfsdemo": {mockBindMount, mockBindSubPathMount},
				}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(HaveLen(2))
				Expect(brokenMounts[0].SourcePath).To(Equal("/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"))
				Expect(brokenMounts[0].MountPath).To(Equal(mockBindMount.MountPath))
				Expect(brokenMounts[0].FilesystemType).To(Equal("fuse.juicefs"))
				Expect(brokenMounts[0].NamespacedDatasetName).To(Equal("default-jfsdemo"))
				Expect(brokenMounts[0].ReadOnly).To(BeFalse())

				Expect(brokenMounts[1].SourcePath).To(Equal("/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse"))
				Expect(brokenMounts[1].MountPath).To(Equal(mockBindSubPathMount.MountPath))
			})
		})

		Context("when bind mounts share same peer groups with global mount", func() {
			It("should not identify them as broken", func() {
				// Make bind mount share same peer group as global mount
				healthyBindMount := &Mount{
					Subtree:        "/",
					MountPath:      "/var/lib/kubelet/pods/1140aa96-18c2-4896-a14f-7e3965a51406/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     peerGroup2, // Same as global mount
					ReadOnly:       false,
				}

				globalMountByName := map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				}
				bindMountByName := map[string][]*Mount{
					"default-jfsdemo": {healthyBindMount},
				}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(BeEmpty())
			})
		})

		Context("when global mount doesn't exist", func() {
			It("should ignore bind mounts without global mount", func() {
				globalMountByName := map[string]*Mount{}
				bindMountByName := map[string][]*Mount{
					"default-jfsdemo": {mockBindMount, mockBindSubPathMount},
				}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(BeEmpty())
			})
		})

		Context("when bind mount list is empty", func() {
			It("should return empty broken mounts", func() {
				globalMountByName := map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				}
				bindMountByName := map[string][]*Mount{}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(BeEmpty())
			})
		})

		Context("when handling subtree paths", func() {
			It("should correctly join subtree with global mount path", func() {
				bindMountWithSubtree := &Mount{
					Subtree:        "/subfolder",
					MountPath:      "/var/lib/kubelet/pods/test/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     peerGroup1, // Different from global
					ReadOnly:       false,
				}

				globalMountByName := map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				}
				bindMountByName := map[string][]*Mount{
					"default-jfsdemo": {bindMountWithSubtree},
				}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(HaveLen(1))
				Expect(brokenMounts[0].SourcePath).To(Equal("/runtime-mnt/juicefs/default/jfsdemo/juicefs-fuse/subfolder"))
			})
		})

		Context("when bind mount has readonly flag", func() {
			It("should preserve readonly flag in broken mount point", func() {
				readOnlyBindMount := &Mount{
					Subtree:        "/",
					MountPath:      "/var/lib/kubelet/pods/test/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     peerGroup1,
					ReadOnly:       true,
				}

				globalMountByName := map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				}
				bindMountByName := map[string][]*Mount{
					"default-jfsdemo": {readOnlyBindMount},
				}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(HaveLen(1))
				Expect(brokenMounts[0].ReadOnly).To(BeTrue())
			})
		})

		Context("when processing multiple broken mounts for same dataset", func() {
			It("should return all broken mounts", func() {
				anotherBrokenMount := &Mount{
					Subtree:        "/",
					MountPath:      "/var/lib/kubelet/pods/another-pod/volumes/kubernetes.io~csi/default-jfsdemo/mount",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     map[int]bool{999: true}, // Different peer group
					ReadOnly:       false,
				}

				globalMountByName := map[string]*Mount{
					"default-jfsdemo": mockGlobalMount,
				}
				bindMountByName := map[string][]*Mount{
					"default-jfsdemo": {mockBindMount, anotherBrokenMount},
				}

				brokenMounts := getBrokenBindMounts(globalMountByName, bindMountByName)

				Expect(brokenMounts).To(HaveLen(2))
			})
		})
	})

	Describe("getGlobalMounts", func() {
		BeforeEach(func() {
			GinkgoT().Setenv(utils.MountRoot, "/runtime-mnt")
		})

		Context("when processing valid global mounts", func() {
			It("should correctly identify global mounts", func() {
				globalMountByName, err := getGlobalMounts(mockMountPoints)

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(HaveLen(1))
				Expect(globalMountByName).To(HaveKey("default-jfsdemo"))
				Expect(globalMountByName["default-jfsdemo"]).To(Equal(mockGlobalMount))
			})

			It("should handle multiple global mounts", func() {
				anotherGlobalMount := &Mount{
					Subtree:        "/",
					MountPath:      "/runtime-mnt/alluxio/staging/data-cache/alluxio-fuse",
					FilesystemType: "fuse.alluxio",
					PeerGroups:     peerGroup2,
					ReadOnly:       false,
				}

				mountPoints := map[string]*Mount{
					mockGlobalMount.MountPath:    mockGlobalMount,
					anotherGlobalMount.MountPath: anotherGlobalMount,
				}

				globalMountByName, err := getGlobalMounts(mountPoints)

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(HaveLen(2))
				Expect(globalMountByName).To(HaveKey("default-jfsdemo"))
				Expect(globalMountByName).To(HaveKey("staging-data-cache"))
			})
		})

		Context("when mount path doesn't contain mount root", func() {
			It("should not include it in global mounts", func() {
				mountPoints := map[string]*Mount{
					"/other/path/mount": {
						Subtree:        "/",
						MountPath:      "/other/path/mount",
						FilesystemType: "fuse.juicefs",
						PeerGroups:     peerGroup2,
						ReadOnly:       false,
					},
				}

				globalMountByName, err := getGlobalMounts(mountPoints)

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(BeEmpty())
			})
		})

		Context("when mount path has insufficient fields", func() {
			It("should skip mounts with less than 6 path components", func() {
				mountPoints := map[string]*Mount{
					"/runtime-mnt/test": {
						Subtree:        "/",
						MountPath:      "/runtime-mnt/test",
						FilesystemType: "fuse.juicefs",
						PeerGroups:     peerGroup2,
						ReadOnly:       false,
					},
					"/runtime-mnt/juicefs/default": {
						Subtree:        "/",
						MountPath:      "/runtime-mnt/juicefs/default",
						FilesystemType: "fuse.juicefs",
						PeerGroups:     peerGroup2,
						ReadOnly:       false,
					},
				}

				globalMountByName, err := getGlobalMounts(mountPoints)

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(BeEmpty())
			})
		})

		Context("when mount map is empty", func() {
			It("should return empty global mount map", func() {
				globalMountByName, err := getGlobalMounts(map[string]*Mount{})

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(BeEmpty())
			})
		})

		Context("when mount root has custom value", func() {
			It("should use the custom mount root to filter mounts", func() {
				GinkgoT().Setenv(utils.MountRoot, "/custom-mount-root")

				customGlobalMount := &Mount{
					Subtree:        "/",
					MountPath:      "/custom-mount-root/juicefs/default/jfsdemo/juicefs-fuse",
					FilesystemType: "fuse.juicefs",
					PeerGroups:     peerGroup2,
					ReadOnly:       false,
				}

				mountPoints := map[string]*Mount{
					customGlobalMount.MountPath: customGlobalMount,
					mockGlobalMount.MountPath:   mockGlobalMount, // This shouldn't match
				}

				globalMountByName, err := getGlobalMounts(mountPoints)

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(HaveLen(1))
				Expect(globalMountByName).To(HaveKey("default-jfsdemo"))
				Expect(globalMountByName["default-jfsdemo"].MountPath).To(Equal("/custom-mount-root/juicefs/default/jfsdemo/juicefs-fuse"))
			})
		})

		Context("when extracting namespace and dataset name", func() {
			It("should correctly parse namespace-dataset pairs from mount path", func() {
				testMounts := map[string]*Mount{
					"/runtime-mnt/juicefs/prod/analytics/juicefs-fuse": {
						Subtree:        "/",
						MountPath:      "/runtime-mnt/juicefs/prod/analytics/juicefs-fuse",
						FilesystemType: "fuse.juicefs",
						PeerGroups:     peerGroup2,
					},
					"/runtime-mnt/alluxio/dev/cache-data/alluxio-fuse": {
						Subtree:        "/",
						MountPath:      "/runtime-mnt/alluxio/dev/cache-data/alluxio-fuse",
						FilesystemType: "fuse.alluxio",
						PeerGroups:     peerGroup2,
					},
				}

				globalMountByName, err := getGlobalMounts(testMounts)

				Expect(err).NotTo(HaveOccurred())
				Expect(globalMountByName).To(HaveKey("prod-analytics"))
				Expect(globalMountByName).To(HaveKey("dev-cache-data"))
			})
		})
	})

	Describe("MountPoint struct", func() {
		It("should properly represent a mount point with all fields", func() {
			mp := MountPoint{
				SourcePath:            "/source/path",
				MountPath:             "/mount/path",
				FilesystemType:        "fuse.juicefs",
				ReadOnly:              true,
				Count:                 5,
				NamespacedDatasetName: "ns-dataset",
			}

			Expect(mp.SourcePath).To(Equal("/source/path"))
			Expect(mp.MountPath).To(Equal("/mount/path"))
			Expect(mp.FilesystemType).To(Equal("fuse.juicefs"))
			Expect(mp.ReadOnly).To(BeTrue())
			Expect(mp.Count).To(Equal(5))
			Expect(mp.NamespacedDatasetName).To(Equal("ns-dataset"))
		})
	})
})
