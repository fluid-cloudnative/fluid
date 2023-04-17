package utils

import (
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
)

// GetExclusiveKey gets exclusive key
func GetExclusiveKey() string {
	return common.FluidExclusiveKey
}

// GetExclusiveValue gets exclusive value
func GetExclusiveValue(namespace, name string) string {
	return fmt.Sprintf("%s_%s", namespace, name)
}
