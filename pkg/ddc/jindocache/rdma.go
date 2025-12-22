package jindocache

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (e *JindoCacheEngine) RdmaEnabled(runtime *datav1alpha1.JindoRuntime) bool {
	if rpcStr, exists := runtime.Annotations[common.LabelAnnotationRPCProtocol]; exists {
		switch rpcStr {
		case common.RPCProtocolRDMA:
			return true
		case common.RPCProtocolTCP:
			return false
		default:
			e.Log.Info("WARNING: unrecognized rpc protocol defined in runtime's annotation, expect to be either \"tcp\" or \"rdma\". Default to use \"tcp\" to continue", "key", common.LabelAnnotationRPCProtocol, "value", rpcStr)
			return false
		}
	}

	// default to false
	return false
}
