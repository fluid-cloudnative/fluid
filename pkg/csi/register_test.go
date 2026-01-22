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

package csi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	fluidconfig "github.com/fluid-cloudnative/fluid/pkg/csi/config"
)

// mockManager implements manager.Manager interface for testing
type mockManager struct {
	manager.Manager
}

func (m *mockManager) Add(runnable manager.Runnable) error {
	return nil
}

func (m *mockManager) Elected() <-chan struct{} {
	return nil
}

func (m *mockManager) AddMetricsServerExtraHandler(path string, handler http.Handler) error {
	return nil
}

func (m *mockManager) AddHealthzCheck(name string, check healthz.Checker) error {
	return nil
}

func (m *mockManager) AddReadyzCheck(name string, check healthz.Checker) error {
	return nil
}

func (m *mockManager) Start(ctx context.Context) error {
	return nil
}

func (m *mockManager) GetConfig() *rest.Config {
	return nil
}

func (m *mockManager) GetScheme() *runtime.Scheme {
	return nil
}

func (m *mockManager) GetClient() client.Client {
	return nil
}

func (m *mockManager) GetFieldIndexer() client.FieldIndexer {
	return nil
}

func (m *mockManager) GetCache() cache.Cache {
	return nil
}

func (m *mockManager) GetEventRecorderFor(name string) record.EventRecorder {
	return nil
}

func (m *mockManager) GetRESTMapper() meta.RESTMapper {
	return nil
}

func (m *mockManager) GetAPIReader() client.Reader {
	return nil
}

func (m *mockManager) GetWebhookServer() webhook.Server {
	return nil
}

func (m *mockManager) GetLogger() logr.Logger {
	return logr.Discard()
}

func (m *mockManager) GetControllerOptions() config.Controller {
	return config.Controller{}
}

func TestCSI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSI Setup Suite")
}

var _ = Describe("SetupWithManager", func() {
	var (
		mockMgr          manager.Manager
		ctx              fluidconfig.RunningContext
		originalRegs     map[string]registrationFuncs
		enabledCalled    bool
		registerCalled   bool
		registerError    error
	)

	BeforeEach(func() {
		// Initialize mock manager
		mockMgr = &mockManager{}

		// Save original registrations with deep copy to prevent test pollution
		originalRegs = make(map[string]registrationFuncs, len(registraions))
		for k, v := range registraions {
			originalRegs[k] = v
		}

		// Reset test flags
		enabledCalled = false
		registerCalled = false
		registerError = nil

		// Create Test Context
		ctx = fluidconfig.RunningContext{}
	})

	AfterEach(func() {
		// Restore original registrations
		registraions = originalRegs
	})

	Context("when all components are enabled and register successfully", func() {
		BeforeEach(func() {
			registraions = map[string]registrationFuncs{
				"test-component": {
					enabled: func() bool {
						enabledCalled = true
						return true
					},
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						registerCalled = true
						return nil
					},
				},
			}
		})

		It("should call enabled and register functions", func() {
			err := SetupWithManager(mockMgr, ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(enabledCalled).To(BeTrue())
			Expect(registerCalled).To(BeTrue())
		})
	})

	Context("when a component is disabled", func() {
		BeforeEach(func() {
			registraions = map[string]registrationFuncs{
				"disabled-component": {
					enabled: func() bool {
						enabledCalled = true
						return false
					},
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						registerCalled = true
						return nil
					},
				},
			}
		})

		It("should not call register function", func() {
			err := SetupWithManager(mockMgr, ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(enabledCalled).To(BeTrue())
			Expect(registerCalled).To(BeFalse())
		})
	})

	Context("when registration fails", func() {
		BeforeEach(func() {
			registerError = errors.New("registration failed")
			registraions = map[string]registrationFuncs{
				"failing-component": {
					enabled: func() bool {
						return true
					},
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						return registerError
					},
				},
			}
		})

		It("should return the error", func() {
			err := SetupWithManager(mockMgr, ctx)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(registerError))
		})
	})

	Context("when multiple components are registered", func() {
		var (
			component1Enabled  bool
			component2Enabled  bool
			component1Reg      bool
			component2Reg      bool
		)

		BeforeEach(func() {
			component1Enabled = false
			component2Enabled = false
			component1Reg = false
			component2Reg = false

			registraions = map[string]registrationFuncs{
				"component-1": {
					enabled: func() bool {
						component1Enabled = true
						return true
					},
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						component1Reg = true
						return nil
					},
				},
				"component-2": {
					enabled: func() bool {
						component2Enabled = true
						return true
					},
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						component2Reg = true
						return nil
					},
				},
			}
		})

		It("should register all enabled components", func() {
			err := SetupWithManager(mockMgr, ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(component1Enabled).To(BeTrue())
			Expect(component2Enabled).To(BeTrue())
			Expect(component1Reg).To(BeTrue())
			Expect(component2Reg).To(BeTrue())
		})
	})

	Context("when mixing enabled and disabled components", func() {
		var (
			enabledReg  bool
			disabledReg bool
		)

		BeforeEach(func() {
			enabledReg = false
			disabledReg = false

			registraions = map[string]registrationFuncs{
				"enabled-component": {
					enabled: func() bool { return true },
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						enabledReg = true
						return nil
					},
				},
				"disabled-component": {
					enabled: func() bool { return false },
					register: func(mgr manager.Manager, ctx fluidconfig.RunningContext) error {
						disabledReg = true
						return nil
					},
				},
			}
		})

		It("should only register enabled components", func() {
			err := SetupWithManager(mockMgr, ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(enabledReg).To(BeTrue())
			Expect(disabledReg).To(BeFalse())
		})
	})

	Context("when registrations map is empty", func() {
		BeforeEach(func() {
			registraions = map[string]registrationFuncs{}
		})

		It("should return without error", func() {
			err := SetupWithManager(mockMgr, ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("init function", func() {
	It("should initialize registrations map with all components", func() {
		Expect(registraions).NotTo(BeNil())
		Expect(registraions).To(And(
			HaveLen(3),
			HaveKey("plugins"),
			HaveKey("recover"),
			HaveKey("updatedbconf"),
		))
	})

	It("should have valid enabled and register functions for all components", func() {
		for name, funcs := range registraions {
			Expect(funcs.enabled).NotTo(BeNil(), "enabled function should not be nil for "+name)
			Expect(funcs.register).NotTo(BeNil(), "register function should not be nil for "+name)
		}
	})
})