/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("JindoCacheEngine Build and Precheck", func() {
	const (
		engineName      = "hbase"
		engineNamespace = "fluid"
	)

	var (
		namespacedName types.NamespacedName
		runtimeObj     *datav1alpha1.JindoRuntime
		reconcileCtx   cruntime.ReconcileRequestContext
		fakeClient     client.Client
	)

	buildRuntime := func() *datav1alpha1.JindoRuntime {
		return &datav1alpha1.JindoRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      engineName,
				Namespace: engineNamespace,
			},
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Replicas: 1,
				},
				Fuse: datav1alpha1.JindoFuseSpec{},
			},
		}
	}

	newContext := func() cruntime.ReconcileRequestContext {
		return cruntime.ReconcileRequestContext{
			NamespacedName: namespacedName,
			Client:         fakeClient,
			Log:            fake.NullLogger(),
			RuntimeType:    common.JindoRuntime,
			EngineImpl:     common.JindoRuntime,
			Runtime:        runtimeObj,
		}
	}

	BeforeEach(func() {
		namespacedName = types.NamespacedName{Name: engineName, Namespace: engineNamespace}
		runtimeObj = buildRuntime()
		fakeClient = fake.NewFakeClientWithScheme(testScheme, runtimeObj.DeepCopy())
		reconcileCtx = newContext()
	})

	Describe("Build", func() {
		It("should build a template engine for a valid Jindo runtime", func() {
			engine, err := Build("testId", reconcileCtx)

			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
			Expect(engine.ID()).To(Equal("testId"))
		})

		It("should fail when reconcile context runtime is nil", func() {
			reconcileCtx.Runtime = nil

			engine, err := Build("testId", reconcileCtx)

			Expect(err).To(MatchError("engine hbase is failed to parse"))
			Expect(engine).To(BeNil())
		})

		It("should fail when reconcile context runtime is not a Jindo runtime", func() {
			reconcileCtx.Runtime = &datav1alpha1.AlluxioRuntime{}

			engine, err := Build("testId", reconcileCtx)

			Expect(err).To(MatchError("engine hbase is failed to parse"))
			Expect(engine).To(BeNil())
		})

		It("should fail when runtime info cannot be loaded from the client", func() {
			fakeClient = fake.NewFakeClientWithScheme(testScheme)
			reconcileCtx = newContext()

			engine, err := Build("testId", reconcileCtx)

			Expect(err).To(MatchError("engine hbase failed to get runtime info"))
			Expect(engine).To(BeNil())
		})
	})

	Describe("Precheck", func() {
		It("should return true when the runtime exists", func() {
			found, err := Precheck(fakeClient, namespacedName)

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
		})

		It("should return false when the runtime does not exist", func() {
			found, err := Precheck(fakeClient, types.NamespacedName{Name: "missing", Namespace: engineNamespace})

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("Build runtime info behavior", func() {
		It("should accept a runtime with no annotations or owner references in unit tests", func() {
			runtimeObj = &datav1alpha1.JindoRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      engineName,
					Namespace: engineNamespace,
				},
			}
			fakeClient = fake.NewFakeClientWithScheme(testScheme, runtimeObj.DeepCopy(), &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: engineNamespace},
			})
			reconcileCtx = newContext()

			engine, err := Build("testId", reconcileCtx)

			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
		})
	})
})
