package goosefs

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func TestTransformAPIGateway(t *testing.T) {
	var engine = &GooseFSEngine{}
	var tests = []struct {
		runtime *datav1alpha1.GooseFSRuntime
		value   *GooseFS
	}{
		{
			runtime: &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			value: &GooseFS{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
						Enabled: false,
					},
				},
			},
			value: &GooseFS{
				APIGateway: APIGateway{
					Enabled: true,
				},
			},
		},
		{
			runtime: nil,
			value: &GooseFS{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.GooseFSRuntime{
				Spec: datav1alpha1.GooseFSRuntimeSpec{
					APIGateway: datav1alpha1.GooseFSCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			value: nil,
		},
	}
	for _, test := range tests {
		err := engine.transformAPIGateway(test.runtime, test.value)
		if test.runtime == nil || test.value == nil {
			if err == nil {
				t.Errorf("should return err if it's possible to lead to nil pointer")
			}
		} else if test.runtime.Spec.APIGateway.Enabled != test.value.APIGateway.Enabled {
			t.Errorf("testcase cannot paas beacuse of wrong result,%t != %t",
				test.runtime.Spec.APIGateway.Enabled,
				test.value.APIGateway.Enabled)
		}
	}

}
