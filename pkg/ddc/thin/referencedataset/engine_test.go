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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("BuildReferenceDatasetThinEngine", func() {
	var (
		testScheme *runtime.Scheme
		testObjs   []runtime.Object
		fakeClient client.Client

		dataset            *datav1alpha1.Dataset
		alluxioRuntime     *datav1alpha1.AlluxioRuntime
		refRuntime         *datav1alpha1.ThinRuntime
		refDataset         *datav1alpha1.Dataset
		multipleRefDataset *datav1alpha1.Dataset
		multipleRefRuntime *datav1alpha1.ThinRuntime
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1.AddToScheme(testScheme)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(appsv1.AddToScheme(testScheme)).To(Succeed())

		testObjs = []runtime.Object{}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done",
				Namespace: "big-data",
			},
			Status: datav1alpha1.DatasetStatus{
				Runtimes: []datav1alpha1.Runtime{
					{
						Name:      "done",
						Namespace: "big-data",
						Type:      common.AlluxioRuntime,
					},
				},
			},
		}

		alluxioRuntime = &datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done",
				Namespace: "big-data",
			},
		}

		refRuntime = &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		}

		refDataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "dataset://big-data/done",
					},
				},
			},
		}

		multipleRefDataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid-mul",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "dataset://big-data/done",
					},
					{
						MountPoint: "http://big-test/done",
					},
				},
			},
		}

		multipleRefRuntime = &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid-mul",
			},
		}

		testObjs = append(testObjs, dataset, refDataset, multipleRefDataset)
		testObjs = append(testObjs, alluxioRuntime, refRuntime, multipleRefRuntime)
		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when building reference dataset thin engine", func() {
		It("should succeed with valid reference dataset", func() {
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "hbase",
					Namespace: "fluid",
				},
				Client:      fakeClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.ThinRuntime,
				Runtime:     refRuntime,
			}

			engine, err := BuildReferenceDatasetThinEngine("success", ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(engine).NotTo(BeNil())
		})

		It("should fail when dataset is not a reference dataset", func() {
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "done",
					Namespace: "big-data",
				},
				Client:      fakeClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.ThinRuntime,
				Runtime:     alluxioRuntime,
			}

			_, err := BuildReferenceDatasetThinEngine("dataset-not-ref", ctx)
			Expect(err).To(HaveOccurred())
		})

		It("should fail when dataset has multiple mount points with different formats", func() {
			ctx := cruntime.ReconcileRequestContext{
				NamespacedName: types.NamespacedName{
					Name:      "hbase",
					Namespace: "fluid-mul",
				},
				Client:      fakeClient,
				Log:         fake.NullLogger(),
				RuntimeType: common.ThinRuntime,
				Runtime:     multipleRefRuntime,
			}

			_, err := BuildReferenceDatasetThinEngine("dataset-with-different-format", ctx)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("ReferenceDatasetEngine Setup", func() {
	var (
		testScheme *runtime.Scheme
		testObjs   []runtime.Object
		fakeClient client.Client

		dataset        *datav1alpha1.Dataset
		alluxioRuntime *datav1alpha1.AlluxioRuntime
		refRuntime     *datav1alpha1.ThinRuntime
		refDataset     *datav1alpha1.Dataset
		configCM       *v1.ConfigMap
		fuseDs         *appsv1.DaemonSet
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1.AddToScheme(testScheme)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(appsv1.AddToScheme(testScheme)).To(Succeed())

		testObjs = []runtime.Object{}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done",
				Namespace: "big-data",
			},
			Status: datav1alpha1.DatasetStatus{
				Runtimes: []datav1alpha1.Runtime{
					{
						Name:      "done",
						Namespace: "big-data",
						Type:      common.AlluxioRuntime,
					},
				},
				DatasetRef: []string{
					"fluid/test",
				},
			},
		}

		alluxioRuntime = &datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done",
				Namespace: "big-data",
			},
		}

		refRuntime = &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		}

		refDataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "dataset://big-data/done",
					},
				},
			},
		}

		configCM = &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      alluxioRuntime.Name + "-config",
				Namespace: alluxioRuntime.Namespace,
			},
			Data: map[string]string{
				"check.sh": "/bin/sh check",
			},
		}

		fuseDs = &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done-fuse",
				Namespace: "big-data",
			},
			Spec: appsv1.DaemonSetSpec{},
		}

		testObjs = append(testObjs, dataset, refDataset, configCM, alluxioRuntime, refRuntime, fuseDs)
		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when setting up reference dataset engine", func() {
		It("should setup successfully and update dataset references", func() {
			engine := &ReferenceDatasetEngine{
				Id:        "",
				Client:    fakeClient,
				Log:       logr.Logger{},
				name:      refRuntime.Name,
				namespace: refRuntime.Namespace,
			}

			ctx := cruntime.ReconcileRequestContext{
				Dataset: refDataset,
				Client:  fakeClient,
			}

			ready, err := engine.Setup(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(ready).To(BeTrue())

			// Verify dataset reference was added
			updatedDataset := &datav1alpha1.Dataset{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: alluxioRuntime.Namespace,
				Name:      alluxioRuntime.Name,
			}, updatedDataset)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataset.Status.DatasetRef).To(ContainElement(base.GetDatasetRefName(engine.name, engine.namespace)))

			// Verify configmap was copied
			cmList := &v1.ConfigMapList{}
			err = fakeClient.List(context.TODO(), cmList, &client.ListOptions{Namespace: engine.namespace})
			Expect(err).NotTo(HaveOccurred())
			Expect(cmList.Items).To(HaveLen(1))
		})
	})
})

