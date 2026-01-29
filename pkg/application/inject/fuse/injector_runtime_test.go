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

package fuse

import (
	"encoding/json"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inject List Tests", Label("pkg.application.inject.fuse.inject_list_test.go"), func() {
	var (
		injector                  *Injector
		fakeClient                client.Client
		runtimeObjs               []runtime.Object
		hostPathCharDev           = corev1.HostPathCharDev
		hostPathDirectoryOrCreate = corev1.HostPathDirectoryOrCreate
		bTrue                     = true
	)

	BeforeEach(func() {
		runtimeObjs = []runtime.Object{}
	})

	Context("InjectList with duplicate PVC name", func() {
		var (
			dataset      *datav1alpha1.Dataset
			pv           *corev1.PersistentVolume
			pvc          *corev1.PersistentVolumeClaim
			fuse         *appsv1.DaemonSet
			podList      *corev1.List
			runtimeInfos map[string]base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			dataset = &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate",
					Namespace: "big-data",
				},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate",
					Namespace: "big-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate",
				},
			}

			fuse = &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
									},
									Command: []string{"/entrypoint.sh"},
									Image:   "duplicate-pvc-name",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "duplicate",
											MountPath: "/mnt/disk1",
										},
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "duplicate",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/mnt/disk1",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/big-data/duplicate",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "duplicate",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			runtimeObjs = append(runtimeObjs, fuse, pv, pvc, dataset)
			fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, runtimeObjs...)

			runtimeInfos = map[string]base.RuntimeInfoInterface{}
			runtimeInfo, err := base.BuildRuntimeInfo("duplicate", "big-data", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetAPIReader(fakeClient)
			runtimeInfos["duplicate"] = runtimeInfo

			podList = &corev1.List{}
			raw, err := json.Marshal(&pod)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})

			injector = NewInjector(fakeClient)
		})

		It("should inject the pod list successfully", func() {
			out, err := injector.Inject(podList, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
		})
	})

	Context("InjectList with multiple pods", func() {
		var (
			dataset      *datav1alpha1.Dataset
			pv           *corev1.PersistentVolume
			pvc          *corev1.PersistentVolumeClaim
			fuse         *appsv1.DaemonSet
			podList      *corev1.List
			runtimeInfos map[string]base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			dataset = &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "default",
				},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default-test-dataset",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/thin/default/test-dataset/thin-fuse",
								common.VolumeAttrMountType: common.ThinRuntime,
							},
						},
					},
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "default",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "default-test-dataset",
				},
			}

			fuse = &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset-fuse",
					Namespace: "default",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "fuse",
									Image:   "test-image",
									Command: []string{"/entrypoint.sh"},
									SecurityContext: &corev1.SecurityContext{
										Privileged: ptr.To(true),
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
							},
						},
					},
				},
			}

			pod1 := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-1",
					Namespace: "default",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test-image",
							Name:  "test-container",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "test-dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-dataset",
								},
							},
						},
					},
				},
			}

			pod2 := pod1.DeepCopy()
			pod2.Name = "test-pod-2"

			runtimeObjs = append(runtimeObjs, fuse, pv, pvc, dataset)
			fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, runtimeObjs...)

			runtimeInfos = map[string]base.RuntimeInfoInterface{}
			runtimeInfo, err := base.BuildRuntimeInfo("test-dataset", "default", common.ThinRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetAPIReader(fakeClient)
			runtimeInfos["test-dataset"] = runtimeInfo

			podList = &corev1.List{}
			raw1, err := json.Marshal(&pod1)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw1})

			raw2, err := json.Marshal(pod2)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw2})

			injector = NewInjector(fakeClient)
		})
	})

	Context("InjectList with invalid JSON", func() {
		It("should return error when pod JSON is invalid", func() {
			podList := &corev1.List{
				Items: []runtime.RawExtension{
					{Raw: []byte("invalid json")},
				},
			}

			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme)
			injector = NewInjector(fakeClient)

			_, err := injector.Inject(podList, runtimeInfos)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("InjectList with empty runtime infos", func() {
		var (
			podList *corev1.List
		)

		BeforeEach(func() {
			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test-image",
							Name:  "test-container",
						},
					},
				},
			}

			podList = &corev1.List{}
			raw, err := json.Marshal(&pod)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})

			fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme)
			injector = NewInjector(fakeClient)
		})

		It("should handle empty runtime infos", func() {
			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			out, err := injector.Inject(podList, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
		})
	})
})

