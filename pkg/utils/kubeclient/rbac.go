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
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func DeleteServiceAccount(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	found := false

	sa := &corev1.ServiceAccount{}
	if err = client.Get(context.TODO(), key, sa); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the serviceAccount due to it's not found", "name", name,
				"namespace", namespace)
			found = false
			err = nil
		} else {
			return err
		}
	} else {
		found = true
	}
	if found {
		err = client.Delete(context.TODO(), sa)
	}

	return err
}

func DeleteRole(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	found := false

	role := &rbacv1.Role{}
	if err = client.Get(context.TODO(), key, role); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the role due to it's not found", "name", name,
				"namespace", namespace)
			found = false
			err = nil
		} else {
			return err
		}
	} else {
		found = true
	}
	if found {
		err = client.Delete(context.TODO(), role)
	}

	return err
}

func DeleteRoleBinding(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	found := false

	roleBinding := &rbacv1.RoleBinding{}
	if err = client.Get(context.TODO(), key, roleBinding); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the rolebinding due to it's not found", "name", name,
				"namespace", namespace)
			found = false
			err = nil
		} else {
			return err
		}
	} else {
		found = true
	}
	if found {
		err = client.Delete(context.TODO(), roleBinding)
	}

	return err
}
