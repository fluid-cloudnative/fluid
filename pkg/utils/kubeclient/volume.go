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

	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsPersistentVolumeExist
func IsPersistentVolumeExist(client client.Client, name string, annotations map[string]string) (found bool, err error) {
	key := types.NamespacedName{
		Name: name,
	}

	pv := &v1.PersistentVolume{}

	err = client.Get(context.TODO(), key, pv)
	if err != nil {
		if apierrs.IsNotFound(err) {
			found = false
			err = nil
		}
	} else if len(pv.Annotations) == 0 {
		found = false
	} else {
		for k, v := range annotations {
			value, _ := pv.Annotations[k]
			if value != v {
				found = false
				log.Info("The expected pv's annotation doesn't equal to what it has", "key", k,
					"expectedValue", v,
					"actualValue", value)
				return
			}
		}
		log.Info("The persistentVolume exist", "name", name,
			"annotaitons", annotations)
		found = true
	}

	return found, err
}

// IsPersistentVolumeClaim
func IsPersistentVolumeClaim(client client.Client, name, namespace string, annotations map[string]string) (found bool, err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	pvc := &v1.PersistentVolumeClaim{}
	err = client.Get(context.TODO(), key, pvc)
	if err != nil {
		if apierrs.IsNotFound(err) {
			found = false
			err = nil
		}
	} else if len(pvc.Annotations) == 0 {
		found = false
	} else {
		for k, v := range annotations {
			value, _ := pvc.Annotations[k]
			if value != v {
				found = false
				log.Info("The expected pvc's annotation doesn't equal to what it has", "key", k,
					"expectedValue", v,
					"actualValue", value)
				return
			}
		}
		log.Info("The persistentVolume exist", "name", name,
			"annotaitons", annotations)
		found = true
	}

	return found, err

}

// DeletePersistentVolume deletes volume
func DeletePersistentVolume(client client.Client, name string) (err error) {
	key := types.NamespacedName{
		Name: name,
	}
	found := false
	pv := &v1.PersistentVolume{}
	if err = client.Get(context.TODO(), key, pv); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the PersistentVolume due to it's not found", "name", name)
			found = false
			err = nil
		} else {
			return
		}
	} else {
		found = true
	}
	if found {
		err = client.Delete(context.TODO(), pv)
		if err != nil && !apierrs.IsNotFound(err) {
			return fmt.Errorf("error deleting pv %s: %s", name, err.Error())
		}
	}

	return
}

// DeletePersistentVolumeClaim deletes volume claim
func DeletePersistentVolumeClaim(client client.Client, name, namespace string) (err error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}

	found := false
	pvc := &v1.PersistentVolumeClaim{}
	if err = client.Get(context.TODO(), key, pvc); err != nil {
		if apierrs.IsNotFound(err) {
			log.V(1).Info("SKip deleteing the PersistentVolumeClaim due to it's not found", "name", name,
				"namespace", namespace)
			found = false
			err = nil
		} else {
			return
		}
	} else {
		found = true
	}
	if found {
		err = client.Delete(context.TODO(), pvc)
		if err != nil && !apierrs.IsNotFound(err) {
			return fmt.Errorf("error deleting pvc %s: %s", name, err.Error())
		}
	}

	return
}
