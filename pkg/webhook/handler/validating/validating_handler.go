package validating

import (
	"context"
	"encoding/json"
	"net/http"

	v1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ValidatingHandler implements admission webhook for validating Fluid CRDs.
type ValidatingHandler struct {
	decoder *admission.Decoder
}

func NewValidatingHandler() *ValidatingHandler {
	return &ValidatingHandler{}
}

func (h *ValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	// Try to decode into a known type (Dataset) when possible
	var ds v1alpha1.Dataset
	if h.decoder != nil {
		if err := h.decoder.Decode(req, &ds); err == nil {
			// Perform Dataset-specific validations
			// Require either Mounts or Runtimes to be present
			if len(ds.Spec.Mounts) == 0 && len(ds.Spec.Runtimes) == 0 {
				return admission.Denied("dataset.spec must contain at least one mount or runtime")
			}

			// Validate mounts
			for _, m := range ds.Spec.Mounts {
				if m.MountPoint == "" {
					return admission.Denied("mount.mountPoint must not be empty")
				}
			}

			// Validate runtimes
			for _, r := range ds.Spec.Runtimes {
				if r.Name == "" || r.Namespace == "" {
					return admission.Denied("runtime entries must include name and namespace")
				}
			}

			return admission.Allowed("dataset validation passed")
		}
	}

	// Fallback generic validation: ensure object contains metadata.name
	var obj map[string]interface{}
	if err := json.Unmarshal(req.Object.Raw, &obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	metadata, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return admission.Denied("metadata.name is required")
	}
	name, ok := metadata["name"].(string)
	if !ok || name == "" {
		return admission.Denied("metadata.name is required")
	}

	return admission.Allowed("validation passed")
}

func (h *ValidatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
