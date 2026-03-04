package validating

import (
    "context"
    "encoding/json"
    "net/http"

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
    // Generic validation: ensure object contains metadata.name
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

    // Passed basic validation
    return admission.Allowed("validation passed")
}

func (h *ValidatingHandler) InjectDecoder(d *admission.Decoder) error {
    h.decoder = d
    return nil
}
