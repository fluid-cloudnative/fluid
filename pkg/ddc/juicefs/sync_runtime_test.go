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

package juicefs

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
)

func TestJuiceFSxEngine_syncWorkerSpec(t *testing.T) {
	cms := []corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "emtpy-worker-script",
				Namespace: "default",
			},
			Data: map[string]string{
				"script.sh": `#!/bin/bash

    if [ enterprise = community ]; then
    echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs format start."
    /usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
    elif [ ! -f /root/.juicefs/test-fluid-2.conf ]; then
    echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs auth start."
    /usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
    fi

    echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs mount start."
    /sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee
`,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "same-worker-script",
				Namespace: "default",
			},
			Data: map[string]string{
				"script.sh": `#!/bin/bash

    if [ enterprise = community ]; then
    echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs format start."
    /usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
    elif [ ! -f /root/.juicefs/test-fluid-2.conf ]; then
    echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs auth start."
    /usr/bin/juicefs auth --token=${TOKEN} --accesskey=${ACCESS_KEY} --secretkey=${SECRET_KEY} --bucket=http://test4.minio.default.svc.cluster.local:9000 test-fluid-2
    fi

    echo "$(date '+%Y/%m/%d %H:%M:%S').$(printf "%03d" $(($(date '+%N')/1000))) juicefs mount start."
    /sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee
`,
			},
		},
	}
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
		name      string
		namespace string
	}
	type args struct {
		ctx    cruntime.ReconcileRequestContext
		worker *appsv1.StatefulSet
		value  *JuiceFS
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantChanged  bool
		wantErr      bool
		wantResource corev1.ResourceRequirements
	}{
		{
			name: "Not resource for juicefs runtime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JuiceFSRuntime{},
			}, args: args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-worker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Name: "juicefs",
								}},
							},
						},
					},
				},
				value: &JuiceFS{
					Worker: Worker{
						Command: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee",
					},
				},
			},
			wantChanged: false,
			wantErr:     false,
		}, {
			name: "worker not found",
			fields: fields{
				name:      "noworker",
				namespace: "default",
				runtime:   &datav1alpha1.JuiceFSRuntime{},
			}, args: args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "noworker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{},
				},
				value: &JuiceFS{
					Worker: Worker{
						Command: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee",
					},
				},
			},
			wantChanged: false,
			wantErr:     true,
		}, {
			name: "worker not change",
			fields: fields{
				name:      "same",
				namespace: "default",
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: common.Memory,
									Quota:      &res,
								},
							},
						},
						Worker: datav1alpha1.JuiceFSCompTemplateSpec{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: res,
								},
							},
						},
					},
				},
			}, args: args{
				worker: &appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "same-worker",
						Namespace: "default",
					}, Spec: appsv1.StatefulSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "worker",
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("100m"),
												corev1.ResourceMemory: resource.MustParse("320Gi"),
											},
										},
									},
								},
							},
						},
					},
				},
				value: &JuiceFS{
					Worker: Worker{
						Resources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "320Gi",
							},
						},
						Command: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o subdir=/demo,cache-size=2048,free-space-ratio=0.1,cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee",
					},
				},
			},
			wantChanged: false,
			wantErr:     false,
			wantResource: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("320Gi"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			//runtimeObjs = append(runtimeObjs, tt.args.worker.DeepCopy())

			s := runtime.NewScheme()
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.worker)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			runtimeObjs = append(runtimeObjs, tt.args.worker)
			for _, cm := range cms {
				runtimeObjs = append(runtimeObjs, cm.DeepCopy())
			}
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JuiceFSEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Log:       fake.NullLogger(),
				Client:    client,
			}
			gotChanged, err := e.syncWorkerSpec(tt.args.ctx, tt.fields.runtime, &JuiceFS{}, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Testcase %s JuiceFSEngine.syncWorkerSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("Testcase %s JuiceFSEngine.syncWorkerSpec() = %v, want %v. got sts resources %v after updated, want %v",
					tt.name,
					gotChanged,
					tt.wantChanged,
					tt.args.worker.Spec.Template.Spec.Containers[0].Resources,
					tt.wantResource,
				)
			}

		})
	}
}

