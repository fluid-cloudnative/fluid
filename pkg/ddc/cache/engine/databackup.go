package engine

import (
	fluiderrors "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *CacheEngine) generateDataBackupValueFile(ctx cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	return "", fluiderrors.NewNotSupported(
		schema.GroupResource{
			Group:    object.GetObjectKind().GroupVersionKind().Group,
			Resource: object.GetObjectKind().GroupVersionKind().Kind,
		}, "CacheRuntime["+e.name+"]")
}
