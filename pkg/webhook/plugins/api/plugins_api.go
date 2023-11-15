/*
Copyright 2021 The Fluid Authors.

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

package api

import (
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MutatingHandler defines the interface for mutating a pod, implementations should be thread(goroutine) safe.
type MutatingHandler interface {
	// Mutate injects affinity info into pod
	// if a plugin return true, it means that no need to call other plugins
	// map[string]base.RuntimeInfoInterface's key is pvcName
	Mutate(*corev1.Pod, map[string]base.RuntimeInfoInterface) (shouldStop bool, err error)
	// GetName returns the name of plugin
	GetName() string
}

// RegistryHandler record the active plugins
// including two kinds: plugins for pod with no dataset mounted and with dataset mounted
type RegistryHandler interface {
	GetPodWithoutDatasetHandler() []MutatingHandler
	GetPodWithDatasetHandler() []MutatingHandler
	GetServerlessPodWithDatasetHandler() []MutatingHandler
	GetServerlessPodWithoutDatasetHandler() []MutatingHandler
}

// HandlerFactory is a function that builds a MutatingHandler.
type HandlerFactory = func(client client.Client, args string) MutatingHandler

type Registry map[string]HandlerFactory

func (r Registry) Register(name string, factory HandlerFactory) error {
	if _, ok := r[name]; ok {
		return fmt.Errorf("a plugin named %v already exists", name)
	}
	r[name] = factory
	return nil
}
