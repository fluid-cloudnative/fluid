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

package thin

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func buildEngineTestContext(client ctrlclient.Client, runtimeObj ctrlclient.Object, name, namespace string) cruntime.ReconcileRequestContext {
	return cruntime.ReconcileRequestContext{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
		Client:      client,
		Log:         fake.NullLogger(),
		RuntimeType: common.ThinRuntime,
		Runtime:     runtimeObj,
	}
}

var _ = Describe("Thin Engine Build", Label("pkg.ddc.thin.engine_test.go"), func() {
	Describe("Build", func() {
		It("should return an error when runtime is nil", func() {
			client := fake.NewFakeClientWithScheme(testScheme)

			engine, err := Build("test-id", cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{Name: "hbase", Namespace: "fluid"},
				Client:         client,
				Log:            fake.NullLogger(),
				RuntimeType:    common.ThinRuntime,
				Runtime:        nil,
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("runtime is nil"))
			Expect(engine).To(BeNil())
		})

		It("should return an error when runtime type conversion fails", func() {
			alluxioRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
			}
			client := fake.NewFakeClientWithScheme(testScheme, alluxioRuntime)

			engine, err := Build("test-id", buildEngineTestContext(client, alluxioRuntime, "hbase", "fluid"))

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("type conversion"))
			Expect(engine).To(BeNil())
		})

		It("should build a template engine for a normal thin runtime when runtime and profile exist", func() {
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fluid"}}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
			}
			runtimeProfile := &datav1alpha1.ThinRuntimeProfile{
				ObjectMeta: metav1.ObjectMeta{Name: "test-profile"},
				Spec:       datav1alpha1.ThinRuntimeProfileSpec{FileSystemType: "test-fstype"},
			}
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec: datav1alpha1.ThinRuntimeSpec{
					ThinRuntimeProfileName: "test-profile",
					Fuse:                   datav1alpha1.ThinFuseSpec{},
				},
				Status: datav1alpha1.RuntimeStatus{
					CacheStates: map[common.CacheStateName]string{common.Cached: "true"},
				},
			}
			worker := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase-worker", Namespace: "fluid"},
			}
			client := fake.NewFakeClientWithScheme(testScheme, namespace, dataset, runtimeProfile, thinRuntime, worker)

			engine, err := Build("test-id", buildEngineTestContext(client, thinRuntime, "hbase", "fluid"))

			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
		})

		It("should delegate to the reference dataset branch when profile name is empty", func() {
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fluid"}}
			physicalDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"},
				Status: datav1alpha1.DatasetStatus{
					Runtimes: []datav1alpha1.Runtime{{Name: "done", Namespace: "big-data", Type: common.AlluxioRuntime}},
				},
			}
			physicalRuntime := &datav1alpha1.AlluxioRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"},
			}
			refRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
			}
			refDataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://big-data/done"}},
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, namespace, physicalDataset, physicalRuntime, refRuntime, refDataset)

			engine, err := Build("test-id", buildEngineTestContext(client, refRuntime, "hbase", "fluid"))

			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
		})

		It("should return an error when thin runtime profile lookup fails", func() {
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fluid"}}
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
			}
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec: datav1alpha1.ThinRuntimeSpec{
					ThinRuntimeProfileName: "missing-profile",
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, namespace, dataset, thinRuntime)

			engine, err := Build("test-id", buildEngineTestContext(client, thinRuntime, "hbase", "fluid"))

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error when getting thinruntime profile missing-profile"))
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
			Expect(engine).To(BeNil())
		})

		It("should return an error when runtime info bootstrap cannot fetch the thin runtime", func() {
			namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fluid"}}
			runtimeProfile := &datav1alpha1.ThinRuntimeProfile{
				ObjectMeta: metav1.ObjectMeta{Name: "test-profile"},
				Spec:       datav1alpha1.ThinRuntimeProfileSpec{FileSystemType: "test-fstype"},
			}
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec: datav1alpha1.ThinRuntimeSpec{
					ThinRuntimeProfileName: "test-profile",
				},
			}
			client := fake.NewFakeClientWithScheme(testScheme, namespace, runtimeProfile)

			engine, err := Build("test-id", buildEngineTestContext(client, thinRuntime, "hbase", "fluid"))

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("engine hbase failed to get runtime info"))
			Expect(engine).To(BeNil())
		})
	})

	Describe("Precheck", func() {
		It("should return found true when the thin runtime exists", func() {
			thinRuntime := &datav1alpha1.ThinRuntime{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
			}
			client := fake.NewFakeClientWithScheme(testScheme, thinRuntime)

			found, err := Precheck(client, types.NamespacedName{Name: "hbase", Namespace: "fluid"})

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeTrue())
		})

		It("should return found false when the thin runtime does not exist", func() {
			client := fake.NewFakeClientWithScheme(testScheme)

			found, err := Precheck(client, types.NamespacedName{Name: "missing", Namespace: "fluid"})

			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeFalse())
		})
	})

	Describe("CheckReferenceDatasetRuntime", func() {
		DescribeTable("should match the current thin runtime profile-name contract",
			func(profileName string, expected bool) {
				thinRuntime := &datav1alpha1.ThinRuntime{
					ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
					Spec:       datav1alpha1.ThinRuntimeSpec{ThinRuntimeProfileName: profileName},
				}

				isRef, err := CheckReferenceDatasetRuntime(cruntime.ReconcileRequestContext{Context: context.TODO()}, thinRuntime)

				Expect(err).NotTo(HaveOccurred())
				Expect(isRef).To(Equal(expected))
			},
			Entry("empty profile name means reference dataset runtime", "", true),
			Entry("non-empty profile name means normal thin runtime", "test-profile", false),
		)
	})
})
