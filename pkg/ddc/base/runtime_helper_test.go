/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package base

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

// func TestGetTemplateToInjectForFuse(t *testing.T) {
// 	type runtimeInfo struct {
// 		name        string
// 		namespace   string
// 		runtimeType string
// 	}
// 	type testCase struct {
// 		name      string
// 		dataset   *datav1alpha1.Dataset
// 		pvcName   string
// 		info      runtimeInfo
// 		pv        *corev1.PersistentVolume
// 		pvc       *corev1.PersistentVolumeClaim
// 		fuse      *appsv1.DaemonSet
// 		expectErr bool
// 	}

// 	hostPathCharDev := corev1.HostPathCharDev
// 	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
// 	bTrue := true

// 	testcases := []testCase{
// 		{
// 			name: "jindo",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata",
// 					Namespace: "big-data",
// 				},
// 			},
// 			info: runtimeInfo{
// 				name:        "mydata",
// 				namespace:   "big-data",
// 				runtimeType: common.JindoRuntime,
// 			},
// 			pvcName: "mydata",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-mydata",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/mydata/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			expectErr: false,
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-mydata",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata-jindofs-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "mydata-pvc-name",
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									}, VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "mydata",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "mydata",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/mnt/disk1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/mydata",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "alluxio",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "dataset1",
// 					Namespace: "big-data",
// 				},
// 			}, info: runtimeInfo{
// 				name:        "dataset1",
// 				namespace:   "big-data",
// 				runtimeType: common.AlluxioRuntime,
// 			},
// 			expectErr: false,
// 			pvcName:   "dataset1",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-dataset1",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "dataset1",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-dataset1",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "dataset1-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "test",
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									},
// 									VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "data",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "data",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime_mnt/dataset1",
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/dataset1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}, {
// 			name: "custome_envs",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "customizedenv",
// 					Namespace: "big-data",
// 				},
// 			}, info: runtimeInfo{
// 				name:        "customizedenv",
// 				namespace:   "big-data",
// 				runtimeType: common.JindoRuntime,
// 			},
// 			pvcName: "customizedenv",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-customizedenv",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/customizedenv/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "customizedenv",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-customizedenv",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "customizedenv-jindofs-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "customizedenv-pvc-name",
// 									Env: []corev1.EnvVar{
// 										{
// 											Name:  "FLUID_FUSE_MOUNTPOINT",
// 											Value: "/jfs/jindofs-fuse",
// 										},
// 									},
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									}, VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "customizedenv",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "customizedenv",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/mnt/disk1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/customizedenv",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}, {
// 			name: "nofounddataset",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "nofounddataset",
// 					Namespace: "big-data",
// 				},
// 			}, info: runtimeInfo{
// 				name:        "nofounddataset",
// 				namespace:   "big-data",
// 				runtimeType: common.JindoRuntime,
// 			},
// 			expectErr: true,
// 			pvcName:   "nofounddataset",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-nofounddataset",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/nofounddataset/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "nofounddataset1",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-nofounddataset",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "nofounddataset-jindofs-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "nofounddataset-pvc-name",
// 									Env: []corev1.EnvVar{
// 										{
// 											Name:  "FLUID_FUSE_MOUNTPOINT",
// 											Value: "/jfs/jindofs-fuse",
// 										},
// 									},
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									}, VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "nofounddataset",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "nofounddataset",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/mnt/disk1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/nofounddataset",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	objs := []runtime.Object{}
// 	s := runtime.NewScheme()
// 	_ = corev1.AddToScheme(s)
// 	_ = datav1alpha1.AddToScheme(s)
// 	_ = appsv1.AddToScheme(s)
// 	for _, testcase := range testcases {
// 		objs = append(objs, testcase.fuse, testcase.pv, testcase.pvc, testcase.dataset)
// 	}

// 	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

// 	options := common.FuseSidecarInjectOption{
// 		EnableCacheDir:            true,
// 		EnableUnprivilegedSidecar: false,
// 	}

// 	for _, testcase := range testcases {
// 		info := testcase.info
// 		runtimeInfo, err := BuildRuntimeInfo(info.name, info.namespace, info.runtimeType, datav1alpha1.TieredStore{})
// 		if err != nil {
// 			t.Errorf("testcase %s failed due to error %v", testcase.name, err)
// 		}
// 		runtimeInfo.SetClient(fakeClient)
// 		_, err = runtimeInfo.GetTemplateToInjectForFuse(testcase.pvcName, testcase.pvc.Namespace, options)
// 		if (err == nil) == testcase.expectErr {
// 			t.Errorf("testcase %s failed due to expecting want error: %v error %v", testcase.name, testcase.expectErr, err)
// 		}
// 	}
// }

