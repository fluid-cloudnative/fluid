/*
  Copyright 2022 The Fluid Authors.

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

package efc

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
	_ = datav1alpha1.AddToScheme(testScheme)
	_ = appsv1.AddToScheme(testScheme)
}

const (
	engineTestNamespace = "fluid"
	engineTestName      = "hbase"
	engineTestID        = "testId"
)

func newEngineRuntime() *datav1alpha1.EFCRuntime {
	return &datav1alpha1.EFCRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engineTestName,
			Namespace: engineTestNamespace,
		},
		Spec: datav1alpha1.EFCRuntimeSpec{
			Fuse: datav1alpha1.EFCFuseSpec{
				CleanPolicy: datav1alpha1.OnDemandCleanPolicy,
			},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
}

func newEngineContext(clientObjs ...runtime.Object) cruntime.ReconcileRequestContext {
	runtime := newEngineRuntime()
	if len(clientObjs) == 0 {
		clientObjs = append(clientObjs, runtime.DeepCopy())
	}

	return cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      engineTestName,
			Namespace: engineTestNamespace,
		},
		Client:      fake.NewFakeClientWithScheme(testScheme, clientObjs...),
		Log:         fake.NullLogger(),
		RuntimeType: common.EFCRuntime,
		Runtime:     runtime,
	}
}

var _ = Describe("EFCEngine", func() {
	Describe("Build", func() {
		It("builds an engine when the runtime can be resolved", func() {
			engine, err := Build(engineTestID, newEngineContext())

			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
		})

		DescribeTable("returns an error for invalid runtime input",
			func(runtimeObj client.Object, expectedError string) {
				ctx := newEngineContext()
				ctx.Runtime = runtimeObj

				engine, err := Build(engineTestID, ctx)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(expectedError))
				Expect(engine).To(BeNil())
			},
			Entry("when runtime is nil", nil, "engine hbase is failed to parse"),
			Entry("when runtime has the wrong type", &datav1alpha1.JindoRuntime{}, "engine hbase is failed to parse"),
		)

		It("returns an error when runtime info cannot be resolved", func() {
			ctx := newEngineContext()
			ctx.Client = fake.NewFakeClientWithScheme(testScheme)

			engine, err := Build(engineTestID, ctx)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("engine hbase failed to get runtime info"))
			Expect(engine).To(BeNil())
		})
	})

	DescribeTable("Precheck",
		func(clientObjs []runtime.Object, expectedFound bool) {
			found, err := Precheck(
				fake.NewFakeClientWithScheme(testScheme, clientObjs...),
				types.NamespacedName{Name: engineTestName, Namespace: engineTestNamespace},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(Equal(expectedFound))
		},
		Entry("returns true when the runtime exists", []runtime.Object{newEngineRuntime().DeepCopy()}, true),
		Entry("returns false when the runtime does not exist", []runtime.Object{}, false),
	)
})
