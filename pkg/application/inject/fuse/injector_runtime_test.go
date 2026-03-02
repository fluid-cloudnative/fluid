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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

var _ = Describe("Fuse Injector", func() {
	var (
		hostPathCharDev           = corev1.HostPathCharDev
		hostPathDirectoryOrCreate = corev1.HostPathDirectoryOrCreate
		bTrue                     = true
	)

	Describe("InjectList", func() {
		var (
			fakeClient   client.Client
			injector     *Injector
			runtimeInfos map[string]base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			objs := []runtime.Object{}
			s := runtime.NewScheme()
			Expect(corev1.AddToScheme(s)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
			Expect(appsv1.AddToScheme(s)).To(Succeed())

			fakeClient = fake.NewFakeClientWithScheme(s, objs...)
			injector = NewInjector(fakeClient)
			runtimeInfos = make(map[string]base.RuntimeInfoInterface)
		})

		Context("when injecting fuse sidecar into pod list", func() {
			var (
				dataset *datav1alpha1.Dataset
				pv      *corev1.PersistentVolume
				pvc     *corev1.PersistentVolumeClaim
				fuse    *appsv1.DaemonSet
				pods    []corev1.Pod
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

				pods = []corev1.Pod{
					{
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
					},
				}

				// Recreate fake client with objects
				objs := []runtime.Object{fuse, pv, pvc, dataset}
				s := runtime.NewScheme()
				Expect(corev1.AddToScheme(s)).To(Succeed())
				Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
				Expect(appsv1.AddToScheme(s)).To(Succeed())
				fakeClient = fake.NewFakeClientWithScheme(s, objs...)
				injector = NewInjector(fakeClient)

				// Setup runtime info
				runtimeInfo, err := base.BuildRuntimeInfo("duplicate", "big-data", common.JindoRuntime)
				Expect(err).NotTo(HaveOccurred())
				runtimeInfo.SetAPIReader(fakeClient)
				runtimeInfos["duplicate"] = runtimeInfo
			})

			It("should successfully inject without error", func() {
				podList := &corev1.List{}
				for _, pod := range pods {
					raw, err := json.Marshal(&pod)
					Expect(err).NotTo(HaveOccurred())
					podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})
				}

				_, err := injector.Inject(podList, runtimeInfos)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("InjectUnstructured", func() {
		var (
			fakeClient   client.Client
			injector     *Injector
			runtimeInfos map[string]base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			objs := []runtime.Object{}
			s := runtime.NewScheme()
			Expect(corev1.AddToScheme(s)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
			Expect(appsv1.AddToScheme(s)).To(Succeed())

			fakeClient = fake.NewFakeClientWithScheme(s, objs...)
			injector = NewInjector(fakeClient)
			runtimeInfos = make(map[string]base.RuntimeInfoInterface)

			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetAPIReader(fakeClient)
			runtimeInfos["test"] = runtimeInfo
		})

		Context("when injecting into unstructured object", func() {
			It("should return not implemented error", func() {
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
	})

	Describe("InjectObject", func() {
		var (
			fakeClient   client.Client
			injector     *Injector
			runtimeInfos map[string]base.RuntimeInfoInterface
		)

		BeforeEach(func() {
			objs := []runtime.Object{}
			s := runtime.NewScheme()
			Expect(corev1.AddToScheme(s)).To(Succeed())
			Expect(datav1alpha1.AddToScheme(s)).To(Succeed())
			Expect(appsv1.AddToScheme(s)).To(Succeed())

			fakeClient = fake.NewFakeClientWithScheme(s, objs...)
			injector = NewInjector(fakeClient)
			runtimeInfos = make(map[string]base.RuntimeInfoInterface)

			runtimeInfo, err := base.BuildRuntimeInfo("test", "default", common.JindoRuntime)
			Expect(err).NotTo(HaveOccurred())
			runtimeInfo.SetAPIReader(fakeClient)
			runtimeInfos["test"] = runtimeInfo
		})

		Context("when injecting into deployment object", func() {
			It("should return not implemented error", func() {
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
	})
})
