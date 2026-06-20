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

package alluxio

import (
	"context"
	"sync"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/controllers"
	"github.com/fluid-cloudnative/fluid/pkg/ddc"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	basemock "github.com/fluid-cloudnative/fluid/pkg/ddc/base/mock"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

// makeReconciler creates a RuntimeReconciler with a fake client for unit tests.
func makeReconciler(s *runtime.Scheme, objs ...runtime.Object) *RuntimeReconciler {
	c := fake.NewFakeClientWithScheme(s, objs...)
	r := &RuntimeReconciler{
		Scheme:  s,
		mutex:   &sync.Mutex{},
		engines: map[string]base.Engine{},
	}
	r.RuntimeReconciler = controllers.NewRuntimeReconciler(r, c, fake.NullLogger(), record.NewFakeRecorder(10))
	return r
}

var _ = Describe("implement", func() {
	const (
		rtName      = "test-alluxio"
		rtNamespace = "default"
	)

	var (
		s *runtime.Scheme
	)

	BeforeEach(func() {
		s = runtime.NewScheme()
		Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
	})

	Describe("getRuntime", func() {
		It("returns the AlluxioRuntime when it exists", func() {
			rt := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      rtName,
					Namespace: rtNamespace,
				},
			}
			r := makeReconciler(s, rt)
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Name: rtName, Namespace: rtNamespace},
				Log:            fake.NullLogger(),
				Client:         r.Client,
			}

			got, err := r.getRuntime(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).ToNot(BeNil())
			Expect(got.Name).To(Equal(rtName))
			Expect(got.Namespace).To(Equal(rtNamespace))
		})

		It("returns an error when the AlluxioRuntime does not exist", func() {
			r := makeReconciler(s)
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Name: "missing", Namespace: rtNamespace},
				Log:            fake.NullLogger(),
				Client:         r.Client,
			}

			_, err := r.getRuntime(ctx)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RemoveEngine", func() {
		It("removes an engine entry from the engines map", func() {
			r := makeReconciler(s)
			nsn := types.NamespacedName{Name: rtName, Namespace: rtNamespace}
			id := ddc.GenerateEngineID(nsn)

			// Inject a non-nil stub engine so RemoveEngine has a real entry to delete.
			mockCtrl := gomock.NewController(GinkgoT())
			var stubEngine base.Engine = basemock.NewMockEngine(mockCtrl)
			r.mutex.Lock()
			r.engines[id] = stubEngine
			r.mutex.Unlock()

			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: nsn,
				Log:            fake.NullLogger(),
			}
			r.RemoveEngine(ctx)

			r.mutex.Lock()
			_, exists := r.engines[id]
			r.mutex.Unlock()
			Expect(exists).To(BeFalse())
		})

		It("is a no-op when no engine exists for the id", func() {
			r := makeReconciler(s)
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Name: "no-such", Namespace: rtNamespace},
				Log:            fake.NullLogger(),
			}
			// Must not panic.
			Expect(func() { r.RemoveEngine(ctx) }).ToNot(Panic())
		})
	})

	Describe("GetOrCreateEngine", func() {
		It("returns an error when no matching engine builder is registered", func() {
			r := makeReconciler(s)
			ctx := cruntime.ReconcileRequestContext{
				Context:        context.Background(),
				NamespacedName: types.NamespacedName{Name: rtName, Namespace: rtNamespace},
				Log:            fake.NullLogger(),
				EngineImpl:     "unknown-impl",
				Client:         r.Client,
			}
			_, err := r.GetOrCreateEngine(ctx)
			Expect(err).To(HaveOccurred())
		})
	})
})
