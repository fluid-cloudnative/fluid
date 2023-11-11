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

package utils

import (
	"context"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewRuntimeCondition creates a new Cache condition.
func NewRuntime(name, namespace string, category common.Category, runtimeType string, replicas int32) data.Runtime {
	return data.Runtime{
		Name:      name,
		Namespace: namespace,
		Category:  category,
		// Engine:    engine,
		Type:           runtimeType,
		MasterReplicas: replicas,
	}
}

// AddRuntimesIfNotExist adds newRuntime to runtimes and return the updated runtime slice
func AddRuntimesIfNotExist(runtimes []data.Runtime, newRuntime data.Runtime) (updatedRuntimes []data.Runtime) {
	categoryMap := map[common.Category]bool{}
	for _, runtime := range runtimes {
		categoryMap[runtime.Category] = true
	}

	if _, found := categoryMap[newRuntime.Category]; !found {
		updatedRuntimes = append(runtimes, newRuntime)
	} else {
		updatedRuntimes = runtimes
		log.V(1).Info("No need to add new runtime to dataset", "type", newRuntime.Category)
	}

	return updatedRuntimes
}

// GetAlluxioRuntime gets Alluxio Runtime object with the given name and namespace
func GetAlluxioRuntime(client client.Client, name, namespace string) (*data.AlluxioRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.AlluxioRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetJindoRuntime gets Jindo Runtime object with the given name and namespace
func GetJindoRuntime(client client.Client, name, namespace string) (*data.JindoRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.JindoRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetGooseFSRuntime gets GooseFS Runtime object with the given name and namespace
func GetGooseFSRuntime(client client.Client, name, namespace string) (*data.GooseFSRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.GooseFSRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetJuiceFSRuntime gets JuiceFS Runtime object with the given name and namespace
func GetJuiceFSRuntime(client client.Client, name, namespace string) (*data.JuiceFSRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.JuiceFSRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func GetThinRuntime(client client.Client, name, namespace string) (*data.ThinRuntime, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.ThinRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

// GetEFCRuntime gets EFC Runtime object with the given name and namespace
func GetEFCRuntime(client client.Client, name, namespace string) (*data.EFCRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime data.EFCRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func GetThinRuntimeProfile(client client.Client, name string) (*data.ThinRuntimeProfile, error) {
	key := types.NamespacedName{
		Name: name,
	}
	var runtimeProfile data.ThinRuntimeProfile
	if err := client.Get(context.TODO(), key, &runtimeProfile); err != nil {
		return nil, err
	}

	return &runtimeProfile, nil
}
