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

package cacheruntime

import (
	"context"
	"fmt"
	"net/http"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var log = ctrl.Log.WithName("cacheruntime-validating-handler")

// SpecUpdateHandler denies an UPDATE of a CacheRuntime that changes a spec
// field the controller cannot apply to the running components.
type SpecUpdateHandler struct {
	decoder *admission.Decoder
}

var _ admission.Handler = &SpecUpdateHandler{}

// Setup implements common.AdmissionHandler. The handler compares the two
// objects carried by the request and never reads the cluster, so the clients
// are not kept.
func (h *SpecUpdateHandler) Setup(_ client.Client, _ client.Reader, decoder *admission.Decoder) {
	h.decoder = decoder
}

// Handle implements admission.Handler.
func (h *SpecUpdateHandler) Handle(_ context.Context, req admission.Request) admission.Response {
	if req.Operation != admissionv1.Update {
		return admission.Allowed("only updates of a CacheRuntime are validated")
	}
	// The webhook rule only selects the main resource, so this never matches a
	// status or scale write. Kept so that a wider rule cannot reach the spec
	// comparison with a Scale object.
	if req.SubResource != "" {
		return admission.Allowed("subresource writes of a CacheRuntime are not validated")
	}

	oldRuntime := &datav1alpha1.CacheRuntime{}
	if err := h.decoder.DecodeRaw(req.OldObject, oldRuntime); err != nil {
		log.Error(err, "unable to decode the old CacheRuntime from the request", "namespace", req.Namespace, "name", req.Name)
		return admission.Errored(http.StatusBadRequest, err)
	}
	newRuntime := &datav1alpha1.CacheRuntime{}
	if err := h.decoder.DecodeRaw(req.Object, newRuntime); err != nil {
		log.Error(err, "unable to decode the new CacheRuntime from the request", "namespace", req.Namespace, "name", req.Name)
		return admission.Errored(http.StatusBadRequest, err)
	}

	errs := ValidateSpecUpdate(&oldRuntime.Spec, &newRuntime.Spec)
	if len(errs) == 0 {
		return admission.Allowed("")
	}

	log.Info("denied a CacheRuntime update that changes immutable spec fields", "namespace", req.Namespace, "name", req.Name, "errors", errs.ToAggregate().Error())
	return admission.Denied(fmt.Sprintf("%s: %s", UpdatableFieldsHint, errs.ToAggregate().Error()))
}