func TestJuiceFSxEngine_syncFuseSpec(t *testing.T) {
	res := resource.MustParse("320Gi")
	type fields struct {
		runtime   *datav1alpha1.JuiceFSRuntime
		helmValue *corev1.ConfigMap
		name      string
		namespace string
	}
	type args struct {
		ctx  cruntime.ReconcileRequestContext
		fuse *appsv1.DaemonSet
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantChanged  bool
		wantErr      bool
		wantResource corev1.ResourceRequirements
	}{
		{
			name: "Not resource for juicefs runtime",
			fields: fields{
				name:      "emtpy",
				namespace: "default",
				runtime:   &datav1alpha1.JuiceFSRuntime{},
				helmValue: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-juicefs-values",
						Namespace: "default",
					},
					Data: map[string]string{
						"data": "",
					},
				},
			}, args: args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "emtpy-fuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     false,
		}, {
			name: "fuse not found",
			fields: fields{
				name:      "nofuse",
				namespace: "default",
				runtime:   &datav1alpha1.JuiceFSRuntime{},
				helmValue: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nofuse-juicefs-values",
						Namespace: "default",
					},
					Data: map[string]string{
						"data": "",
					},
				},
			}, args: args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nofuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{},
				},
			},
			wantChanged: false,
			wantErr:     true,
		}, {
			name: "fuse not change",
			fields: fields{
				name:      "same",
				namespace: "default",
				runtime: &datav1alpha1.JuiceFSRuntime{
					Spec: datav1alpha1.JuiceFSRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{
								{
									MediumType: common.Memory,
									Quota:      &res,
								},
							},
						},
						Fuse: datav1alpha1.JuiceFSFuseSpec{
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
				helmValue: &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "same-juicefs-values",
						Namespace: "default",
					},
					Data: map[string]string{
						"data": "",
					},
				},
			}, args: args{
				fuse: &appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "same-fuse",
						Namespace: "default",
					}, Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name: "fuse",
										Resources: corev1.ResourceRequirements{
											Requests: corev1.ResourceList{
												corev1.ResourceCPU:    resource.MustParse("100m"),
												corev1.ResourceMemory: resource.MustParse("1Gi"),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantChanged: false,
			wantErr:     false,
			wantResource: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("100m"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}

			s := runtime.NewScheme()
			tt.fields.runtime.SetName(tt.fields.name)
			tt.fields.runtime.SetNamespace(tt.fields.namespace)
			s.AddKnownTypes(appsv1.SchemeGroupVersion, tt.args.fuse)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(corev1.SchemeGroupVersion, tt.fields.helmValue)

			_ = corev1.AddToScheme(s)
			runtimeObjs = append(runtimeObjs, tt.fields.runtime)
			runtimeObjs = append(runtimeObjs, tt.fields.helmValue)
			runtimeObjs = append(runtimeObjs, tt.args.fuse)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)

			e := &JuiceFSEngine{
				runtime:    tt.fields.runtime,
				name:       tt.fields.name,
				namespace:  tt.fields.namespace,
				Log:        fake.NullLogger(),
				Client:     client,
				engineImpl: "juicefs",
			}
			value := &JuiceFS{
				Fuse: Fuse{},
			}
			gotChanged, err := e.syncFuseSpec(tt.args.ctx, tt.fields.runtime, &JuiceFS{}, value)
			if (err != nil) != tt.wantErr {
				t.Errorf("testcase %s: JuiceFSEngine.syncFuseSpec() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if gotChanged != tt.wantChanged {
				t.Errorf("testcase %s JuiceFSEngine.syncFuseSpec() = %v, want %v. got sts resources %v after updated, want %v",
					tt.name,
					gotChanged,
					tt.wantChanged,
					tt.args.fuse.Spec.Template.Spec.Containers[0].Resources,
					tt.wantResource,
				)
			}

		})
	}
}

