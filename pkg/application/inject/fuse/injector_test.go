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

package fuse

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/google/go-cmp/cmp"
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
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "duplicate"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
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
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-dataset1",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset1"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							}, Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "data-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime_mnt/dataset1",
								},
							}}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
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
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/customizedenv/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "customizedenv"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
									Name:      "customizedenv-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/customizedenv",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "customizedenv-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_conflict_name",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset-conflict",
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
									Name:      "data-0",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-0",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset-conflict",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-dataset-conflict",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset-conflict/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset-conflict",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-dataset-conflict",
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset-conflict-jindofs-fuse",
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
											Path: "/runtime_mnt/dataset-conflict",
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
											Path: "/runtime-mnt/jindo/big-data/dataset-conflict",
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
				"dataset-conflict": {
					name:        "dataset-conflict",
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset-conflict"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fluid-",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							}, Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
						}, {
							Image: "test",
							Name:  "test",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "data-0",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset-conflict/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset-conflict",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "fluid-",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime_mnt/dataset-conflict",
								},
							}}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
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
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType)
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
			t.Errorf("testcase %s failed, diff between want and got is: %v", testcase.name, cmp.Diff(gotMetaObj, wantMetaObj))
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
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
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
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantVolume, gotVolume))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the volume %s", testcase.name, k)
			}
		}
	}
}

func TestSkipInjectPod(t *testing.T) {
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
		want    *corev1.Pod
		infos   map[string]runtimeInfo
		wantErr error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	bTrue := true

	testcases := []testCase{
		{
			name: "inject_pod_with_injection_done_label",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-2",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-fuse-sidecar-2",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/fuse-sidecar-2/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-2",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-fuse-sidecar-2",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-2-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "fuse-sidecar-2-pvc-name",
							Name:  common.FuseContainerName,
						}, {
							Image: "fuse-sidecar-2-pvc-name",
							Name:  "fuse-sidecar-2-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fuse-sidecar-2",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "fuse-sidecar-2",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "fuse-sidecar-2",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-2-jindofs-fuse",
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
									Image:   "fuse-sidecar-2-pvc-name",
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
											Path: "/runtime-mnt/jindo/big-data/fuse-sidecar-2",
											Type: &hostPathDirectoryOrCreate,
										},
									},
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fuse-sidecar-2-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						common.InjectSidecarDone: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "fuse-sidecar-2-pvc-name",
							Name:  common.FuseContainerName,
						}, {
							Image: "fuse-sidecar-2-pvc-name",
							Name:  "fuse-sidecar-2-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "fuse-sidecar-2",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "fuse-sidecar-2",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "fuse-sidecar-2",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"fuse-sidecar-2": {
					name:        "fuse-sidecar-2",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
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
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType)
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
			t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantMetaObj, gotMetaObj))
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
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
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
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantVolume, gotVolume))
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

