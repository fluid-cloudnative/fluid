package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"testing"
)

func TestTransformAPIGateway(t *testing.T) {
	var engine = &AlluxioEngine{}
	var tests = []struct {
		runtime *datav1alpha1.AlluxioRuntime
		value   *Alluxio
	}{
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: true,
					},
				},
			},
			value: &Alluxio{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
						Enabled: false,
					},
				},
			},
			value: &Alluxio{
				APIGateway: APIGateway{
					Enabled: true,
				},
			},
		},
		{
			runtime: nil,
			value: &Alluxio{
				APIGateway: APIGateway{
					Enabled: false,
				},
			},
		},
		{
			runtime: &datav1alpha1.AlluxioRuntime{
				Spec: datav1alpha1.AlluxioRuntimeSpec{
					APIGateway: datav1alpha1.AlluxioCompTemplateSpec{
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