func TestJuiceFSEngine_isVolumeMountsChanged(t *testing.T) {
	type args struct {
		crtVolumeMounts     []corev1.VolumeMount
		runtimeVolumeMounts []corev1.VolumeMount
	}
	tests := []struct {
		name                string
		args                args
		wantChanged         bool
		wantNewVolumeMounts []corev1.VolumeMount
	}{
		{
			name: "test-false",
			args: args{
				crtVolumeMounts:     []corev1.VolumeMount{{Name: "test", MountPath: "/data"}},
				runtimeVolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/data"}},
			},
			wantChanged:         false,
			wantNewVolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/data"}},
		},
		{
			name: "test-changed",
			args: args{
				crtVolumeMounts:     []corev1.VolumeMount{{Name: "test", MountPath: "/data"}},
				runtimeVolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/data2"}},
			},
			wantChanged:         true,
			wantNewVolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/data2"}},
		},
		{
			name: "test-override",
			args: args{
				crtVolumeMounts:     []corev1.VolumeMount{{Name: "test", MountPath: "/data"}},
				runtimeVolumeMounts: []corev1.VolumeMount{{Name: "test2", MountPath: "/data2"}},
			},
			wantChanged:         true,
			wantNewVolumeMounts: []corev1.VolumeMount{{Name: "test2", MountPath: "/data2"}},
		},
		{
			name: "test-add",
			args: args{
				crtVolumeMounts:     []corev1.VolumeMount{{Name: "test", MountPath: "/data"}},
				runtimeVolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/data"}, {Name: "test2", MountPath: "/data2"}},
			},
			wantChanged:         true,
			wantNewVolumeMounts: []corev1.VolumeMount{{Name: "test", MountPath: "/data"}, {Name: "test2", MountPath: "/data2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewVolumeMounts := j.isVolumeMountsChanged(tt.args.crtVolumeMounts, tt.args.runtimeVolumeMounts)
			if gotChanged != tt.wantChanged {
				t.Errorf("isVolumeMountsChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if len(gotNewVolumeMounts) != len(tt.wantNewVolumeMounts) {
				t.Errorf("isVolumeMountsChanged() gotNewVolumeMounts = %v, want %v", gotNewVolumeMounts, tt.wantNewVolumeMounts)
			}
			for _, n := range gotNewVolumeMounts {
				got := false
				for _, w := range tt.wantNewVolumeMounts {
					if n.Name == w.Name && n.MountPath == w.MountPath {
						got = true
						break
					}
				}
				if !got {
					t.Errorf("isVolumeMountsChanged() gotNewVolumeMounts = %v, want %v", gotNewVolumeMounts, tt.wantNewVolumeMounts)
				}
			}
		})
	}
}

func TestJuiceFSEngine_isEnvsChanged(t *testing.T) {
	type args struct {
		crtEnvs     []corev1.EnvVar
		runtimeEnvs []corev1.EnvVar
	}
	tests := []struct {
		name        string
		args        args
		wantChanged bool
		wantNewEnvs []corev1.EnvVar
	}{
		{
			name: "test-false",
			args: args{
				crtEnvs:     []corev1.EnvVar{{Name: "test", Value: "test"}},
				runtimeEnvs: []corev1.EnvVar{{Name: "test", Value: "test"}},
			},
			wantChanged: false,
			wantNewEnvs: []corev1.EnvVar{{Name: "test", Value: "test"}},
		},
		{
			name: "test-changed",
			args: args{
				crtEnvs:     []corev1.EnvVar{{Name: "test", Value: "test"}},
				runtimeEnvs: []corev1.EnvVar{{Name: "test", Value: "test2"}},
			},
			wantChanged: true,
			wantNewEnvs: []corev1.EnvVar{{Name: "test", Value: "test2"}},
		},
		{
			name: "test-override",
			args: args{
				crtEnvs:     []corev1.EnvVar{{Name: "test", Value: "test"}},
				runtimeEnvs: []corev1.EnvVar{{Name: "test2", Value: "test2"}},
			},
			wantChanged: true,
			wantNewEnvs: []corev1.EnvVar{{Name: "test2", Value: "test2"}},
		},
		{
			name: "test-add",
			args: args{
				crtEnvs:     []corev1.EnvVar{{Name: "test", Value: "test"}},
				runtimeEnvs: []corev1.EnvVar{{Name: "test", Value: "test"}, {Name: "test2", Value: "test2"}},
			},
			wantChanged: true,
			wantNewEnvs: []corev1.EnvVar{{Name: "test", Value: "test"}, {Name: "test2", Value: "test2"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewEnvs := j.isEnvsChanged(tt.args.crtEnvs, tt.args.runtimeEnvs)
			if gotChanged != tt.wantChanged {
				t.Errorf("isEnvsChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if len(gotNewEnvs) != len(tt.wantNewEnvs) {
				t.Errorf("isEnvsChanged() gotNewEnvs = %v, want %v", gotNewEnvs, tt.wantNewEnvs)
			}
			for _, n := range gotNewEnvs {
				got := false
				for _, w := range tt.wantNewEnvs {
					if n.Name == w.Name && n.Value == w.Value {
						got = true
						break
					}
				}
				if !got {
					t.Errorf("isEnvsChanged() gotNewEnvs = %v, want %v", gotNewEnvs, tt.wantNewEnvs)
				}
			}
		})
	}
}

func TestJuiceFSEngine_isResourcesChanged(t *testing.T) {
	type args struct {
		crtResources     corev1.ResourceRequirements
		runtimeResources corev1.ResourceRequirements
	}
	tests := []struct {
		name             string
		args             args
		wantChanged      bool
		wantNewResources corev1.ResourceRequirements
	}{
		{
			name: "test-false",
			args: args{
				crtResources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("320Gi"),
					},
				},
				runtimeResources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("320Gi"),
					},
				},
			},
			wantChanged: false,
			wantNewResources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("320Gi"),
				},
			},
		},
		{
			name: "test-changed",
			args: args{
				crtResources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("320Gi"),
					},
				},
				runtimeResources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1000m"),
						corev1.ResourceMemory: resource.MustParse("3200Gi"),
					},
				},
			},
			wantChanged: true,
			wantNewResources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1000m"),
					corev1.ResourceMemory: resource.MustParse("3200Gi"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewResources := j.isResourcesChanged(tt.args.crtResources, tt.args.runtimeResources)
			if gotChanged != tt.wantChanged {
				t.Errorf("isResourcesChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if !reflect.DeepEqual(gotNewResources, tt.wantNewResources) {
				t.Errorf("isResourcesChanged() gotNewResources = %v, want %v", gotNewResources, tt.wantNewResources)
			}
		})
	}
}

