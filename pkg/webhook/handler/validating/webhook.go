package validating

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var HandlerMap = map[string]common.AdmissionHandler{
	"/validate": &Handler{},
}

type Handler struct {
	*ValidatingHandler
}

func (h *Handler) Setup(c client.Client, apiReader client.Reader, decoder *admission.Decoder) {
	h.ValidatingHandler = NewValidatingHandler()
	h.ValidatingHandler.InjectDecoder(decoder)
}