var _ = Describe("Inject Unstructured Tests", Label("pkg.application.inject.fuse.inject_unstructured_test.go"), func() {
	var (
		injector     *Injector
		fakeClient   client.Client
		runtimeInfos map[string]base.RuntimeInfoInterface
	)

	BeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme)
		injector = NewInjector(fakeClient)

		runtimeInfos = map[string]base.RuntimeInfoInterface{}
		runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.JindoRuntime)
		Expect(err).NotTo(HaveOccurred())
		runtimeInfo.SetAPIReader(fakeClient)
		runtimeInfos["test"] = runtimeInfo
	})

	Context("InjectUnstructured with Pod", func() {
		It("should return 'not implemented' error", func() {
			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "duplicate",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}

			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
			Expect(err).NotTo(HaveOccurred())

			_, err = injector.Inject(&unstructured.Unstructured{Object: object}, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})
	})

	Context("InjectUnstructured with various objects", func() {
		It("should return 'not implemented' error for StatefulSet", func() {
			sts := appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sts",
					Namespace: "default",
				},
			}

			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&sts)
			Expect(err).NotTo(HaveOccurred())

			_, err = injector.Inject(&unstructured.Unstructured{Object: object}, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})

		It("should return 'not implemented' error for DaemonSet", func() {
			ds := appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ds",
					Namespace: "default",
				},
			}

			object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&ds)
			Expect(err).NotTo(HaveOccurred())

			_, err = injector.Inject(&unstructured.Unstructured{Object: object}, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})
	})
})

var _ = Describe("Inject Object Tests", Label("pkg.application.inject.fuse.inject_object_test.go"), func() {
	var (
		injector     *Injector
		fakeClient   client.Client
		runtimeInfos map[string]base.RuntimeInfoInterface
	)

	BeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme)
		injector = NewInjector(fakeClient)

		runtimeInfos = map[string]base.RuntimeInfoInterface{}
		runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.JindoRuntime)
		Expect(err).NotTo(HaveOccurred())
		runtimeInfo.SetAPIReader(fakeClient)
		runtimeInfos["test"] = runtimeInfo
	})

	Context("InjectObject with Deployment", func() {
		It("should return 'not implemented' error", func() {
			deploy := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "details-v1",
					Namespace: "default",
					Labels: map[string]string{
						"app": "details",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "details"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "details"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "duplicate-pvc-name",
									Name:  "duplicate-pvc-name",
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "duplicate",
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "duplicate",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: "duplicate",
											ReadOnly:  true,
										},
									},
								},
							},
						},
					},
				},
			}

			_, err := injector.Inject(&deploy, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})
	})

	Context("InjectObject with various Kubernetes objects", func() {
		It("should return 'not implemented' error for StatefulSet", func() {
			sts := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-sts",
					Namespace: "default",
				},
				Spec: appsv1.StatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}

			_, err := injector.Inject(sts, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})

		It("should return 'not implemented' error for DaemonSet", func() {
			ds := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ds",
					Namespace: "default",
				},
				Spec: appsv1.DaemonSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}

			_, err := injector.Inject(ds, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})

		It("should return 'not implemented' error for ReplicaSet", func() {
			rs := &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-rs",
					Namespace: "default",
				},
				Spec: appsv1.ReplicaSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "test",
									Image: "test:latest",
								},
							},
						},
					},
				},
			}

			_, err := injector.Inject(rs, runtimeInfos)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not implemented"))
		})
	})
})

var _ = Describe("Inject Error Handling Tests", Label("pkg.application.inject.fuse.inject_error_test.go"), func() {
	var (
		injector   *Injector
		fakeClient client.Client
	)

	BeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme)
		injector = NewInjector(fakeClient)
	})

	Context("BuildRuntimeInfo with different runtime types", func() {
		It("should build runtime info successfully for valid runtime types", func() {
			// BuildRuntimeInfo doesn't actually error for unknown runtime types
			// It just creates a generic runtime info
			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfo).NotTo(BeNil())
		})
	})

	Context("Inject with nil or empty objects", func() {
		It("should handle empty runtime infos without error", func() {
			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
						},
					},
				},
			}

			podList := &corev1.List{}
			raw, err := json.Marshal(&pod)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})

			// Empty runtime infos should work fine - no injection needed
			out, err := injector.Inject(podList, map[string]base.RuntimeInfoInterface{})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
		})
	})
})

