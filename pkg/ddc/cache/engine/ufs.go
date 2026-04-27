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
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
)

func (e *CacheEngine) PrepareUFS(mountUfs *datav1alpha1.ExecutionCommonEntry, value *common.CacheRuntimeValue) error {
	// execute mount command in master pod
	if mountUfs == nil {
		return nil
	}
	podName, containerName, err := e.getMasterPodInfo(value)
	if err != nil {
		return err
	}

	fileUtils := newCacheFileUtils(podName, containerName, e.namespace, e.Log)

	// at least 20 seconds
	timeoutSeconds := max(mountUfs.TimeoutSeconds, 20)

	err = fileUtils.Mount(mountUfs.Command, time.Duration(timeoutSeconds)*time.Second)
	if err != nil {
		return err
	}

	return nil
}
