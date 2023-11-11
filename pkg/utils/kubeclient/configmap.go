/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package kubeclient

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"

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

func CopyConfigMap(client client.Client, src types.NamespacedName, dst types.NamespacedName, reference metav1.OwnerReference) error {
	found, err := IsConfigMapExist(client, dst.Name, dst.Namespace)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// copy configmap
	srcConfigMap, err := GetConfigmapByName(client, src.Name, src.Namespace)
	if err != nil {
		return err
	}
	// if the source dataset configmap not found, return error and requeue
	if srcConfigMap == nil {
		return fmt.Errorf("runtime configmap %v do not exist", src)
	}
	// create the virtual dataset configmap if not exist
	copiedConfigMap := srcConfigMap.DeepCopy()

	dstConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            dst.Name,
			Namespace:       dst.Namespace,
			Labels:          copiedConfigMap.Labels,
			Annotations:     copiedConfigMap.Annotations,
			OwnerReferences: []metav1.OwnerReference{reference},
		},
		Data: copiedConfigMap.Data,
	}

	err = client.Create(context.TODO(), dstConfigMap)
	if err != nil {
		if otherErr := utils.IgnoreAlreadyExists(err); otherErr != nil {
			return err
		}
	}
	return nil
}

func UpdateConfigMap(client client.Client, cm *v1.ConfigMap) error {
	err := client.Update(context.TODO(), cm)
	return err
}
