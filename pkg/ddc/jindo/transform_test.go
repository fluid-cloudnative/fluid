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

package jindo

import (
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
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
		engine := &JindoEngine{Log: fake.NullLogger()}
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
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Secret: "secret",
			},
		}, &datav1alpha1.Dataset{
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{{
					MountPoint: "local:///mnt/test",
					Name:       "test",
					Path:       "/",
				}},
			}}, &Jindo{}, "registry.cn-shanghai.aliyuncs.com/jindofs/smartdata:3.8.0"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: fake.NullLogger()}
		imageR, tagR, _ := engine.getSmartDataConfigs()
		registryVersion := imageR + ":" + tagR
		if registryVersion != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Fuse.RunAs)
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
		engine := &JindoEngine{Log: fake.NullLogger()}
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
		engine := &JindoEngine{Log: fake.NullLogger()}
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
		engine := &JindoEngine{Log: fake.NullLogger()}
		engine.transformNetworkMode(test.runtime, test.jindoValue)
		test.jindoValue.Master.ReplicaCount = 3
		err := engine.allocatePorts(test.jindoValue)
		if test.jindoValue.Master.Port.Rpc != test.expect && err != nil {
			t.Errorf("expected value %v, but got %v, and err %v", test.expect, test.jindoValue.Master.Port.Rpc, err)
		}
	}
}
