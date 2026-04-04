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

package utils

import (
	"context"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewRuntimeCondition creates a new Cache condition.
func NewRuntime(name, namespace string, category common.Category, runtimeType string, replicas int32) datav1alpha1.Runtime {
	return datav1alpha1.Runtime{
		Name:      name,
		Namespace: namespace,
		Category:  category,
		// Engine:    engine,
		Type:           runtimeType,
		MasterReplicas: replicas,
	}
}

// AddRuntimesIfNotExist adds newRuntime to runtimes and return the updated runtime slice
func AddRuntimesIfNotExist(runtimes []datav1alpha1.Runtime, newRuntime datav1alpha1.Runtime) (updatedRuntimes []datav1alpha1.Runtime) {
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
func GetAlluxioRuntime(client client.Reader, name, namespace string) (*datav1alpha1.AlluxioRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime datav1alpha1.AlluxioRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetJindoRuntime gets Jindo Runtime object with the given name and namespace
func GetJindoRuntime(client client.Reader, name, namespace string) (*datav1alpha1.JindoRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime datav1alpha1.JindoRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetGooseFSRuntime gets GooseFS Runtime object with the given name and namespace
func GetGooseFSRuntime(client client.Reader, name, namespace string) (*datav1alpha1.GooseFSRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime datav1alpha1.GooseFSRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

// GetJuiceFSRuntime gets JuiceFS Runtime object with the given name and namespace
func GetJuiceFSRuntime(client client.Reader, name, namespace string) (*datav1alpha1.JuiceFSRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime datav1alpha1.JuiceFSRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func GetThinRuntime(client client.Reader, name, namespace string) (*datav1alpha1.ThinRuntime, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime datav1alpha1.ThinRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

// GetEFCRuntime gets EFC Runtime object with the given name and namespace
func GetEFCRuntime(client client.Reader, name, namespace string) (*datav1alpha1.EFCRuntime, error) {

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var runtime datav1alpha1.EFCRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func GetThinRuntimeProfile(client client.Reader, name string) (*datav1alpha1.ThinRuntimeProfile, error) {
	key := types.NamespacedName{
		Name: name,
	}
	var runtimeProfile datav1alpha1.ThinRuntimeProfile
	if err := client.Get(context.TODO(), key, &runtimeProfile); err != nil {
		return nil, err
	}

	return &runtimeProfile, nil
}

// GetVineyardRuntime gets Vineyard Runtime object with the given name and namespace
func GetVineyardRuntime(client client.Reader, name, namespace string) (*datav1alpha1.VineyardRuntime, error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	var vineyardRuntime datav1alpha1.VineyardRuntime
	if err := client.Get(context.TODO(), key, &vineyardRuntime); err != nil {
		return nil, err
	}

	return &vineyardRuntime, nil
}

func GetCacheRuntime(client client.Reader, name, namespace string) (*datav1alpha1.CacheRuntime, error) {
	key := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	var runtime datav1alpha1.CacheRuntime
	if err := client.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}

	return &runtime, nil
}

func GetCacheRuntimeClass(client client.Client, name string) (*datav1alpha1.CacheRuntimeClass, error) {
	key := types.NamespacedName{
		Name: name,
	}
	var runtimeClass datav1alpha1.CacheRuntimeClass
	if err := client.Get(context.TODO(), key, &runtimeClass); err != nil {
		return nil, err
	}

	return &runtimeClass, nil
}
