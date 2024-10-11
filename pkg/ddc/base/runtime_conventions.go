package base

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (info *RuntimeInfo) GetWorkerStatefulsetName() string {
	switch info.runtimeType {
	case common.JindoRuntime, common.JindoCacheEngineImpl, common.JindoFSxEngineImpl:
		return info.name + "-jindofs-worker"
	default:
		return info.name + "-worker"
	}
}
