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

package base_test

import (
	"errors"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	enginemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/types"
)

func newSetupEngine(impl *enginemock.MockImplement) (*base.TemplateEngine, cruntime.ReconcileRequestContext) {
	ctx := cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "dataset-0",
		},
		Log:         fake.NullLogger(),
		RuntimeType: "alluxio",
		Runtime:     &datav1alpha1.AlluxioRuntime{},
	}

	return base.NewTemplateEngine(impl, "setup-test", ctx), ctx
}

func assertSetupResult(t *testing.T, ready bool, err error, expectedReady bool, expectedErr error) {
	t.Helper()
	if ready != expectedReady {
		t.Fatalf("unexpected ready state, got %v, want %v", ready, expectedReady)
	}

	if expectedErr == nil && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expectedErr != nil && !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error, got %v, want %v", err, expectedErr)
	}
}

func TestSetupShouldCheckUFSError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	impl := enginemock.NewMockImplement(ctrl)
	engine, ctx := newSetupEngine(impl)
	testErr := errors.New("should-check-ufs failed")

	gomock.InOrder(
		impl.EXPECT().ShouldSetupMaster().Return(false, nil).Times(1),
		impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1),
		impl.EXPECT().ShouldCheckUFS().Return(false, testErr).Times(1),
	)

	ready, err := engine.Setup(ctx)
	assertSetupResult(t, ready, err, false, testErr)
}

func TestSetupPrepareUFSError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	impl := enginemock.NewMockImplement(ctrl)
	engine, ctx := newSetupEngine(impl)
	testErr := errors.New("prepare-ufs failed")

	gomock.InOrder(
		impl.EXPECT().ShouldSetupMaster().Return(false, nil).Times(1),
		impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1),
		impl.EXPECT().ShouldCheckUFS().Return(true, nil).Times(1),
		impl.EXPECT().PrepareUFS().Return(testErr).Times(1),
	)

	ready, err := engine.Setup(ctx)
	assertSetupResult(t, ready, err, false, testErr)
}

func TestSetupWorkersSetupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	impl := enginemock.NewMockImplement(ctrl)
	engine, ctx := newSetupEngine(impl)
	testErr := errors.New("setup-workers failed")

	gomock.InOrder(
		impl.EXPECT().ShouldSetupMaster().Return(false, nil).Times(1),
		impl.EXPECT().CheckMasterReady().Return(true, nil).Times(1),
		impl.EXPECT().ShouldCheckUFS().Return(false, nil).Times(1),
		impl.EXPECT().ShouldSetupWorkers().Return(true, nil).Times(1),
		impl.EXPECT().SetupWorkers().Return(testErr).Times(1),
	)

	ready, err := engine.Setup(ctx)
	assertSetupResult(t, ready, err, false, testErr)
}
