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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t ThinEngine) CheckWorkersReady() (ready bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (t ThinEngine) ShouldSetupWorkers() (should bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (t ThinEngine) SetupWorkers() (err error) {
	//TODO implement me
	panic("implement me")
}

// getWorkerSelectors gets the selector of the worker
func (t *ThinEngine) getWorkerSelectors() string {
	labels := map[string]string{
		"release":   t.name,
		PodRoleType: WorkerPodRole,
		"app":       common.ThinRuntime,
	}
	labelSelector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	selectorValue := ""
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		t.Log.Error(err, "Failed to parse the labelSelector of the runtime", "labels", labels)
	} else {
		selectorValue = selector.String()
	}
	return selectorValue
}
