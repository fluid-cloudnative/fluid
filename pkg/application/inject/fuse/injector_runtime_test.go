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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestInjectList(t *testing.T) {
	type runtimeInfo struct {
		name        string
		namespace   string
		runtimeType string
	}
	type testCase struct {
		name    string
		in      []corev1.Pod
		dataset *datav1alpha1.Dataset
		pv      *corev1.PersistentVolume
		pvc     *corev1.PersistentVolumeClaim
		fuse    *appsv1.DaemonSet
		infos   map[string]runtimeInfo
		wantErr error
	}

	hostPathCharDev := corev1.HostPathCharDev
	hostPathDirectoryOrCreate := corev1.HostPathDirectoryOrCreate
	bTrue := true

	testcases := []testCase{
		{
			name: "inject_list",
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
			in: []corev1.Pod{{
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
				}},
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

		podList := &corev1.List{}
		for _, pod := range testcase.in {
			raw, err := json.Marshal(&pod)
			if err != nil {
				t.Fatalf("Could not create parse pod: %v", err)
			}
			podList.Items = append(podList.Items, runtime.RawExtension{Raw: raw})
		}

		_, err := injector.Inject(podList, runtimeInfos)
		if err != nil {
			if testcase.wantErr == nil {
				t.Errorf("testcase %s failed, Got error %v", testcase.name, err)
			} else {
				continue
			}
		}
	}
}

func TestInjectUnstructured(t *testing.T) {
	var (
		name        = "test"
		namespace   = "default"
		runtimeType = common.JindoRuntime
	)
	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	runtimeInfos := map[string]base.RuntimeInfoInterface{}

	runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, runtimeType)
	if err != nil {
		t.Errorf("testcase %s failed due to error %v", name, err)
	}
	runtimeInfo.SetClient(fakeClient)
	runtimeInfos[name] = runtimeInfo
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
			}}}

	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	if err != nil {
		t.Errorf("testcase %s failed due to %v", name, err)
	}
	injector := NewInjector(fakeClient)
	_, err = injector.Inject(&unstructured.Unstructured{Object: object}, runtimeInfos)
	if err == nil || err.Error() != "not implemented" {
		t.Errorf("testcase %s for unstructured is not implemented and error is %v", name, err)
	}

}

func TestInjectObject(t *testing.T) {
	var (
		name        = "test"
		namespace   = "default"
		runtimeType = common.JindoRuntime
	)
	objs := []runtime.Object{}
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = datav1alpha1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	fakeClient := fake.NewFakeClientWithScheme(s, objs...)
	runtimeInfos := map[string]base.RuntimeInfoInterface{}

	runtimeInfo, err := base.BuildRuntimeInfo(name, namespace, runtimeType)
	if err != nil {
		t.Errorf("testcase %s failed due to error %v", name, err)
	}
	runtimeInfo.SetClient(fakeClient)
	runtimeInfos[name] = runtimeInfo

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
				}, Spec: corev1.PodSpec{
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
					}},
			},
		},
	}

	injector := NewInjector(fakeClient)
	_, err = injector.Inject(&deploy, runtimeInfos)
	if err == nil || err.Error() != "not implemented" {
		t.Errorf("testcase %s for unstructured is not implemented and err is %v", name, err)
	}

}
