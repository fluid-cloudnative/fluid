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

func (e *CacheEngine) SetupMasterComponent(masterValue *common.CacheRuntimeComponentValue) (bool, error) {
	shouldSetupMaster, err := e.shouldSetupMaster()
	if err != nil {
		return false, err
	}
	if shouldSetupMaster {
		if err = e.setupMasterInternal(masterValue); err != nil {
			e.Log.Error(err, "failed to setup master")
			return false, err
		}
	}

	return true, nil
}

func (e *CacheEngine) shouldSetupMaster() (bool, error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return false, err
	}

	if runtime.Status.Master.Phase == datav1alpha1.RuntimePhaseNone {
		return true, nil
	}

	return false, nil
}

func (e *CacheEngine) setupMasterInternal(masterValue *common.CacheRuntimeComponentValue) error {
	return newNotImplementError("setupMasterInternal")
}
