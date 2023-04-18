/*
Copyright 2023 The Fluid Authors.

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

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// get service given name and namespace of the service.
func GetServiceByName(client client.Client, name, namespace string) (service *corev1.Service, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	service = &corev1.Service{}

	if err = client.Get(context.TODO(), key, service); err != nil {
		if apierrs.IsNotFound(err) {
			err = nil
			service = nil
		}
		return service, err
	}

	return
}
