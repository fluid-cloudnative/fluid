package builder

import "github.com/fluid-cloudnative/fluid/pkg/ctrl"

type ServerlessHandler struct {
}

func (h *ServerlessHandler) Visit(*Info, ctrl.Helper, error) (err error) {
	return
}
