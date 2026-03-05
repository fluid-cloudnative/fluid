package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

type fakeEngineCore struct {
	id string
}

func (e *fakeEngineCore) ID() string { return e.id }

func (e *fakeEngineCore) Shutdown() error { return nil }

func (e *fakeEngineCore) Setup(ctx cruntime.ReconcileRequestContext) (bool, error) {
	return true, nil
}

func (e *fakeEngineCore) CreateVolume() error { return nil }

func (e *fakeEngineCore) DeleteVolume() error { return nil }

func (e *fakeEngineCore) Sync(ctx cruntime.ReconcileRequestContext) error { return nil }

func (e *fakeEngineCore) Validate(ctx cruntime.ReconcileRequestContext) error { return nil }

func (e *fakeEngineCore) Operate(ctx cruntime.ReconcileRequestContext, opStatus *datav1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}
