/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (e *CacheEngine) SetupClientComponent(clientValue *common.CacheRuntimeComponentValue) (bool, error) {
	shouldSetupClient, err := e.ShouldSetupClient()
	if err != nil {
		return false, err
	}
	if shouldSetupClient {
		if err = e.SetupClientInternal(clientValue); err != nil {
			e.Log.Error(err, "failed to setup client")
			return false, err
		}
	}

	return true, nil
}

func (e *CacheEngine) ShouldSetupClient() (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	if runtime.Status.Client.Phase == datav1alpha1.RuntimePhaseNone {
		return true, nil
	}

	return false, nil
}

func (e *CacheEngine) SetupClientInternal(clientValue *common.CacheRuntimeComponentValue) error {
	return newNotImplementError("SetupClientInternal")
}
