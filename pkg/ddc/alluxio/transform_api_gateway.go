package alluxio

import (
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// transformAPIGateway transforms the given value
func (e *AlluxioEngine) transformAPIGateway(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {

	if runtime.Spec.APIGateway.Enabled {
		setDefaultProperties(runtime, value, "alluxio.proxy.web.port", strconv.Itoa(value.APIGateway.Ports.Rest))
	}

	value.APIGateway.Enabled = runtime.Spec.APIGateway.Enabled
	return nil
}
