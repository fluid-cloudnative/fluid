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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	v1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestValidatingHandler_EmptyRaw(t *testing.T) {
	h := newHandler(t)

	// DELETE requests may have empty Object.Raw — should be allowed.
	req := admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Object: runtime.RawExtension{Raw: nil},
		},
	}
	resp := h.Handle(context.Background(), req)
	assert.True(t, resp.Allowed)
}

func TestValidatingHandler_Dataset(t *testing.T) {
	h := newHandler(t)

	tests := []struct {
		name    string
		ds      *v1alpha1.Dataset
		allowed bool
		message string
	}{
		{
			name:    "empty mounts and runtimes is allowed (mounts are optional)",
			ds:      &v1alpha1.Dataset{ObjectMeta: metav1.ObjectMeta{Name: "ds"}},
			allowed: true,
		},
		{
			name: "valid mount is allowed",
			ds: &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "ds"},
				Spec:       v1alpha1.DatasetSpec{Mounts: []v1alpha1.Mount{{MountPoint: "/data"}}},
			},
			allowed: true,
		},
		{
			name: "empty mountPoint is denied",
			ds: &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "ds"},
				Spec:       v1alpha1.DatasetSpec{Mounts: []v1alpha1.Mount{{MountPoint: ""}}},
			},
			allowed: false,
			message: "mount.mountPoint must not be empty",
		},
		{
			name: "runtime missing namespace is denied",
			ds: &v1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "ds"},
				Spec:       v1alpha1.DatasetSpec{Runtimes: []v1alpha1.Runtime{{Name: "alluxio"}}},
			},
			allowed: false,
			message: "runtime entries must include name and namespace",
		},
		{
			name:    "missing metadata.name is denied",
			ds:      &v1alpha1.Dataset{},
			allowed: false,
			message: "metadata.name is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set TypeMeta so the decoder can recognize the GVK.
			tc.ds.TypeMeta = metav1.TypeMeta{
				APIVersion: "data.fluid.io/v1alpha1",
				Kind:       "Dataset",
			}
			raw, err := json.Marshal(tc.ds)
			assert.NoError(t, err)

			req := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					Object: runtime.RawExtension{Raw: raw},
				},
			}
			resp := h.Handle(context.Background(), req)
			assert.Equal(t, tc.allowed, resp.Allowed, "test %q", tc.name)
			if tc.message != "" {
				assert.Contains(t, resp.Result.Message, tc.message)
			}
		})
	}
}

// newHandler creates a ValidatingHandler wired with a decoder for v1alpha1.
func newHandler(t *testing.T) *ValidatingHandler {
	t.Helper()
	scheme := runtime.NewScheme()
	assert.NoError(t, v1alpha1.AddToScheme(scheme))
	h := NewValidatingHandler()
	h.InjectDecoder(admission.NewDecoder(scheme))
	return h
}