var _ = Describe("Additional Coverage Tests", Label("pkg.application.inject.fuse.additional_coverage_test.go"), func() {
	var (
		injector   *Injector
		fakeClient client.Client
	)

	BeforeEach(func() {
		fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme)
		injector = NewInjector(fakeClient)
	})

	Context("Edge cases for different runtime types", func() {
		It("should work with AlluxioRuntime", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.AlluxioRuntime)
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfo).NotTo(BeNil())
		})

		It("should work with JuiceFSRuntime", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.JuiceFSRuntime)
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfo).NotTo(BeNil())
		})

		It("should work with GooseFSRuntime", func() {
			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.GooseFSRuntime)
			Expect(err).NotTo(HaveOccurred())
			Expect(runtimeInfo).NotTo(BeNil())
		})
	})

	Context("InjectList with different pod configurations", func() {
		var (
			dataset *datav1alpha1.Dataset
			pv      *corev1.PersistentVolume
			pvc     *corev1.PersistentVolumeClaim
			fuse    *appsv1.DaemonSet
		)

		BeforeEach(func() {
			dataset = &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "test-ns",
				},
			}

			pv = &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-ns-test-dataset",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/alluxio/test-ns/test-dataset/alluxio-fuse",
								common.VolumeAttrMountType: common.AlluxioRuntime,
							},
						},
					},
				},
			}

			pvc = &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset",
					Namespace: "test-ns",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "test-ns-test-dataset",
				},
			}

			hostPathCharDev := corev1.HostPathCharDev
			fuse = &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-dataset-alluxio-fuse",
					Namespace: "test-ns",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "fuse",
									Image:   "alluxio-image",
									Command: []string{"/entrypoint.sh"},
									SecurityContext: &corev1.SecurityContext{
										Privileged: ptr.To(true),
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
							},
						},
					},
				},
			}

			objs := []runtime.Object{dataset, pv, pvc, fuse}
			fakeClient = fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, objs...)
			injector = NewInjector(fakeClient)
		})

		It("should inject pod with no labels", func() {
			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-no-labels",
					Namespace: "test-ns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test-image",
							Name:  "test-container",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "test-dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "test-dataset",
								},
							},
						},
					},
				},
			}

			podList := &corev1.List{}
			raw, err := json.Marshal(&pod)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})

			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			runtimeInfo, err := base.BuildRuntimeInfo("test-dataset", "test-ns", common.AlluxioRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetAPIReader(fakeClient)
			runtimeInfos["test-dataset"] = runtimeInfo

			out, err := injector.Inject(podList, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
		})
	})

	Context("Testing Injector constructor", func() {
		It("should create a new Injector with valid client", func() {
			newInjector := NewInjector(fakeClient)
			Expect(newInjector).NotTo(BeNil())
			Expect(newInjector.client).To(Equal(fakeClient))
		})
	})

	Context("Complex injection scenarios", func() {
		It("should handle pod with both init and regular containers", func() {
			dataset := &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complex-dataset",
					Namespace: "complex-ns",
				},
			}

			pv := &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "complex-ns-complex-dataset",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/complex-ns/complex-dataset/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			}

			pvc := &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complex-dataset",
					Namespace: "complex-ns",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "complex-ns-complex-dataset",
				},
			}

			hostPathCharDev := corev1.HostPathCharDev
			hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
			fuse := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complex-dataset-jindofs-fuse",
					Namespace: "complex-ns",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:    "fuse",
									Image:   "jindo-image",
									Command: []string{"/entrypoint.sh"},
									Args:    []string{"-oroot_ns=jindo"},
									SecurityContext: &corev1.SecurityContext{
										Privileged: ptr.To(true),
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										},
										{
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "fuse-device",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/dev/fuse",
											Type: &hostPathCharDev,
										},
									},
								},
								{
									Name: "jindofs-fuse-mount",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime-mnt/jindo/complex-ns/complex-dataset",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			}

			objs := []runtime.Object{dataset, pv, pvc, fuse}
			complexClient := fake.NewFakeClientWithScheme(datav1alpha1.UnitTestScheme, objs...)
			complexInjector := NewInjector(complexClient)

			pod := corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "complex-pod",
					Namespace: "complex-ns",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "init-image",
							Name:  "init",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "complex-dataset",
									MountPath: "/init-data",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Image: "main-image",
							Name:  "main",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "complex-dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "complex-dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "complex-dataset",
								},
							},
						},
					},
				},
			}

			podList := &corev1.List{}
			raw, err := json.Marshal(&pod)
			Expect(err).NotTo(HaveOccurred())
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})

			runtimeInfos := map[string]base.RuntimeInfoInterface{}
			runtimeInfo, err := base.BuildRuntimeInfo("complex-dataset", "complex-ns", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetAPIReader(complexClient)
			runtimeInfos["complex-dataset"] = runtimeInfo

			out, err := complexInjector.Inject(podList, runtimeInfos)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
		})
	})
})
