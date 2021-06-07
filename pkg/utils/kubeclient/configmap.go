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

	"k8s.io/api/core/v1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsConfigMapExist checks if the configMap exists given its name and namespace.
func IsConfigMapExist(client client.Client, name, namespace string) (found bool, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	cm := &v1.ConfigMap{}

	if err = client.Get(context.TODO(), key, cm); err != nil {
		if apierrs.IsNotFound(err) {
			found = false
			err = nil
		}
	} else {
		found = true
	}
	return found, err
}

// GetConfigmapByName gets configmap with given name and namespace of the configmap.
func GetConfigmapByName(client client.Client, name, namespace string) (configmap *v1.ConfigMap, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	configmap = &v1.ConfigMap{}

	if err = client.Get(context.TODO(), key, configmap); err != nil {
		if apierrs.IsNotFound(err) {
			err = nil
			configmap = nil
		}
		return configmap, err
	}

	return
}

// DeleteConfigMap deletes the configmap given its name and namespace if the configmap exists.
func DeleteConfigMap(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	found := false

	cm := &v1.ConfigMap{}
	if err = client.Get(context.TODO(), key, cm); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the configmap due to it's not found", "name", name,
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
		err = client.Delete(context.TODO(), cm)
	}

	return err
}
