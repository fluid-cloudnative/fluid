package validating

import (
	"context"
	"testing"

	"encoding/json"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	v1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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

func TestValidatingHandler_Dataset(t *testing.T) {
	h := NewValidatingHandler()
	// prepare decoder with scheme so that typed decoding works
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("failed to add v1alpha1 to scheme: %v", err)
	}
	dec := admission.NewDecoder(scheme)
	h.InjectDecoder(dec)

	// Dataset missing mounts and runtimes -> denied
	ds := &v1alpha1.Dataset{}
	raw, _ := json.Marshal(ds)
	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{Raw: raw},
		},
	}
	resp := h.Handle(context.Background(), req)
	assert.False(t, resp.Allowed)

	// Dataset with a valid mount -> allowed
	ds = &v1alpha1.Dataset{}
	ds.Spec.Mounts = []v1alpha1.Mount{{MountPoint: "/data"}}
	raw, _ = json.Marshal(ds)
	req.AdmissionRequest.Object.Raw = raw
	resp = h.Handle(context.Background(), req)
	assert.True(t, resp.Allowed)

	// Dataset with invalid mount (too short) -> denied
	ds = &v1alpha1.Dataset{}
	ds.Spec.Mounts = []v1alpha1.Mount{{MountPoint: "abc"}}
	raw, _ = json.Marshal(ds)
	req.AdmissionRequest.Object.Raw = raw
	resp = h.Handle(context.Background(), req)
	assert.False(t, resp.Allowed)
}
