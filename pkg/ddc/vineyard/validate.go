/*
Copyright 2024 The Fluid Authors.

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

package vineyard

import (
	fluiderrs "github.com/fluid-cloudnative/fluid/pkg/errors"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func (e *VineyardEngine) Validate(ctx cruntime.ReconcileRequestContext) (err error) {
	// XXXEngine.runtimeInfo must have full information about the bound dataset for further reconcilation.
	// getRuntimeInfo() here is a refresh to make sure the information is correctly set
	runtimeInfo, err := e.getRuntimeInfo()
	if err != nil {
		return err
	}

	if len(runtimeInfo.GetOwnerDatasetUID()) == 0 {
		return fluiderrs.NewTemporaryValidationFailed("OwnerDatasetUID is not set in runtime info, this is usually a temporary state, retrying")
	}

	// TODO: impl validation logic for VineyardEngine
	return nil
}
