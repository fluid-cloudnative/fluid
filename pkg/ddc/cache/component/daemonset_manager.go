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

type DaemonSetManager struct {
	client client.Client
}

func newDaemonSetManager(client client.Client) *DaemonSetManager {
	return &DaemonSetManager{client: client}
}

func (s *DaemonSetManager) Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if err := s.reconcileDaemonSet(ctx, component); err != nil {
		return err
	}

	return reconcileService(ctx, s.client, component)
}

func (s *DaemonSetManager) GetNodeAffinity(identity *common.ComponentIdentity) (*corev1.NodeAffinity, error) {
	ds, err := kubeclient.GetDaemonset(s.client, identity.Name, identity.Namespace)
	if err != nil {
		return nil, err
	}

	affinity := kubeclient.MergeNodeSelectorAndNodeAffinity(ds.Spec.Template.Spec.NodeSelector, ds.Spec.Template.Spec.Affinity)
	return affinity, nil
}

func (s *DaemonSetManager) GetPodSpec(ctx context.Context, identity *common.ComponentIdentity) (*corev1.PodSpec, error) {
	ds, err := kubeclient.GetDaemonset(s.client, identity.Name, identity.Namespace)
	if err != nil {
		return nil, err
	}

	return ds.Spec.Template.Spec.DeepCopy(), nil
}

func (s *DaemonSetManager) reconcileDaemonSet(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	logger := log.FromContext(ctx)
	logger.Info("start to reconciling ds workload")

	ds := &appsv1.DaemonSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, ds)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// return if already created
	if err == nil {
		return nil
	}
	// create the daemonset
	ds = s.constructDaemonSet(component)
	err = s.client.Create(ctx, ds)
	if err != nil {
		return err
	}
	logger.Info("create ds workload succeed")
	return nil
}
func (s *DaemonSetManager) constructDaemonSet(component *common.CacheRuntimeComponentValue) *appsv1.DaemonSet {
	matchLabels := getCommonLabelsFromComponent(component)

	podTemplateSpec := component.PodTemplateSpec
	podTemplateSpec.Labels = utils.UnionMapsWithOverride(podTemplateSpec.Labels, matchLabels)

	trueVar := true
	ds := &appsv1.DaemonSet{
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
		Spec: appsv1.DaemonSetSpec{
			Template: podTemplateSpec,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		},
	}
	return ds
}

func (s *DaemonSetManager) ConstructComponentStatus(ctx context.Context, identity *common.ComponentIdentity) (datav1alpha1.RuntimeComponentStatus, error) {
	logger := log.FromContext(ctx)
	logger.Info("start to ConstructComponentStatus")

	ds := &appsv1.DaemonSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: identity.Name, Namespace: identity.Namespace}, ds)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get component: %s/%s", identity.Namespace, identity.Name))
		return datav1alpha1.RuntimeComponentStatus{}, err
	}

	desiredReplicas := ds.Status.DesiredNumberScheduled
	readyReplicas := ds.Status.NumberReady

	runtimePhase := datav1alpha1.RuntimePhaseNotReady
	if desiredReplicas == readyReplicas {
		runtimePhase = datav1alpha1.RuntimePhaseReady
	}

	return datav1alpha1.RuntimeComponentStatus{
		Phase:               runtimePhase,
		DesiredReplicas:     desiredReplicas,
		CurrentReplicas:     ds.Status.CurrentNumberScheduled,
		AvailableReplicas:   ds.Status.NumberAvailable,
		UnavailableReplicas: ds.Status.NumberUnavailable,
		ReadyReplicas:       readyReplicas,
	}, nil
}

// SyncComponentSpec synchronizes component specification changes to the DaemonSet.
// This supports in-place update for compatible fields (image, resources) without deleting the DaemonSet;
// Kubernetes\' DaemonSet controller handles rolling out the updated pod template to each node\'s pod automatically.
// Note: Replicas is not applicable to DaemonSet (replica count is determined by node count) and is ignored if set.
func (s *DaemonSetManager) SyncComponentSpec(ctx context.Context, identity *common.ComponentIdentity, newSpec ComponentSpec) error {
	logger := log.FromContext(ctx)
	logger.Info("start syncing component spec", "component", identity.Name)

	ds := &appsv1.DaemonSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: identity.Name, Namespace: identity.Namespace}, ds)
	if err != nil {
		logger.Error(err, "failed to get daemonset")
		return err
	}

	if len(ds.Spec.Template.Spec.Containers) == 0 {
		return fmt.Errorf("no containers found in daemonset %s/%s", identity.Namespace, identity.Name)
	}

	dsToUpdate := ds.DeepCopy()
	needsUpdate := false

	if s.updateImage(dsToUpdate, newSpec.Version, logger) {
		needsUpdate = true
	}

	if s.updateResources(dsToUpdate, newSpec.Resources, logger) {
		needsUpdate = true
	}

	if !needsUpdate {
		logger.Info("no spec changes detected, skip update")
		return nil
	}

	patch := client.MergeFrom(ds)
	err = s.client.Patch(ctx, dsToUpdate, patch)
	if err != nil {
		logger.Error(err, "failed to patch daemonset")
		return err
	}

	logger.Info("successfully patched daemonset with new spec")
	return nil
}

// updateImage updates container image if changed. Returns true if update is needed.
func (s *DaemonSetManager) updateImage(ds *appsv1.DaemonSet, version datav1alpha1.VersionSpec, logger logr.Logger) bool {
	if len(ds.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	container := &ds.Spec.Template.Spec.Containers[0]
	currentImage := container.Image

	if version.Image == "" || version.ImageTag == "" {
		return false
	}

	newImage := version.Image + ":" + version.ImageTag

	if currentImage != newImage {
		logger.Info("image changed, will update", "old", currentImage, "new", newImage)
		container.Image = newImage
		return true
	}

	return false
}

// updateResources updates container resources if changed. Returns true if update is needed.
func (s *DaemonSetManager) updateResources(ds *appsv1.DaemonSet, resources corev1.ResourceRequirements, logger logr.Logger) bool {
	if len(ds.Spec.Template.Spec.Containers) == 0 {
		return false
	}

	container := &ds.Spec.Template.Spec.Containers[0]

	if !reflect.DeepEqual(container.Resources, resources) {
		logger.Info("resources changed, will update")
		container.Resources = *resources.DeepCopy()
		return true
	}

	return false
}
