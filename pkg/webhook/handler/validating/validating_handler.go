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

package validating

import (
	"context"

	v1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ValidatingHandler implements admission webhook for validating Fluid CRDs.
type ValidatingHandler struct {
	decoder *admission.Decoder
}

func NewValidatingHandler(decoder *admission.Decoder) *ValidatingHandler {
	return &ValidatingHandler{decoder: decoder}
}

func (h *ValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	// DELETE requests may have an empty Object.Raw; allow them through.
	if len(req.Object.Raw) == 0 {
		return admission.Allowed("no object to validate")
	}

	// Decode into a Dataset when possible.
	var ds v1alpha1.Dataset
	if h.decoder != nil {
		if err := h.decoder.Decode(req, &ds); err == nil {
			return h.validateDataset(&ds)
		}
	}

	// For unknown types, allow through (skeleton — extend as needed).
	return admission.Allowed("validation passed")
}

func (h *ValidatingHandler) validateDataset(ds *v1alpha1.Dataset) admission.Response {
	// Mounts is optional (e.g. Vineyard runtime), so we only validate
	// individual entries when present.
	for _, m := range ds.Spec.Mounts {
		if m.MountPoint == "" {
			return admission.Denied("mount.mountPoint must not be empty")
		}
	}

	// Validate runtime entries when present.
	for _, r := range ds.Spec.Runtimes {
		if r.Name == "" || r.Namespace == "" {
			return admission.Denied("runtime entries must include name and namespace")
		}
	}

	// Require metadata.name (also covers generateName-only submissions).
	if ds.Name == "" {
		return admission.Denied("metadata.name is required")
	}

	return admission.Allowed("dataset validation passed")
}