func TestJuiceFSEngine_isVolumesChanged(t *testing.T) {
	type args struct {
		crtVolumes     []corev1.Volume
		runtimeVolumes []corev1.Volume
	}
	tests := []struct {
		name           string
		args           args
		wantChanged    bool
		wantNewVolumes []corev1.Volume
	}{
		{
			name: "test-false",
			args: args{
				crtVolumes:     []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test"}}}},
				runtimeVolumes: []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test"}}}},
			},
			wantChanged:    false,
			wantNewVolumes: []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test"}}}},
		},
		{
			name: "test-changed",
			args: args{
				crtVolumes:     []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test"}}}},
				runtimeVolumes: []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test2"}}}},
			},
			wantChanged:    true,
			wantNewVolumes: []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test2"}}}},
		},
		{
			name: "test-deleted",
			args: args{
				crtVolumes:     []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test"}}}},
				runtimeVolumes: []corev1.Volume{},
			},
			wantChanged:    true,
			wantNewVolumes: []corev1.Volume{},
		},
		{
			name: "test-new",
			args: args{
				crtVolumes: []corev1.Volume{{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test"}}}},
				runtimeVolumes: []corev1.Volume{
					{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test2"}}},
					{Name: "test2", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test2"}}},
				},
			},
			wantChanged: true,
			wantNewVolumes: []corev1.Volume{
				{Name: "test", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test2"}}},
				{Name: "test2", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/test2"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewVolumes := j.isVolumesChanged(tt.args.crtVolumes, tt.args.runtimeVolumes)
			if gotChanged != tt.wantChanged {
				t.Errorf("isVolumesChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if len(gotNewVolumes) != len(tt.wantNewVolumes) {
				t.Errorf("isVolumesChanged() gotNewVolumes = %v, want %v", gotNewVolumes, tt.wantNewVolumes)
			}
			for _, n := range gotNewVolumes {
				got := false
				for _, w := range tt.wantNewVolumes {
					if n.Name == w.Name && reflect.DeepEqual(n.VolumeSource, w.VolumeSource) {
						got = true
						break
					}
				}
				if !got {
					t.Errorf("isVolumesChanged() gotNewVolumes = %v, want %v", gotNewVolumes, tt.wantNewVolumes)
				}
			}
		})
	}
}

func TestJuiceFSEngine_isImageChanged(t *testing.T) {
	type args struct {
		crtImage     string
		runtimeImage string
	}
	tests := []struct {
		name        string
		args        args
		wantChanged bool
		wantImage   string
	}{
		{
			name: "test-false",
			args: args{
				crtImage:     "juicedata/juicefs-fuse:ee-4.9.6",
				runtimeImage: "juicedata/juicefs-fuse:ee-4.9.6",
			},
			wantChanged: false,
			wantImage:   "juicedata/juicefs-fuse:ee-4.9.6",
		},
		{
			name: "test-true",
			args: args{
				crtImage:     "juicedata/juicefs-fuse:ee-4.9.6",
				runtimeImage: "juicedata/juicefs-fuse:ee-4.9.10",
			},
			wantChanged: true,
			wantImage:   "juicedata/juicefs-fuse:ee-4.9.10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewImage := j.isImageChanged(tt.args.crtImage, tt.args.runtimeImage)
			if gotChanged != tt.wantChanged {
				t.Errorf("isImageChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if gotNewImage != tt.wantImage {
				t.Errorf("isImageChanged() gotNewHostNetwork = %v, want %v", gotNewImage, tt.wantImage)
			}
		})
	}
}

func TestJuiceFSEngine_isNodeSelectorChanged(t *testing.T) {
	type args struct {
		crtNodeSelector     map[string]string
		runtimeNodeSelector map[string]string
	}
	tests := []struct {
		name                string
		args                args
		wantChanged         bool
		wantNewNodeSelector map[string]string
	}{
		{
			name: "test-false",
			args: args{
				crtNodeSelector:     map[string]string{"test": "test"},
				runtimeNodeSelector: map[string]string{"test": "test"},
			},
			wantChanged:         false,
			wantNewNodeSelector: map[string]string{"test": "test"},
		},
		{
			name: "test-true",
			args: args{
				crtNodeSelector:     map[string]string{"test": "test"},
				runtimeNodeSelector: map[string]string{"test": "abc"},
			},
			wantChanged:         true,
			wantNewNodeSelector: map[string]string{"test": "abc"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewNodeSelector := j.isNodeSelectorChanged(tt.args.crtNodeSelector, tt.args.runtimeNodeSelector)
			if gotChanged != tt.wantChanged {
				t.Errorf("isNodeSelectorChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if !reflect.DeepEqual(gotNewNodeSelector, tt.wantNewNodeSelector) {
				t.Errorf("isNodeSelectorChanged() gotNewNodeSelector = %v, want %v", gotNewNodeSelector, tt.wantNewNodeSelector)
			}
		})
	}
}

func TestJuiceFSEngine_isCommandChanged(t *testing.T) {
	type args struct {
		crtCommand     string
		runtimeCommand string
	}
	tests := []struct {
		name           string
		args           args
		wantChanged    bool
		wantNewCommand string
	}{
		{
			name: "test-false",
			args: args{
				crtCommand:     "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o cache-dir=/dev/shm,verbose,foreground,no-update,cache-group=default-jfsdemo-ee,subdir=/demo,cache-size=1024,free-space-ratio=0.1",
				runtimeCommand: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o cache-dir=/dev/shm,verbose,foreground,no-update,cache-group=default-jfsdemo-ee,subdir=/demo,cache-size=1024,free-space-ratio=0.1",
			},
			wantChanged:    false,
			wantNewCommand: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o cache-dir=/dev/shm,verbose,foreground,no-update,cache-group=default-jfsdemo-ee,subdir=/demo,cache-size=1024,free-space-ratio=0.1",
		},
		{
			name: "test-true",
			args: args{
				crtCommand:     "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o cache-dir=/dev/shm,foreground,no-update,cache-group=default-jfsdemo-ee,subdir=/demo,cache-size=1024,free-space-ratio=0.1",
				runtimeCommand: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o cache-dir=/dev/shm,verbose,foreground,no-update,cache-group=default-jfsdemo-ee,subdir=/demo,cache-size=1024,free-space-ratio=0.1",
			},
			wantChanged:    true,
			wantNewCommand: "/sbin/mount.juicefs test-fluid-2 /runtime-mnt/juicefs/default/jfsdemo-ee/juicefs-fuse -o cache-dir=/dev/shm,verbose,foreground,no-update,cache-group=default-jfsdemo-ee,subdir=/demo,cache-size=1024,free-space-ratio=0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewCommand := j.isCommandChanged(tt.args.crtCommand, tt.args.runtimeCommand)
			if gotChanged != tt.wantChanged {
				t.Errorf("isCommandChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if gotNewCommand != tt.wantNewCommand {
				t.Errorf("isCommandChanged() gotNewCommand = %v, want %v", gotNewCommand, tt.wantNewCommand)
			}
		})
	}
}

func TestJuiceFSEngine_isLabelsChanged(t *testing.T) {
	type args struct {
		crtLabels     map[string]string
		runtimeLabels map[string]string
	}
	tests := []struct {
		name          string
		args          args
		wantChanged   bool
		wantNewLabels map[string]string
	}{
		{
			name: "test-false",
			args: args{
				crtLabels:     map[string]string{"test": "abc"},
				runtimeLabels: map[string]string{"test": "abc"},
			},
			wantChanged:   false,
			wantNewLabels: map[string]string{"test": "abc"},
		},
		{
			name: "test-changed",
			args: args{
				crtLabels:     map[string]string{"test": "abc"},
				runtimeLabels: map[string]string{"test": "def"},
			},
			wantChanged:   true,
			wantNewLabels: map[string]string{"test": "def"},
		},
		{
			name: "test-add",
			args: args{
				crtLabels:     map[string]string{"test": "abc"},
				runtimeLabels: map[string]string{"test": "abc", "test2": "def"},
			},
			wantChanged:   true,
			wantNewLabels: map[string]string{"test": "abc", "test2": "def"},
		},
		{
			name: "test-remove",
			args: args{
				crtLabels:     map[string]string{"test": "abc"},
				runtimeLabels: map[string]string{},
			},
			wantChanged:   true,
			wantNewLabels: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewLabels := j.isLabelsChanged(tt.args.crtLabels, tt.args.runtimeLabels)
			if gotChanged != tt.wantChanged {
				t.Errorf("isLabelsChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if !isMapEqual(gotNewLabels, tt.wantNewLabels) {
				t.Errorf("isLabelsChanged() gotNewLabels = %v, want %v", gotNewLabels, tt.wantNewLabels)
			}
		})
	}
}

func TestJuiceFSEngine_isAnnotationsChanged(t *testing.T) {
	type args struct {
		crtAnnotations     map[string]string
		runtimeAnnotations map[string]string
	}
	tests := []struct {
		name               string
		args               args
		wantChanged        bool
		wantNewAnnotations map[string]string
	}{
		{
			name: "test-false",
			args: args{
				crtAnnotations:     map[string]string{"test": "abc"},
				runtimeAnnotations: map[string]string{"test": "abc"},
			},
			wantChanged:        false,
			wantNewAnnotations: map[string]string{"test": "abc"},
		},
		{
			name: "test-changed",
			args: args{
				crtAnnotations:     map[string]string{"test": "abc"},
				runtimeAnnotations: map[string]string{"test": "def"},
			},
			wantChanged:        true,
			wantNewAnnotations: map[string]string{"test": "def"},
		},
		{
			name: "test-new",
			args: args{
				crtAnnotations:     map[string]string{"test": "abc"},
				runtimeAnnotations: map[string]string{"test": "abc", "test2": "def"},
			},
			wantChanged:        true,
			wantNewAnnotations: map[string]string{"test": "abc", "test2": "def"},
		},
		{
			name: "test-remove",
			args: args{
				crtAnnotations:     map[string]string{"test": "abc"},
				runtimeAnnotations: map[string]string{},
			},
			wantChanged:        true,
			wantNewAnnotations: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JuiceFSEngine{
				Log: fake.NullLogger(),
			}
			gotChanged, gotNewAnnotations := j.isAnnotationsChanged(tt.args.crtAnnotations, tt.args.runtimeAnnotations)
			if gotChanged != tt.wantChanged {
				t.Errorf("isAnnotationsChanged() gotChanged = %v, want %v", gotChanged, tt.wantChanged)
			}
			if !isMapEqual(gotNewAnnotations, tt.wantNewAnnotations) {
				t.Errorf("isAnnotationsChanged() gotNewAnnotations = %v, want %v", gotNewAnnotations, tt.wantNewAnnotations)
			}
		})
	}
}

var _ = Describe("JuiceFSEngine.checkAndSetFuseChanges()", func() {
	var engine *JuiceFSEngine

	BeforeEach(func() {
		engine = &JuiceFSEngine{
			name:      "test",
			namespace: "default",
		}
	})

	Context("When checking Fuse changes", func() {
		var (
			currentValue *JuiceFS
			latestValue  *JuiceFS
			runtime      *datav1alpha1.JuiceFSRuntime
			fuseToUpdate *appsv1.DaemonSet
		)
		BeforeEach(func() {
			currentValue = constructBaseRuntimeValue()
			latestValue = constructBaseRuntimeValue()
			fuseToUpdate = constructBaseFuseDaemonset()

			runtime = &datav1alpha1.JuiceFSRuntime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Spec: datav1alpha1.JuiceFSRuntimeSpec{
					Fuse: datav1alpha1.JuiceFSFuseSpec{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
					},
				},
			}
		})

		Context("When latestValue chnaged because runtime.spec changed (i.e. user update/patch runtime.spec)", func() {
			Context("Detect fuse nodeSelector changes", func() {
				cases := []struct {
					caseText            string
					latestNodeSelectors map[string]string
					changed             bool
				}{
					{
						caseText:            "should have no nodeSelector changed",
						latestNodeSelectors: map[string]string{"key1": "value1"},
						changed:             false,
					},
					{
						caseText:            "should add a new nodeSelector",
						latestNodeSelectors: map[string]string{"key1": "value1", "key2": "value2"},
						changed:             true,
					},
					{
						caseText:            "should remove a new nodeSelector",
						latestNodeSelectors: map[string]string{},
						changed:             true,
					},
					{
						caseText:            "should update a old nodeSelector",
						latestNodeSelectors: map[string]string{"key1": "new-value"},
						changed:             true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.NodeSelector = c.latestNodeSelectors
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						// Make sure helm related configurations are not touched
						Expect(fuseToUpdate.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue("helm_key1", "helm_value"))
						for k, v := range c.latestNodeSelectors {
							Expect(fuseToUpdate.Spec.Template.Spec.NodeSelector).To(HaveKeyWithValue(k, v))
						}
					})
				}
			})

			Context("detect volume chnages", func() {
				cases := []struct {
					caseText      string
					latestVolumes []corev1.Volume
					changed       bool
				}{
					{
						caseText: "should have no volume changed",
						latestVolumes: []corev1.Volume{
							{
								Name: "test-volume",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
						changed: false,
					},
					{
						caseText: "should add a new volume",
						latestVolumes: []corev1.Volume{
							{
								Name: "test-volume",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
							{
								Name: "new-volume",
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
						changed: true,
					},
					{
						caseText:      "should remove a volume",
						latestVolumes: []corev1.Volume{},
						changed:       true,
					},
					{
						caseText: "should update a volume",
						latestVolumes: []corev1.Volume{
							{
								Name: "test-volume",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/tmp",
									},
								},
							},
						},
						changed: true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.Volumes = c.latestVolumes
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						// Make sure helm related configurations are not touched
						Expect(fuseToUpdate.Spec.Template.Spec.Volumes).To(ContainElement(corev1.Volume{
							Name: "helm-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							}}))

						for _, vol := range c.latestVolumes {
							Expect(fuseToUpdate.Spec.Template.Spec.Volumes).To(ContainElement(vol))
						}
					})
				}
			})

			Context("detect label changes", func() {
				cases := []struct {
					caseText     string
					latestLabels map[string]string
					changed      bool
				}{
					{
						caseText:     "should have no label changed",
						latestLabels: map[string]string{"label1": "value1"},
						changed:      false,
					},
					{
						caseText:     "should add a new label",
						latestLabels: map[string]string{"label1": "value1", "label2": "value2"},
						changed:      true,
					},
					{
						caseText:     "should remove a label",
						latestLabels: map[string]string{},
						changed:      true,
					},
					{
						caseText:     "should update a label",
						latestLabels: map[string]string{"label1": "new-value"},
						changed:      true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.Labels = c.latestLabels
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						// Make sure helm related configurations are not touched
						Expect(fuseToUpdate.Spec.Template.Labels).To(HaveKeyWithValue("helm_label", "helm_value"))
						for k, v := range c.latestLabels {
							Expect(fuseToUpdate.Spec.Template.Labels).To(HaveKeyWithValue(k, v))
						}
					})
				}
			})

			Context("detect annotations changes", func() {
				cases := []struct {
					caseText          string
					latestAnnotations map[string]string
					changed           bool
				}{
					{
						caseText:          "should have no annotation changed",
						latestAnnotations: map[string]string{"annotation1": "value1"},
						changed:           false,
					},
					{
						caseText:          "should add a new annotation",
						latestAnnotations: map[string]string{"annotation1": "value1", "annotation2": "value2"},
						changed:           true,
					},
					{
						caseText:          "should remove an annotation",
						latestAnnotations: map[string]string{},
						changed:           true,
					},
					{
						caseText:          "should update an annotation",
						latestAnnotations: map[string]string{"annotation1": "new-value"},
						changed:           true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.Annotations = c.latestAnnotations
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						// Make sure helm related configurations are not touched
						Expect(fuseToUpdate.Spec.Template.Annotations).To(HaveKeyWithValue("helm_annotation", "helm_value"))
						for k, v := range c.latestAnnotations {
							Expect(fuseToUpdate.Spec.Template.Annotations).To(HaveKeyWithValue(k, v))
						}
					})
				}
			})

			Context("detect resources chnages", func() {
				cases := []struct {
					caseText         string
					latestResources  common.Resources
					runtimeResources corev1.ResourceRequirements
					changed          bool
				}{
					{
						caseText: "should have no resource changed",
						latestResources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
							Limits: common.ResourceList{
								corev1.ResourceCPU:    "200m",
								corev1.ResourceMemory: "200Mi",
							},
						},
						runtimeResources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
						changed: false,
					},
					{
						caseText: "should change resource requests",
						latestResources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "200m",
								corev1.ResourceMemory: "200Mi",
							},
							Limits: common.ResourceList{
								corev1.ResourceCPU:    "200m",
								corev1.ResourceMemory: "200Mi",
							},
						},
						runtimeResources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("200Mi"),
							},
						},
						changed: true,
					},
					{
						caseText: "should change resource limits",
						latestResources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "100m",
								corev1.ResourceMemory: "100Mi",
							},
							Limits: common.ResourceList{
								corev1.ResourceCPU:    "400m",
								corev1.ResourceMemory: "400Mi",
							},
						},
						runtimeResources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("100Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("400m"),
								corev1.ResourceMemory: resource.MustParse("400Mi"),
							},
						},
						changed: true,
					},
					{
						caseText: "should change both resource requests and limits",
						latestResources: common.Resources{
							Requests: common.ResourceList{
								corev1.ResourceCPU:    "300m",
								corev1.ResourceMemory: "300Mi",
							},
							Limits: common.ResourceList{
								corev1.ResourceCPU:    "600m",
								corev1.ResourceMemory: "600Mi",
							},
						},
						runtimeResources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("300m"),
								corev1.ResourceMemory: resource.MustParse("300Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("600m"),
								corev1.ResourceMemory: resource.MustParse("600Mi"),
							},
						},
						changed: true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.Resources = c.latestResources
						runtime.Spec.Fuse.Resources = c.runtimeResources
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						if c.changed {
							Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].Resources).To(Equal(c.runtimeResources))
						}
					})
				}
			})

			Context("detect environment variable changes", func() {
				cases := []struct {
					caseText   string
					latestEnvs []corev1.EnvVar
					changed    bool
				}{
					{
						caseText: "should have no env changed",
						latestEnvs: []corev1.EnvVar{
							{
								Name:  "ENV1",
								Value: "value1",
							},
						},
						changed: false,
					},
					{
						caseText: "should add a new env",
						latestEnvs: []corev1.EnvVar{
							{
								Name:  "ENV1",
								Value: "value1",
							},
							{
								Name:  "ENV2",
								Value: "value2",
							},
						},
						changed: true,
					},
					{
						caseText:   "should remove an env",
						latestEnvs: []corev1.EnvVar{},
						changed:    true,
					},
					{
						caseText: "should update an env",
						latestEnvs: []corev1.EnvVar{
							{
								Name:  "ENV1",
								Value: "new-value",
							},
						},
						changed: true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.Envs = c.latestEnvs
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						// Make sure helm related configurations are not touched
						Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].Env).To(ContainElement(corev1.EnvVar{
							Name:  "HELM_ENV1",
							Value: "HELM_value1",
						}))

						for _, env := range c.latestEnvs {
							Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].Env).To(ContainElement(env))
						}
					})
				}
			})

			Context("detect volume mount changes", func() {
				cases := []struct {
					caseText           string
					latestVolumeMounts []corev1.VolumeMount
					changed            bool
				}{
					{
						caseText: "should have no volume mount changed",
						latestVolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test-volume",
								MountPath: "/test",
							},
						},
						changed: false,
					},
					{
						caseText: "should add a new volume mount",
						latestVolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test-volume",
								MountPath: "/test",
							},
							{
								Name:      "new-volume",
								MountPath: "/new-path",
							},
						},
						changed: true,
					},
					{
						caseText:           "should remove a volume mount",
						latestVolumeMounts: []corev1.VolumeMount{},
						changed:            true,
					},
					{
						caseText: "should update a volume mount",
						latestVolumeMounts: []corev1.VolumeMount{
							{
								Name:      "test-volume",
								MountPath: "/updated-path",
							},
						},
						changed: true,
					},
				}

				for _, c := range cases {
					It(c.caseText, func() {
						latestValue.Fuse.VolumeMounts = c.latestVolumeMounts
						changed, _ := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)
						Expect(changed).To(Equal(c.changed))
						// Make sure helm related configurations are not touched
						Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(corev1.VolumeMount{
							Name:      "helm-volume",
							MountPath: "/test_helm",
						}))

						for _, vm := range c.latestVolumeMounts {
							Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].VolumeMounts).To(ContainElement(vm))
						}
					})
				}
			})

			Context("detect image changes", func() {
				It("should not do anything if runtime has no image or imageTag set", func() {
					runtime.Spec.Fuse.Image = ""
					runtime.Spec.Fuse.ImageTag = ""

					// latestValue changes probably because our default image chnages
					latestValue.Fuse.Image = "juicefs/fuse"
					latestValue.Fuse.ImageTag = "v2.0.0"

					changed, generationNeedUpdate := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)

					Expect(changed).To(BeFalse())
					Expect(generationNeedUpdate).To(BeFalse())
					Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].Image).To(Equal(fmt.Sprintf("%s:%s", currentValue.Fuse.Image, currentValue.Fuse.ImageTag)))
				})

				It("should keep image unchanged if runtime.spec.fuse image info is not changed", func() {
					runtime.Spec.Fuse.Image = currentValue.Fuse.Image
					runtime.Spec.Fuse.ImageTag = currentValue.Fuse.ImageTag
					latestValue.Fuse.Image = currentValue.Fuse.Image
					latestValue.Fuse.ImageTag = currentValue.Fuse.ImageTag

					changed, generationNeedUpdate := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)

					Expect(changed).To(BeFalse())
					Expect(generationNeedUpdate).To(BeFalse())
					Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].Image).To(Equal(fmt.Sprintf("%s:%s", currentValue.Fuse.Image, currentValue.Fuse.ImageTag)))
				})

				It("Should detect image changes and require generation update", func() {
					latestValue.Fuse.Image = "juicefs/fuse"
					latestValue.Fuse.ImageTag = "v2.0.0"
					runtime.Spec.Fuse.Image = "juicefs/fuse"
					runtime.Spec.Fuse.ImageTag = "v2.0.0"

					changed, generationNeedUpdate := engine.checkAndSetFuseChanges(currentValue, latestValue, runtime, fuseToUpdate)

					fmt.Println(fuseToUpdate.Spec.Template.Spec.Containers[0].Image)
					Expect(changed).To(BeTrue())
					Expect(generationNeedUpdate).To(BeTrue())
					Expect(fuseToUpdate.Spec.Template.Spec.Containers[0].Image).To(Equal("juicefs/fuse:v2.0.0"))
				})
			})
		})
	})
})

