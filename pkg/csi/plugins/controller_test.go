/*
Copyright 2022 The Fluid Authors.

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
	"context"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ = Describe("Controller", func() {
	var (
		ctx context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("ControllerGetVolume", func() {
		It("should return Unimplemented error", func() {
			cs := &controllerServer{}
			req := &csi.ControllerGetVolumeRequest{
				VolumeId: "test-volume",
			}

			resp, err := cs.ControllerGetVolume(ctx, req)

			Expect(resp).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
		})
	})

	Describe("CreateVolume", func() {
		var (
			driver *csicommon.CSIDriver
			cs     *controllerServer
		)

		When("volume creation is successful", func() {
			BeforeEach(func() {
				driver = csicommon.NewCSIDriver("test-driver", "1.0.0", "test-node")
				driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
					csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				})
				cs = &controllerServer{
					DefaultControllerServer: csicommon.NewDefaultControllerServer(driver),
				}
			})

			It("should create volume with valid request", func() {
				req := &csi.CreateVolumeRequest{
					Name: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: 1073741824, // 1GB
					},
					Parameters: map[string]string{
						"param1": "value1",
					},
				}

				resp, err := cs.CreateVolume(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Volume).NotTo(BeNil())
				Expect(resp.Volume.VolumeId).To(Equal("test-volume"))
				Expect(resp.Volume.CapacityBytes).To(Equal(req.CapacityRange.RequiredBytes))
				Expect(resp.Volume.VolumeContext["param1"]).To(Equal("value1"))
			})

			It("should handle long volume name requiring sanitization", func() {
				req := &csi.CreateVolumeRequest{
					Name: strings.Repeat("a", 70), // Name longer than 63 characters
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: 1073741824,
					},
				}

				resp, err := cs.CreateVolume(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Volume).NotTo(BeNil())
				// SHA1 hash will be 40 characters
				Expect(resp.Volume.VolumeId).To(Equal("ed6c69d9e8b4373af86303dfaa3528dfbc129902"))
			})

			It("should convert uppercase volume name to lowercase", func() {
				req := &csi.CreateVolumeRequest{
					Name: "TEST-VOLUME",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: 1073741824,
					},
				}

				resp, err := cs.CreateVolume(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Volume).NotTo(BeNil())
				Expect(resp.Volume.VolumeId).To(Equal("test-volume"))
			})
		})

		When("volume creation fails", func() {
			BeforeEach(func() {
				driver = csicommon.NewCSIDriver("test-driver", "1.0.0", "test-node")
				driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
					csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				})
				cs = &controllerServer{
					DefaultControllerServer: csicommon.NewDefaultControllerServer(driver),
				}
			})

			It("should return error for empty volume name", func() {
				req := &csi.CreateVolumeRequest{
					Name: "",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
				}

				resp, err := cs.CreateVolume(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})

			It("should return error for missing volume capabilities", func() {
				req := &csi.CreateVolumeRequest{
					Name:               "test-volume",
					VolumeCapabilities: nil,
				}

				resp, err := cs.CreateVolume(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})

			It("should return error for invalid controller service request", func() {
				// Create driver without CREATE_DELETE_VOLUME capability
				driverWithoutCap := csicommon.NewCSIDriver("test-driver", "1.0.0", "test-node")
				csWithoutCap := &controllerServer{
					DefaultControllerServer: csicommon.NewDefaultControllerServer(driverWithoutCap),
				}

				req := &csi.CreateVolumeRequest{
					Name: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
				}

				resp, err := csWithoutCap.CreateVolume(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})
		})
	})

	Describe("DeleteVolume", func() {
		var (
			driver *csicommon.CSIDriver
			cs     *controllerServer
		)

		When("volume deletion is successful", func() {
			BeforeEach(func() {
				driver = csicommon.NewCSIDriver("test-driver", "1.0.0", "test-node")
				driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
					csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				})
				cs = &controllerServer{
					DefaultControllerServer: csicommon.NewDefaultControllerServer(driver),
				}
			})

			It("should delete volume with valid request", func() {
				req := &csi.DeleteVolumeRequest{
					VolumeId: "test-volume",
				}

				resp, err := cs.DeleteVolume(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
			})
		})

		When("volume deletion fails", func() {
			BeforeEach(func() {
				driver = csicommon.NewCSIDriver("test-driver", "1.0.0", "test-node")
				driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
					csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				})
				cs = &controllerServer{
					DefaultControllerServer: csicommon.NewDefaultControllerServer(driver),
				}
			})

			It("should return error for empty volume ID", func() {
				req := &csi.DeleteVolumeRequest{
					VolumeId: "",
				}

				resp, err := cs.DeleteVolume(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})

			It("should return error for invalid controller service request", func() {
				// Create driver without CREATE_DELETE_VOLUME capability
				driverWithoutCap := csicommon.NewCSIDriver("test-driver", "1.0.0", "test-node")
				csWithoutCap := &controllerServer{
					DefaultControllerServer: csicommon.NewDefaultControllerServer(driverWithoutCap),
				}

				req := &csi.DeleteVolumeRequest{
					VolumeId: "test-volume",
				}

				resp, err := csWithoutCap.DeleteVolume(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})
		})
	})

	Describe("ValidateVolumeCapabilities", func() {
		var cs *controllerServer

		BeforeEach(func() {
			cs = &controllerServer{}
		})

		When("capabilities are valid", func() {
			It("should confirm valid single node writer capability", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Confirmed).NotTo(BeNil())
			})
		})

		When("capabilities are invalid", func() {
			It("should return error for empty volume ID", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId: "",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
					},
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})

			It("should return error for missing volume capabilities", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId:           "test-volume",
					VolumeCapabilities: nil,
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).To(HaveOccurred())
				Expect(status.Code(err)).To(Equal(codes.InvalidArgument))
				Expect(resp).To(BeNil())
			})
		})

		When("capabilities are unsupported", func() {
			It("should reject multi node reader only", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
							},
						},
					},
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Message).To(Equal("Only single node writer is supported"))
				Expect(resp.Confirmed).To(BeNil())
			})

			It("should reject multi node multi writer", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
							},
						},
					},
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Message).To(Equal("Only single node writer is supported"))
				Expect(resp.Confirmed).To(BeNil())
			})

			It("should reject single node reader only", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY,
							},
						},
					},
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Message).To(Equal("Only single node writer is supported"))
				Expect(resp.Confirmed).To(BeNil())
			})

			It("should reject multiple capabilities with unsupported mode", func() {
				req := &csi.ValidateVolumeCapabilitiesRequest{
					VolumeId: "test-volume",
					VolumeCapabilities: []*csi.VolumeCapability{
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
							},
						},
						{
							AccessMode: &csi.VolumeCapability_AccessMode{
								Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
							},
						},
					},
				}

				resp, err := cs.ValidateVolumeCapabilities(ctx, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Message).To(Equal("Only single node writer is supported"))
				Expect(resp.Confirmed).To(BeNil())
			})
		})
	})

	Describe("ControllerExpandVolume", func() {
		It("should return Unimplemented error", func() {
			cs := &controllerServer{}
			req := &csi.ControllerExpandVolumeRequest{
				VolumeId: "test-volume",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 2147483648, // 2GB
				},
			}

			resp, err := cs.ControllerExpandVolume(ctx, req)

			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.Unimplemented))
			Expect(resp).NotTo(BeNil())
		})
	})

	Describe("SanitizeVolumeID", func() {
		DescribeTable("sanitizing volume IDs",
			func(input string, expected string) {
				result := sanitizeVolumeID(input)
				Expect(result).To(Equal(expected))
				Expect(len(result)).To(BeNumerically("<=", 63))
				Expect(result).To(Equal(strings.ToLower(result)))
			},
			Entry("lowercase conversion", "TEST-Volume", "test-volume"),
			Entry("short name no hash", "short", "short"),
			Entry("exactly 63 characters", strings.Repeat("a", 63), strings.Repeat("a", 63)),
			Entry("64 characters triggers hash", strings.Repeat("a", 64), "0098ba824b5c16427bd7a1122a5a442a25ec644d"),
			Entry("100 characters triggers hash", strings.Repeat("B", 100), "dd3b750e3862ec12c332c6d7c11150288471d0c9"),
			Entry("special characters", "Volume-With-Dashes-And-Numbers-123", "volume-with-dashes-and-numbers-123"),
			Entry("empty string", "", ""),
			Entry("mixed case long string", "ThisIsAVeryLongVolumeNameThatExceedsSixtyThreeCharactersInLength", "cfe80810fef3112acc6c4c8b90e032204766d8b0"),
		)
	})
})