var _ = Describe("ReferenceDatasetEngine Shutdown", func() {
	var (
		testScheme *runtime.Scheme
		testObjs   []runtime.Object
		fakeClient client.Client

		dataset        *datav1alpha1.Dataset
		alluxioRuntime *datav1alpha1.AlluxioRuntime
		refRuntime     *datav1alpha1.ThinRuntime
		refDataset     *datav1alpha1.Dataset
	)

	BeforeEach(func() {
		testScheme = runtime.NewScheme()
		Expect(v1.AddToScheme(testScheme)).To(Succeed())
		Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
		Expect(appsv1.AddToScheme(testScheme)).To(Succeed())

		testObjs = []runtime.Object{}

		dataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done",
				Namespace: "big-data",
			},
			Status: datav1alpha1.DatasetStatus{
				Runtimes: []datav1alpha1.Runtime{
					{
						Name:      "done",
						Namespace: "big-data",
						Type:      common.AlluxioRuntime,
					},
				},
				DatasetRef: []string{
					"fluid/hbase",
					"fluid/test",
				},
			},
		}

		alluxioRuntime = &datav1alpha1.AlluxioRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "done",
				Namespace: "big-data",
			},
		}

		refRuntime = &datav1alpha1.ThinRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
		}

		refDataset = &datav1alpha1.Dataset{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "hbase",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "dataset://big-data/done",
					},
				},
			},
		}

		testObjs = append(testObjs, dataset, refDataset)
		testObjs = append(testObjs, alluxioRuntime, refRuntime)
		fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
	})

	Context("when shutting down reference dataset engine", func() {
		It("should remove dataset reference successfully", func() {
			engine := &ReferenceDatasetEngine{
				Id:        "",
				Client:    fakeClient,
				Log:       logr.Logger{},
				name:      refRuntime.Name,
				namespace: refRuntime.Namespace,
			}

			err := engine.Shutdown()
			Expect(err).NotTo(HaveOccurred())

			// Verify dataset reference was removed
			updatedDataset := &datav1alpha1.Dataset{}
			err = fakeClient.Get(context.TODO(), types.NamespacedName{
				Namespace: engine.physicalRuntimeInfo.GetNamespace(),
				Name:      engine.physicalRuntimeInfo.GetName(),
			}, updatedDataset)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedDataset.Status.DatasetRef).NotTo(ContainElement(base.GetDatasetRefName(engine.name, engine.namespace)))
		})
	})
})
