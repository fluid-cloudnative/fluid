package validating

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestValidatingHandler_Basic(t *testing.T) {
	h := NewValidatingHandler()

	// Valid object
	obj := []byte(`{"metadata":{"name":"test-dataset"}}`)
	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{Raw: obj},
		},
	}
	resp := h.Handle(context.Background(), req)
	assert.True(t, resp.Allowed)

	// Missing metadata.name
	obj = []byte(`{"metadata":{}}`)
	req.AdmissionRequest.Object.Raw = obj
	resp = h.Handle(context.Background(), req)
	assert.False(t, resp.Allowed)
	assert.Contains(t, resp.Result.Message, "metadata.name is required")
}