func TestInjectPodWithMultiplePVC(t *testing.T) {
	type runtimeInfo struct {
		name        string
		namespace   string
		runtimeType string
	}
	type testCase struct {
		name        string
		in          *corev1.Pod
		dataset     []*datav1alpha1.Dataset
		pv          []*corev1.PersistentVolume
		pvc         []*corev1.PersistentVolumeClaim
		fuse        []*appsv1.DaemonSet
		infos       map[string]runtimeInfo
		numPvcMount int
		want        *corev1.Pod
		wantErr     error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
	bTrue := true
	var mode int32 = 0755

	testcases := []testCase{
		{
			name: "inject_pod_with_duplicate_pvc_volumes",
			dataset: []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "duplicate",
						Namespace: "big-data",
					},
				},
			},
			pv: []*corev1.PersistentVolume{
				{
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
				},
			},
			pvc: []*corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "duplicate",
						Namespace: "big-data",
					}, Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-duplicate",
					},
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
									Name:      "duplicate1",
									MountPath: "/data1",
								},
								{
									Name:      "duplicate2",
									MountPath: "/data2",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate1",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "duplicate",
									ReadOnly:  true,
								},
							},
						},
						{
							Name: "duplicate2",
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
			fuse: []*appsv1.DaemonSet{
				{
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
			},
			infos: map[string]runtimeInfo{
				"duplicate": {
					name:        "duplicate",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			numPvcMount: 1,
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "duplicate"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
									Name:             "duplicate1",
									MountPath:        "/data1",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:             "duplicate2",
									MountPath:        "/data2",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								},
							},
						},
						{
							Name: "duplicate2",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_multiple_pvc_volumes",
			dataset: []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset1",
						Namespace: "big-data",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset2",
						Namespace: "big-data",
					},
				},
			},
			pv: []*corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset1",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset2",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset2/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
			},
			pvc: []*corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset1",
						Namespace: "big-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset2",
						Namespace: "big-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset2",
					},
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multiple-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "multiple-pvc-name",
							Name:  "multiple-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset1",
									MountPath: "/data1",
								},
								{
									Name:      "dataset2",
									MountPath: "/data2",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset2",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset2",
									ReadOnly:  true,
								},
							},
						},
						{
							Name: "dataset1",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset1",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: []*appsv1.DaemonSet{
				{
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
										Image:   "multiple-pvc-name",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk1",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk1",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset2-jindofs-fuse",
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
										Image:   "multiple-pvc-name",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk1",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk1",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
												Path: "/runtime-mnt/jindo/big-data/dataset2",
												Type: &hostPathDirectoryOrCreate,
											},
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
				"dataset2": {
					name:        "dataset2",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			numPvcMount: 2,
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "multiple-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset1"),
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-1"): fmt.Sprintf("%s-%s", "big-data", "dataset2"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-1",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "multiple-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-1",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "jindofs-fuse-device-1",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-1",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-1",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "multiple-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "jindofs-fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Image: "multiple-pvc-name",
							Name:  "multiple-pvc-name",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "dataset1",
									MountPath:        "/data1",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:             "dataset2",
									MountPath:        "/data2",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								},
							},
						},
						{
							Name: "dataset2",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset2/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "jindofs-fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "cachedir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "jindofs-fuse-device-1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset2",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "cachedir-1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-1",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
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
		for _, obj := range testcase.fuse {
			objs = append(objs, obj)
		}
		for _, obj := range testcase.pv {
			objs = append(objs, obj)
		}
		for _, obj := range testcase.pvc {
			objs = append(objs, obj)
		}
		for _, obj := range testcase.dataset {
			objs = append(objs, obj)
		}
	}

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	for _, testcase := range testcases {
		injector := NewInjector(fakeClient)

		runtimeInfos := map[string]base.RuntimeInfoInterface{}
		for pvc, info := range testcase.infos {
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType)
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
			t.Errorf("testcase %s failed, diff between wantMetaObj and gotMetaObj: %v", testcase.name, cmp.Diff(wantMetaObj, gotMetaObj))
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
			t.Errorf("testcase %s failed, want containers length %d, Got containers length  %d", testcase.name, len(wantContainerMap), len(gotContainerMap))
		}

		for k, wantContainer := range wantContainerMap {
			if strings.HasPrefix(k, common.FuseContainerName) {
				var exists bool
				tempWant := wantContainer.DeepCopy()
				tempWant.Name = ""
				for _, gotContainer := range gotContainers {
					tempGot := gotContainer.DeepCopy()
					tempGot.Name = ""

					if reflect.DeepEqual(tempGot, tempWant) {
						exists = true
					}
				}

				if !exists {
					want, err := yaml.Marshal(wantContainer)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}
					t.Errorf("testcase %s failed, want container: %v, but not found in containers", testcase.name, string(want))
				}
			} else if gotContainer, found := gotContainerMap[k]; found {
				if !reflect.DeepEqual(wantContainer, gotContainer) {
					t.Errorf("testcase %s failed, diff between wantContainer and gotContainer: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
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

		wantVolumesTotal := len(testcase.in.Spec.Volumes) + testcase.numPvcMount + 1
		for _, injectedFuse := range testcase.fuse {
			for _, wantVolume := range injectedFuse.Spec.Template.Spec.Volumes {
				wantTemp := wantVolume.DeepCopy()
				wantTemp.Name = ""
				var exists bool
				for _, gotVolume := range gotVolumes {
					gotTemp := gotVolume.DeepCopy()
					gotTemp.Name = ""
					if reflect.DeepEqual(wantTemp, gotTemp) {
						exists = true
						break
					}
				}

				if !exists {
					want, err := yaml.Marshal(wantVolumes)
					if err != nil {
						t.Errorf("testcase %s failed due to %v", testcase.name, err)
					}
					t.Errorf("testcase %s failed, wantVolume: %s, but not found in gotVolumes", testcase.name, string(want))
				}
			}
			wantVolumesTotal += len(injectedFuse.Spec.Template.Spec.Volumes)
		}

		if wantVolumesTotal != len(gotVolumes) {
			t.Errorf("testcase %s failed, want volume length: %d, got volume length: %d", testcase.name, wantVolumesTotal, len(gotVolumes))
		}
	}
}

func TestInjectPodWithDatasetSubPath(t *testing.T) {
	type runtimeInfo struct {
		name        string
		namespace   string
		runtimeType string
	}
	type testCase struct {
		name           string
		in             *corev1.Pod
		dataset        *datav1alpha1.Dataset
		pv             *corev1.PersistentVolume
		pvc            *corev1.PersistentVolumeClaim
		subPathDataset *datav1alpha1.Dataset
		subPathPv      *corev1.PersistentVolume
		subPathPvc     *corev1.PersistentVolumeClaim
		fuse           *appsv1.DaemonSet
		infos          map[string]runtimeInfo
		want           *corev1.Pod
		wantErr        error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
	bTrue := true
	var mode int32 = 0755

	testcases := []testCase{
		{
			name: "inject_pod_with_subpath_dataset",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dataset1",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-dataset1",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
			subPathDataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subpath",
					Namespace: "ref",
				},
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{
						{
							MountPoint: "dataset://big-data/dataset1/path-a",
						},
					},
				},
			},
			subPathPv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ref-subpath",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath:    "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								common.VolumeAttrMountType:    common.JindoRuntime,
								common.VolumeAttrFluidSubPath: "path-a",
							},
						},
					},
				},
			},
			subPathPvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subpath",
					Namespace: "ref",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "ref-subpath",
				},
			},

			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subpath-fuse",
					Namespace: "ref",
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
				"subpath": {
					name:        "subpath",
					namespace:   "ref",
					runtimeType: common.ThinRuntime,
				},
			},

			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ref",
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
									ClaimName: "subpath",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "ref",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "ref", "subpath"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							}, Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo path-a >> /proc/1/fd/1",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse/path-a",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "data-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime_mnt/dataset1",
								},
							}}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
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
		objs = append(objs, testcase.fuse, testcase.pv, testcase.pvc, testcase.dataset, testcase.subPathDataset, testcase.subPathPv, testcase.subPathPvc)
	}

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	for _, testcase := range testcases {
		injector := NewInjector(fakeClient)

		runtimeInfos := map[string]base.RuntimeInfoInterface{}
		for pvc, info := range testcase.infos {
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType)
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
			t.Errorf("testcase %s failed, diff between wantMetaObj and gotMetaObj: %v", testcase.name, cmp.Diff(wantMetaObj, gotMetaObj))
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
					t.Errorf("testcase %s failed, diff between wantContainer and gotContainer: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
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
					t.Errorf("testcase %s failed, diff between wantVolume and gotVolume: %v", testcase.name, cmp.Diff(wantVolume, gotVolume))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the volume %s", testcase.name, k)
			}
		}
	}
}

