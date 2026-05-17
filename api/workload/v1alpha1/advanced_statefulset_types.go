/*
Copyright 2026 The Fluid Authors.
Copyright 2024 The Kruise Authors.

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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strings"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ============================================================
// Well-known labels and annotations
// ============================================================

const (
	ControllerRevisionHashLabelKey = "workload.fluid.io/controller-revision-hash"
	ReservedPodLabelKey            = "workload.fluid.io/united-deployment-reserved-pod"
	SubSetNameLabelKey             = "workload.fluid.io/subset-name"
	SpecifiedDeleteKey             = "workload.fluid.io/specified-delete"
	AnnotationSubsetPatchKey       = "workload.fluid.io/subset-patch"

	// ContainerLaunchBarrierEnvName is the env name used to indicate container launch priority barrier.
	ContainerLaunchBarrierEnvName = "KRUISE_CONTAINER_LAUNCH_BARRIER"
)

// ============================================================
// InPlace Update types
// ============================================================

const (
	InPlaceUpdateReady v1.PodConditionType = "InPlaceUpdateReady"

	InPlaceUpdateStateKey    string = "workload.fluid.io/inplace-update-state"
	InPlaceUpdateStateKeyOld string = "inplace-update-state"

	InPlaceUpdateGraceKey    string = "workload.fluid.io/inplace-update-grace"
	InPlaceUpdateGraceKeyOld string = "inplace-update-grace"

	RuntimeContainerMetaKey = "workload.fluid.io/runtime-containers-meta"
)

// InPlaceUpdateState is the state of an in-place update, recorded in pod annotations.
type InPlaceUpdateState struct {
	Revision                 string                                  `json:"revision"`
	UpdateTimestamp          metav1.Time                             `json:"updateTimestamp"`
	LastContainerStatuses    map[string]InPlaceUpdateContainerStatus `json:"lastContainerStatuses"`
	UpdateEnvFromMetadata    bool                                    `json:"updateEnvFromMetadata,omitempty"`
	UpdateResources          bool                                    `json:"updateResources,omitempty"`
	UpdateImages             bool                                    `json:"updateImages,omitempty"`
	NextContainerImages      map[string]string                       `json:"nextContainerImages,omitempty"`
	NextContainerRefMetadata map[string]metav1.ObjectMeta            `json:"nextContainerRefMetadata,omitempty"`
	NextContainerResources   map[string]v1.ResourceRequirements      `json:"nextContainerResources,omitempty"`
	PreCheckBeforeNext       *InPlaceUpdatePreCheckBeforeNext        `json:"preCheckBeforeNext,omitempty"`
	ContainerBatchesRecord   []InPlaceUpdateContainerBatch           `json:"containerBatchesRecord,omitempty"`
}

// InPlaceUpdatePreCheckBeforeNext specifies containers that must be ready before the next batch update.
type InPlaceUpdatePreCheckBeforeNext struct {
	ContainersRequiredReady []string `json:"containersRequiredReady,omitempty"`
}

// InPlaceUpdateContainerBatch records a batch of containers updated at a given time.
type InPlaceUpdateContainerBatch struct {
	Timestamp  metav1.Time `json:"timestamp"`
	Containers []string    `json:"containers"`
}

// InPlaceUpdateContainerStatus records the image ID of a container before in-place update.
type InPlaceUpdateContainerStatus struct {
	ImageID string `json:"imageID,omitempty"`
}

// InPlaceUpdateStrategy defines the strategy for in-place updates.
type InPlaceUpdateStrategy struct {
	GracePeriodSeconds int32 `json:"gracePeriodSeconds,omitempty"`
}

// GetInPlaceUpdateState returns the in-place update state annotation value.
func GetInPlaceUpdateState(obj metav1.Object) (string, bool) {
	if v, ok := obj.GetAnnotations()[InPlaceUpdateStateKey]; ok {
		return v, ok
	}
	v, ok := obj.GetAnnotations()[InPlaceUpdateStateKeyOld]
	return v, ok
}

// GetInPlaceUpdateGrace returns the in-place update grace annotation value.
func GetInPlaceUpdateGrace(obj metav1.Object) (string, bool) {
	if v, ok := obj.GetAnnotations()[InPlaceUpdateGraceKey]; ok {
		return v, ok
	}
	v, ok := obj.GetAnnotations()[InPlaceUpdateGraceKeyOld]
	return v, ok
}

// RemoveInPlaceUpdateGrace removes the in-place update grace annotations from the object.
func RemoveInPlaceUpdateGrace(obj metav1.Object) {
	delete(obj.GetAnnotations(), InPlaceUpdateGraceKey)
	delete(obj.GetAnnotations(), InPlaceUpdateGraceKeyOld)
}

// RuntimeContainerMetaSet holds metadata for all containers at runtime.
type RuntimeContainerMetaSet struct {
	Containers []RuntimeContainerMeta `json:"containers"`
}

// RuntimeContainerMeta holds runtime metadata for a single container.
type RuntimeContainerMeta struct {
	Name         string                 `json:"name"`
	ContainerID  string                 `json:"containerID"`
	RestartCount int32                  `json:"restartCount"`
	Hashes       RuntimeContainerHashes `json:"hashes"`
}

// RuntimeContainerHashes holds hash values used to detect container spec changes.
type RuntimeContainerHashes struct {
	PlainHash                    uint64 `json:"plainHash"`
	ExtractedEnvFromMetadataHash uint64 `json:"extractedEnvFromMetadataHash,omitempty"`
}

// GetRuntimeContainerMetaSet parses the runtime container meta annotation.
func GetRuntimeContainerMetaSet(obj metav1.Object) (*RuntimeContainerMetaSet, error) {
	str, ok := obj.GetAnnotations()[RuntimeContainerMetaKey]
	if !ok {
		return nil, nil
	}
	s := RuntimeContainerMetaSet{}
	if err := json.Unmarshal([]byte(str), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ============================================================
// Lifecycle types
// ============================================================

const (
	LifecycleStateKey                                = "lifecycle.workload.fluid.io/state"
	LifecycleTimestampKey                            = "lifecycle.workload.fluid.io/timestamp"
	LifecycleStatePreparingNormal LifecycleStateType = "PreparingNormal"
	LifecycleStateNormal          LifecycleStateType = "Normal"
	LifecycleStatePreparingUpdate LifecycleStateType = "PreparingUpdate"
	LifecycleStateUpdating        LifecycleStateType = "Updating"
	LifecycleStateUpdated         LifecycleStateType = "Updated"
	LifecycleStatePreparingDelete LifecycleStateType = "PreparingDelete"
)

// LifecycleStateType represents the lifecycle state of a pod.
type LifecycleStateType string

// Lifecycle defines hooks for pod lifecycle events.
type Lifecycle struct {
	// PreDelete is the hook before a pod is deleted.
	// +optional
	PreDelete *LifecycleHook `json:"preDelete,omitempty"`
	// InPlaceUpdate is the hook during in-place updates.
	// +optional
	InPlaceUpdate *LifecycleHook `json:"inPlaceUpdate,omitempty"`
	// PreNormal is the hook before a pod transitions to normal state.
	// +optional
	PreNormal *LifecycleHook `json:"preNormal,omitempty"`
}

// LifecycleHook defines label/finalizer handlers for a lifecycle event.
type LifecycleHook struct {
	// LabelsHandler holds label key-value pairs to set on the pod during this lifecycle hook.
	// +optional
	LabelsHandler map[string]string `json:"labelsHandler,omitempty"`
	// FinalizersHandler holds finalizers to add on the pod during this lifecycle hook.
	// +optional
	FinalizersHandler []string `json:"finalizersHandler,omitempty"`
	// MarkPodNotReady marks the pod as not ready during this lifecycle hook.
	// +optional
	MarkPodNotReady bool `json:"markPodNotReady,omitempty"`
}

// ============================================================
// Pod Readiness Gate
// ============================================================

const (
	// KruisePodReadyConditionType is the pod condition type for kruise readiness gate.
	KruisePodReadyConditionType v1.PodConditionType = "KruisePodReady"
	// InPlaceUpdateStrategyKruisePodReadyConditionType is an alias for KruisePodReadyConditionType.
	InPlaceUpdateStrategyKruisePodReadyConditionType = KruisePodReadyConditionType
)

// ============================================================
// Pod Unavailable Label
// ============================================================

const (
	PubUnavailablePodLabelPrefix = "unavailable-pod.fluid.io/"
)

// HasUnavailableLabel returns true if the given labels contain an unavailable pod label.
func HasUnavailableLabel(labels map[string]string) bool {
	if len(labels) == 0 {
		return false
	}
	for key := range labels {
		if strings.HasPrefix(key, PubUnavailablePodLabelPrefix) {
			return true
		}
	}
	return false
}

// ============================================================
// Update Priority types
// ============================================================

// UpdatePriorityStrategy defines how to select pods to update in priority order.
type UpdatePriorityStrategy struct {
	// OrderPriority specifies the ordered key for pod update priority.
	// +optional
	OrderPriority []UpdatePriorityOrderTerm `json:"orderPriority,omitempty"`
	// WeightPriority specifies the weight-based priority for pod update.
	// +optional
	WeightPriority []UpdatePriorityWeightTerm `json:"weightPriority,omitempty"`
}

// UpdatePriorityOrderTerm defines a single ordered key for pod update priority.
type UpdatePriorityOrderTerm struct {
	OrderedKey string `json:"orderedKey"`
}

// UpdatePriorityWeightTerm defines a weight-based priority for pods matching a selector.
type UpdatePriorityWeightTerm struct {
	Weight        int32                `json:"weight"`
	MatchSelector metav1.LabelSelector `json:"matchSelector"`
}

// FieldsValidation validates the UpdatePriorityStrategy fields.
func (strategy *UpdatePriorityStrategy) FieldsValidation() error {
	if strategy == nil {
		return nil
	}
	if len(strategy.WeightPriority) > 0 && len(strategy.OrderPriority) > 0 {
		return fmt.Errorf("only one of weightPriority and orderPriority can be used")
	}
	for _, w := range strategy.WeightPriority {
		if w.Weight < 0 || w.Weight > 100 {
			return fmt.Errorf("weight must be valid number in the range 1-100")
		}
		if w.MatchSelector.Size() == 0 {
			return fmt.Errorf("selector can not be empty")
		}
		if _, err := metav1.LabelSelectorAsSelector(&w.MatchSelector); err != nil {
			return fmt.Errorf("invalid selector %v", err)
		}
	}
	for _, o := range strategy.OrderPriority {
		if len(o.OrderedKey) == 0 {
			return fmt.Errorf("order key can not be empty")
		}
	}
	return nil
}

// ============================================================
// Scatter Strategy
// ============================================================

// UpdateScatterStrategy defines how to scatter pod updates across different label values.
type UpdateScatterStrategy []UpdateScatterTerm

// UpdateScatterTerm defines a single label key-value pair for scatter strategy.
type UpdateScatterTerm struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FieldsValidation validates the UpdateScatterStrategy fields.
func (strategy UpdateScatterStrategy) FieldsValidation() error {
	if len(strategy) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(strategy))
	for _, term := range strategy {
		if term.Key == "" {
			return fmt.Errorf("key should not be empty")
		}
		id := term.Key + ":" + term.Value
		if _, ok := m[id]; !ok {
			m[id] = struct{}{}
		} else {
			return fmt.Errorf("duplicated key=%v value=%v", term.Key, term.Value)
		}
	}
	return nil
}

// ============================================================
// AdvancedStatefulSet types
// ============================================================

const (
	// MaxMinReadySeconds is the maximum value for minReadySeconds in rolling update.
	MaxMinReadySeconds = 300

	// FailedCreatePod is the condition type for failed pod creation.
	FailedCreatePod apps.StatefulSetConditionType = "FailedCreatePod"
	// FailedUpdatePod is the condition type for failed pod update.
	FailedUpdatePod apps.StatefulSetConditionType = "FailedUpdatePod"
)

// VolumeClaimUpdateStrategyType defines the strategy for updating volume claims.
// +enum
type VolumeClaimUpdateStrategyType string

const (
	// OnPodRollingUpdateVolumeClaimUpdateStrategyType updates volume claims during pod rolling update.
	OnPodRollingUpdateVolumeClaimUpdateStrategyType VolumeClaimUpdateStrategyType = "OnPodRollingUpdate"
	// OnPVCDeleteVolumeClaimUpdateStrategyType updates volume claims when PVC is deleted.
	OnPVCDeleteVolumeClaimUpdateStrategyType VolumeClaimUpdateStrategyType = "OnDelete"
)

// VolumeClaimStatus records the compatibility status of a volume claim template.
type VolumeClaimStatus struct {
	VolumeClaimName         string `json:"volumeClaimName"`
	CompatibleReplicas      int32  `json:"compatibleReplicas"`
	CompatibleReadyReplicas int32  `json:"compatibleReadyReplicas"`
}

// VolumeClaimUpdateStrategy defines the strategy for updating volume claim templates.
type VolumeClaimUpdateStrategy struct {
	Type VolumeClaimUpdateStrategyType `json:"type,omitempty"`
}

// PodUpdateStrategyType defines the strategy for updating pods in-place.
// +enum
type PodUpdateStrategyType string

const (
	// RecreatePodUpdateStrategyType deletes and recreates the pod on update.
	RecreatePodUpdateStrategyType PodUpdateStrategyType = "ReCreate"
	// InPlaceIfPossiblePodUpdateStrategyType updates pods in-place if possible, falls back to recreate.
	InPlaceIfPossiblePodUpdateStrategyType PodUpdateStrategyType = "InPlaceIfPossible"
	// InPlaceOnlyPodUpdateStrategyType requires in-place update; fails if not possible.
	InPlaceOnlyPodUpdateStrategyType PodUpdateStrategyType = "InPlaceOnly"
)

// PersistentVolumeClaimRetentionPolicyType defines the retention policy for PVCs.
// +enum
type PersistentVolumeClaimRetentionPolicyType string

const (
	// RetainPersistentVolumeClaimRetentionPolicyType retains the PVC after pod deletion/scale.
	RetainPersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Retain"
	// DeletePersistentVolumeClaimRetentionPolicyType deletes the PVC after pod deletion/scale.
	DeletePersistentVolumeClaimRetentionPolicyType PersistentVolumeClaimRetentionPolicyType = "Delete"
)

// StatefulSetPersistentVolumeClaimRetentionPolicy describes the policy for PVC retention.
type StatefulSetPersistentVolumeClaimRetentionPolicy struct {
	// WhenDeleted specifies what happens to PVCs when the AdvancedStatefulSet is deleted.
	WhenDeleted PersistentVolumeClaimRetentionPolicyType `json:"whenDeleted,omitempty"`
	// WhenScaled specifies what happens to PVCs when the AdvancedStatefulSet is scaled down.
	WhenScaled PersistentVolumeClaimRetentionPolicyType `json:"whenScaled,omitempty"`
}

// StatefulSetOrdinals defines the start ordinal for pod naming.
type StatefulSetOrdinals struct {
	// Start is the starting ordinal for pod naming.
	// +optional
	Start int32 `json:"start" protobuf:"varint,1,opt,name=start"`
}

// StatefulSetScaleStrategy defines the strategy for scaling AdvancedStatefulSet pods.
type StatefulSetScaleStrategy struct {
	// MaxUnavailable is the maximum number of pods that can be unavailable during scaling.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
}

// UnorderedUpdateStrategy allows updating pods without strict ordering.
type UnorderedUpdateStrategy struct {
	// PriorityStrategy defines the priority order for pod updates.
	// +optional
	PriorityStrategy *UpdatePriorityStrategy `json:"priorityStrategy,omitempty"`
}

// RollingUpdateStatefulSetStrategy defines the rolling update strategy for AdvancedStatefulSet.
type RollingUpdateStatefulSetStrategy struct {
	// Partition indicates the ordinal at which the AdvancedStatefulSet should be partitioned for updates.
	// +optional
	Partition *int32 `json:"partition,omitempty"`
	// MaxUnavailable is the maximum number of pods that can be unavailable during rolling update.
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
	// PodUpdatePolicy indicates the policy for pod updates.
	// +optional
	PodUpdatePolicy PodUpdateStrategyType `json:"podUpdatePolicy,omitempty"`
	// Paused indicates that the AdvancedStatefulSet is paused.
	// +optional
	Paused bool `json:"paused,omitempty"`
	// UnorderedUpdate allows updating pods out of order.
	// +optional
	UnorderedUpdate *UnorderedUpdateStrategy `json:"unorderedUpdate,omitempty"`
	// InPlaceUpdateStrategy defines the in-place update strategy.
	// +optional
	InPlaceUpdateStrategy *InPlaceUpdateStrategy `json:"inPlaceUpdateStrategy,omitempty"`
	// MinReadySeconds is the minimum number of seconds a pod must be ready before being considered available.
	// +optional
	MinReadySeconds *int32 `json:"minReadySeconds,omitempty"`
}

// StatefulSetUpdateStrategy defines how a AdvancedStatefulSet is updated.
type StatefulSetUpdateStrategy struct {
	// Type indicates the type of the StatefulSetUpdateStrategy.
	// +optional
	Type apps.StatefulSetUpdateStrategyType `json:"type,omitempty"`
	// RollingUpdate is used to communicate parameters when Type is RollingUpdateStatefulSetStrategyType.
	// +optional
	RollingUpdate *RollingUpdateStatefulSetStrategy `json:"rollingUpdate,omitempty"`
}

// AdvancedStatefulSetSpec defines the desired state of AdvancedStatefulSet.
type AdvancedStatefulSetSpec struct {
	// Replicas is the desired number of replicas of the given Template.
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Selector is a label query over pods that should match the replica count.
	Selector *metav1.LabelSelector `json:"selector"`
	// Template is the object that describes the pod that will be created.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Template v1.PodTemplateSpec `json:"template"`
	// VolumeClaimTemplates is a list of claims that pods are allowed to reference.
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	VolumeClaimTemplates []v1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty"`
	// VolumeClaimUpdateStrategy defines the strategy for updating volume claim templates.
	// +optional
	VolumeClaimUpdateStrategy VolumeClaimUpdateStrategy `json:"volumeClaimUpdateStrategy,omitempty"`
	// ServiceName is the name of the service that governs this AdvancedStatefulSet.
	// +optional
	ServiceName string `json:"serviceName,omitempty"`
	// PodManagementPolicy controls how pods are created during initial scale up,
	// when replacing pods on nodes, or when scaling down.
	// +optional
	PodManagementPolicy apps.PodManagementPolicyType `json:"podManagementPolicy,omitempty"`
	// UpdateStrategy indicates the StatefulSetUpdateStrategy that will be employed.
	UpdateStrategy StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`
	// RevisionHistoryLimit is the maximum number of revisions that will be maintained
	// in the AdvancedStatefulSet's revision history.
	// +optional
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`
	// ReserveOrdinals is the list of ordinals to skip when assigning pod ordinals.
	ReserveOrdinals []intstr.IntOrString `json:"reserveOrdinals,omitempty"`
	// Lifecycle defines the lifecycle hooks for pods.
	// +optional
	Lifecycle *Lifecycle `json:"lifecycle,omitempty"`
	// ScaleStrategy defines the strategy for scaling AdvancedStatefulSet pods.
	// +optional
	ScaleStrategy *StatefulSetScaleStrategy `json:"scaleStrategy,omitempty"`
	// PersistentVolumeClaimRetentionPolicy describes the policy used for PVCs created from
	// the AdvancedStatefulSet VolumeClaimTemplates.
	// +optional
	PersistentVolumeClaimRetentionPolicy *StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty"`
	// Ordinals controls the numbering of replica indices in a AdvancedStatefulSet.
	// +optional
	Ordinals *StatefulSetOrdinals `json:"ordinals,omitempty"`
}

// AdvancedStatefulSetStatus defines the observed state of AdvancedStatefulSet.
type AdvancedStatefulSetStatus struct {
	// ObservedGeneration is the most recent generation observed for this AdvancedStatefulSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// Replicas is the number of Pods created by the AdvancedStatefulSet controller.
	Replicas int32 `json:"replicas"`
	// ReadyReplicas is the number of Pods created by the AdvancedStatefulSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas"`
	// AvailableReplicas is the number of Pods created by the AdvancedStatefulSet controller that have been Ready
	// for minReadySeconds.
	AvailableReplicas int32 `json:"availableReplicas"`
	// CurrentReplicas is the number of Pods created by the AdvancedStatefulSet controller from the AdvancedStatefulSet version
	// indicated by currentRevision.
	CurrentReplicas int32 `json:"currentReplicas"`
	// UpdatedReplicas is the number of Pods created by the AdvancedStatefulSet controller from the AdvancedStatefulSet version
	// indicated by updateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas"`
	// UpdatedReadyReplicas is the number of updated Pods that have a Ready Condition.
	// +optional
	UpdatedReadyReplicas int32 `json:"updatedReadyReplicas,omitempty"`
	// UpdatedAvailableReplicas is the number of updated Pods that have been Ready for minReadySeconds.
	// +optional
	UpdatedAvailableReplicas int32 `json:"updatedAvailableReplicas,omitempty"`
	// CurrentRevision, if not empty, indicates the version of the AdvancedStatefulSet used to generate Pods in the sequence
	// [0,currentReplicas).
	// +optional
	CurrentRevision string `json:"currentRevision,omitempty"`
	// UpdateRevision, if not empty, indicates the version of the AdvancedStatefulSet used to generate Pods in the sequence
	// [replicas-updatedReplicas,replicas)
	// +optional
	UpdateRevision string `json:"updateRevision,omitempty"`
	// CollisionCount is the count of hash collisions for the AdvancedStatefulSet.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty"`
	// Conditions represent the latest available observations of a AdvancedStatefulSet's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []apps.StatefulSetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
	// LabelSelector is the label selector for pods scaling purposes.
	// +optional
	LabelSelector string `json:"labelSelector,omitempty"`
	// VolumeClaims records the compatibility status of each volume claim template.
	// +optional
	VolumeClaims []VolumeClaimStatus `json:"volumeClaims,omitempty"`
}

// +genclient
// +genclient:method=GetScale,verb=get,subresource=scale,result=k8s.io/api/autoscaling/v1.Scale
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.labelSelector
// +kubebuilder:resource:shortName=asts,categories=fluid
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="DESIRED",type="integer",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="CURRENT",type="integer",JSONPath=".status.replicas"
// +kubebuilder:printcolumn:name="UPDATED",type="integer",JSONPath=".status.updatedReplicas"
// +kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.readyReplicas"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="CONTAINERS",type="string",priority=1,JSONPath=".spec.template.spec.containers[*].name"
// +kubebuilder:printcolumn:name="IMAGES",type="string",priority=1,JSONPath=".spec.template.spec.containers[*].image"

// AdvancedStatefulSet is the Schema for the statefulsets API
type AdvancedStatefulSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AdvancedStatefulSetSpec   `json:"spec,omitempty"`
	Status AdvancedStatefulSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AdvancedStatefulSetList contains a list of AdvancedStatefulSet
type AdvancedStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AdvancedStatefulSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AdvancedStatefulSet{}, &AdvancedStatefulSetList{})
}
