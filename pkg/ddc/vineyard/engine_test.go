/*
Copyright 2024 The Fluid Authors.
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

package vineyard

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	engineTestNamespace = "fluid"
	engineTestName      = "hbase"
	engineTestID        = "testId"
)

var _ = Describe("VineyardEngine", Label("pkg.ddc.vineyard.engine_test.go"), func() {
	Describe("Build", func() {
		Context("when runtime is a valid VineyardRuntime", func() {
			It("should build the template engine", func() {
				vineyardRuntime := newEngineTestVineyardRuntime()
				fakeClient := fake.NewFakeClientWithScheme(testScheme, newEngineTestObjects(vineyardRuntime)...)

				engine, err := Build(engineTestID, newEngineTestContext(fakeClient, vineyardRuntime))

				Expect(err).NotTo(HaveOccurred())
				Expect(engine).NotTo(BeNil())
			})
		})

		Context("when runtime is nil", func() {
			It("should return a parse error", func() {
				fakeClient := fake.NewFakeClientWithScheme(testScheme)

				engine, err := Build(engineTestID, newEngineTestContext(fakeClient, nil))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("engine hbase is failed to parse"))
				Expect(engine).To(BeNil())
			})
		})

		Context("when runtime has the wrong type", func() {
			It("should return a parse error", func() {
				dataset := newEngineTestDataset()
				fakeClient := fake.NewFakeClientWithScheme(testScheme)

				engine, err := Build(engineTestID, newEngineTestContext(fakeClient, dataset))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("engine hbase is failed to parse"))
				Expect(engine).To(BeNil())
			})
		})

		Context("when runtime info cannot be loaded", func() {
			It("should return a runtime info error", func() {
				vineyardRuntime := newEngineTestVineyardRuntime()
				fakeClient := fake.NewFakeClientWithScheme(testScheme)

				engine, err := Build(engineTestID, newEngineTestContext(fakeClient, vineyardRuntime))

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("engine hbase failed to get runtime info"))
				Expect(engine).To(BeNil())
			})
		})
	})

	Describe("Precheck", func() {
		Context("when the VineyardRuntime exists", func() {
			It("should return found true without error", func() {
				vineyardRuntime := newEngineTestVineyardRuntime()
				fakeClient := fake.NewFakeClientWithScheme(testScheme, vineyardRuntime)

				found, err := Precheck(fakeClient, types.NamespacedName{Name: engineTestName, Namespace: engineTestNamespace})

				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("when the VineyardRuntime does not exist", func() {
			It("should return found false without error", func() {
				fakeClient := fake.NewFakeClientWithScheme(testScheme)

				found, err := Precheck(fakeClient, types.NamespacedName{Name: engineTestName, Namespace: engineTestNamespace})

				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})
	})
})

func newEngineTestContext(fakeClient client.Client, runtime client.Object) cruntime.ReconcileRequestContext {
	return cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      engineTestName,
			Namespace: engineTestNamespace,
		},
		Client:      fakeClient,
		Log:         fake.NullLogger(),
		RuntimeType: "vineyard",
		Runtime:     runtime,
	}
}

func newEngineTestObjects(vineyardRuntime *datav1alpha1.VineyardRuntime) []k8sruntime.Object {
	return []k8sruntime.Object{
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: engineTestNamespace}},
		newEngineTestDataset(),
		vineyardRuntime,
		&appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      engineTestName + "-worker",
				Namespace: engineTestNamespace,
			},
		},
	}
}

func newEngineTestDataset() *datav1alpha1.Dataset {
	return &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engineTestName,
			Namespace: engineTestNamespace,
		},
	}
}

func newEngineTestVineyardRuntime() *datav1alpha1.VineyardRuntime {
	return &datav1alpha1.VineyardRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      engineTestName,
			Namespace: engineTestNamespace,
		},
		Spec: datav1alpha1.VineyardRuntimeSpec{
			Master: datav1alpha1.MasterSpec{
				VineyardCompTemplateSpec: datav1alpha1.VineyardCompTemplateSpec{
					Replicas: 1,
				},
			},
		},
		Status: datav1alpha1.RuntimeStatus{
			CacheStates: map[common.CacheStateName]string{
				common.Cached: "true",
			},
		},
	}
}
