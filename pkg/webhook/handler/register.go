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
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Register registers the handlers to the manager
func Register(mgr manager.Manager, client client.Client, log logr.Logger) {
	server := mgr.GetWebhookServer()
	filterActiveHandlers()
	for path, handler := range HandlerMap {
		handler.Setup(client)
		server.Register(path, &webhook.Admission{Handler: handler})
		log.Info("Registered webhook handler", "path", path)
	}
}
