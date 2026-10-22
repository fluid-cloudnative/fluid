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

package engine

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CacheEngine Tests", Label("pkg.ddc.cache.engine.engine_test.go"), func() {
	It("should build CacheEngine and get correct warning interval and ID", func() {
		fakeClient := fake.NewFakeClientWithScheme(CacheEngineTestScheme)
		recorder := record.NewFakeRecorder(10)
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Name:      "test-cache",
				Namespace: "default",
			},
			Context:     context.TODO(),
			Log:         ctrl.Log.WithName("test"),
			Client:      fakeClient,
			Recorder:    recorder,
			RuntimeType: common.CacheRuntime,
			EngineImpl:  common.CacheEngineImpl,
		}

		engineObj, err := Build("test-id", ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(engineObj).NotTo(BeNil())
		Expect(engineObj.ID()).To(Equal("test-id"))

		engine, ok := engineObj.(*CacheEngine)
		Expect(ok).To(BeTrue())
		Expect(engine.name).To(Equal("test-cache"))
		Expect(engine.namespace).To(Equal("default"))
		Expect(engine.tieredStoreWarningInterval).To(Equal(defaultTieredStoreWarningInterval))
		Expect(engine.getTieredStoreWarningInterval()).To(Equal(defaultTieredStoreWarningInterval))

		// When custom tieredStoreWarningInterval is set
		engine.tieredStoreWarningInterval = 5 * time.Minute
		Expect(engine.getTieredStoreWarningInterval()).To(Equal(5 * time.Minute))

		// When tieredStoreWarningInterval is 0, fall back to default
		engine.tieredStoreWarningInterval = 0
		Expect(engine.getTieredStoreWarningInterval()).To(Equal(defaultTieredStoreWarningInterval))

		// Precheck test
		found, err := Precheck(fakeClient, types.NamespacedName{Name: "not-found", Namespace: "default"})
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
	})
})
