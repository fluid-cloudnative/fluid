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
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type DaemonSetManager struct {
	scheme *runtime.Scheme
	client client.Client
}

func newDaemonSetManager(scheme *runtime.Scheme, client client.Client) *DaemonSetManager {
	return &DaemonSetManager{scheme: scheme, client: client}
}

func (s *DaemonSetManager) Reconciler(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if err := s.reconcileDaemonSet(ctx, component); err != nil {
		return err
	}

	return s.reconcileService(ctx, component)
}

func (s *DaemonSetManager) reconcileDaemonSet(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	logger := log.FromContext(ctx)
	logger.Info("start to reconciling dst workload")

	ds := &appsv1.DaemonSet{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Name, Namespace: component.Namespace}, ds)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// return if already created
	if err == nil {
		return nil
	}
	// create the stateful set
	ds = s.constructDaemonSet(component)
	err = s.client.Create(ctx, ds)
	if err != nil {
		return err
	}
	logger.Info("create sts workload succeed")
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
func (s *DaemonSetManager) reconcileService(ctx context.Context, component *common.CacheRuntimeComponentValue) error {
	if component.Service == nil {
		return nil
	}
	logger := log.FromContext(ctx)
	logger.Info("start to reconciling headless service")

	svc := &corev1.Service{}
	err := s.client.Get(ctx, types.NamespacedName{Name: component.Service.Name, Namespace: component.Namespace}, svc)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// return if already created
	if err == nil {
		return nil
	}
	svc = s.constructService(component)
	err = s.client.Create(ctx, svc)
	if err != nil {
		return err
	}
	logger.Info("create headless service succeed")
	return nil
}

func (s *DaemonSetManager) constructService(component *common.CacheRuntimeComponentValue) *corev1.Service {
	matchLabels := getCommonLabelsFromComponent(component)

	trueVar := true
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      component.Service.Name,
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
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			Selector:                 matchLabels,
			PublishNotReadyAddresses: true,
		},
	}
	return svc
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

	return datav1alpha1.RuntimeComponentStatus{
		Phase:               datav1alpha1.RuntimePhaseReady,
		DesiredReplicas:     ds.Status.DesiredNumberScheduled,
		CurrentReplicas:     ds.Status.CurrentNumberScheduled,
		AvailableReplicas:   ds.Status.NumberAvailable,
		UnavailableReplicas: ds.Status.NumberUnavailable,
		ReadyReplicas:       ds.Status.NumberReady,
	}, nil
}