func TestInjectPodUnprivileged(t *testing.T) {
	type runtimeInfo struct {
		name        string
		namespace   string
		runtimeType string
	}
	type testCase struct {
		name    string
		in      *corev1.Pod
		dataset []*datav1alpha1.Dataset
		pv      []*corev1.PersistentVolume
		pvc     []*corev1.PersistentVolumeClaim
		fuse    []*appsv1.DaemonSet
		infos   map[string]runtimeInfo
		want    *corev1.Pod
		wantErr error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	mountPropagationHostToContainer := corev1.MountPropagationHostToContainer
	bTrue := true
	bFalse := false
	var mode int32 = 0755

	testcases := []testCase{
		{
			name: "inject_pod_unprivileged",
			dataset: []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset",
						Namespace: "big-data",
					},
				},
			},
			pv: []*corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
			},
			pvc: []*corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset",
						Namespace: "big-data",
					}, Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset",
					},
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unprivileged-pvc-pod",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "unprivileged-pvc-pod",
							Name:  "unprivileged-pvc-pod",
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
									ClaimName: "dataset",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: []*appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-jindofs-fuse",
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
										Image:   "unprivileged-pvc-pod",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk1",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk1",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
												Path: "/runtime-mnt/jindo/big-data/dataset",
												Type: &hostPathDirectoryOrCreate,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"dataset": {
					name:        "dataset",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unprivileged-pvc-pod",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "unprivileged-pvc-pod",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-0",
									MountPath: "/mnt/disk1",
								},
								{
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Image: "unprivileged-pvc-pod",
							Name:  "unprivileged-pvc-pod",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "dataset",
									MountPath:        "/data",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "cachedir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_unprivileged_multiple_pvc",
			dataset: []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset1",
						Namespace: "big-data",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset2",
						Namespace: "big-data",
					},
				},
			},
			pv: []*corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset1",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset2",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset2/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
			},
			pvc: []*corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset1",
						Namespace: "big-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset2",
						Namespace: "big-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset2",
					},
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unprivileged-pvc-pod",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "unprivileged-pvc-pod",
							Name:  "unprivileged-pvc-pod",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset1",
									MountPath: "/data1",
								},
								{
									Name:      "dataset2",
									MountPath: "/data2",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset1",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset1",
									ReadOnly:  true,
								},
							},
						},
						{
							Name: "dataset2",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset2",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: []*appsv1.DaemonSet{
				{
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
										Image:   "unprivileged-pvc-pod",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset2-jindofs-fuse",
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
										Image:   "unprivileged-pvc-pod",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
												Path: "/runtime-mnt/jindo/big-data/dataset2",
												Type: &hostPathDirectoryOrCreate,
											},
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
				"dataset2": {
					name:        "dataset2",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unprivileged-pvc-pod",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset1"),
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-1"): fmt.Sprintf("%s-%s", "big-data", "dataset2"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-1",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "unprivileged-pvc-pod",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-1",
									MountPath: "/mnt/disk",
								},
								{
									Name:      "default-check-mount-1",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "unprivileged-pvc-pod",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-0",
									MountPath: "/mnt/disk",
								},
								{
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Image: "unprivileged-pvc-pod",
							Name:  "unprivileged-pvc-pod",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "dataset1",
									MountPath:        "/data1",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:             "dataset2",
									MountPath:        "/data2",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								},
							},
						},
						{
							Name: "dataset2",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset2/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "cachedir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "cachedir-1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "check-mount-1",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_unprivileged_multiple_pvc_with_poststart_hook",
			dataset: []*datav1alpha1.Dataset{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-a",
						Namespace: "big-data",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-b",
						Namespace: "big-data",
					},
				},
			},
			pv: []*corev1.PersistentVolume{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset-a",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset-a/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "big-data-dataset-b",
					},
					Spec: corev1.PersistentVolumeSpec{
						PersistentVolumeSource: corev1.PersistentVolumeSource{
							CSI: &corev1.CSIPersistentVolumeSource{
								Driver: "fuse.csi.fluid.io",
								VolumeAttributes: map[string]string{
									common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset-b/jindofs-fuse",
									common.VolumeAttrMountType: common.JindoRuntime,
								},
							},
						},
					},
				},
			},
			pvc: []*corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-a",
						Namespace: "big-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset-a",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-b",
						Namespace: "big-data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "big-data-dataset-b",
					},
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unprivileged-pvc-pod",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
						common.InjectAppPostStart:            common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "unprivileged-pvc-pod",
							Name:  "unprivileged-pvc-pod",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dataset-a",
									MountPath: "/data1",
								},
								{
									Name:      "dataset-b",
									MountPath: "/data2",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset-a",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset-a",
									ReadOnly:  true,
								},
							},
						},
						{
							Name: "dataset-b",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "dataset-b",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: []*appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-a-jindofs-fuse",
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
										Image:   "unprivileged-pvc-pod",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
												Path: "/runtime-mnt/jindo/big-data/dataset-a",
												Type: &hostPathDirectoryOrCreate,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dataset-b-jindofs-fuse",
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
										Image:   "unprivileged-pvc-pod",
										SecurityContext: &corev1.SecurityContext{
											Privileged: &bTrue,
										}, VolumeMounts: []corev1.VolumeMount{
											{
												Name:      "cachedir",
												MountPath: "/mnt/disk",
											}, {
												Name:      "jindofs-fuse-device",
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
										Name: "cachedir",
										VolumeSource: corev1.VolumeSource{
											HostPath: &corev1.HostPathVolumeSource{
												Path: "/mnt/disk",
												Type: &hostPathDirectoryOrCreate,
											},
										}},
									{
										Name: "jindofs-fuse-device",
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
												Path: "/runtime-mnt/jindo/big-data/dataset-b",
												Type: &hostPathDirectoryOrCreate,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			infos: map[string]runtimeInfo{
				"dataset-a": {
					name:        "dataset-a",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
				"dataset-b": {
					name:        "dataset-b",
					namespace:   "big-data",
					runtimeType: common.JindoRuntime,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "unprivileged-pvc-pod",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
						common.InjectAppPostStart:            common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset-a"),
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-1"): fmt.Sprintf("%s-%s", "big-data", "dataset-b"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-1",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "unprivileged-pvc-pod",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-1",
									MountPath: "/mnt/disk",
								},
								{
									Name:      "default-check-mount-1",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "unprivileged-pvc-pod",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "cachedir-0",
									MountPath: "/mnt/disk",
								},
								{
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							},
						},
						{
							Image: "unprivileged-pvc-pod",
							Name:  "unprivileged-pvc-pod",
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"bash", "-c", "time /check-fluid-mount-ready.sh /data1:/data2 jindo:jindo >> /proc/1/fd/1"},
									},
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:             "dataset-a",
									MountPath:        "/data1",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:             "dataset-b",
									MountPath:        "/data2",
									MountPropagation: &mountPropagationHostToContainer,
								},
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dataset-a",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset-a/jindofs-fuse",
								},
							},
						},
						{
							Name: "dataset-b",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset-b/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "cachedir-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "cachedir-1",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						},
						{
							Name: "default-check-mount-1",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
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
		for _, obj := range testcase.fuse {
			objs = append(objs, obj)
		}
		for _, obj := range testcase.pv {
			objs = append(objs, obj)
		}
		for _, obj := range testcase.pvc {
			objs = append(objs, obj)
		}
		for _, obj := range testcase.dataset {
			objs = append(objs, obj)
		}
	}

	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

	for _, testcase := range testcases {
		injector := NewInjector(fakeClient)

		runtimeInfos := map[string]base.RuntimeInfoInterface{}
		for pvc, info := range testcase.infos {
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType)
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
			t.Errorf("testcase %s failed, diff between wantMetaObj and gotMetaObj: %v", testcase.name, cmp.Diff(wantMetaObj, gotMetaObj))
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
			t.Errorf("testcase %s failed, want containers length %d, Got containers length  %d", testcase.name, len(wantContainerMap), len(gotContainerMap))
		}

		for k, wantContainer := range wantContainerMap {
			if strings.HasPrefix(k, common.FuseContainerName) {
				var exists bool
				tempWant := wantContainer.DeepCopy()
				tempWant.Name = ""
				for _, gotContainer := range gotContainers {
					tempGot := gotContainer.DeepCopy()
					tempGot.Name = ""

					if reflect.DeepEqual(tempGot, tempWant) {
						exists = true
					}
				}

				if !exists {
					want, err := yaml.Marshal(wantContainer)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}
					t.Errorf("testcase %s failed, want container: %v, but not found in containers", testcase.name, string(want))
				}
			} else if gotContainer, found := gotContainerMap[k]; found {
				if gotContainer.Lifecycle != nil && wantContainer.Lifecycle != nil {
					if gotContainer.Lifecycle.PostStart != nil && wantContainer.Lifecycle.PostStart != nil {
						if gotContainer.Lifecycle.PostStart.Exec != nil && wantContainer.Lifecycle.PostStart.Exec != nil {
							equal := comparePostStartExecCommands(gotContainer.Lifecycle.PostStart.Exec, wantContainer.Lifecycle.PostStart.Exec)
							if !equal {
								t.Errorf("testcase %s failed, want poststart %v, got poststart %v", testcase.name, wantContainer.Lifecycle.PostStart.Exec, gotContainer.Lifecycle.PostStart.Exec)
							}
							// ignore post start exec when checking deep equal
							wantContainer.Lifecycle.PostStart.Exec = nil
							gotContainer.Lifecycle.PostStart.Exec = nil
						}
					}
				}
				if !reflect.DeepEqual(wantContainer, gotContainer) {
					want, err := yaml.Marshal(wantContainer)
					if err != nil {
						t.Errorf("testcase %s failed,  due to %v", testcase.name, err)
					}

					outYaml, err := yaml.Marshal(gotContainer)
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

		//wantVolumesTotal := len(testcase.in.Spec.Volumes) + testcase.numPvcMount
		for _, injectedFuse := range testcase.fuse {
			for _, wantVolume := range injectedFuse.Spec.Template.Spec.Volumes {
				// Skip check for volumes like "<runtime>-fuse-mount" and "<runtime>-fuse-device"
				if wantVolume.VolumeSource.HostPath != nil &&
					(strings.HasPrefix(wantVolume.VolumeSource.HostPath.Path, "/dev") ||
						strings.HasPrefix(wantVolume.VolumeSource.HostPath.Path, "/runtime-mnt")) {
					continue
				}
				wantTemp := wantVolume.DeepCopy()
				wantTemp.Name = ""
				var exists bool
				for _, gotVolume := range gotVolumes {
					gotTemp := gotVolume.DeepCopy()
					gotTemp.Name = ""
					if reflect.DeepEqual(wantTemp, gotTemp) {
						exists = true
						break
					}
				}

				if !exists {
					want, err := yaml.Marshal(wantVolumes)
					if err != nil {
						t.Errorf("testcase %s failed due to %v", testcase.name, err)
					}
					t.Errorf("testcase %s failed, wantVolume: %s, but not found in gotVolumes", testcase.name, string(want))
				}
			}
			//wantVolumesTotal += len(injectedFuse.Spec.Template.Spec.Volumes)
		}

	}
}

func TestInjectPodWithInitContainer(t *testing.T) {
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
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
					InitContainers: []corev1.Container{
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "init-fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "duplicate"),
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: common.InitFuseContainerName + "-0",
							// Args: []string{
							// 	"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							// },
							// Command: []string{"/entrypoint.sh"},
							Args:    []string{"2s"},
							Command: []string{"sleep"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								},
								{
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To[int32](mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To[int32](mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_init_container_success",
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
					InitContainers: []corev1.Container{
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
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"):      fmt.Sprintf("%s-%s", "big-data", "dataset1"),
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "init-fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "dataset1"),
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: common.InitFuseContainerName + "-0",
							// Args: []string{
							// 	"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							// },
							// Command: []string{"/entrypoint.sh"},
							Args:    []string{"2s"},
							Command: []string{"sleep"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "test",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "data-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
									ReadOnly:  true,
									MountPath: "/check-mount.sh",
									SubPath:   "check-mount.sh",
								},
							}, Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To[int32](mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/dataset1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "data-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime_mnt/dataset1",
								},
							}}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To[int32](mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
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
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/customizedenv/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
					InitContainers: []corev1.Container{
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
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"):      fmt.Sprintf("%s-%s", "big-data", "customizedenv"),
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "init-fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "customizedenv"),
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name: common.InitFuseContainerName + "-0",
							// Args: []string{
							// 	"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							// },
							// Command: []string{"/entrypoint.sh"},
							Args:    []string{"2s"},
							Command: []string{"sleep"},
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
									Name:      "customizedenv-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
									Name:      "customizedenv-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To[int32](mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/customizedenv",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "customizedenv-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To[int32](mode),
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
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType)
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
			t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantMetaObj, gotMetaObj))
		}

		gotContainers := out.Spec.Containers
		gotInitContainers := out.Spec.InitContainers
		gotVolumes := out.Spec.Volumes
		// gotContainers := out.
		// , gotVolumes, err := getInjectPiece(out)
		// if err != nil {
		// 	t.Errorf("testcase %s failed due to inject error %v", testcase.name, err)
		// }

		wantContainers := testcase.want.Spec.Containers
		wantVolumes := testcase.want.Spec.Volumes

		wantInitContainers := testcase.want.Spec.InitContainers

		gotInitContainerMap := makeContainerMap(gotInitContainers)
		wantInitContainerMap := makeContainerMap(wantInitContainers)

		if len(gotInitContainerMap) != len(wantInitContainerMap) {
			t.Errorf("testcase %s failed, want Initcontainers length %d, Got Initcontainers length  %d", testcase.name, len(gotInitContainerMap), len(wantInitContainerMap))
		}

		for k, wantInitContainer := range wantInitContainerMap {
			if gotInitContainer, found := gotInitContainerMap[k]; found {
				if !reflect.DeepEqual(wantInitContainer, gotInitContainer) {
					t.Errorf("testcase %s failed, diff between wantInitContainer and gotInitContainer: %v", testcase.name, cmp.Diff(wantInitContainer, gotInitContainer))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the Initcontainer %s", testcase.name, k)
			}
		}

		gotContainerMap := makeContainerMap(gotContainers)
		wantContainerMap := makeContainerMap(wantContainers)

		if len(gotContainerMap) != len(wantContainerMap) {
			t.Errorf("testcase %s failed, want containers length %d, Got containers length  %d", testcase.name, len(gotContainerMap), len(wantContainerMap))
		}

		for k, wantContainer := range wantContainerMap {
			if gotContainer, found := gotContainerMap[k]; found {
				if !reflect.DeepEqual(wantContainer, gotContainer) {
					t.Errorf("testcase %s failed, diff between wantContainer and gotContainer: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
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
					t.Errorf("testcase %s failed, diff between wantVolume and gotVolume: %v", testcase.name, cmp.Diff(wantVolume, gotVolume))
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

func TestInjectPodWithEnabledFUSEMetrics(t *testing.T) {
	type runtimeInfo struct {
		name         string
		namespace    string
		runtimeType  string
		scrapeTarget string
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
	bFalse := false
	var mode int32 = 0755

	testcases := []testCase{
		{
			name: "inject_pod_with_enabled_fuse_metrics",
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
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
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
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
									},
									Ports: []corev1.ContainerPort{
										{Name: "jindo-metrics", ContainerPort: 15000},
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
					name:         "duplicate",
					namespace:    "big-data",
					runtimeType:  common.JindoRuntime,
					scrapeTarget: base.MountModeSelectAll,
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "duplicate"),
					},
					Annotations: map[string]string{
						common.AnnotationPrometheusFuseMetricsScrapeKey: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
							},
							Ports: []corev1.ContainerPort{
								{Name: "jindo-metrics", ContainerPort: 15000},
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
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
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
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
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_scrape_target_mount_pod_only",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate2",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate2",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate2/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate2",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate2",
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
									ClaimName: "duplicate2",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate2-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
									},
									Ports: []corev1.ContainerPort{
										{Name: "jindo-metrics", ContainerPort: 15000},
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
											Path: "/runtime-mnt/jindo/big-data/duplicate2",
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
				"duplicate2": {
					name:         "duplicate2",
					namespace:    "big-data",
					runtimeType:  common.JindoRuntime,
					scrapeTarget: string(base.MountPodMountMode),
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "duplicate2"),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Ports:   []corev1.ContainerPort{},
							Command: []string{"/entrypoint.sh"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bTrue,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "jindofs-fuse-mount-0",
									MountPath: "/jfs",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate2/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "jindofs-fuse-mount-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate2",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "inject_pod_with_enabled_fuse_metrics_unprivileged",
			dataset: &datav1alpha1.Dataset{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate3",
					Namespace: "big-data",
				},
			},
			pv: &corev1.PersistentVolume{
				ObjectMeta: metav1.ObjectMeta{
					Name: "big-data-duplicate3",
				},
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeSource: corev1.PersistentVolumeSource{
						CSI: &corev1.CSIPersistentVolumeSource{
							Driver: "fuse.csi.fluid.io",
							VolumeAttributes: map[string]string{
								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/duplicate3/jindofs-fuse",
								common.VolumeAttrMountType: common.JindoRuntime,
							},
						},
					},
				},
			},
			pvc: &corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate3",
					Namespace: "big-data",
				}, Spec: corev1.PersistentVolumeClaimSpec{
					VolumeName: "big-data-duplicate3",
				},
			},
			in: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
					},
					Annotations: nil,
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
									ClaimName: "duplicate3",
									ReadOnly:  true,
								},
							},
						},
					},
				},
			},
			fuse: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate3-jindofs-fuse",
					Namespace: "big-data",
				},
				Spec: appsv1.DaemonSetSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name: "fuse",
									Args: []string{
										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
									},
									Ports: []corev1.ContainerPort{
										{Name: "jindo-metrics", ContainerPort: 15000},
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
											Path: "/runtime-mnt/jindo/big-data/duplicate3",
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
				"duplicate3": {
					name:         "duplicate3",
					namespace:    "big-data",
					runtimeType:  common.JindoRuntime,
					scrapeTarget: string(base.SidecarMountMode),
				},
			},
			want: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "duplicate-pvc-name",
					Namespace: "big-data",
					Labels: map[string]string{
						common.InjectFuseSidecar:             common.True,
						common.InjectUnprivilegedFuseSidecar: common.True,
						fmt.Sprintf("%s%s", common.LabelContainerDatasetMappingKeyPrefix, "fluid-fuse-0"): fmt.Sprintf("%s-%s", "big-data", "duplicate3"),
					},
					Annotations: map[string]string{
						common.AnnotationPrometheusFuseMetricsScrapeKey: common.True,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: common.FuseContainerName + "-0",
							Args: []string{
								"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000", "-ometrics_port=15000",
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{
											// "/check-mount.sh",
											// "/jfs",
											// "jindo",
											"bash",
											"-c",
											"time /check-mount.sh /jfs jindo  >> /proc/1/fd/1",
										},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{Name: "jindo-metrics", ContainerPort: 15000},
							},
							Command: []string{"/entrypoint.sh"},
							Image:   "duplicate-pvc-name",
							SecurityContext: &corev1.SecurityContext{
								Privileged: &bFalse,
							}, VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "duplicate-0",
									MountPath: "/mnt/disk1",
								}, {
									Name:      "fuse-device-0",
									MountPath: "/dev/fuse",
								}, {
									Name:      "default-check-mount-0",
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
								{
									Name:      "check-fluid-mount-ready",
									ReadOnly:  true,
									MountPath: "/check-fluid-mount-ready.sh",
									SubPath:   "check-fluid-mount-ready.sh",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "duplicate",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/runtime-mnt/jindo/big-data/duplicate3/jindofs-fuse",
								},
							},
						},
						{
							Name: "check-fluid-mount-ready",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "check-fluid-mount-ready",
									},
									DefaultMode: ptr.To(mode),
								},
							},
						},
						{
							Name: "fuse-device-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev/fuse",
									Type: &hostPathCharDev,
								},
							},
						},
						{
							Name: "duplicate-0",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/mnt/disk1",
									Type: &hostPathDirectoryOrCreate,
								},
							},
						}, {
							Name: "default-check-mount-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "jindo-default-check-mount",
									},
									DefaultMode: ptr.To(mode),
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
			opts := []base.RuntimeInfoOption{
				base.WithClientMetrics(datav1alpha1.ClientMetrics{
					Enabled:      true,
					ScrapeTarget: info.scrapeTarget,
				}),
			}
			runtimeInfo, err := base.BuildRuntimeInfo(info.name, info.namespace, info.runtimeType, opts...)
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
			t.Errorf("testcase %s failed, diff between want and got is: %v", testcase.name, cmp.Diff(gotMetaObj, wantMetaObj))
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
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantContainer, gotContainer))
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
					t.Errorf("testcase %s failed, diff between want and got: %v", testcase.name, cmp.Diff(wantVolume, gotVolume))
				}
			} else {
				t.Errorf("testcase %s failed due to missing the volume %s", testcase.name, k)
			}
		}
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

func comparePostStartExecCommands(exec1, exec2 *corev1.ExecAction) (equal bool) {
	if len(exec1.Command) != len(exec2.Command) {
		return false
	}

	for ci := range exec1.Command {
		subCmd1 := exec1.Command[ci]
		subCmd2 := exec2.Command[ci]
		if strings.Contains(subCmd1, " ") {
			parameters1 := strings.Split(subCmd1, " ")
			parameters2 := strings.Split(subCmd2, " ")
			if len(parameters1) != len(parameters2) {
				return false
			}
			for pi := range parameters1 {
				if strings.Contains(parameters1[pi], ":") {
					tokens1 := strings.Split(parameters1[pi], ":")
					tokens2 := strings.Split(parameters2[pi], ":")

					if len(tokens1) != len(tokens2) {
						return false
					}

					for _, token := range tokens1 {
						if !utils.ContainsString(tokens2, token) {
							return false
						}
					}
				} else {
					if !reflect.DeepEqual(parameters1[pi], parameters2[pi]) {
						return false
					}
				}
			}
		}
	}

	return true
}
