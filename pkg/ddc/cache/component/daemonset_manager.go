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

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
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

func (s *DaemonSetManager) ConstructComponentStatus(ctx context.Context, component *common.CacheRuntimeComponentValue) (datav1alpha1.RuntimeComponentStatus, error) {
	logger := log.FromContext(ctx)
	logger.Info("start to ConstructComponentStatus")

	ds := &appsv1.DaemonSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, ds)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get component: %s/%s", component.Namespace, component.Name))
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
