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

// Package validating registers the validating admission webhook handlers.
package validating

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/handler/validating/cacheruntime"
)

// +kubebuilder:webhook:path=/validate-data-fluid-io-v1alpha1-cacheruntime,mutating=false,failurePolicy=fail,sideEffects=None,admissionReviewVersions=v1;v1beta1,groups=data.fluid.io,resources=cacheruntimes,verbs=update,versions=v1alpha1,name=update.cacheruntime.fluid.io

var (
	// HandlerMap contains the validating admission webhook handlers, keyed by path.
	HandlerMap = map[string]common.AdmissionHandler{
		common.WebhookValidateCacheRuntimePath: &cacheruntime.SpecUpdateHandler{},
	}
)
