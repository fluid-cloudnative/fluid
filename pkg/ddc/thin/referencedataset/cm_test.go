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
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ConfigMap Operations", func() {
	Describe("copyFuseDaemonSetForRefDataset", func() {
		var (
			testScheme *runtime.Scheme
			testObjs   []runtime.Object
			fakeClient client.Client
		)

		BeforeEach(func() {
			testScheme = runtime.NewScheme()
			Expect(corev1.AddToScheme(testScheme)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())
			Expect(appsv1.AddToScheme(testScheme)).To(Succeed())

			testObjs = []runtime.Object{}
		})

		Context("when copying fuse daemonset for AlluxioRuntime", func() {
			It("should copy daemonset successfully", func() {
				sourceDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "alluxio-fuse",
						Namespace: "source-ns",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "fuse",
										Image: "alluxio:latest",
									},
								},
							},
						},
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, sourceDaemonSet)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				runtimeInfo, err := base.BuildRuntimeInfo("alluxio", "source-ns", common.AlluxioRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = copyFuseDaemonSetForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())

				var dsList appsv1.DaemonSetList
				err = fakeClient.List(context.TODO(), &dsList, client.InNamespace("ref-ns"))
				Expect(err).NotTo(HaveOccurred())
				Expect(dsList.Items).To(HaveLen(1))
				Expect(dsList.Items[0].Name).To(Equal("ref-dataset-fuse"))
				Expect(dsList.Items[0].Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("fluid.io/fuse-balloon", "true"))
			})
		})

		Context("when copying fuse daemonset for JindoRuntime", func() {
			It("should copy daemonset with correct naming convention", func() {
				sourceDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jindo-jindofs-fuse",
						Namespace: "source-ns",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "fuse",
										Image: "jindo:latest",
									},
								},
								NodeSelector: map[string]string{
									"existing-key": "existing-value",
								},
							},
						},
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, sourceDaemonSet)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				runtimeInfo, err := base.BuildRuntimeInfo("jindo", "source-ns", common.JindoRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = copyFuseDaemonSetForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())

				var dsList appsv1.DaemonSetList
				err = fakeClient.List(context.TODO(), &dsList, client.InNamespace("ref-ns"))
				Expect(err).NotTo(HaveOccurred())
				Expect(dsList.Items).To(HaveLen(1))
				Expect(dsList.Items[0].Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("fluid.io/fuse-balloon", "true"))
				Expect(dsList.Items[0].Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("existing-key", "existing-value"))
			})
		})

		Context("when daemonset already exists", func() {
			It("should not return error", func() {
				sourceDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "alluxio-fuse",
						Namespace: "source-ns",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "fuse",
									},
								},
							},
						},
					},
				}

				existingDaemonSet := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset-fuse",
						Namespace: "ref-ns",
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, sourceDaemonSet, existingDaemonSet)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				runtimeInfo, err := base.BuildRuntimeInfo("alluxio", "source-ns", common.AlluxioRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = copyFuseDaemonSetForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when source daemonset does not exist", func() {
			It("should return error", func() {
				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
				}

				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)

				runtimeInfo, err := base.BuildRuntimeInfo("alluxio", "source-ns", common.AlluxioRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = copyFuseDaemonSetForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("createConfigMapForRefDataset", func() {
		var (
			testScheme *runtime.Scheme
			testObjs   []runtime.Object
			fakeClient client.Client
			engine     *ReferenceDatasetEngine
		)

		BeforeEach(func() {
			testScheme = runtime.NewScheme()
			Expect(corev1.AddToScheme(testScheme)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(testScheme)).To(Succeed())

			testObjs = []runtime.Object{}
			engine = &ReferenceDatasetEngine{
				Log:       fake.NullLogger(),
				name:      "test-engine",
				namespace: "test-ns",
			}
		})

		Context("when physical runtime is AlluxioRuntime", func() {
			It("should copy config configmap successfully", func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "alluxio-config",
						Namespace: "source-ns",
					},
					Data: map[string]string{
						"config": "test-config",
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, configMap)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("alluxio", "source-ns", common.AlluxioRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())

				var cmList corev1.ConfigMapList
				err = fakeClient.List(context.TODO(), &cmList, client.InNamespace("ref-ns"))
				Expect(err).NotTo(HaveOccurred())
				Expect(cmList.Items).To(HaveLen(1))
				Expect(cmList.Items[0].Name).To(Equal("alluxio-config"))
			})
		})

		Context("when physical runtime is JuiceFSRuntime", func() {
			It("should copy fuse-script configmap successfully", func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "juicefs-fuse-script",
						Namespace: "source-ns",
					},
					Data: map[string]string{
						"script": "test-script",
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, configMap)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("juicefs", "source-ns", common.JuiceFSRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())

				var cmList corev1.ConfigMapList
				err = fakeClient.List(context.TODO(), &cmList, client.InNamespace("ref-ns"))
				Expect(err).NotTo(HaveOccurred())
				Expect(cmList.Items).To(HaveLen(1))
			})
		})

		Context("when physical runtime is GooseFSRuntime", func() {
			It("should copy config configmap successfully", func() {
				configMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "goosefs-config",
						Namespace: "source-ns",
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, configMap)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("goosefs", "source-ns", common.GooseFSRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())

				var cmList corev1.ConfigMapList
				err = fakeClient.List(context.TODO(), &cmList, client.InNamespace("ref-ns"))
				Expect(err).NotTo(HaveOccurred())
				Expect(cmList.Items).To(HaveLen(1))
			})
		})

		Context("when physical runtime is JindoRuntime", func() {
			It("should copy both client and jindofs configmaps successfully", func() {
				clientConfigMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jindo-jindofs-client-config",
						Namespace: "source-ns",
					},
				}

				jindoConfigMap := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jindo-jindofs-config",
						Namespace: "source-ns",
					},
				}

				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
						UID:       types.UID("test-uid"),
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "data.fluid.io/v1alpha1",
						Kind:       "Dataset",
					},
				}

				testObjs = append(testObjs, clientConfigMap, jindoConfigMap)
				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("jindo", "source-ns", common.JindoRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())

				var cmList corev1.ConfigMapList
				err = fakeClient.List(context.TODO(), &cmList, client.InNamespace("ref-ns"))
				Expect(err).NotTo(HaveOccurred())
				Expect(cmList.Items).To(HaveLen(2))
			})
		})

		Context("when physical runtime is EFCRuntime", func() {
			It("should skip and not return error", func() {
				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
					},
				}

				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("efc", "source-ns", common.EFCRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when physical runtime is ThinRuntime", func() {
			It("should skip and not return error", func() {
				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
					},
				}

				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("thin", "source-ns", common.ThinRuntime)
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when physical runtime type is unsupported", func() {
			It("should return error", func() {
				refDataset := &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-dataset",
						Namespace: "ref-ns",
					},
				}

				fakeClient = fake.NewFakeClientWithScheme(testScheme, testObjs...)
				engine.Client = fakeClient

				runtimeInfo, err := base.BuildRuntimeInfo("unknown", "source-ns", "UnknownRuntime")
				Expect(err).NotTo(HaveOccurred())

				err = engine.createConfigMapForRefDataset(fakeClient, refDataset, runtimeInfo)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fail to get configmap for runtime type"))
			})
		})
	})
})