func constructBaseRuntimeValue() *JuiceFS {
	return &JuiceFS{
		Fuse: Fuse{
			NodeSelector: map[string]string{
				"key1": "value1",
			},
			Volumes: []corev1.Volume{
				{
					Name: "test-volume",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
			Labels: map[string]string{
				"label1": "value1",
			},
			Annotations: map[string]string{
				"annotation1": "value1",
			},
			Envs: []corev1.EnvVar{
				{
					Name:  "ENV1",
					Value: "value1",
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "test-volume",
					MountPath: "/test",
				},
			},
			Resources: common.Resources{
				Requests: common.ResourceList{
					corev1.ResourceCPU:    "100m",
					corev1.ResourceMemory: "100Mi",
				},
				Limits: common.ResourceList{
					corev1.ResourceCPU:    "200m",
					corev1.ResourceMemory: "200Mi",
				},
			},
			Image:    "juicefs/fuse",
			ImageTag: "v1.0.0",
		},
	}
}

func constructBaseFuseDaemonset() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-fuse",
			Namespace: "default",
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"label1":     "value1",
						"helm_label": "helm_value",
					},
					Annotations: map[string]string{
						"annotation1":     "value1",
						"helm_annotation": "helm_value",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "helm-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					NodeSelector: map[string]string{
						"key1":      "value1",
						"helm_key1": "helm_value",
					},
					Containers: []corev1.Container{
						{
							Name:  JuiceFSFuseContainerName,
							Image: "juicefs/fuse:v1.0.0",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("100Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("200m"),
									corev1.ResourceMemory: resource.MustParse("200Mi"),
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "ENV1",
									Value: "value1",
								},
								{
									Name:  "HELM_ENV1",
									Value: "HELM_value1",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-volume",
									MountPath: "/test",
								},
								{
									Name:      "helm-volume",
									MountPath: "/test_helm",
								},
							},
						},
					},
				},
			},
		},
	}
}
