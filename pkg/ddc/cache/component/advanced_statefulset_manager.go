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
	"fmt"
	"reflect"

	workloadv1alpha1 "github.com/fluid-cloudnative/advanced-statefulset/api/workload/v1alpha1"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type AdvancedStatefulSetManager struct {
	client client.Client
}

func newAdvancedStatefulSetManager(client client.Client) *AdvancedStatefulSetManager {
	return &AdvancedStatefulSetManager{client: client}
}

func (s *AdvancedStatefulSetManager) Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if err := s.reconcileStatefulSet(ctx, component); err != nil {
		return err
	}

	return reconcileService(ctx, s.client, component)
}

func (s *AdvancedStatefulSetManager) GetNodeAffinity(identity *common.ComponentIdentity) (*corev1.NodeAffinity, error) {
	asts := &workloadv1alpha1.AdvancedStatefulSet{}
	err := s.client.Get(context.TODO(), types.NamespacedName{Name: identity.Name, Namespace: identity.Namespace}, asts)
	if err != nil {
		return nil, err
	}

	affinity := kubeclient.MergeNodeSelectorAndNodeAffinity(asts.Spec.Template.Spec.NodeSelector, asts.Spec.Template.Spec.Affinity)
	return affinity, nil
}

func (s *AdvancedStatefulSetManager) reconcileStatefulSet(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	logger := log.FromContext(ctx)
	logger.Info("start to reconciling advanced statefulset workload")

	asts := &workloadv1alpha1.AdvancedStatefulSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, asts)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// if already created, update it
	if err == nil {
		return nil
	}
	// create the advanced stateful set
	asts = s.constructAdvancedStatefulSet(component)
	err = s.client.Create(ctx, asts)
	if err != nil {
		return err
	}
	logger.Info("create advanced statefulset workload succeed")
	return nil
}
func (s *AdvancedStatefulSetManager) constructAdvancedStatefulSet(component *common.CacheRuntimeComponentValue) *workloadv1alpha1.AdvancedStatefulSet {
	matchLabels := getCommonLabelsFromComponent(component)

	podTemplateSpec := component.PodTemplateSpec
	podTemplateSpec.Labels = utils.UnionMapsWithOverride(podTemplateSpec.Labels, matchLabels)

	trueVar := true

	// Configure rolling update strategy with in-place update support
	rollingUpdateStrategy := &workloadv1alpha1.RollingUpdateStatefulSetStrategy{
		PodUpdatePolicy: workloadv1alpha1.InPlaceIfPossiblePodUpdateStrategyType,
	}

	asts := &workloadv1alpha1.AdvancedStatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.Name,
			Namespace: component.Namespace,
			Labels:    matchLabels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         component.Owner.APIVersion,
					Kind:               component.Owner.Kind,
					Name:               component.Owner.Name,
					UID:                types.UID(component.Owner.UID),
					BlockOwnerDeletion: &trueVar,
					Controller:         &trueVar,
				},
			},
		},
		Spec: workloadv1alpha1.AdvancedStatefulSetSpec{
			Replicas:            &component.Replicas,
			Template:            podTemplateSpec,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			UpdateStrategy: workloadv1alpha1.StatefulSetUpdateStrategy{
				Type:          appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: rollingUpdateStrategy,
			},
		},
	}

	// Set ServiceName if service is configured
	if component.Service != nil {
		asts.Spec.ServiceName = component.Service.Name
	}

	return asts
}

func (s *AdvancedStatefulSetManager) ConstructComponentStatus(ctx context.Context, identity *common.ComponentIdentity) (datav1alpha1.RuntimeComponentStatus, error) {
	logger := log.FromContext(ctx)
	logger.Info("start to ConstructComponentStatus")

	asts := &workloadv1alpha1.AdvancedStatefulSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: identity.Name, Namespace: identity.Namespace}, asts)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get component: %s/%s", identity.Namespace, identity.Name))
		return datav1alpha1.RuntimeComponentStatus{}, err
	}

	desiredReplicas := int32(0)
	if asts.Spec.Replicas != nil {
		desiredReplicas = *asts.Spec.Replicas
	}
	readyReplicas := asts.Status.ReadyReplicas

	runtimePhase := datav1alpha1.RuntimePhaseNotReady
	if desiredReplicas == readyReplicas {
		runtimePhase = datav1alpha1.RuntimePhaseReady
	}

	// AvailableReplicas can be greater than CurrentReplicas (Kubernetes API allows this)
	unavailableReplicas := asts.Status.CurrentReplicas - asts.Status.AvailableReplicas
	if unavailableReplicas < 0 {
		unavailableReplicas = 0
	}

	return datav1alpha1.RuntimeComponentStatus{
		Phase:               runtimePhase,
		DesiredReplicas:     desiredReplicas,
		CurrentReplicas:     asts.Status.CurrentReplicas,
		AvailableReplicas:   asts.Status.AvailableReplicas,
		UnavailableReplicas: unavailableReplicas,
		ReadyReplicas:       readyReplicas,
	}, nil
}

