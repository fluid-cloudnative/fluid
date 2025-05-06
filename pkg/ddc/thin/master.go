/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (t ThinEngine) CheckMasterReady() (bool, error) {
	if _, err := kubeclient.GetDaemonset(t.Client, t.getFuseName(), t.namespace); err != nil {
		return false, err
	}

	return true, nil
}

func (t ThinEngine) ShouldSetupMaster() (should bool, err error) {
	runtime, err := t.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.FusePhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}
	return
}

func (t ThinEngine) SetupMaster() (err error) {
	fuseName := t.getFuseName()
	// 1. Setup
	_, err = kubeclient.GetDaemonset(t.Client, fuseName, t.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		//3.1. Is not found error
		t.Log.Info("SetupMaster", "fuse", fuseName)
		return t.setupMasterInternal()
	} else if err != nil {
		//3.2. Other errors
		return
	}

	// 2.The fuse has been set up
	t.Log.V(1).Info("The fuse has been set.")
	return
}