// func TestGetTemplateToInjectForFuseForCacheDir(t *testing.T) {
// 	type runtimeInfo struct {
// 		name        string
// 		namespace   string
// 		runtimeType string
// 	}
// 	type testCase struct {
// 		name           string
// 		dataset        *datav1alpha1.Dataset
// 		pvcName        string
// 		enableCacheDir bool
// 		info           runtimeInfo
// 		pv             *corev1.PersistentVolume
// 		pvc            *corev1.PersistentVolumeClaim
// 		fuse           *appsv1.DaemonSet
// 		expectErr      bool
// 	}

// 	hostPathCharDev := corev1.HostPathCharDev
// 	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
// 	bTrue := true

// 	testcases := []testCase{
// 		{
// 			name: "jindo",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata",
// 					Namespace: "big-data",
// 				},
// 			},
// 			info: runtimeInfo{
// 				name:        "mydata",
// 				namespace:   "big-data",
// 				runtimeType: common.JindoRuntime,
// 			},
// 			pvcName: "mydata",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-mydata",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/mydata/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			enableCacheDir: false,
// 			expectErr:      false,
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-mydata",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata-jindofs-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "mydata-pvc-name",
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									}, VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "datavolume-1",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "datavolume-1",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/mnt/disk1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/mydata",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			name: "alluxio",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "dataset1",
// 					Namespace: "big-data",
// 				},
// 			}, info: runtimeInfo{
// 				name:        "dataset1",
// 				namespace:   "big-data",
// 				runtimeType: common.AlluxioRuntime,
// 			},
// 			expectErr: false,
// 			pvcName:   "dataset1",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-dataset1",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/dataset1/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "dataset1",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-dataset1",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "dataset1-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "test",
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									},
// 									VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "data",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "data",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime_mnt/dataset1",
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/dataset1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		}, {
// 			name: "custome_envs",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "customizedenv",
// 					Namespace: "big-data",
// 				},
// 			}, info: runtimeInfo{
// 				name:        "customizedenv",
// 				namespace:   "big-data",
// 				runtimeType: common.JindoRuntime,
// 			},
// 			pvcName: "customizedenv",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-customizedenv",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/customizedenv/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "customizedenv",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-customizedenv",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "customizedenv-jindofs-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "customizedenv-pvc-name",
// 									Env: []corev1.EnvVar{
// 										{
// 											Name:  "FLUID_FUSE_MOUNTPOINT",
// 											Value: "/jfs/jindofs-fuse",
// 										},
// 									},
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 									}, VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "customizedenv",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "customizedenv",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/mnt/disk1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									}},
// 								{
// 									Name: "fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/customizedenv",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	objs := []runtime.Object{}
// 	s := runtime.NewScheme()
// 	_ = corev1.AddToScheme(s)
// 	_ = datav1alpha1.AddToScheme(s)
// 	_ = appsv1.AddToScheme(s)
// 	for _, testcase := range testcases {
// 		objs = append(objs, testcase.fuse, testcase.pv, testcase.pvc, testcase.dataset)
// 	}

// 	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

// 	for _, testcase := range testcases {
// 		info := testcase.info
// 		options := common.FuseSidecarInjectOption{
// 			EnableCacheDir:            testcase.enableCacheDir,
// 			EnableUnprivilegedSidecar: false,
// 		}
// 		runtimeInfo, err := BuildRuntimeInfo(info.name, info.namespace, info.runtimeType, datav1alpha1.TieredStore{})
// 		if err != nil {
// 			t.Errorf("testcase %s failed due to error %v", testcase.name, err)
// 		}
// 		runtimeInfo.SetClient(fakeClient)
// 		_, err = runtimeInfo.GetTemplateToInjectForFuse(testcase.pvcName, testcase.pvc.Namespace, options)
// 		if (err == nil) == testcase.expectErr {
// 			t.Errorf("testcase %s failed due to expecting want error: %v error %v", testcase.name, testcase.expectErr, err)
// 		}
// 	}
// }

