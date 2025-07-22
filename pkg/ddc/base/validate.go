package base

import (
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
)

func ValidateRuntimeInfo(runtimeInfo RuntimeInfoInterface) (err error) {
	if len(runtimeInfo.GetOwnerDatasetUID()) == 0 {
		return fluiderrs.NewTemporaryValidationFailed("OwnerDatasetUID is not set in runtime info, this is usually a temporary state, retrying")
	}

	if runtimeInfo.IsExclusive() == nil {
		return fluiderrs.NewTemporaryValidationFailed("exclusive mode is not set in runtime info, this is usually a temporary state, retrying")
	}

	return nil
}
