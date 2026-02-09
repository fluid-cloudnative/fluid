/*
Copyright 2021 The Fluid Authors.

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

package base

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("RuntimeHelper", func() {

	Describe("getFuseDaemonset", func() {
		var (
			runtimeInfo RuntimeInfo
			scheme      *runtime.Scheme
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(scheme)).To(Succeed())
			Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		})

		Context("when the Alluxio runtime fuse daemonset exists", func() {
			BeforeEach(func() {
				ds := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "alluxio-fuse",
						Namespace: "default",
					},
				}
				fakeClient := fake.NewFakeClientWithScheme(scheme, ds)
				runtimeInfo = RuntimeInfo{
					name:        "alluxio",
					namespace:   "default",
					runtimeType: common.AlluxioRuntime,
				}
				runtimeInfo.SetAPIReader(fakeClient)
			})

			It("should retrieve the fuse daemonset successfully", func() {
				ds, err := runtimeInfo.getFuseDaemonset()
				Expect(err).NotTo(HaveOccurred())
				Expect(ds).NotTo(BeNil())
				Expect(ds.Name).To(Equal("alluxio-fuse"))
			})
		})

		Context("when the Jindo runtime fuse daemonset exists", func() {
			BeforeEach(func() {
				ds := &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "jindo-jindofs-fuse",
						Namespace: "default",
					},
				}
				fakeClient := fake.NewFakeClientWithScheme(scheme, ds)
				runtimeInfo = RuntimeInfo{
					name:        "jindo",
					namespace:   "default",
					runtimeType: common.JindoRuntime,
				}
				runtimeInfo.SetAPIReader(fakeClient)
			})

			It("should retrieve the fuse daemonset successfully", func() {
				ds, err := runtimeInfo.getFuseDaemonset()
				Expect(err).NotTo(HaveOccurred())
				Expect(ds).NotTo(BeNil())
				Expect(ds.Name).To(Equal("jindo-jindofs-fuse"))
			})
		})

		Context("when no API client is set", func() {
			BeforeEach(func() {
				runtimeInfo = RuntimeInfo{
					name:        "noclient",
					namespace:   "default",
					runtimeType: common.JindoRuntime,
				}
			})

			It("should return an error", func() {
				ds, err := runtimeInfo.getFuseDaemonset()
				Expect(err).To(HaveOccurred())
				Expect(ds).To(BeNil())
			})
		})
	})

	Describe("getMountInfo", func() {
		const testNamespace = "default"

		var (
			runtimeInfo RuntimeInfo
			scheme      *runtime.Scheme
			objs        []runtime.Object
		)

		BeforeEach(func() {
			scheme = runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())

			objs = []runtime.Object{
				&corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fluid-dataset",
						Namespace: testNamespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "default-fluid-dataset",
					},
				},
				&corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "nonfluidpvc",
						Namespace:   testNamespace,
						Annotations: common.GetExpectedFluidAnnotations(),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "nonfluidpv",
					},
				},
				&corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "nopv",
						Namespace:   testNamespace,
						Annotations: common.GetExpectedFluidAnnotations(),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "nopv",
					},
				},
				&corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "fluid-dataset-subpath",
						Namespace:   testNamespace,
						Annotations: common.GetExpectedFluidAnnotations(),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "default-fluid-dataset-subpath",
					},
				},
				&corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{Name: "default-fluid-dataset"},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
				&corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{Name: "nonfluidpv", Annotations: common.GetExpectedFluidAnnotations()},
					Spec:       corev1.PersistentVolumeSpec{},
				},
				&corev1.PersistentVolume{
					ObjectMeta: metav1.ObjectMeta{Name: "default-fluid-dataset-subpath"},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath:    "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
									common.VolumeAttrMountType:    common.JindoRuntime,
									common.VolumeAttrFluidSubPath: "subtest",
								},
							},
						},
					},
				},
			}
		})

		Context("when the volume claim does not exist", func() {
			BeforeEach(func() {
				fakeClient := fake.NewFakeClientWithScheme(scheme, objs...)
				runtimeInfo = RuntimeInfo{
					name:        "notExist",
					namespace:   testNamespace,
					runtimeType: common.JindoRuntime,
					apiReader:   fakeClient,
				}
			})

			It("should return an error", func() {
				path, mountType, subpath, err := runtimeInfo.getMountInfo()
				Expect(err).To(HaveOccurred())
				Expect(path).To(BeEmpty())
				Expect(mountType).To(BeEmpty())
				Expect(subpath).To(BeEmpty())
			})
		})

		Context("when the PV is not a Fluid PV", func() {
			BeforeEach(func() {
				fakeClient := fake.NewFakeClientWithScheme(scheme, objs...)
				runtimeInfo = RuntimeInfo{
					name:        "nonfluidpvc",
					namespace:   testNamespace,
					runtimeType: common.JindoRuntime,
					apiReader:   fakeClient,
				}
			})

			It("should return an error", func() {
				path, mountType, subpath, err := runtimeInfo.getMountInfo()
				Expect(err).To(HaveOccurred())
				Expect(path).To(BeEmpty())
				Expect(mountType).To(BeEmpty())
				Expect(subpath).To(BeEmpty())
			})
		})

		Context("when the PV is a valid Fluid PV", func() {
			BeforeEach(func() {
				fakeClient := fake.NewFakeClientWithScheme(scheme, objs...)
				runtimeInfo = RuntimeInfo{
					name:        "fluid-dataset",
					namespace:   testNamespace,
					runtimeType: common.JindoRuntime,
					apiReader:   fakeClient,
				}
			})

			It("should return the correct mount info", func() {
				path, mountType, subpath, err := runtimeInfo.getMountInfo()
				Expect(err).NotTo(HaveOccurred())
				Expect(path).To(Equal("/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse"))
				Expect(mountType).To(Equal(common.JindoRuntime))
				Expect(subpath).To(BeEmpty())
			})
		})

		Context("when the PVC has no bound PV", func() {
			BeforeEach(func() {
				fakeClient := fake.NewFakeClientWithScheme(scheme, objs...)
				runtimeInfo = RuntimeInfo{
					name:        "nopv",
					namespace:   testNamespace,
					runtimeType: common.JindoRuntime,
					apiReader:   fakeClient,
				}
			})

			It("should return an error", func() {
				path, mountType, subpath, err := runtimeInfo.getMountInfo()
				Expect(err).To(HaveOccurred())
				Expect(path).To(BeEmpty())
				Expect(mountType).To(BeEmpty())
				Expect(subpath).To(BeEmpty())
			})
		})

		Context("when the PV has a subpath configured", func() {
			BeforeEach(func() {
				fakeClient := fake.NewFakeClientWithScheme(scheme, objs...)
				runtimeInfo = RuntimeInfo{
					name:        "fluid-dataset-subpath",
					namespace:   testNamespace,
					runtimeType: common.JindoRuntime,
					apiReader:   fakeClient,
				}
			})

			It("should return the correct mount info with subpath", func() {
				path, mountType, subpath, err := runtimeInfo.getMountInfo()
				Expect(err).NotTo(HaveOccurred())
				Expect(path).To(Equal("/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse"))
				Expect(mountType).To(Equal(common.JindoRuntime))
				Expect(subpath).To(Equal("subtest"))
			})
		})
	})
})
