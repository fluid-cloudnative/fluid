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

package utils

import (
	"reflect"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestTransformRequirementsToResources(t *testing.T) {
	testCases := map[string]struct {
		required corev1.ResourceRequirements
		wantRes  common.Resources
	}{
		"test resource transform case 1": {
			required: mockRequiredResource(
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
				corev1.ResourceList{"cpu": resource.MustParse("200m")},
			),
			wantRes: common.Resources{
				Requests: common.ResourceList{"cpu": "100m"},
				Limits:   common.ResourceList{"cpu": "200m"},
			},
		},
		"test resource transform case 2": {
			required: mockRequiredResource(
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
				corev1.ResourceList{},
			),
			wantRes: common.Resources{
				Requests: common.ResourceList{"cpu": "100m"},
				Limits:   common.ResourceList{},
			},
		},
		"test resource transform case 3": {
			required: mockRequiredResource(
				corev1.ResourceList{"memory": resource.MustParse("100Gi"), "cpu": resource.MustParse("100m")},
				corev1.ResourceList{"memory": resource.MustParse("600Gi"), "cpu": resource.MustParse("600m")},
			),
			wantRes: common.Resources{
				Requests: common.ResourceList{"memory": "100Gi", "cpu": "100m"},
				Limits:   common.ResourceList{"memory": "600Gi", "cpu": "600m"},
			},
		},
		"test resource transform case 4": {
			required: mockRequiredResource(
				corev1.ResourceList{},
				corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
			),
			wantRes: common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{"nvidia.com/gpu": "1"},
			},
		},
		"test resource transform case 5": {
			required: mockRequiredResource(
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
				corev1.ResourceList{"cpu": resource.MustParse("200m"), "nvidia.com/gpu": resource.MustParse("1")},
			),
			wantRes: common.Resources{
				Requests: common.ResourceList{"cpu": "100m"},
				Limits:   common.ResourceList{"cpu": "200m", "nvidia.com/gpu": "1"},
			},
		},
		"test resource transform case 6": {
			required: mockRequiredResource(
				corev1.ResourceList{},
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
			),
			wantRes: common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{"cpu": "100m"},
			},
		},
	}

	for k, item := range testCases {
		got := TransformCoreV1ResourcesToInternalResources(item.required)
		if !reflect.DeepEqual(got, item.wantRes) {
			t.Errorf("%s check failure,want:%v,got:%v", k, item.wantRes, got)
		}
	}
}

func TestTransformInternalResourcesToCoreV1Resources(t *testing.T) {
	testCases := map[string]struct {
		internal common.Resources
		want     corev1.ResourceRequirements
		wantErr  bool
	}{
		"requests and limits": {
			internal: common.Resources{
				Requests: common.ResourceList{"cpu": "100m", "memory": "1Gi"},
				Limits:   common.ResourceList{"cpu": "2", "memory": "4Gi"},
			},
			want: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("4Gi"),
				},
			},
		},
		"empty lists become nil so that the result equals a zero-valued requirement": {
			internal: common.Resources{
				Requests: common.ResourceList{},
				Limits:   common.ResourceList{},
			},
			want: corev1.ResourceRequirements{},
		},
		"extended resource": {
			internal: common.Resources{
				Limits: common.ResourceList{"nvidia.com/gpu": "1"},
			},
			want: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
			},
		},
		"invalid quantity in requests": {
			internal: common.Resources{
				Requests: common.ResourceList{"cpu": "not-a-quantity"},
			},
			wantErr: true,
		},
		"invalid quantity in limits": {
			internal: common.Resources{
				Limits: common.ResourceList{"memory": "10VeryBig"},
			},
			wantErr: true,
		},
	}

	for name, item := range testCases {
		got, err := TransformInternalResourcesToCoreV1Resources(item.internal)
		if item.wantErr {
			if err == nil {
				t.Errorf("%s: expected an error but got none", name)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error %v", name, err)
			continue
		}
		if !reflect.DeepEqual(got, item.want) {
			t.Errorf("%s check failure, want: %v, got: %v", name, item.want, got)
		}
	}
}

// TestTransformResourcesRoundTrip makes sure a value that went through the Helm values ConfigMap
// can be compared against the live workload it was rendered into.
func TestTransformResourcesRoundTrip(t *testing.T) {
	for _, original := range []corev1.ResourceRequirements{
		{
			Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1000m")},
			Limits:   corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("4")},
		},
		{
			Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("1Gi")},
		},
	} {
		roundTripped, err := TransformInternalResourcesToCoreV1Resources(
			TransformCoreV1ResourcesToInternalResources(original))
		if err != nil {
			t.Fatalf("unexpected error round tripping %v: %v", original, err)
		}
		if !ResourceRequirementsEqual(original, roundTripped) {
			t.Errorf("round trip changed the resources, want: %v, got: %v", original, roundTripped)
		}
	}
}

func mockRequiredResource(req, limit corev1.ResourceList) corev1.ResourceRequirements {
	res := corev1.ResourceRequirements{}
	if len(req) > 0 {
		res.Requests = req
	}
	if len(limit) > 0 {
		res.Limits = limit
	}
	return res
}

func TestResourceRequirementsEqual(t *testing.T) {
	type args struct {
		source corev1.ResourceRequirements
		target corev1.ResourceRequirements
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "memory resource emty and nil",
			args: args{
				source: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
				}, target: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("100m"),
						corev1.ResourceMemory: resource.MustParse("0"),
					}, Limits: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
				},
			}, want: true,
		}, {
			name: "no limit",
			args: args{
				source: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{},
					Requests: corev1.ResourceList{},
				}, target: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("0"),
					},
				},
			}, want: true,
		}, {
			name: "no resources",
			args: args{
				source: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{},
					Requests: corev1.ResourceList{},
				}, target: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("0"),
					},
				},
			}, want: true,
		}, {
			name: "resource list is different",
			args: args{
				source: corev1.ResourceRequirements{
					Limits:   corev1.ResourceList{},
					Requests: corev1.ResourceList{},
				}, target: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("10Gi"),
					},
				},
			}, want: false,
		}, {
			name: "resource value is different",
			args: args{
				source: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("10Gi"),
					},
				}, target: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("20Gi"),
					},
				},
			}, want: false,
		}, {
			name: "resource value is different",
			args: args{
				source: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("10Gi"),
					},
				}, target: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("100m"),
					},
				},
			}, want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResourceRequirementsEqual(tt.args.source, tt.args.target); got != tt.want {
				t.Errorf("testcase %s ResourceRequirementsEqual() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
