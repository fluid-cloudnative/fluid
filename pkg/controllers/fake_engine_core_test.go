/*
Copyright 2026 The Fluid Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
