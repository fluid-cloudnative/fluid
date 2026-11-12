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

	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// reconcileService reconciles the headless service for a component
func reconcileService(ctx context.Context, c client.Client, component *common.CacheRuntimeComponentValue) error {
	if component.Service == nil {
		return nil
	}
	logger := log.FromContext(ctx)
	logger.Info("start to reconciling headless service")

	svc := &corev1.Service{}
	err := c.Get(ctx, types.NamespacedName{Name: component.Service.Name, Namespace: component.Namespace}, svc)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	// return if already created
	if err == nil {
		return nil
	}
	svc = constructService(component)
	err = c.Create(ctx, svc)
	if err != nil {
		return err
	}
	logger.Info("create headless service succeed")
	return nil
}

// constructService constructs a headless service for a component
func constructService(component *common.CacheRuntimeComponentValue) *corev1.Service {
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
