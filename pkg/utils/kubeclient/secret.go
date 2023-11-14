/*
Copyright 2023 The Fluid Author.

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
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetSecret gets the secret.
// It returns a pointer to the secret if successful.
func GetSecret(client client.Client, name, namespace string) (*v1.Secret, error) {

	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	var secret v1.Secret
	if err := client.Get(context.TODO(), key, &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}

func CreateSecret(client client.Client, secret *v1.Secret) error {
	if err := client.Create(context.TODO(), secret); err != nil {
		return err
	}
	return nil
}

func UpdateSecret(client client.Client, secret *v1.Secret) error {
	if err := client.Update(context.TODO(), secret); err != nil {
		return err
	}
	return nil
}

func CopySecretToNamespace(client client.Client, from types.NamespacedName, to types.NamespacedName, ownerReference *common.OwnerReference) error {
	if _, err := GetSecret(client, to.Name, to.Namespace); err == nil {
		return nil
	}

	secret, err := GetSecret(client, from.Name, from.Namespace)
	if err != nil {
		return err
	}

	secretToCreate := &v1.Secret{}
	secretToCreate.Namespace = to.Namespace
	secretToCreate.Name = to.Name
	secretToCreate.Data = secret.Data
	secretToCreate.StringData = secret.StringData
	secretToCreate.Labels = map[string]string{}
	secretToCreate.Labels["fluid.io/copied-from"] = fmt.Sprintf("%s_%s", from.Namespace, from.Name)
	if ownerReference != nil {
		secretToCreate.OwnerReferences = append(secretToCreate.OwnerReferences, metav1.OwnerReference{
			APIVersion:         ownerReference.APIVersion,
			Kind:               ownerReference.Kind,
			Name:               ownerReference.Name,
			UID:                types.UID(ownerReference.UID),
			Controller:         &ownerReference.Controller,
			BlockOwnerDeletion: &ownerReference.BlockOwnerDeletion,
		})
	}

	if err = CreateSecret(client, secretToCreate); err != nil {
		return err
	}

	return nil
}
