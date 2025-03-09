package engine

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func (e *CacheEngine) transformNodeSelector(runtimeComponentCommonSpec *datav1alpha1.CacheRuntimeComponentCommonSpec) map[string]string {
	properties := map[string]string{}
	if runtimeComponentCommonSpec.NodeSelector != nil {
		properties = runtimeComponentCommonSpec.NodeSelector
	}
	return properties
}
