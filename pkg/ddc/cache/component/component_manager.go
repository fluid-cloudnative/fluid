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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ComponentManager interface {
	Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error
	ConstructComponentStatus(todo context.Context, identity *common.ComponentIdentity) (datav1alpha1.RuntimeComponentStatus, error)
	GetNodeAffinity(identity *common.ComponentIdentity) (*corev1.NodeAffinity, error)
	// SyncComponentSpec synchronizes component specification changes to the workload
	SyncComponentSpec(ctx context.Context, identity *common.ComponentIdentity, newSpec ComponentSpec) error
}

// ComponentSpec represents the specification that can be synchronized to a component
// This structure groups all fields that support in-place update
type ComponentSpec struct {
	// Replicas is the desired number of replicas (optional, nil means no change)
	Replicas *int32
	// Version contains image and pull policy information
	Version datav1alpha1.VersionSpec
	// Resources contains CPU and memory resource requirements
	Resources corev1.ResourceRequirements
}

func NewComponentHelper(componentType common.ComponentType, client client.Client) ComponentManager {
	// Master and Worker use AdvancedStatefulSet, Client uses DaemonSet
	switch componentType {
	case common.ComponentTypeMaster, common.ComponentTypeWorker:
		return newAdvancedStatefulSetManager(client)
	case common.ComponentTypeClient:
		return newDaemonSetManager(client)
	default:
		// Default to AdvancedStatefulSetManager for unknown types
		return newAdvancedStatefulSetManager(client)
	}
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
