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
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"gopkg.in/yaml.v3"
	utilpointer "k8s.io/utils/pointer"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestInjectPod(t *testing.T) {
	type runtimeInfo struct {
		name        string
		namespace   string
		runtimeType string
	}
	type testCase struct {
		name    string
		in      *corev1.Pod
		dataset *datav1alpha1.Dataset
		pv      *corev1.PersistentVolume
		pvc     *corev1.PersistentVolumeClaim
		fuse    *appsv1.DaemonSet
		infos   map[string]runtimeInfo
		want    *corev1.Pod
		wantErr error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
	bTrue := true
	var mode int32 = 0755

	testcases := []testCase{
		{
			name: "inject_pod_with_duplicate_volumemount_name",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.FluidPath: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								common.MountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate",
				},
			},
			in: &corev1.Pod{
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
			fuse: &appsv1.DaemonSet{
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
									}, VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "duplicate",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
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
									}},
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
			},
			infos: map[string]runtimeInfo{
				"duplicate": {
					name:        "duplicate",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName,
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fluid-ate",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount",
									MountPath: "/jfs",
								}, {
									Name:      "check-mount",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						}, {
							Image: "duplicate-pvc-name",
							Name:  "duplicate-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "duplicate",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
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
						}, {
							Name: "fluid-ate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "check-mount",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "duplicate-jindo-check-mount",
									},
									DefaultMode: utilpointer.Int32Ptr(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_success",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset1",
					Namespace: "big-data",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset1",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			}, pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-dataset1",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.FluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								common.MountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset1",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-dataset1",
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset1-jindofs-fuse",
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
									Image:   "test",
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "data",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "data",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/runtime_mnt/dataset1",
										},
									}},
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
											Path: "/runtime-mnt/jindo/big-data/dataset1",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"dataset1": {
					name:        "dataset1",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName,
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount",
									MountPath: "/jfs",
								}, {
									Name:      "check-mount",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							}, Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo >> /proc/1/fd/1",
										},
									},
								},
							},
						}, {
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "dataset",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
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
									Path: "/runtime-mnt/jindo/big-data/dataset1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime_mnt/dataset1",
								},
							}}, {
							Name: "check-mount",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "dataset1-jindo-check-mount",
									},
									DefaultMode: utilpointer.Int32Ptr(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		}, {
			name: "inject_pod_with_customizedenv_volumemount_name",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customizedenv",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-customizedenv",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.FluidPath: "/runtime-mnt/jindo/big-data/customizedenv/jindofs-fuse",
								common.MountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customizedenv",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-customizedenv",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customizedenv-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "customizedenv-pvc-name",
							Name:  "customizedenv-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "customizedenv",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "customizedenv",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "customizedenv",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customizedenv-jindofs-fuse",
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
									Image:   "customizedenv-pvc-name",
									Env: []corev1.EnvVar{
										{
											Name:  "FLUID_FUSE_MOUNTPOINT",
											Value: "/jfs/jindofs-fuse",
										},
									},
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									}, VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "customizedenv",
											MountPath: "/mnt/disk1",
										}, {
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
											Name:      "jindofs-fuse-mount",
											MountPath: "/jfs",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "customizedenv",
									VolumeSource: corev1.VolumeSource{
										HostPath: &corev1.HostPathVolumeSource{
											Path: "/mnt/disk1",
											Type: &hostPathDirectoryOrCreate,
										},
									}},
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
											Path: "/runtime-mnt/jindo/big-data/customizedenv",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"customizedenv": {
					name:        "customizedenv",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "customizedenv-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName,
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs/jindofs-fuse jindo >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "customizedenv-pvc-name",
							Env: []corev1.EnvVar{
								{
									Name:  "FLUID_FUSE_MOUNTPOINT",
									Value: "/jfs/jindofs-fuse",
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fluid-izedenv",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount",
									MountPath: "/jfs",
								}, {
									Name:      "check-mount",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						}, {
							Image: "customizedenv-pvc-name",
							Name:  "customizedenv-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "customizedenv",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "customizedenv",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/customizedenv/jindofs-fuse",
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
									Path: "/runtime-mnt/jindo/big-data/customizedenv",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "fluid-izedenv",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "check-mount",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "customizedenv-jindo-check-mount",
									},
									DefaultMode: utilpointer.Int32Ptr(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		}, {
			name: "inject_pod_with_fuse_sidecar",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-fuse-sidecar",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.FluidPath: "/runtime-mnt/jindo/big-data/fuse-sidecar/jindofs-fuse",
								common.MountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-fuse-sidecar",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "fuse-sidecar-pvc-name",
							Name:  common.FuseContainerName,
						}, {
							Image: "fuse-sidecar-pvc-name",
							Name:  "fuse-sidecar-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fuse-sidecar",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "fuse-sidecar",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "fuse-sidecar",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-jindofs-fuse",
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
									Image:   "fuse-sidecar-pvc-name",
									Env: []corev1.EnvVar{
										{
											Name:  "FLUID_FUSE_MOUNTPOINT",
											Value: "/jfs/jindofs-fuse",
										},
									},
									SecurityContext: &corev1.SecurityContext{
										Privileged: &bTrue,
									}, VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "fuse-device",
											MountPath: "/dev/fuse",
										}, {
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
											Path: "/runtime-mnt/jindo/big-data/fuse-sidecar",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"fuse-sidecar": {
					name:        "fuse-sidecar",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "fuse-sidecar-pvc-name",
							Name:  common.FuseContainerName,
						}, {
							Image: "fuse-sidecar-pvc-name",
							Name:  "fuse-sidecar-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "fuse-sidecar",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "fuse-sidecar",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/fuse-sidecar/jindofs-fuse",
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
									Path: "/runtime-mnt/jindo/big-data/fuse-sidecar",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "check-mount",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "fuse-sidecar-jindo-check-mount",
									},
									DefaultMode: utilpointer.Int32Ptr(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	for _, testcase := range testcases {
		objs = append(objs, testcase.fuse, testcase.pv, testcase.pvc, testcase.dataset)
	}

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	for _, testcase := range testcases {
		injector := NewInjector(fakeClient)

		runtimeInfos := map[string]base.RuntimeInfoInterface{}
		for pvc, info := range testcase.infos {
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType, datav1alpha1.TieredStore{})
			if err != nil {
				t.Errorf("testcase %s failed due to error %v", testcase.name, err)
			}
			runtimeInfo.SetClient(fakeClient)
			runtimeInfos[pvc] = runtimeInfo
		}

		out, err := injector.InjectPod(testcase.in, runtimeInfos)
		if err != nil {
			if testcase.wantErr == nil {
				t.Errorf("testcase %s failed, Got error %v", testcase.name, err)
			} else {
				continue
			}
		}

		gotMetaObj := out.ObjectMeta
		wantMetaObj := testcase.want.ObjectMeta

		if !reflect.DeepEqual(gotMetaObj, wantMetaObj) {

			want, err := yaml.Marshal(wantMetaObj)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			outYaml, err := yaml.Marshal(gotMetaObj)
			if err != nil {
				t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
			}

			t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
		}

		gotContainers := out.Spec.Containers
		gotVolumes := out.Spec.Volumes
		// gotContainers := out.
		// , gotVolumes, err := getInjectPiece(out)
		// if err != nil {
		// 	t.Errorf("testcase %s failed due to inject error %v", testcase.name, err)
		// }

		wantContainers := testcase.want.Spec.Containers
		wantVolumes := testcase.want.Spec.Volumes

		gotContainerMap := makeContainerMap(gotContainers)
		wantContainerMap := makeContainerMap(wantContainers)

		if len(gotContainerMap) != len(wantContainerMap) {
			t.Errorf("testcase %s failed, want containers length %d, Got containers length  %d", testcase.name, len(gotContainerMap), len(wantContainerMap))
		}

		for k, wantContainer := range wantContainerMap {
			if gotContainer, found := gotContainerMap[k]; found {
				if !reflect.DeepEqual(wantContainer, gotContainer) {
					want, err := yaml.Marshal(wantContainers)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}

					outYaml, err := yaml.Marshal(gotContainers)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}

					t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the container %s", testcase.name, k)
			}
		}

		gotVolumeMap := makeVolumeMap(gotVolumes)
		wantVolumeMap := makeVolumeMap(wantVolumes)
		if len(gotVolumeMap) != len(wantVolumeMap) {
			gotVolumeKeys := keys(gotVolumeMap)
			wantVolumeKeys := keys(wantVolumeMap)
			t.Errorf("testcase %s failed, got volumes length %d with keys %v, want volumes length  %d with keys %v", testcase.name, len(gotVolumeMap),
				gotVolumeKeys, len(wantVolumeMap), wantVolumeKeys)
		}

		for k, wantVolume := range wantVolumeMap {
			if gotVolume, found := gotVolumeMap[k]; found {
				if !reflect.DeepEqual(wantVolume, gotVolume) {
					want, err := yaml.Marshal(wantVolume)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}

					outYaml, err := yaml.Marshal(gotVolume)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}

					t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the volume %s", testcase.name, k)
			}
		}

		// if !reflect.DeepEqual(gotVolumeMap, wantVolumeMap) {
		// 	want, err := yaml.Marshal(wantVolumes)
		// 	if err != nil {
		// 		t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
		// 	}

		// 	outYaml, err := yaml.Marshal(gotVolumes)
		// 	if err != nil {
		// 		t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
		// 	}

		// 	t.Errorf("testcase %s failed, want %v, Got  %v", testcase.name, string(want), string(outYaml))
		// }

	}
}

func makeContainerMap(containers []corev1.Container) (containerMap map[string]corev1.Container) {
	containerMap = map[string]corev1.Container{}
	for _, c := range containers {
		containerMap[c.Name] = c
	}
	return
}

func makeVolumeMap(volumes []corev1.Volume) (volumeMap map[string]corev1.Volume) {
	volumeMap = map[string]corev1.Volume{}
	for _, v := range volumes {
		volumeMap[v.Name] = v
	}
	return
}

func keys(vMap interface{}) (keys []string) {
	switch v := vMap.(type) {
	case map[string]corev1.Volume:
		for k := range v {
			keys = append(keys, k)
		}
	case map[string]corev1.Container:
		for k := range v {
			keys = append(keys, k)
		}
	}

	return
}
