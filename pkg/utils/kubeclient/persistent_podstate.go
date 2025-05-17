/*
Copyright 2025 The Fluid Authors.

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

package kubeclient

import (
	"context"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetPersistentPodState gets PersistentPodState with given name and namespace of the configmap.
func GetPersistentPodState(client client.Client, name, namespace string) (persistentPodState *v1alpha1.PersistentPodState, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	persistentPodState = &v1alpha1.PersistentPodState{}

	if err = client.Get(context.TODO(), key, persistentPodState); err != nil {
		if apierrs.IsNotFound(err) {
			err = nil
			persistentPodState = nil
		}
		return persistentPodState, err
	}

	return
}

func UpdatePersistentPodStateStatus(client client.Client, persistentPodState *v1alpha1.PersistentPodState) (err error) {
	err = client.Status().Update(context.TODO(), persistentPodState)
	return
}
