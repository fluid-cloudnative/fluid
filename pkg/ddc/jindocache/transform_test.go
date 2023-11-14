/*
Copyright 2023 The Fluid Authors.

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

package jindocache

import (
	"os"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/net"
)

func TestTransformTolerations(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     int
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				Master: datav1alpha1.JindoCompTemplateSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "master",
						Operator: "Equals",
						Value:    "true",
					}},
				},
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "worker",
						Operator: "Equals",
						Value:    "true",
					}},
				},
				Fuse: datav1alpha1.JindoFuseSpec{
					Tolerations: []corev1.Toleration{{
						Key:      "fuse",
						Operator: "Equals",
						Value:    "true",
					}},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
				Tolerations: []corev1.Toleration{{
					Key:      "jindo",
					Operator: "Equals",
					Value:    "true",
				}},
			}}, &Jindo{}, 2,
		},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		engine.transformTolerations(test.dataset, test.runtime, test.jindoValue)
		if len(test.jindoValue.Master.Tolerations) != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.Tolerations)
		}
		if len(test.jindoValue.Worker.Tolerations) != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.Tolerations)
		}
		if len(test.jindoValue.Fuse.Tolerations) != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Master.Tolerations)
		}
	}
}

func TestParseSmartDataImage(t *testing.T) {
	var tests = []struct {
		runtime               *datav1alpha1.JindoRuntime
		dataset               *datav1alpha1.Dataset
		jindoValue            *Jindo
		expect                string
		expectImagePullPolicy string
		expectDnsServer       string
	}{
		{
			runtime: &datav1alpha1.JindoRuntime{
				Spec: datav1alpha1.JindoRuntimeSpec{
					Secret: "secret",
				}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "local:///mnt/test",
						Name:       "test",
						Path:       "/",
					}},
				}},
			jindoValue:            &Jindo{},
			expect:                "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:5.0.0",
			expectImagePullPolicy: "Always",
			expectDnsServer:       "1.1.1.1",
		},
		{
			runtime: &datav1alpha1.JindoRuntime{
				Spec: datav1alpha1.JindoRuntimeSpec{
					Secret: "secret",
					JindoVersion: datav1alpha1.VersionSpec{
						Image:           "jindofs/smartdata",
						ImageTag:        "testtag",
						ImagePullPolicy: "IfNotPresent",
					},
				}},
			dataset: &datav1alpha1.Dataset{
				Spec: datav1alpha1.DatasetSpec{
					Mounts: []datav1alpha1.Mount{{
						MountPoint: "local:///mnt/test",
						Name:       "test",
						Path:       "/",
					}},
				}},
			jindoValue:            &Jindo{},
			expect:                "jindofs/smartdata:testtag",
			expectImagePullPolicy: "IfNotPresent",
			expectDnsServer:       "1.1.1.1",
		},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		smartdataConfig := engine.getSmartDataConfigs(test.runtime)
		registryVersion := smartdataConfig.image + ":" + smartdataConfig.imageTag
		if registryVersion != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, registryVersion)
		}
		if smartdataConfig.imagePullPolicy != test.expectImagePullPolicy {
			t.Errorf("expected imagePullPolicy %v, but got %v", test.expectImagePullPolicy, smartdataConfig.imagePullPolicy)
		}
		if smartdataConfig.dnsServer != test.expectDnsServer {
			t.Errorf("expected dnsServer %v, but got %v", test.expectDnsServer, smartdataConfig.dnsServer)
		}
	}
}

func TestTransformHostNetWork(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     bool
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, true,
		},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				NetworkMode: "HostNetwork",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, true,
		},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				NetworkMode: "ContainerNetwork",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, false,
		},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		engine.transformNetworkMode(test.runtime, test.jindoValue)
		if test.jindoValue.UseHostNetwork != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.UseHostNetwork)
		}
	}

	var errortests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     bool
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				NetworkMode: "Non",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, false,
		},
	}
	for _, test := range errortests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		engine.transformNetworkMode(test.runtime, test.jindoValue)
		if test.jindoValue.UseHostNetwork != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.UseHostNetwork)
		}
	}
}

func TestTransformAllocatePorts(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     int
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				NetworkMode: "ContainerNetwork",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, 8101,
		},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
						High:       "0.8",
						Low:        "0.1",
					}},
				},
				NetworkMode: "ContainerNetwork",
				Replicas:    3,
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
				}},
			}}, &Jindo{}, 8101,
		},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		engine.transformNetworkMode(test.runtime, test.jindoValue)
		test.jindoValue.Master.ReplicaCount = 3
		err := engine.allocatePorts(test.jindoValue)
		if test.jindoValue.Master.Port.Rpc != test.expect && err != nil {
			t.Errorf("expected value %v, but got %v, and err %v", test.expect, test.jindoValue.Master.Port.Rpc, err)
		}
	}
}

// func TestTransformResource(t *testing.T) {
// 	resources := corev1.ResourceRequirements{}
// 	resources.Limits = make(corev1.ResourceList)
// 	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

// 	result := resource.MustParse("200Gi")
// 	var tests = []struct {
// 		runtime      *datav1alpha1.JindoRuntime
// 		jindoValue   *Jindo
// 		userQuatas   string
// 		expectMaster string
// 		expectWorker string
// 	}{
// 		{&datav1alpha1.JindoRuntime{
// 			Spec: datav1alpha1.JindoRuntimeSpec{
// 				Secret: "secret",
// 				TieredStore: datav1alpha1.TieredStore{
// 					Levels: []datav1alpha1.Level{{
// 						MediumType: common.Memory,
// 						Quota:      &result,
// 						High:       "0.8",
// 						Low:        "0.1",
// 					}},
// 				},
// 				NetworkMode: "ContainerNetwork",
// 			},
// 		}, &Jindo{}, "200g", "30Gi", "200Gi",
// 		},
// 	}
// 	for _, test := range tests {
// 		engine := &JindoCacheEngine{Log: fake.NullLogger()}
// 		engine.transformResources(test.runtime, test.jindoValue, test.userQuatas)
// 		if test.jindoValue.Master.Resources.Requests.Memory != test.expectMaster ||
// 			test.jindoValue.Worker.Resources.Requests.Memory != test.expectWorker {
// 			t.Errorf("expected master value %v, worker value %v,  but got %v and %v", test.expectMaster, test.expectWorker, test.jindoValue.Master.Resources.Requests.Memory, test.jindoValue.Worker.Resources.Requests.Memory)
// 		}
// 	}
// }

func TestJindoCacheEngine_transformMasterResources(t *testing.T) {
	os.Setenv("USE_DEFAULT_MEM_LIMIT", "true")
	type fields struct {
		name      string
		namespace string
	}
	type args struct {
		runtime    *datav1alpha1.JindoRuntime
		value      *Jindo
		userQuotas string
	}
	quotas := []resource.Quantity{resource.MustParse("200Gi"), resource.MustParse("10Gi")}
	tests := []struct {
		name                string
		fields              fields
		args                args
		wantErr             bool
		wantRuntimeResource corev1.ResourceRequirements
		wantValue           Resources
	}{
		// TODO: Add test cases.
		{
			name: "runtime_resource_is_null",
			fields: fields{
				name:      "testNull",
				namespace: "default",
			},
			args: args{
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "testNull",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						TieredStore: datav1alpha1.TieredStore{
							Levels: []datav1alpha1.Level{{
								MediumType: common.Memory,
								Quota:      &quotas[0],
								High:       "0.8",
								Low:        "0.1",
							}},
						},
					},
				},
				value:      &Jindo{},
				userQuotas: "200g",
			},
			wantErr: false,
			wantRuntimeResource: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("30Gi"),
				},
			},
			wantValue: Resources{
				Requests: Resource{
					Memory: "30Gi",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.args.runtime.DeepCopy())
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.args.runtime)
			_ = corev1.AddToScheme(s)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoCacheEngine{
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
			}

			if err := e.transformMasterResources(tt.args.runtime, tt.args.value, tt.args.userQuotas); (err != nil) != tt.wantErr {
				t.Errorf("JindoCacheEngine.transformMasterResources() error = %v, wantErr %v", err, tt.wantErr)
			}

			runtime, err := e.getRuntime()
			if err != nil {
				t.Errorf("JindoCacheEngine.getRUntime() error = %v", err)
			}

			if !ResourceRequirementsEqual(tt.wantRuntimeResource, runtime.Spec.Master.Resources) {
				t.Errorf("JindoCacheEngine.transformMasterResources() runtime = %v, wantRuntime %v", runtime.Spec.Master.Resources, tt.wantRuntimeResource)
			}

			if !reflect.DeepEqual(tt.wantValue, tt.args.value.Master.Resources) {
				t.Errorf("JindoCacheEngine.transformMasterResources() value = %v, wantValue %v", tt.args.value.Master.Resources, tt.wantValue)
			}

		})
	}
}

func ResourceRequirementsEqual(source corev1.ResourceRequirements,
	target corev1.ResourceRequirements) bool {
	return resourceListsEqual(source.Requests, target.Requests) &&
		resourceListsEqual(source.Limits, target.Limits)
}

func resourceListsEqual(a corev1.ResourceList, b corev1.ResourceList) bool {
	a = withoutZeroElems(a)
	b = withoutZeroElems(b)
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		vb, found := b[k]
		if !found {
			return false
		}
		if v.Cmp(vb) != 0 {
			return false
		}
	}
	return true
}

func withoutZeroElems(input corev1.ResourceList) (output corev1.ResourceList) {
	output = corev1.ResourceList{}
	for k, v := range input {
		if !v.IsZero() {
			output[k] = v
		}
	}
	return
}

func TestJindoCacheEngine_transform(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
		dataset   *datav1alpha1.Dataset
	}
	type args struct {
		runtime *datav1alpha1.JindoRuntime
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue *Jindo
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "fuseOnly",
			fields: fields{
				name:      "test",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
				},
			},
		},
		{
			name: "pvc",
			fields: fields{
				name:      "test",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
							},
						},
					},
				},
			},
		},
		{
			name: "pvc-subpath",
			fields: fields{
				name:      "test",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data/subpath",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.fields.runtime.DeepCopy())
			runtimeObjs = append(runtimeObjs, tt.fields.dataset.DeepCopy())
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.dataset)
			s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.DatasetList{})
			_ = corev1.AddToScheme(s)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoCacheEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}
			tt.args.runtime = tt.fields.runtime
			err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
			if err != nil {
				t.Fatalf("failed to set up runtime port allocator due to %v", err)
			}
			_, err = e.transform(tt.args.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoCacheEngine.transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestJindoCacheEngine_transformDeployMode(t *testing.T) {

	type args struct {
		runtime *datav1alpha1.JindoRuntime
		value   *Jindo
	}
	tests := []struct {
		name string
		args args
		want *Jindo
	}{
		// TODO: Add test cases.
		{
			name: "replicas is 1, enabled",
			args: args{
				runtime: &datav1alpha1.JindoRuntime{},
				value:   &Jindo{},
			},
			want: &Jindo{
				Master: Master{
					ServiceCount: 1,
					ReplicaCount: 1,
				},
			},
		}, {
			name: "replicas is 1, disabled",
			args: args{
				runtime: &datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Disabled: true,
						},
					},
				},
				value: &Jindo{},
			},
			want: &Jindo{
				Master: Master{
					ServiceCount: 1,
					ReplicaCount: 0,
				},
			},
		}, {
			name: "replicas is 1, disabled",
			args: args{
				runtime: &datav1alpha1.JindoRuntime{
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Disabled: true,
							Replicas: 3,
						},
					},
				},
				value: &Jindo{},
			},
			want: &Jindo{
				Master: Master{
					ServiceCount: 3,
					ReplicaCount: 0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JindoCacheEngine{
				runtime: tt.args.runtime,
			}
			tt.args.value.Master.ReplicaCount = e.transformReplicasCount(tt.args.runtime)
			tt.args.value.Master.ServiceCount = e.transformReplicasCount(tt.args.runtime)
			e.transformDeployMode(tt.args.runtime, tt.args.value)
			if !reflect.DeepEqual(tt.want.Master, tt.args.value.Master) {
				t.Errorf("Testcase %s failed, wanted %v, got %v.",
					tt.name,
					tt.want.Master,
					tt.args.value.Master)
			}
		})
	}
}

func TestJindoCacheEngine_transformPodMetadata(t *testing.T) {
	engine := &JindoCacheEngine{Log: fake.NullLogger()}

	type testCase struct {
		Name    string
		Runtime *datav1alpha1.JindoRuntime
		Value   *Jindo

		wantValue *Jindo
	}

	testCases := []testCase{
		{
			Name: "set_common_labels_and_annotations",
			Runtime: &datav1alpha1.JindoRuntime{
				Spec: datav1alpha1.JindoRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
				},
			},
			Value: &Jindo{},
			wantValue: &Jindo{
				Master: Master{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
				Worker: Worker{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
				Fuse: Fuse{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
			},
		},
		{
			Name: "set_master_and_workers_labels_and_annotations",
			Runtime: &datav1alpha1.JindoRuntime{
				Spec: datav1alpha1.JindoRuntimeSpec{
					PodMetadata: datav1alpha1.PodMetadata{
						Labels:      map[string]string{"common-key": "common-value"},
						Annotations: map[string]string{"common-annotation": "val"},
					},
					Master: datav1alpha1.JindoCompTemplateSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "master-value"},
							Annotations: map[string]string{"common-annotation": "master-val"},
						},
					},
					Worker: datav1alpha1.JindoCompTemplateSpec{
						PodMetadata: datav1alpha1.PodMetadata{
							Labels:      map[string]string{"common-key": "worker-value"},
							Annotations: map[string]string{"common-annotation": "worker-val"},
						},
					},
				},
			},
			Value: &Jindo{},
			wantValue: &Jindo{
				Master: Master{
					Labels:      map[string]string{"common-key": "master-value"},
					Annotations: map[string]string{"common-annotation": "master-val"}},
				Worker: Worker{
					Labels:      map[string]string{"common-key": "worker-value"},
					Annotations: map[string]string{"common-annotation": "worker-val"},
				},
				Fuse: Fuse{
					Labels:      map[string]string{"common-key": "common-value"},
					Annotations: map[string]string{"common-annotation": "val"},
				},
			},
		},
	}

	for _, tt := range testCases {
		err := engine.transformPodMetadata(tt.Runtime, tt.Value)
		if err != nil {
			t.Fatalf("test name: %s. Expect err = nil, but got err = %v", tt.Name, err)
		}

		if !reflect.DeepEqual(tt.Value, tt.wantValue) {
			t.Fatalf("test name: %s. Expect value %v, but got %v", tt.Name, tt.wantValue, tt.Value)
		}
	}
}

func TestTransformLogConfig(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				LogConfig: map[string]string{"logger.level": "6"},
				Fuse: datav1alpha1.JindoFuseSpec{
					LogConfig: map[string]string{"logger.level": "6"},
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "6"},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		engine.transformLogConfig(test.runtime, test.jindoValue)
		if test.jindoValue.LogConfig["logger.level"] != test.expect || test.jindoValue.FuseLogConfig["logger.level"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
		}
	}
}

func TestJindoCacheEngine_transformEnvVariables(t *testing.T) {
	type args struct {
		runtime *datav1alpha1.JindoRuntime
		value   *Jindo
	}
	tests := []struct {
		name             string
		args             args
		expectMasterEnvs map[string]string
		expectWorkerEnvs map[string]string
		expectFuseEnvs   map[string]string
	}{
		{
			name: "no_env_variable",
			args: args{
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-no-env",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				value: &Jindo{},
			},
			expectMasterEnvs: nil,
			expectWorkerEnvs: nil,
			expectFuseEnvs:   nil,
		},
		{
			name: "all_env_variable_set",
			args: args{
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-all-env-set",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{
						Master: datav1alpha1.JindoCompTemplateSpec{
							Env: map[string]string{
								"test-master": "foo",
							},
						},
						Worker: datav1alpha1.JindoCompTemplateSpec{
							Env: map[string]string{
								"test-worker": "bar",
							},
						},
						Fuse: datav1alpha1.JindoFuseSpec{
							Env: map[string]string{
								"test-fuse": "test",
							},
						},
					},
				},
				value: &Jindo{},
			},
			expectMasterEnvs: map[string]string{"test-master": "foo"},
			expectWorkerEnvs: map[string]string{"test-worker": "bar"},
			expectFuseEnvs:   map[string]string{"test-fuse": "test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &JindoCacheEngine{
				name:      tt.args.runtime.Name,
				namespace: tt.args.runtime.Namespace,
				Log:       fake.NullLogger(),
			}
			e.transformEnvVariables(tt.args.runtime, tt.args.value)
			if !reflect.DeepEqual(tt.args.value.Master.Env, tt.expectMasterEnvs) {
				t.Fatalf("testcase %s failed: failed to transform env variable for master, expect: %v, got %v", tt.name, tt.args.value.Master.Env, tt.expectMasterEnvs)
			}

			if !reflect.DeepEqual(tt.args.value.Worker.Env, tt.expectWorkerEnvs) {
				t.Fatalf("testcase %s failed: failed to transform env variable for worker, expect: %v, got %v", tt.name, tt.args.value.Worker.Env, tt.expectWorkerEnvs)
			}

			if !reflect.DeepEqual(tt.args.value.Fuse.Env, tt.expectFuseEnvs) {
				t.Fatalf("testcase %s failed: failed to transform env variable for fuse, expect: %v, got %v", tt.name, tt.args.value.Fuse.Env, tt.expectFuseEnvs)
			}
		})
	}
}

func TestCheckIfSupportSecretMount(t *testing.T) {
	var tests = []struct {
		runtime      *datav1alpha1.JindoRuntime
		dataset      *datav1alpha1.Dataset
		smartdataTag string
		fuseTag      string
		expect       bool
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, "4.5.2", "4.5.2", false},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, "4.5.2", "5.0.0", false},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, "5.0.0", "5.0.0", true},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, "5.0.0", "5.0.0", true},
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Master: datav1alpha1.JindoCompTemplateSpec{
					Disabled: true,
				},
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Disabled: true,
				},
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, "4.5.2", "5.0.0", true},
	}
	for _, test := range tests {
		engine := &JindoCacheEngine{Log: fake.NullLogger()}
		result := engine.checkIfSupportSecretMount(test.runtime, test.smartdataTag, test.fuseTag)
		if result != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, result)
		}
	}
}

func TestJindoCacheEngine_transformPolicy(t *testing.T) {
	type fields struct {
		runtime   *datav1alpha1.JindoRuntime
		name      string
		namespace string
		dataset   *datav1alpha1.Dataset
	}
	type args struct {
		runtime *datav1alpha1.JindoRuntime
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantValue *Jindo
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name: "WRITE_THROUGH_ALWAYS",
			fields: fields{
				name:      "test",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
								Options: map[string]string{
									"writePolicy": "WRITE_THROUGH",
									"metaPolicy":  "ALWAYS",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CACHE_ONLY_ONCE",
			fields: fields{
				name:      "test",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data",
								Options: map[string]string{
									"writePolicy": "CACHE_ONLY",
									"metaPolicy":  "ONCE",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "CACHE_ONLY_ALWAYS",
			fields: fields{
				name:      "test",
				namespace: "default",
				runtime: &datav1alpha1.JindoRuntime{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.JindoRuntimeSpec{},
				},
				dataset: &datav1alpha1.Dataset{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "default",
					},
					Spec: datav1alpha1.DatasetSpec{
						Mounts: []datav1alpha1.Mount{
							{
								MountPoint: "pvc://data/subpath",
								Options: map[string]string{
									"writePolicy": "CACHE_ONLY",
									"metaPolicy":  "ALWAYS",
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			runtimeObjs := []runtime.Object{}
			runtimeObjs = append(runtimeObjs, tt.fields.runtime.DeepCopy())
			runtimeObjs = append(runtimeObjs, tt.fields.dataset.DeepCopy())
			s := runtime.NewScheme()
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.runtime)
			s.AddKnownTypes(datav1alpha1.GroupVersion, tt.fields.dataset)
			s.AddKnownTypes(datav1alpha1.GroupVersion, &datav1alpha1.DatasetList{})
			_ = corev1.AddToScheme(s)
			client := fake.NewFakeClientWithScheme(s, runtimeObjs...)
			e := &JindoCacheEngine{
				runtime:   tt.fields.runtime,
				name:      tt.fields.name,
				namespace: tt.fields.namespace,
				Client:    client,
				Log:       fake.NullLogger(),
			}
			tt.args.runtime = tt.fields.runtime
			err := portallocator.SetupRuntimePortAllocator(client, &net.PortRange{Base: 10, Size: 100}, "bitmap", GetReservedPorts)
			if err != nil {
				t.Fatalf("failed to set up runtime port allocator due to %v", err)
			}
			_, err = e.transform(tt.args.runtime)
			if (err != nil) != tt.wantErr {
				t.Errorf("JindoCacheEngine.transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
