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

package csi

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/fluid-cloudnative/fluid/pkg/csi/config"
)

func TestCSI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CSI Setup Suite")
}

var _ = Describe("SetupWithManager", func() {
	var (
		mockMgr          manager.Manager
		ctx              config.RunningContext
		originalRegs     map[string]registrationFuncs
		enabledCalled    bool
		registerCalled   bool
		registerError    error
	)

	BeforeEach(func() {
		// Save original registrations
		originalRegs = registraions

		// Reset test flags
		enabledCalled = false
		registerCalled = false
		registerError = nil

		// Create test context
		ctx = config.RunningContext{}
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
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
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
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
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
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
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
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
						component1Reg = true
						return nil
					},
				},
				"component-2": {
					enabled: func() bool {
						component2Enabled = true
						return true
					},
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
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
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
						enabledReg = true
						return nil
					},
				},
				"disabled-component": {
					enabled: func() bool { return false },
					register: func(mgr manager.Manager, ctx config.RunningContext) error {
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
		Expect(registraions).To(HaveKey("plugins"))
		Expect(registraions).To(HaveKey("recover"))
		Expect(registraions).To(HaveKey("updatedbconf"))
	})

	It("should have valid enabled and register functions for all components", func() {
		for name, funcs := range registraions {
			Expect(funcs.enabled).NotTo(BeNil(), "enabled function should not be nil for "+name)
			Expect(funcs.register).NotTo(BeNil(), "register function should not be nil for "+name)
		}
	})
})