// func TestGetTemplateToInjectForFuseWithVirtualFuseDevice(t *testing.T) {
// 	type runtimeInfo struct {
// 		name        string
// 		namespace   string
// 		runtimeType string
// 	}
// 	type testCase struct {
// 		name                          string
// 		dataset                       *datav1alpha1.Dataset
// 		pvcName                       string
// 		enableCacheDir                bool
// 		enableUnprivilegedFuseSidecar bool
// 		info                          runtimeInfo
// 		pv                            *corev1.PersistentVolume
// 		pvc                           *corev1.PersistentVolumeClaim
// 		fuse                          *appsv1.DaemonSet
// 		expectErr                     bool
// 	}

// 	hostPathCharDev := corev1.HostPathCharDev
// 	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
// 	bTrue := true

// 	var containerCheckers []func(template *common.FuseInjectionTemplate) bool
// 	// Check mutated container's security context
// 	containerCheckers = append(containerCheckers, func(template *common.FuseInjectionTemplate) bool {
// 		for _, capAdd := range template.FuseContainer.SecurityContext.Capabilities.Add {
// 			if capAdd == "SYS_ADMIN" {
// 				return false
// 			}
// 		}

// 		return *template.FuseContainer.SecurityContext.Privileged == false
// 	})
// 	// Check mutated container's volumes
// 	containerCheckers = append(containerCheckers, func(template *common.FuseInjectionTemplate) bool {
// 		for _, vol := range template.VolumesToAdd {
// 			if utils.ContainsString(hostMountNames, vol.Name) || utils.ContainsString(hostFuseDeviceNames, vol.Name) {
// 				return false
// 			}
// 		}

// 		for _, volMount := range template.FuseContainer.VolumeMounts {
// 			if utils.ContainsString(hostMountNames, volMount.Name) || utils.ContainsString(hostFuseDeviceNames, volMount.Name) {
// 				return false
// 			}
// 		}

// 		return true
// 	})
// 	// Check mutated container's resources
// 	containerCheckers = append(containerCheckers, func(template *common.FuseInjectionTemplate) bool {
// 		if len(template.FuseContainer.Resources.Limits) == 0 {
// 			return false
// 		}
// 		if _, ok := template.FuseContainer.Resources.Limits[corev1.ResourceName(common.DefaultFuseDeviceResourceName)]; !ok {
// 			return false
// 		}
// 		if len(template.FuseContainer.Resources.Requests) == 0 {
// 			return false
// 		}
// 		if _, ok := template.FuseContainer.Resources.Requests[corev1.ResourceName(common.DefaultFuseDeviceResourceName)]; !ok {
// 			return false
// 		}

// 		return true
// 	})

// 	testcases := []testCase{
// 		{
// 			name: "jindo",
// 			dataset: &datav1alpha1.Dataset{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata",
// 					Namespace: "big-data",
// 				},
// 			},
// 			info: runtimeInfo{
// 				name:        "mydata",
// 				namespace:   "big-data",
// 				runtimeType: common.JindoRuntime,
// 			},
// 			pvcName: "mydata",
// 			pv: &corev1.PersistentVolume{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name: "big-data-mydata",
// 				},
// 				Spec: corev1.PersistentVolumeSpec{
// 					PersistentVolumeSource: corev1.PersistentVolumeSource{
// 						CSI: &corev1.CSIPersistentVolumeSource{
// 							Driver: "fuse.csi.fluid.io",
// 							VolumeAttributes: map[string]string{
// 								common.VolumeAttrFluidPath: "/runtime-mnt/jindo/big-data/mydata/jindofs-fuse",
// 								common.VolumeAttrMountType: common.JindoRuntime,
// 							},
// 						},
// 					},
// 				},
// 			},
// 			enableCacheDir:                false,
// 			enableUnprivilegedFuseSidecar: true,
// 			expectErr:                     false,
// 			pvc: &corev1.PersistentVolumeClaim{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata",
// 					Namespace: "big-data",
// 				}, Spec: corev1.PersistentVolumeClaimSpec{
// 					VolumeName: "big-data-mydata",
// 				},
// 			},
// 			fuse: &appsv1.DaemonSet{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Name:      "mydata-jindofs-fuse",
// 					Namespace: "big-data",
// 				},
// 				Spec: appsv1.DaemonSetSpec{
// 					Template: corev1.PodTemplateSpec{
// 						Spec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								{
// 									Name: "fuse",
// 									Args: []string{
// 										"-oroot_ns=jindo", "-okernel_cache", "-oattr_timeout=9000", "-oentry_timeout=9000",
// 									},
// 									Command: []string{"/entrypoint.sh"},
// 									Image:   "mydata-pvc-name",
// 									SecurityContext: &corev1.SecurityContext{
// 										Privileged: &bTrue,
// 										Capabilities: &corev1.Capabilities{
// 											Add: []corev1.Capability{"SYS_ADMIN"},
// 										},
// 									}, VolumeMounts: []corev1.VolumeMount{
// 										{
// 											Name:      "datavolume-1",
// 											MountPath: "/mnt/disk1",
// 										}, {
// 											Name:      "jindofs-fuse-device",
// 											MountPath: "/dev/fuse",
// 										}, {
// 											Name:      "jindofs-fuse-mount",
// 											MountPath: "/jfs",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								{
// 									Name: "datavolume-1",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/mnt/disk1",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									}},
// 								{
// 									Name: "jindofs-fuse-device",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/dev/fuse",
// 											Type: &hostPathCharDev,
// 										},
// 									},
// 								},
// 								{
// 									Name: "jindofs-fuse-mount",
// 									VolumeSource: corev1.VolumeSource{
// 										HostPath: &corev1.HostPathVolumeSource{
// 											Path: "/runtime-mnt/jindo/big-data/mydata",
// 											Type: &hostPathDirectoryOrCreate,
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	objs := []runtime.Object{}
// 	s := runtime.NewScheme()
// 	_ = corev1.AddToScheme(s)
// 	_ = datav1alpha1.AddToScheme(s)
// 	_ = appsv1.AddToScheme(s)
// 	for _, testcase := range testcases {
// 		objs = append(objs, testcase.fuse, testcase.pv, testcase.pvc, testcase.dataset)
// 	}

