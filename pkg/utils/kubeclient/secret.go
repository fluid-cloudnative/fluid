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
