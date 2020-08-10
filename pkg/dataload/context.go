package dataload

import (
	"context"
	"github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

// Context for processing requests of dataload reconciling
type ReconcileRequestContext struct {
	context.Context
	Log logr.Logger
	types.NamespacedName
	Recorder record.EventRecorder
	DataLoad v1alpha1.AlluxioDataLoad
}
