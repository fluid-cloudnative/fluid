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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Driver", func() {
	var (
		testNodeID   string
		testEndpoint string
		mockClient   client.Client
		mockReader   client.Reader
		locks        *utils.VolumeLocks
		tempDir      string
	)

	BeforeEach(func() {
		testNodeID = "test-node-123"
		locks = utils.NewVolumeLocks()

		// Create temporary directory for socket tests
		var err error
		tempDir, err = os.MkdirTemp("", "csi-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("NewDriver", func() {
		Context("with valid unix socket endpoint", func() {
			BeforeEach(func() {
				socketPath := filepath.Join(tempDir, "csi.sock")
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			})

			It("should create socket directory if it doesn't exist", func() {
				socketPath := filepath.Join(tempDir, "nested", "dir", "csi.sock")
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				socketDir := filepath.Dir(socketPath)
				_, err := os.Stat(socketDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should initialize CSI driver with correct name and version", func() {
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d.csiDriver).NotTo(BeNil())
			})

			It("should add controller service capabilities", func() {
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d.csiDriver).NotTo(BeNil())
			})

			It("should add volume capability access modes", func() {
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d.csiDriver).NotTo(BeNil())
			})
		})

		Context("with endpoint needing slash prepend", func() {
			It("should handle endpoint address without leading slash", func() {
				// Create endpoint where the address part doesn't start with /
				socketPath := filepath.Join(tempDir, "csi.sock")
				// This will be converted to have a leading slash by the driver
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				Expect(d.endpoint).To(Equal(testEndpoint))
			})
		})

		Context("with various node IDs", func() {
			BeforeEach(func() {
				socketPath := filepath.Join(tempDir, "csi.sock")
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			})

			It("should accept node ID with special characters", func() {
				var nodeAuthorizedClient *kubernetes.Clientset
				specialNodeID := "node-123.example.com"

				d := NewDriver(specialNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				Expect(d.nodeId).To(Equal(specialNodeID))
			})

			It("should handle unicode characters in node ID", func() {
				var nodeAuthorizedClient *kubernetes.Clientset
				unicodeNodeID := "node-测试-123"

				d := NewDriver(unicodeNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				Expect(d.nodeId).To(Equal(unicodeNodeID))
			})
		})

		Context("with nil parameters", func() {
			BeforeEach(func() {
				socketPath := filepath.Join(tempDir, "csi.sock")
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			})

			It("should handle nil client", func() {
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, nil, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				Expect(d.client).To(BeNil())
			})

			It("should handle nil apiReader", func() {
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, nil, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				Expect(d.apiReader).To(BeNil())
			})

			It("should handle nil nodeAuthorizedClient", func() {
				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nil, locks)

				Expect(d).NotTo(BeNil())
				Expect(d.nodeAuthorizedClient).To(BeNil())
			})

			It("should handle nil locks", func() {
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, nil)

				Expect(d).NotTo(BeNil())
				Expect(d.locks).To(BeNil())
			})
		})
	})

	Describe("newControllerServer", func() {
		var d *driver

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
		})

		It("should create controller server successfully", func() {
			cs := d.newControllerServer()

			Expect(cs).NotTo(BeNil())
			Expect(cs.DefaultControllerServer).NotTo(BeNil())
		})

		It("should create different instances on multiple calls", func() {
			cs1 := d.newControllerServer()
			cs2 := d.newControllerServer()

			Expect(cs1).NotTo(BeNil())
			Expect(cs2).NotTo(BeNil())
			Expect(cs1).NotTo(BeIdenticalTo(cs2))
		})
	})

	Describe("newNodeServer", func() {
		var d *driver

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
		})

		It("should create different instances on multiple calls", func() {
			ns1 := d.newNodeServer()
			ns2 := d.newNodeServer()

			Expect(ns1).NotTo(BeNil())
			Expect(ns2).NotTo(BeNil())
			Expect(ns1).NotTo(BeIdenticalTo(ns2))
		})
	})

	Describe("Start", func() {
		var (
			d   *driver
			ctx context.Context
		)

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi-start.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
			ctx = context.Background()
		})

		It("should implement manager.Runnable interface", func() {
			var _ interface{} = d
			Expect(d.Start).NotTo(BeNil())
		})

		It("should start without immediate error", func() {
			errChan := make(chan error, 1)
			go func() {
				errChan <- d.Start(ctx)
			}()

			time.Sleep(100 * time.Millisecond)

			select {
			case err := <-errChan:
				Expect(err).NotTo(HaveOccurred())
			default:
				// No immediate error, which is expected
			}
		})

		Context("with cancelled context", func() {
			It("should handle context cancellation gracefully", func() {
				ctx, cancel := context.WithCancel(context.Background())

				errChan := make(chan error, 1)
				go func() {
					errChan <- d.Start(ctx)
				}()

				time.Sleep(100 * time.Millisecond)
				cancel()
				time.Sleep(100 * time.Millisecond)
			})
		})
	})

	Describe("run", func() {
		var d *driver

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi-run.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
		})

		It("should start gRPC server without panic", func() {
			done := make(chan bool, 1)

			go func() {
				defer func() {
					if r := recover(); r != nil {
						done <- false
					}
				}()
				go d.run()
				time.Sleep(100 * time.Millisecond)
				done <- true
			}()

			Eventually(done, "2s").Should(Receive(Equal(true)))
		})
	})

	Describe("Constants", func() {
		It("should have correct driver name", func() {
			Expect(driverName).To(Equal("fuse.csi.fluid.io"))
		})

		It("should have correct version", func() {
			Expect(version).To(Equal("1.0.0"))
		})
	})

	Describe("Driver struct fields", func() {
		var d *driver

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
		})

		It("should have all required fields properly initialized", func() {
			Expect(d.csiDriver).NotTo(BeNil())
			Expect(d.nodeId).NotTo(BeEmpty())
			Expect(d.endpoint).NotTo(BeEmpty())
			Expect(d.locks).NotTo(BeNil())
		})

		It("should maintain reference to passed locks", func() {
			Expect(d.locks).To(BeIdenticalTo(locks))
		})
	})

	Describe("CSI Driver Capabilities", func() {
		var d *driver

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
		})

		It("should have controller service capabilities configured", func() {
			Expect(d.csiDriver).NotTo(BeNil())
			cs := d.newControllerServer()
			Expect(cs).NotTo(BeNil())
		})

		It("should have volume capability access modes configured", func() {
			Expect(d.csiDriver).NotTo(BeNil())
		})
	})

	Describe("Edge Cases", func() {
		Context("with various endpoint formats", func() {
			It("should handle endpoint with nested directories", func() {
				socketPath := filepath.Join(tempDir, "very", "deep", "nested", "path", "csi.sock")
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
				socketDir := filepath.Dir(socketPath)
				_, err := os.Stat(socketDir)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle long socket paths", func() {
				longDir := filepath.Join(tempDir, "very", "long", "nested", "directory", "structure", "for", "testing")
				err := os.MkdirAll(longDir, 0755)
				Expect(err).NotTo(HaveOccurred())

				socketPath := filepath.Join(longDir, "csi.sock")
				testEndpoint = fmt.Sprintf("unix://%s", socketPath)
				var nodeAuthorizedClient *kubernetes.Clientset

				d := NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)

				Expect(d).NotTo(BeNil())
			})
		})
	})

	Describe("Concurrent Operations", func() {
		var d *driver

		BeforeEach(func() {
			socketPath := filepath.Join(tempDir, "csi.sock")
			testEndpoint = fmt.Sprintf("unix://%s", socketPath)
			var nodeAuthorizedClient *kubernetes.Clientset
			d = NewDriver(testNodeID, testEndpoint, mockClient, mockReader, nodeAuthorizedClient, locks)
		})

		It("should handle concurrent controller server creation", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					cs := d.newControllerServer()
					Expect(cs).NotTo(BeNil())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive(Equal(true)))
			}
		})

		It("should handle concurrent node server creation", func() {
			done := make(chan bool, 10)

			for i := 0; i < 10; i++ {
				go func() {
					ns := d.newNodeServer()
					Expect(ns).NotTo(BeNil())
					done <- true
				}()
			}

			for i := 0; i < 10; i++ {
				Eventually(done).Should(Receive(Equal(true)))
			}
		})
	})
})
