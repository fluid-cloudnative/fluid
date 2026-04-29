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

package component

import (
	"context"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ComponentManager interface {
	Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error
	ConstructComponentStatus(todo context.Context, value *common.CacheRuntimeComponentValue) (v1alpha1.RuntimeComponentStatus, error)
}

func NewComponentHelper(workloadType metav1.TypeMeta, client client.Client) ComponentManager {
	if workloadType.APIVersion == "apps/v1" {
		if workloadType.Kind == "StatefulSet" {
			return newStatefulSetManager(client)
		} else if workloadType.Kind == "DaemonSet" {
			return newDaemonSetManager(client)
		}
	}

	return newStatefulSetManager(client)
}

// getCommonLabelsFromComponent returns the common labels for component used for stateful
func getCommonLabelsFromComponent(component *common.CacheRuntimeComponentValue) map[string]string {
	// These labels are used as sts.spec.selector which cannot be updated.
	// If changed, may cause all exist runtime failed.
	return map[string]string{
		common.LabelCacheRuntimeName: component.Owner.Name,
		// format: runtimeName-componentType
		common.LabelCacheRuntimeComponentName: component.Name,
	}
}
