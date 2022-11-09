/*

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

	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
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

func CopySecretToNamespace(client client.Client, secretName, fromNamespace, toNamespace string) error {
	if _, err := GetSecret(client, secretName, toNamespace); err == nil {
		// secret already exists, skip
		return nil
	}

	secret, err := GetSecret(client, secretName, fromNamespace)
	if err != nil {
		if apierrs.IsNotFound(err) {
			// original secret not found, ignore & skip
			return nil
		}
		return err
	}

	secretToCreate := secret.DeepCopy()
	secretToCreate.Namespace = toNamespace
	if len(secretToCreate.Labels) == 0 {
		secretToCreate.Labels = map[string]string{}
	}
	secretToCreate.Labels["fluid.io/copied-from"] = fmt.Sprintf("%s_%s", fromNamespace, secretName)

	if err = CreateSecret(client, secretToCreate); err != nil {
		return err
	}

	return nil
}