// 	fakeClient := fake.NewFakeClientWithScheme(s, objs...)

// 	for _, testcase := range testcases {
// 		info := testcase.info
// 		options := common.FuseSidecarInjectOption{
// 			EnableCacheDir:            testcase.enableCacheDir,
// 			EnableUnprivilegedSidecar: testcase.enableUnprivilegedFuseSidecar,
// 		}
// 		runtimeInfo, err := BuildRuntimeInfo(info.name, info.namespace, info.runtimeType, datav1alpha1.TieredStore{})
// 		if err != nil {
// 			t.Errorf("testcase %s failed due to error %v", testcase.name, err)
// 		}
// 		runtimeInfo.SetClient(fakeClient)
// 		template, err := runtimeInfo.GetTemplateToInjectForFuse(testcase.pvcName, testcase.pvc.Namespace, options)
// 		if (err == nil) == testcase.expectErr {
// 			t.Errorf("testcase %s failed due to expecting want error: %v error %v", testcase.name, testcase.expectErr, err)
// 		}

// 		for _, checker := range containerCheckers {
// 			if !checker(template) {
// 				t.Errorf("testcase %s failed due to check failed for template %v", testcase.name, template)
// 			}
// 		}
// 	}
// }

func TestGetFuseDaemonset(t *testing.T) {
	type testCase struct {
		name        string
		namespace   string
		runtimeType string
		ds          *appsv1.DaemonSet
		setClient   bool
		wantErr     bool
	}

	tests := []testCase{
		{
			name:        "alluxio",
			namespace:   "default",
			runtimeType: common.AlluxioRuntime,
			ds: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "alluxio-fuse",
					Namespace: "default",
				},
			},
			setClient: true,
			wantErr:   false,
		}, {
			name:        "jindo",
			namespace:   "default",
			runtimeType: common.JindoRuntime,
			ds: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jindo-jindofs-fuse",
					Namespace: "default",
				},
			},
			setClient: true,
			wantErr:   false,
		}, {
			name:        "noclient",
			namespace:   "default",
			runtimeType: common.JindoRuntime,
			ds: &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jindo-jindofs-fuse",
					Namespace: "default",
				},
			},
			setClient: false,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		var fakeClient client.Client
		if test.setClient {
			objs := []runtime.Object{}
			s := runtime.NewScheme()
			_ = corev1.AddToScheme(s)
			_ = datav1alpha1.AddToScheme(s)
			_ = appsv1.AddToScheme(s)
			objs = append(objs, test.ds)
			fakeClient = fake.NewFakeClientWithScheme(s, objs...)
		}

		runtimeInfo := RuntimeInfo{
			name:        test.name,
			namespace:   test.namespace,
			runtimeType: test.runtimeType,
		}

		if fakeClient != nil {
			runtimeInfo.SetClient(fakeClient)
		}

		_, err := runtimeInfo.getFuseDaemonset()
		if (err == nil) == test.wantErr {
			t.Errorf("testcase %s is failed, want err %v, got err %v", test.name, test.wantErr, err)
		}
	}
}
