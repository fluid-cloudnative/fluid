package serverless

import (
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var log logr.Logger

func init() {
	log = ctrl.Log.WithName("serverless")
}
