package mutating

import "github.com/fluid-cloudnative/fluid/pkg/common"

var (
	// HandlerMap contains admission webhook handlers
	HandlerMap = map[string]common.AdmissionHandler{
		common.WebhookMutateAlluxioRuntimePath: &UpdateAlluxioRuntimeHandler{},
	}
)
