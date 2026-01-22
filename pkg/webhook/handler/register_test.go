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

package handler

import (
	"context"
	"net/http"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// mockAdmissionHandler implements common.AdmissionHandler for testing
type mockAdmissionHandler struct {
	setupCalled bool
}

func (m *mockAdmissionHandler) Setup(client client.Client, reader client.Reader, decoder *admission.Decoder) {
	m.setupCalled = true
}

func (m *mockAdmissionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
		},
	}
}

var _ = Describe("AddHandlers", func() {
	var (
		originalHandlerMap   map[string]common.AdmissionHandler
		originalHandlerGates map[string]GateFunc
	)

	BeforeEach(func() {
		originalHandlerMap = handlerMap
		originalHandlerGates = handlerGates
		handlerMap = map[string]common.AdmissionHandler{}
		handlerGates = map[string]GateFunc{}
	})

	AfterEach(func() {
		handlerMap = originalHandlerMap
		handlerGates = originalHandlerGates
	})

	Context("when adding handlers", func() {
		It("should add single handler with leading slash", func() {
			inputMap := map[string]common.AdmissionHandler{
				"/test-handler": &mockAdmissionHandler{},
			}

			addHandlers(inputMap)

			Expect(handlerMap).To(HaveKey("/test-handler"))
			Expect(handlerMap).To(HaveLen(1))
		})

		It("should add single handler without leading slash", func() {
			inputMap := map[string]common.AdmissionHandler{
				"test-handler": &mockAdmissionHandler{},
			}

			addHandlers(inputMap)

			Expect(handlerMap).To(HaveKey("/test-handler"))
			Expect(handlerMap).To(HaveLen(1))
		})

		It("should add multiple handlers", func() {
			inputMap := map[string]common.AdmissionHandler{
				"/handler1": &mockAdmissionHandler{},
				"handler2":  &mockAdmissionHandler{},
			}

			addHandlers(inputMap)

			Expect(handlerMap).To(HaveKey("/handler1"))
			Expect(handlerMap).To(HaveKey("/handler2"))
			Expect(handlerMap).To(HaveLen(2))
		})
	})

	Context("when adding handlers with empty path", func() {
		It("should skip empty path", func() {
			inputMap := map[string]common.AdmissionHandler{
				"": &mockAdmissionHandler{},
			}

			addHandlers(inputMap)

			Expect(handlerMap).To(BeEmpty())
		})
	})

	Context("when adding handlers with conflicting path", func() {
		It("should panic", func() {
			handlerMap = map[string]common.AdmissionHandler{
				"/existing": &mockAdmissionHandler{},
			}

			inputMap := map[string]common.AdmissionHandler{
				"/existing": &mockAdmissionHandler{},
			}

			Expect(func() {
				addHandlers(inputMap)
			}).To(Panic())
		})
	})
})

var _ = Describe("AddHandlersWithGate", func() {
	var (
		originalHandlerMap   map[string]common.AdmissionHandler
		originalHandlerGates map[string]GateFunc
	)

	BeforeEach(func() {
		originalHandlerMap = handlerMap
		originalHandlerGates = handlerGates
		handlerMap = map[string]common.AdmissionHandler{}
		handlerGates = map[string]GateFunc{}
	})

	AfterEach(func() {
		handlerMap = originalHandlerMap
		handlerGates = originalHandlerGates
	})

	Context("when adding handler with gate", func() {
		It("should add handler and gate function", func() {
			inputMap := map[string]common.AdmissionHandler{
				"/gated-handler": &mockAdmissionHandler{},
			}
			gateFunc := func() bool {
				return true
			}

			addHandlersWithGate(inputMap, gateFunc)

			Expect(handlerMap).To(HaveKey("/gated-handler"))
			Expect(handlerGates).To(HaveKey("/gated-handler"))
		})
	})

	Context("when adding handler without gate", func() {
		It("should add handler without gate function", func() {
			inputMap := map[string]common.AdmissionHandler{
				"/no-gate-handler": &mockAdmissionHandler{},
			}

			addHandlersWithGate(inputMap, nil)

			Expect(handlerMap).To(HaveKey("/no-gate-handler"))
			Expect(handlerGates).To(BeEmpty())
		})
	})
})

