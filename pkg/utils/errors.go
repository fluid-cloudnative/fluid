package utils

import (
	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// LoggingErrorExceptConflict logs error except for updating operation violates with etcd concurrency control
func LoggingErrorExceptConflict(logging logr.Logger, err error, info string, namespacedKey types.NamespacedName) (result error) {
	if apierrs.IsConflict(err) {
		log.Info("Retry later when update operation violates with apiserver concurrency control.",
			"error", err,
			"name", namespacedKey.Name,
			"namespace", namespacedKey.Namespace)
	} else {
		log.Error(err, info, "name", namespacedKey.Name,
			"namespace", namespacedKey.Namespace)
		result = err
	}
	return result
}
