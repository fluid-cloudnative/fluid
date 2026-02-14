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
	"errors"
	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// createTestRunningContext creates a RunningContext for testing
func createTestRunningContext(nodeID, endpoint, kubeletConfigPath string) config.RunningContext {
	return config.RunningContext{
		Config: config.Config{
			NodeId:            nodeID,
			Endpoint:          endpoint,
			KubeletConfigPath: kubeletConfigPath,
		},
	}
}

// mockManager is a mock implementation of manager.Manager for testing
type mockManager struct {
	manager.Manager
	addFunc      func(manager.Runnable) error
	client       client.Client
	apiReader    client.Reader
	addCallCount int
	lastRunnable manager.Runnable
}

func (m *mockManager) Add(runnable manager.Runnable) error {
	m.addCallCount++
	m.lastRunnable = runnable
	if m.addFunc != nil {
		return m.addFunc(runnable)
	}
	return nil
}

func (m *mockManager) GetClient() client.Client {
	return m.client
}

func (m *mockManager) GetAPIReader() client.Reader {
	return m.apiReader
}

func (m *mockManager) Start(ctx context.Context) error {
	return nil
}

var _ = Describe("getNodeAuthorizedClientFromKubeletConfig", func() {
	Context("when kubelet config file does not exist", func() {
		It("should return nil client without error", func() {
			nonExistentPath := "/tmp/non-existent-kubelet-config-test-file-" + GinkgoT().Name()

			client, err := getNodeAuthorizedClientFromKubeletConfig(nonExistentPath)

			Expect(err).To(BeNil())
			Expect(client).To(BeNil())
		})
	})

	Context("table-driven tests", func() {
		DescribeTable("various file scenarios",
			func(setupFunc func() string, expectedError bool, expectedClient bool) {
				path := setupFunc()

				client, err := getNodeAuthorizedClientFromKubeletConfig(path)

				if expectedError {
					Expect(err).ToNot(BeNil())
				} else {
					Expect(err).To(BeNil())
				}

				if expectedClient {
					Expect(client).ToNot(BeNil())
				} else {
					Expect(client).To(BeNil())
				}
			},
			Entry("non-existent file", func() string {
				return "/tmp/non-existent-file-" + GinkgoT().Name()
			}, false, false),
			Entry("empty file exists", func() string {
				tempDir := GinkgoT().TempDir()
				path := filepath.Join(tempDir, "kubelet-config.yaml")
				err := os.WriteFile(path, []byte(""), 0644)
				Expect(err).To(BeNil())
				return path
			}, true, false),
			Entry("directory instead of file", func() string {
				tempDir := GinkgoT().TempDir()
				dirPath := filepath.Join(tempDir, "kubelet-dir")
				err := os.Mkdir(dirPath, 0755)
				Expect(err).To(BeNil())
				return dirPath
			}, true, false),
		)
	})
})

var _ = Describe("Register", func() {
	var (
		mockMgr *mockManager
		ctx     config.RunningContext
	)

	BeforeEach(func() {
		// Create a RunningContext with valid test data
		ctx = createTestRunningContext(
			"test-node-id",
			"unix:///tmp/test-csi.sock",
			"/tmp/non-existent-kubelet-config",
		)
	})

	Context("when registration is successful", func() {
		BeforeEach(func() {
			mockMgr = &mockManager{
				addFunc: func(r manager.Runnable) error {
					return nil
				},
			}
		})

		It("should register the CSI driver without error", func() {
			err := Register(mockMgr, ctx)

			Expect(err).To(BeNil())
			Expect(mockMgr.addCallCount).To(Equal(1))
			Expect(mockMgr.lastRunnable).ToNot(BeNil())
		})
	})

	Context("when Add fails", func() {
		BeforeEach(func() {
			expectedErr := errors.New("failed to add driver")
			mockMgr = &mockManager{
				addFunc: func(r manager.Runnable) error {
					return expectedErr
				},
			}
		})

		It("should return an error", func() {
			err := Register(mockMgr, ctx)

			Expect(err).ToNot(BeNil())
			Expect(mockMgr.addCallCount).To(Equal(1))
		})
	})
})

var _ = Describe("Enabled", func() {
	It("should always return true", func() {
		result := Enabled()
		Expect(result).To(BeTrue())
	})
})
