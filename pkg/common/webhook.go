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

package common

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	WebhookName            = "fluid-pod-admission-webhook"
	WebhookServiceName     = "fluid-pod-admission-webhook"
	WebhookSchedulePodPath = "mutate-fluid-io-v1alpha1-schedulepod"

	CertSecretName = "fluid-webhook-certs"

	PluginProfileConfigMapName = "webhook-plugins"
	PluginProfileKeyName       = "pluginsProfile"
)

// AdmissionHandler wrappers admission.Handler, but adding client-go capablities
type AdmissionHandler interface {
	admission.Handler

	Setup(client client.Client)
}