// SyncComponentSpec synchronizes component specification changes to the AdvancedStatefulSet
// This supports in-place update for compatible fields (e.g., image, resources, env) without pod recreation
// Supported in-place update fields:
// - Container image and ImagePullPolicy
// - Resource requests and limits (CPU, memory)
// - Environment variables
// - Replicas count
// Note: Fields like volumes, volumeMounts, args, ports will trigger pod recreation
func (s *AdvancedStatefulSetManager) SyncComponentSpec(ctx context.Context, identity *common.ComponentIdentity, newSpec ComponentSpec) error {
	logger := log.FromContext(ctx)
	logger.Info("start syncing component spec", "component", identity.Name)

	// Get current AdvancedStatefulSet
	asts := &workloadv1alpha1.AdvancedStatefulSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: identity.Name, Namespace: identity.Namespace}, asts)
	if err != nil {
		logger.Error(err, "failed to get advanced statefulset")
		return err
	}

	// Check if containers exist
	if len(asts.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("no containers found in advanced statefulset %s/%s", identity.Namespace, identity.Name)
	}

	// Create a copy for patching to avoid modifying the original object
	astsToUpdate := asts.DeepCopy()
	needsUpdate := false

	// 1. Update replicas if specified and changed
	if newSpec.Replicas != nil {
		if s.updateReplicas(astsToUpdate, *newSpec.Replicas, logger) {
			needsUpdate = true
		}
	}

	// 2. Update image if specified
	if s.updateImage(astsToUpdate, newSpec.Version, logger) {
		needsUpdate = true
	}

	// 3. Update resources if specified
	if s.updateResources(astsToUpdate, newSpec.Resources, logger) {
		needsUpdate = true
	}

	// Skip patching if no changes detected
	if !needsUpdate {
		logger.Info("no spec changes detected, skip update")
		return nil
	}

	// Create patch using the original object as base
	patch := client.MergeFrom(asts)

	// Apply patch
	err = s.client.Patch(ctx, astsToUpdate, patch)
	if err != nil {
		logger.Error(err, "failed to patch advanced statefulset")
		return err
	}

	logger.Info("successfully patched advanced statefulset with new spec")
	return nil
}

// updateReplicas updates the replica count if changed
// Returns true if update is needed
func (s *AdvancedStatefulSetManager) updateReplicas(asts *workloadv1alpha1.AdvancedStatefulSet, newReplicas int32, logger logr.Logger) bool {
	var oldReplicas int32 = 0
	if asts.Spec.Replicas != nil {
		oldReplicas = *asts.Spec.Replicas
	}
	if oldReplicas != newReplicas {
		logger.Info("replicas changed, will update", "old", oldReplicas, "new", newReplicas)
		asts.Spec.Replicas = &newReplicas
		return true
	}
	return false
}

// updateImage updates container image if changed
// Returns true if update is needed
func (s *AdvancedStatefulSetManager) updateImage(asts *workloadv1alpha1.AdvancedStatefulSet, version datav1alpha1.VersionSpec, logger logr.Logger) bool {
	if len(asts.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	container := &asts.Spec.Template.Spec.Containers[0]
	currentImage := container.Image

	// Both image and imageTag must be specified
	if version.Image == "" || version.ImageTag == "" {
		return false
	}

	// TODO: 之前的镜像如何决定的。。。是否会触发强制更新（如之前用的 runtimeclass ）
	// Build new image: directly concatenate image and imageTag
	newImage := version.Image + ":" + version.ImageTag

	if currentImage != newImage {
		logger.Info("image changed, will update", "old", currentImage, "new", newImage)
		container.Image = newImage
		return true
	}

	return false
}

// updateResources updates container resources
// Returns true if update is needed
func (s *AdvancedStatefulSetManager) updateResources(asts *workloadv1alpha1.AdvancedStatefulSet, resources corev1.ResourceRequirements, logger logr.Logger) bool {
	if len(asts.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	container := &asts.Spec.Template.Spec.Containers[0]

	// Directly compare and replace, nil is also a valid value
	if !reflect.DeepEqual(container.Resources, resources) {
		logger.Info("resources changed, will update")
		container.Resources = *resources.DeepCopy()
		return true
	}

	return false
}
