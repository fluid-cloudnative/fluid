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

package engine

import (
	"fmt"
)

func (e *CacheEngine) getPersistentVolumeName() string {
	return fmt.Sprintf("%s-%s", e.namespace, e.name)
}
func (e *CacheEngine) CreateVolume() (err error) {
	if err = e.createFusePersistentVolume(); err != nil {
		return err
	}

	if err = e.createFusePersistentVolumeClaim(); err != nil {
		return err
	}
	return nil
}

func (e *CacheEngine) DeleteVolume() (err error) {
	if err = e.deleteFusePersistentVolumeClaim(); err != nil {
		return err
	}

	if err = e.deleteFusePersistentVolume(); err != nil {
		return err
	}

	return nil
}

func (e *CacheEngine) createFusePersistentVolume() error {
	// TODO(cache runtime): Implement, refer to JuiceFS Implementation

	return nil
}

func (e *CacheEngine) createFusePersistentVolumeClaim() error {
	// TODO(cache runtime): Implement, refer to JuiceFS Implementation

	return nil
}

func (e *CacheEngine) deleteFusePersistentVolume() error {
	// TODO(cache runtime): Implement, refer to JuiceFS Implementation
	return nil
}

func (e *CacheEngine) deleteFusePersistentVolumeClaim() error {
	// TODO(cache runtime): Implement, refer to JuiceFS Implementation
	return nil
}
