package base

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// GetWorkerStatefulsetName returns the worker StatefulSet name for the runtime.
// This function determines the name of the worker StatefulSet based on the runtime type.
// Jindo-related runtimes use the "-jindofs-worker" suffix, while other runtimes use the "-worker" suffix.
//
// Parameters:
//   - None: This function does not take any parameters.
//
// Returns:
//   - string: The generated worker StatefulSet name for the current runtime.
func (info *RuntimeInfo) GetWorkerStatefulsetName() string {
	switch info.runtimeType {
	// JindoRuntime has a extra "-jindofs" suffix in worker statefulset name.
	case common.JindoRuntime, common.JindoCacheEngineImpl, common.JindoFSxEngineImpl:
		return info.name + "-jindofs-worker"
	default:
		return info.name + "-worker"
	}
}
