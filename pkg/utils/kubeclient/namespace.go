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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnsureNamespace makes sure the namespace exist.
func EnsureNamespace(client client.Client, namespace string) (err error) {
	key := types.NamespacedName{
		Name: namespace,
	}

	var ns v1.Namespace

	if err = client.Get(context.TODO(), key, &ns); err != nil {
		if apierrs.IsNotFound(err) {
			return createNamespace(client, namespace)
		}
	}
	return err
}

func createNamespace(client client.Client, namespace string) error {
	created := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	return client.Create(context.TODO(), created)
}
