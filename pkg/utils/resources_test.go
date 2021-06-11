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
			wantRes: mockCommonResource(
				common.ResourceList{"cpu": "100m"},
				common.ResourceList{"cpu": "200m"},
			),
		},
		"test resource transform case 2": {
			required: mockRequiredResource(
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
				corev1.ResourceList{},
			),
			wantRes: mockCommonResource(
				common.ResourceList{"cpu": "100m"},
				common.ResourceList{},
			),
		},
		"test resource transform case 3": {
			required: mockRequiredResource(
				corev1.ResourceList{"memory": resource.MustParse("100Gi"), "cpu": resource.MustParse("100m")},
				corev1.ResourceList{"memory": resource.MustParse("600Gi"), "cpu": resource.MustParse("600m")},
			),
			wantRes: mockCommonResource(
				common.ResourceList{"memory": "100Gi", "cpu": "100m"},
				common.ResourceList{"memory": "600Gi", "cpu": "600m"},
			),
		},
		"test resource transform case 4": {
			required: mockRequiredResource(
				corev1.ResourceList{},
				corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
			),
			wantRes: mockCommonResource(
				common.ResourceList{},
				common.ResourceList{"nvidia.com/gpu": "1"},
			),
		},
		"test resource transform case 5": {
			required: mockRequiredResource(
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
				corev1.ResourceList{"cpu": resource.MustParse("200m"), "nvidia.com/gpu": resource.MustParse("1")},
			),
			wantRes: mockCommonResource(
				common.ResourceList{"cpu": "100m"},
				common.ResourceList{"cpu": "200m", "nvidia.com/gpu": "1"},
			),
		},
		"test resource transform case 6": {
			required: mockRequiredResource(
				corev1.ResourceList{},
				corev1.ResourceList{"cpu": resource.MustParse("100m")},
			),
			wantRes: mockCommonResource(
				common.ResourceList{},
				common.ResourceList{"cpu": "100m"},
			),
		},
	}

	for k, item := range testCases {
		got := TransformRequirementsToResources(item.required)
		if !reflect.DeepEqual(got, item.wantRes) {
			t.Errorf("%s check failure,want:%v,got:%v", k, item.wantRes, got)
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

func mockCommonResource(req, limit common.ResourceList) common.Resources {
	res := common.Resources{}
	if len(req) > 0 {
		res.Requests = req
	}
	if len(limit) > 0 {
		res.Limits = limit
	}
	return res
}
