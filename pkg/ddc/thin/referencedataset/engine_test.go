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

package referencedataset

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
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

var _ = Describe("ReferenceDatasetEngine", func() {
	Describe("BuildReferenceDatasetThinEngine", func() {
		It("should return expected error state for each case", func() {
			testScheme := runtime.NewScheme()
			_ = v1.AddToScheme(testScheme)
			_ = datav1alpha1.AddToScheme(testScheme)
			_ = appsv1.AddToScheme(testScheme)

			testObjs := []runtime.Object{}

			dataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"},
				Status:     datav1alpha1.DatasetStatus{Runtimes: []datav1alpha1.Runtime{{Name: "done", Namespace: "big-data", Type: common.AlluxioRuntime}}},
			}
			alluxioRuntime := datav1alpha1.AlluxioRuntime{ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"}}

			refRuntime := datav1alpha1.ThinRuntime{ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"}}
			refDataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec:       datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://big-data/done"}}},
			}

			multipleRefDataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid-mul"},
				Spec:       datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://big-data/done"}, {MountPoint: "http://big-test/done"}}},
			}
			multipleRefRuntime := datav1alpha1.ThinRuntime{ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid-mul"}}

			testObjs = append(testObjs, &dataset, &refDataset, &multipleRefDataset)
			testObjs = append(testObjs, &alluxioRuntime, &refRuntime, &multipleRefRuntime)
			c := fake.NewFakeClientWithScheme(testScheme, testObjs...)

			testcases := []struct {
				name    string
				ctx     cruntime.ReconcileRequestContext
				wantErr bool
			}{
				{
					name: "success",
					ctx: cruntime.ReconcileRequestContext{
						NamespacedName: types.NamespacedName{Name: "hbase", Namespace: "fluid"},
						Client:         c,
						Log:            fake.NullLogger(),
						RuntimeType:    common.ThinRuntime,
						Runtime:        &refRuntime,
					},
					wantErr: false,
				},
				{
					name: "dataset-not-ref",
					ctx: cruntime.ReconcileRequestContext{
						NamespacedName: types.NamespacedName{Name: "done", Namespace: "big-data"},
						Client:         c,
						Log:            fake.NullLogger(),
						RuntimeType:    common.ThinRuntime,
						Runtime:        &alluxioRuntime,
					},
					wantErr: true,
				},
				{
					name: "dataset-with-different-format",
					ctx: cruntime.ReconcileRequestContext{
						NamespacedName: types.NamespacedName{Name: "hbase", Namespace: "fluid-mul"},
						Client:         c,
						Log:            fake.NullLogger(),
						RuntimeType:    common.ThinRuntime,
						Runtime:        &multipleRefRuntime,
					},
					wantErr: true,
				},
			}

			for _, tc := range testcases {
				_, err := BuildReferenceDatasetThinEngine(tc.name, tc.ctx)
				Expect(err != nil).To(Equal(tc.wantErr), "case: %s", tc.name)
			}
		})
	})

	Describe("Setup", func() {
		It("should setup and copy configmap", func() {
			testScheme := runtime.NewScheme()
			_ = v1.AddToScheme(testScheme)
			_ = datav1alpha1.AddToScheme(testScheme)
			_ = appsv1.AddToScheme(testScheme)

			dataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"},
				Status: datav1alpha1.DatasetStatus{
					Runtimes:   []datav1alpha1.Runtime{{Name: "done", Namespace: "big-data", Type: common.AlluxioRuntime}},
					DatasetRef: []string{"fluid/test"},
				},
			}
			alluxioRuntime := datav1alpha1.AlluxioRuntime{ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"}}
			refRuntime := datav1alpha1.ThinRuntime{ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"}}
			refDataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec:       datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://big-data/done"}}},
			}
			configCM := v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: alluxioRuntime.Name + "-config", Namespace: alluxioRuntime.Namespace}, Data: map[string]string{"check.sh": "/bin/sh check"}}
			fuseDs := appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "done-fuse", Namespace: "big-data"}}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &dataset, &refDataset, &configCM, &alluxioRuntime, &refRuntime, &fuseDs)

			e := &ReferenceDatasetEngine{Client: fakeClient, name: refRuntime.Name, namespace: refRuntime.Namespace}
			gotReady, err := e.Setup(cruntime.ReconcileRequestContext{Dataset: &refDataset, Client: fakeClient})
			Expect(err).NotTo(HaveOccurred())
			Expect(gotReady).To(BeTrue())

			updatedDataset := &datav1alpha1.Dataset{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Namespace: alluxioRuntime.Namespace, Name: alluxioRuntime.Name}, updatedDataset)
			Expect(err).NotTo(HaveOccurred())
			Expect(utils.ContainsString(updatedDataset.Status.DatasetRef, base.GetDatasetRefName(e.name, e.namespace))).To(BeTrue())

			cmList := &v1.ConfigMapList{}
			err = fakeClient.List(context.TODO(), cmList, &client.ListOptions{Namespace: e.namespace})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(cmList.Items)).To(Equal(1))
		})
	})

	Describe("Shutdown", func() {
		It("should remove dataset ref name", func() {
			testScheme := runtime.NewScheme()
			_ = v1.AddToScheme(testScheme)
			_ = datav1alpha1.AddToScheme(testScheme)
			_ = appsv1.AddToScheme(testScheme)

			dataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"},
				Status: datav1alpha1.DatasetStatus{
					Runtimes:   []datav1alpha1.Runtime{{Name: "done", Namespace: "big-data", Type: common.AlluxioRuntime}},
					DatasetRef: []string{"fluid/hbase", "fluid/test"},
				},
			}
			alluxioRuntime := datav1alpha1.AlluxioRuntime{ObjectMeta: metav1.ObjectMeta{Name: "done", Namespace: "big-data"}}
			refRuntime := datav1alpha1.ThinRuntime{ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"}}
			refDataset := datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{Name: "hbase", Namespace: "fluid"},
				Spec:       datav1alpha1.DatasetSpec{Mounts: []datav1alpha1.Mount{{MountPoint: "dataset://big-data/done"}}},
			}

			fakeClient := fake.NewFakeClientWithScheme(testScheme, &dataset, &refDataset, &alluxioRuntime, &refRuntime)
			e := &ReferenceDatasetEngine{Client: fakeClient, name: refRuntime.Name, namespace: refRuntime.Namespace}

			err := e.Shutdown()
			Expect(err).NotTo(HaveOccurred())

			updatedDataset := &datav1alpha1.Dataset{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{Namespace: e.physicalRuntimeInfo.GetNamespace(), Name: e.physicalRuntimeInfo.GetName()}, updatedDataset)
			Expect(err).NotTo(HaveOccurred())
			Expect(utils.ContainsString(updatedDataset.Status.DatasetRef, base.GetDatasetRefName(e.name, e.namespace))).To(BeFalse())
		})
	})
})
