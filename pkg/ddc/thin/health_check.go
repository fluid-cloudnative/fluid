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
	"k8s.io/client-go/util/retry"
)

func (t ThinEngine) CheckRuntimeHealthy() (err error) {
	// Check the healthy of the fuse
	err = t.checkFuseHealthy()
	if err != nil {
		t.Log.Error(err, "checkFuseHealthy failed")
		return
	}

	return
}

// checkFuseHealthy check fuses number changed
func (t *ThinEngine) checkFuseHealthy() error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		runtime, err := t.getRuntime()
		if err != nil {
			t.Log.Error(err, "Failed to get Runtime", "runtimeName", t.name, "runtimeNamespace", t.namespace)
			return
		}
		err = t.Helper.CheckFuseHealthy(t.Recorder, runtime.DeepCopy(), t.getFuseName())
		if err != nil {
			t.Log.Error(err, "Failed to check runtimeFuse healthy")
		}
		return
	})
}
