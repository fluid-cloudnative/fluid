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
	"context"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

// getRuntime gets thin runtime
func (t *ThinEngine) getRuntime() (*datav1alpha1.ThinRuntime, error) {

	key := types.NamespacedName{
		Name:      t.name,
		Namespace: t.namespace,
	}

	var runtime datav1alpha1.ThinRuntime
	if err := t.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (t *ThinEngine) getFuseDaemonsetName() (dsName string) {
	return t.name + "-fuse"
}
func (t *ThinEngine) getWorkerName() (dsName string) {
	return t.name + "-worker"
}
