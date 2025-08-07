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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/handler/mutating"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type GateFunc func() (enabled bool)

var (
	setupLog = ctrl.Log.WithName("handler")
)

var (
	// HandlerMap contains all admission webhook handlers.
	handlerMap   = map[string]common.AdmissionHandler{}
	handlerGates = map[string]GateFunc{}
)

func init() {
	addHandlers(mutating.HandlerMap)
	// addHandlers(validating.HandlerMap)
}

// Register registers the handlers to the manager
func Register(mgr manager.Manager, log logr.Logger) {
	server := mgr.GetWebhookServer()
	filterActiveHandlers()
	for path, handler := range handlerMap {
		decoder := admission.NewDecoder(mgr.GetScheme())
		handler.Setup(mgr.GetClient(), mgr.GetAPIReader(), decoder)
		server.Register(path, &webhook.Admission{Handler: handler})
		log.Info("Registered webhook handler", "path", path)
	}
}

func addHandlers(m map[string]common.AdmissionHandler) {
	addHandlersWithGate(m, nil)
}

func addHandlersWithGate(m map[string]common.AdmissionHandler, fn GateFunc) {
	for path, handler := range m {
		if len(path) == 0 {
			setupLog.Info("Skip handler with empty path.", "handler", handler)
			continue
		}
		// Ensure path starts with '/', making it absolute.
		if path[0] != '/' {
			path = "/" + path
		}
		_, found := handlerMap[path]
		if found {
			setupLog.Info("error: conflicting webhook builder path in handler map", "path", path)
			panic("Conflicting webhook path detected")
		}
		handlerMap[path] = handler
		if fn != nil {
			handlerGates[path] = fn
		}
	}
}

func filterActiveHandlers() {
	disablePaths := sets.NewString()
	for path := range handlerMap {
		if fn, ok := handlerGates[path]; ok {
			if !fn() {
				disablePaths.Insert(path)
			}
		}
	}
	for _, path := range disablePaths.List() {
		delete(handlerMap, path)
	}
}
