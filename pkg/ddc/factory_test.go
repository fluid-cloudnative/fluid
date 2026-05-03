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

package ddc

import (
	"context"
	"fmt"

	fluidv1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/dataoperation"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type fakeEngine struct {
	id string
}

func (f *fakeEngine) ID() string {
	return f.id
}

func (f *fakeEngine) Shutdown() error {
	return nil
}

func (f *fakeEngine) Setup(ctx cruntime.ReconcileRequestContext) (bool, error) {
	return false, nil
}

func (f *fakeEngine) CreateVolume(ctx context.Context) error {
	return nil
}

func (f *fakeEngine) DeleteVolume(ctx context.Context) error {
	return nil
}

func (f *fakeEngine) Sync(ctx cruntime.ReconcileRequestContext) error {
	return nil
}

func (f *fakeEngine) Validate(ctx cruntime.ReconcileRequestContext) error {
	return nil
}

func (f *fakeEngine) Operate(ctx cruntime.ReconcileRequestContext, opStatus *fluidv1alpha1.OperationStatus, operation dataoperation.OperationInterface) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

var _ = Describe("factory", func() {
	Describe("CreateEngine", func() {
		var (
			originalBuildFuncMap map[string]buildFunc
		)

		BeforeEach(func() {
			originalBuildFuncMap = buildFuncMap
			buildFuncMap = map[string]buildFunc{}
		})

		AfterEach(func() {
			buildFuncMap = originalBuildFuncMap
		})

		It("dispatches using the engine impl and forwards id and context", func() {
			expectedEngine := &fakeEngine{id: "engine-id"}
			ctx := cruntime.ReconcileRequestContext{EngineImpl: common.AlluxioEngineImpl}
			captured := struct {
				id  string
				ctx cruntime.ReconcileRequestContext
			}{}

			buildFuncMap[common.AlluxioEngineImpl] = func(id string, gotCtx cruntime.ReconcileRequestContext) (base.Engine, error) {
				captured.id = id
				captured.ctx = gotCtx
				return expectedEngine, nil
			}

			engine, err := CreateEngine("engine-id", ctx)

			Expect(err).NotTo(HaveOccurred())
			Expect(engine).To(BeIdenticalTo(expectedEngine))
			Expect(captured.id).To(Equal("engine-id"))
			Expect(captured.ctx).To(Equal(ctx))
		})

		It("returns builder errors unchanged", func() {
			expectedErr := fmt.Errorf("builder failed")
			ctx := cruntime.ReconcileRequestContext{EngineImpl: common.JindoFSEngineImpl}

			buildFuncMap[common.JindoFSEngineImpl] = func(id string, gotCtx cruntime.ReconcileRequestContext) (base.Engine, error) {
				return nil, expectedErr
			}

			engine, err := CreateEngine("engine-id", ctx)

			Expect(engine).To(BeNil())
			Expect(err).To(MatchError(expectedErr))
		})

		It("errors on unknown impl and mentions it", func() {
			ctx := cruntime.ReconcileRequestContext{EngineImpl: "mystery"}

			engine, err := CreateEngine("engine-id", ctx)

			Expect(engine).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mystery"))
		})
	})

	Describe("GenerateEngineID", func() {
		It("returns namespace-name", func() {
			id := GenerateEngineID(types.NamespacedName{Namespace: "fluid", Name: "demo"})

			Expect(id).To(Equal("fluid-demo"))
		})
	})

	Describe("InferEngineImpl", func() {
		It("returns default for empty value file configmap", func() {
			impl := InferEngineImpl(fluidv1alpha1.RuntimeStatus{}, common.ThinEngineImpl)

			Expect(impl).To(Equal(common.ThinEngineImpl))
		})

		It("returns the recognized engine from dataset-engine-values", func() {
			impl := InferEngineImpl(fluidv1alpha1.RuntimeStatus{ValueFileConfigmap: "dataset-goosefs-values"}, common.ThinEngineImpl)

			Expect(impl).To(Equal(common.GooseFSEngineImpl))
		})

		It("still works when the dataset name contains hyphens", func() {
			impl := InferEngineImpl(fluidv1alpha1.RuntimeStatus{ValueFileConfigmap: "my-dataset-jindocache-values"}, common.ThinEngineImpl)

			Expect(impl).To(Equal(common.JindoCacheEngineImpl))
		})

		It("falls back to default for malformed configmap names", func() {
			impl := InferEngineImpl(fluidv1alpha1.RuntimeStatus{ValueFileConfigmap: "dataset-alluxio-config"}, common.ThinEngineImpl)

			Expect(impl).To(Equal(common.ThinEngineImpl))
		})

		It("falls back to default for unrecognized engines", func() {
			impl := InferEngineImpl(fluidv1alpha1.RuntimeStatus{ValueFileConfigmap: "dataset-unknown-values"}, common.ThinEngineImpl)

			Expect(impl).To(Equal(common.ThinEngineImpl))
		})
	})

	Describe("buildFuncMap", func() {
		It("registers the expected engines", func() {
			Expect(buildFuncMap).To(HaveKey(common.AlluxioEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.JindoFSEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.JindoFSxEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.JindoCacheEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.GooseFSEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.JuiceFSEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.ThinEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.EFCEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.VineyardEngineImpl))
			Expect(buildFuncMap).To(HaveKey(common.CacheEngineImpl))
		})
	})
})
