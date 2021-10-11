/*

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
	"strings"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestTransformWorker(t *testing.T) {
	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")

	result := resource.MustParse("20Gi")
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     string
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
			}}, &Jindo{}, "1G"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: log.NullLogger{}}
		test.jindoValue.Worker.Port.Rpc = 8001
		test.jindoValue.Worker.Port.Raft = 8002
		metaPath := "/var/lib/docker/meta"
		dataPath := "/var/lib/docker/data"
		userQuotas := "1G"
		err := engine.transformWorker(test.runtime, metaPath, dataPath, userQuotas, test.jindoValue)
		if err != nil {
			t.Errorf("Got err %v", err)
		}
		if test.jindoValue.Worker.WorkerProperties["storage.data-dirs.capacities"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Worker.WorkerProperties["storage.data-dirs.capacities"])
		}
	}
}

func TestTransformWorkerMountPath(t *testing.T) {
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
				}},
			}}, &Jindo{}, "/mnt/disk2"},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: log.NullLogger{}}
		stroagePath := "/mnt/disk1,/mnt/disk2"
		originPath := strings.Split(stroagePath, ",")
		properties := engine.transformWorkerMountPath(originPath)
		if properties["2"] != test.expect {
			t.Errorf("expected value %v, but got %v", test.expect, properties["2"])
		}
	}
}

func TestTransformResourcesForWorkerNoValue(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		jindoValue *Jindo
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{},
		}, &Jindo{
			Properties: map[string]string{},
		}},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: log.NullLogger{}}
		engine.transformResourcesForWorker(test.runtime, test.jindoValue)
		if test.jindoValue.Worker.Resources.Requests.Memory != "" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Requests.Memory)
		}
		if test.jindoValue.Worker.Resources.Requests.CPU != "" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Requests.CPU)
		}
		if test.jindoValue.Worker.Resources.Limits.Memory != "" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Limits.Memory)
		}
		if test.jindoValue.Worker.Resources.Limits.CPU != "" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Limits.CPU)
		}
	}
}

func TestTransformResourcesForWorkerWithValue(t *testing.T) {

	resources := corev1.ResourceRequirements{}
	resources.Limits = make(corev1.ResourceList)
	resources.Limits[corev1.ResourceMemory] = resource.MustParse("2Gi")
	resources.Limits[corev1.ResourceCPU] = resource.MustParse("2")
	resources.Requests = make(corev1.ResourceList)
	resources.Requests[corev1.ResourceMemory] = resource.MustParse("1Gi")
	resources.Requests[corev1.ResourceCPU] = resource.MustParse("1")
	result := resource.MustParse("20Gi")

	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		jindoValue *Jindo
	}{
		{&datav1alpha1.JindoRuntime{
			Spec: datav1alpha1.JindoRuntimeSpec{
				Worker: datav1alpha1.JindoCompTemplateSpec{
					Resources: resources,
				},
				TieredStore: datav1alpha1.TieredStore{
					Levels: []datav1alpha1.Level{{
						MediumType: common.Memory,
						Quota:      &result,
					}},
				},
			},
		}, &Jindo{
			Properties: map[string]string{},
			Master:     Master{},
		}},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: log.NullLogger{}}
		engine.transformResourcesForWorker(test.runtime, test.jindoValue)
		if test.jindoValue.Worker.Resources.Requests.Memory != "1Gi" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Requests.Memory)
		}
		if test.jindoValue.Worker.Resources.Requests.CPU != "1" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Requests.CPU)
		}
		if test.jindoValue.Worker.Resources.Limits.Memory != "2Gi" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Limits.Memory)
		}
		if test.jindoValue.Worker.Resources.Limits.CPU != "2" {
			t.Errorf("expected nil, got %v", test.jindoValue.Worker.Resources.Limits.CPU)
		}
	}
}

func TestTransformWorkerAffinity(t *testing.T) {
	var tests = []struct {
		runtime    *datav1alpha1.JindoRuntime
		dataset    *datav1alpha1.Dataset
		jindoValue *Jindo
		expect     v1.Affinity
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
				},
				},
				NodeAffinity: &datav1alpha1.CacheableNodeAffinity{
					Required: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "type",
										Operator: v1.NodeSelectorOpDoesNotExist,
									},
								},
							},
						},
					},
				},
			}}, &Jindo{
			Worker: Worker{},
		}, v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "type",
									Operator: v1.NodeSelectorOpDoesNotExist,
								},
							},
						},
					},
				},
			},
		}},
	}
	for _, test := range tests {
		engine := &JindoEngine{Log: log.NullLogger{}}
		engine.transformWorkerAffinity(test.dataset, test.runtime, test.jindoValue)
		if test.jindoValue.Worker.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key !=
			test.expect.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key {
			t.Errorf("expected value %v, but got %v", test.expect, test.jindoValue.Worker.Affinity)
		}
	}
}
