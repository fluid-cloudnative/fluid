package utils

import (
	"github.com/go-logr/logr"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

// LoggingErrorExceptConflict logs error except for updating operation violates with etcd concurrency control
func LoggingErrorExceptConflict(logging logr.Logger, err error, info string) (result error) {
	if apierrs.IsConflict(err) {
		log.Info("Retry later when update operation violates with etcd concurrency control.", "error", err)
	} else {
		log.Error(err, info)
		result = err
	}
	return result
}