var _ = Describe("FilterActiveHandlers", func() {
	var (
		originalHandlerMap   map[string]common.AdmissionHandler
		originalHandlerGates map[string]GateFunc
	)

	BeforeEach(func() {
		originalHandlerMap = handlerMap
		originalHandlerGates = handlerGates
	})

	AfterEach(func() {
		handlerMap = originalHandlerMap
		handlerGates = originalHandlerGates
	})

	Context("when filtering disabled handlers", func() {
		It("should keep only enabled handlers", func() {
			handlerMap = map[string]common.AdmissionHandler{
				"/enabled":  &mockAdmissionHandler{},
				"/disabled": &mockAdmissionHandler{},
			}
			handlerGates = map[string]GateFunc{
				"/enabled": func() bool {
					return true
				},
				"/disabled": func() bool {
					return false
				},
			}

			filterActiveHandlers()

			Expect(handlerMap).To(HaveKey("/enabled"))
			Expect(handlerMap).NotTo(HaveKey("/disabled"))
			Expect(handlerMap).To(HaveLen(1))
		})
	})

	Context("when no gates present", func() {
		It("should keep all handlers", func() {
			handlerMap = map[string]common.AdmissionHandler{
				"/handler1": &mockAdmissionHandler{},
				"/handler2": &mockAdmissionHandler{},
			}
			handlerGates = map[string]GateFunc{}

			filterActiveHandlers()

			Expect(handlerMap).To(HaveKey("/handler1"))
			Expect(handlerMap).To(HaveKey("/handler2"))
			Expect(handlerMap).To(HaveLen(2))
		})
	})

	Context("when all handlers disabled", func() {
		It("should remove all handlers", func() {
			handlerMap = map[string]common.AdmissionHandler{
				"/disabled1": &mockAdmissionHandler{},
				"/disabled2": &mockAdmissionHandler{},
			}
			handlerGates = map[string]GateFunc{
				"/disabled1": func() bool {
					return false
				},
				"/disabled2": func() bool {
					return false
				},
			}

			filterActiveHandlers()

			Expect(handlerMap).To(BeEmpty())
		})
	})
})

// mockWebhookServer implements webhook.Server for testing
type mockWebhookServer struct {
	handlers map[string]http.Handler
}

func (m *mockWebhookServer) Register(path string, handler http.Handler) {
	m.handlers[path] = handler
}

func (m *mockWebhookServer) Start(ctx context.Context) error {
	return nil
}

func (m *mockWebhookServer) StartedChecker() error {
	return nil
}

func (m *mockWebhookServer) WebhookMux() *http.ServeMux {
	return nil
}

// mockManager implements manager.Manager for testing
type mockManager struct {
	client client.Client
	scheme *runtime.Scheme
	server *mockWebhookServer
}

func (m *mockManager) Add(manager.Runnable) error                                     { return nil }
func (m *mockManager) Elected() <-chan struct{}                                       { return nil }
func (m *mockManager) SetFields(interface{}) error                                    { return nil }
func (m *mockManager) AddMetricsExtraHandler(path string, handler http.Handler) error { return nil }
func (m *mockManager) AddHealthzCheck(name string, check func(req *http.Request) error) error {
	return nil
}
func (m *mockManager) AddReadyzCheck(name string, check func(req *http.Request) error) error {
	return nil
}
func (m *mockManager) Start(ctx context.Context) error             { return nil }
func (m *mockManager) GetConfig() *runtime.Scheme                  { return m.scheme }
func (m *mockManager) GetScheme() *runtime.Scheme                  { return m.scheme }
func (m *mockManager) GetClient() client.Client                    { return m.client }
func (m *mockManager) GetFieldIndexer() client.FieldIndexer        { return nil }
func (m *mockManager) GetCache() interface{}                       { return nil }
func (m *mockManager) GetEventRecorderFor(name string) interface{} { return nil }
func (m *mockManager) GetRESTMapper() interface{}                  { return nil }
func (m *mockManager) GetAPIReader() client.Reader                 { return m.client }
func (m *mockManager) GetWebhookServer() interface{}               { return m.server }
func (m *mockManager) GetLogger() logr.Logger                      { return ctrl.Log }
func (m *mockManager) GetControllerOptions() interface{}           { return nil }
