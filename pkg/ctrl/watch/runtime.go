package watch

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type runtimeEventHandler struct {
}

func (handler *runtimeEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) bool {
		runtime, ok := e.Object.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if e.Object.GetDeletionTimestamp() != nil {
			return false
		}

		log.V(1).Info("runtimeEventHandler.onCreateFunc", "name", runtime.GetName(), "namespace", runtime.GetNamespace())
		return true
	}
}

func (handler *runtimeEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		runtimeNew, ok := e.ObjectNew.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		runtimeOld, ok := e.ObjectOld.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if runtimeNew.GetResourceVersion() == runtimeOld.GetResourceVersion() {
			log.V(1).Info("runtime.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		log.V(1).Info("runtimeEventHandler.onUpdateFunc", "name", runtimeNew.GetName(), "namespace", runtimeNew.GetNamespace())
		return true
	}
}

func (handler *runtimeEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		runtime, ok := e.Object.(base.RuntimeInterface)
		if !ok {
			log.Info("runtime.onDeleteFunc Skip", "object", e.Object)
			return false
		}

		log.V(1).Info("runtimeEventHandler.onDeleteFunc", "name", runtime.GetName(), "namespace", runtime.GetNamespace())
		return true
	}
}
