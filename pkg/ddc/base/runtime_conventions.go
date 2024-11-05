package base

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (info *RuntimeInfo) GetWorkerStatefulsetName() string {
	switch info.runtimeType {
	// JindoRuntime has a extra "-jindofs" suffix in worker statefulset name.
	case common.JindoRuntime, common.JindoCacheEngineImpl, common.JindoFSxEngineImpl:
		return info.name + "-jindofs-worker"
	default:
		return info.name + "-worker"
	}
}
