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

type StatefulSetManager struct {
	client client.Client
}

func newStatefulSetManager(client client.Client) *StatefulSetManager {
	return &StatefulSetManager{client: client}
}

func (s *StatefulSetManager) Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if err := s.reconcileStatefulSet(ctx, component); err != nil {
		return err
	}

	return reconcileService(ctx, s.client, component)
}

func (s *StatefulSetManager) reconcileStatefulSet(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	logger := log.FromContext(ctx)
	logger.Info("start to reconciling sts workload")

	sts := &appsv1.StatefulSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, sts)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// return if already created
	if err == nil {
		return nil
	}
	// create the stateful set
	sts = s.constructStatefulSet(component)
	err = s.client.Create(ctx, sts)
	if err != nil {
		return err
	}
	logger.Info("create sts workload succeed")
	return nil
}
func (s *StatefulSetManager) constructStatefulSet(component *common.CacheRuntimeComponentValue) *appsv1.StatefulSet {
	matchLabels := getCommonLabelsFromComponent(component)

	podTemplateSpec := component.PodTemplateSpec
	podTemplateSpec.Labels = utils.UnionMapsWithOverride(podTemplateSpec.Labels, matchLabels)

	trueVar := true
	sts := &appsv1.StatefulSet{
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
		Spec: appsv1.StatefulSetSpec{
			Replicas:            &component.Replicas,
			Template:            podTemplateSpec,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		},
	}

	// Set ServiceName if service is configured
	if component.Service != nil {
		sts.Spec.ServiceName = component.Service.Name
	}

	return sts
}

func (s *StatefulSetManager) ConstructComponentStatus(ctx context.Context, component *common.CacheRuntimeComponentValue) (datav1alpha1.RuntimeComponentStatus, error) {
	logger := log.FromContext(ctx)
	logger.Info("start to ConstructComponentStatus")

	sts := &appsv1.StatefulSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, sts)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get component: %s/%s", component.Namespace, component.Name))
		return datav1alpha1.RuntimeComponentStatus{}, err
	}

	desiredReplicas := *sts.Spec.Replicas
	readyReplicas := sts.Status.ReadyReplicas

	runtimePhase := datav1alpha1.RuntimePhaseNotReady
	if desiredReplicas == readyReplicas {
		runtimePhase = datav1alpha1.RuntimePhaseReady
	}

	// AvailableReplicas can be greater than CurrentReplicas (Kubernetes API allows this)
	unavailableReplicas := sts.Status.CurrentReplicas - sts.Status.AvailableReplicas
	if unavailableReplicas < 0 {
		unavailableReplicas = 0
	}

	return datav1alpha1.RuntimeComponentStatus{
		Phase:               runtimePhase,
		DesiredReplicas:     desiredReplicas,
		CurrentReplicas:     sts.Status.CurrentReplicas,
		AvailableReplicas:   sts.Status.AvailableReplicas,
		UnavailableReplicas: unavailableReplicas,
		ReadyReplicas:       readyReplicas,
	}, nil
}
