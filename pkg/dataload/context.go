package dataload

import (
	"context"
	"github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
)

// Context for processing requests of dataload reconciling
type ReconcileRequestContext struct {
	context.Context
	Log      logr.Logger
	DataLoad v1alpha1.AlluxioDataLoad
	types.NamespacedName
}
