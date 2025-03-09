package engine

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (e *CacheEngine) transformHeadlessServiceValue(value *common.CacheRuntimeValue) *common.CacheRuntimeComponentServiceConfig {
	return &common.CacheRuntimeComponentServiceConfig{
		Name: e.getServiceName(value.RuntimeIdentity.Name, "master"),
	}
}
