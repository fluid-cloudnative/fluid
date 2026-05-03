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
	"context"
	stderrors "errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlconfig "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type registerTestManager struct {
	manager.Manager
	addErr      error
	addedRunner manager.Runnable
	client      client.Client
	apiReader   client.Reader
}

func (m *registerTestManager) Add(runnable manager.Runnable) error {
	m.addedRunner = runnable
	return m.addErr
}

func (m *registerTestManager) Elected() <-chan struct{} {
	return nil
}

func (m *registerTestManager) AddMetricsServerExtraHandler(path string, handler http.Handler) error {
	return nil
}

func (m *registerTestManager) AddHealthzCheck(name string, check healthz.Checker) error {
	return nil
}

func (m *registerTestManager) AddReadyzCheck(name string, check healthz.Checker) error {
	return nil
}

func (m *registerTestManager) Start(ctx context.Context) error {
	return nil
}

func (m *registerTestManager) GetConfig() *rest.Config {
	return nil
}

func (m *registerTestManager) GetScheme() *runtime.Scheme {
	return nil
}

func (m *registerTestManager) GetClient() client.Client {
	return m.client
}

func (m *registerTestManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (m *registerTestManager) GetCache() cache.Cache {
	return nil
}

func (m *registerTestManager) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

func (m *registerTestManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (m *registerTestManager) GetAPIReader() client.Reader {
	return m.apiReader
}

func (m *registerTestManager) GetWebhookServer() webhook.Server {
	return nil
}

func (m *registerTestManager) GetLogger() logr.Logger {
	return logr.Discard()
}

func (m *registerTestManager) GetControllerOptions() ctrlconfig.Controller {
	return ctrlconfig.Controller{}
}

var _ = Describe("Register", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "plugins-register-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		}
	})

	Describe("getNodeAuthorizedClientFromKubeletConfig", func() {
		It("should return no client when the kubelet config file does not exist", func() {
			clientset, err := getNodeAuthorizedClientFromKubeletConfig(filepath.Join(tempDir, "missing-kubelet.conf"))

			Expect(err).NotTo(HaveOccurred())
			Expect(clientset).To(BeNil())
		})

		It("should return stat errors for invalid paths", func() {
			clientset, err := getNodeAuthorizedClientFromKubeletConfig(string([]byte{'\x00'}))

			Expect(err).To(HaveOccurred())
			Expect(clientset).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("fail to stat kubelet config file"))
		})
	})

	Describe("Register", func() {
		var (
			mgr *registerTestManager
			ctx config.RunningContext
		)

		BeforeEach(func() {
			mgr = &registerTestManager{}
			ctx = config.RunningContext{
				Config: config.Config{
					NodeId:            "test-node",
					Endpoint:          "unix://" + filepath.Join(tempDir, "csi.sock"),
					KubeletConfigPath: filepath.Join(tempDir, "missing-kubelet.conf"),
				},
				VolumeLocks: utils.NewVolumeLocks(),
			}
		})

		It("should add the constructed driver to the manager", func() {
			err := Register(mgr, ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(mgr.addedRunner).To(BeAssignableToTypeOf(&driver{}))

			addedDriver := mgr.addedRunner.(*driver)
			Expect(addedDriver.nodeId).To(Equal(ctx.NodeId))
			Expect(addedDriver.endpoint).To(Equal(ctx.Endpoint))
			Expect(addedDriver.nodeAuthorizedClient).To(BeNil())
			Expect(addedDriver.locks).To(BeIdenticalTo(ctx.VolumeLocks))
		})

		It("should return manager add errors", func() {
			mgr.addErr = stderrors.New("add failed")

			err := Register(mgr, ctx)

			Expect(err).To(MatchError("add failed"))
		})

		It("should return kubelet client initialization errors", func() {
			invalidKubeletConfigPath := filepath.Join(tempDir, "kubelet.conf")
			Expect(os.WriteFile(invalidKubeletConfigPath, []byte("not-a-kubeconfig"), 0o644)).To(Succeed())
			ctx.KubeletConfigPath = invalidKubeletConfigPath

			err := Register(mgr, ctx)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fail to build kubelet config"))
		})
	})

	Describe("Enabled", func() {
		It("should always enable the CSI plugin", func() {
			Expect(Enabled()).To(BeTrue())
		})
	})
})
