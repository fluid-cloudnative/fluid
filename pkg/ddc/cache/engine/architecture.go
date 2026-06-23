/*
  Copyright 2026 The Fluid Authors.

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

package engine

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// ArchitectureApi defines architecture-aware operations for a CacheRuntime.
// Different runtime topologies (Master-Worker / Workers-Only) are represented by
// different struct implementations, making it easy to extend with new architectures.
type ArchitectureApi interface {
	// GetExecutionPodInfo resolves the target pod name and container name used to execute
	// runtime operations (e.g. report summary, ufs mount).
	GetExecutionPodInfo() (podName string, containerName string, err error)

	// GetExecutionEntries returns the execution entries for this architecture.
	GetExecutionEntries() *datav1alpha1.ExecutionEntries

	// IsMountUFSSupported returns whether UFS mounting is supported on this architecture.
	IsMountUFSSupported() bool
}

// resolveArchitectureApi inspects the runtime class topology and runtime spec
// to decide which architecture handler should be used.
//
// Resolution rules:
//   - If runtimeClass.Topology.Master is configured AND runtime.Spec.Master is not
//     disabled, MasterWorkerArchitecture is returned.
//   - Otherwise, WorkersOnlyArchitecture is returned.
func resolveArchitectureApi(name, namespace string, runtime *datav1alpha1.CacheRuntime, runtimeClass *datav1alpha1.CacheRuntimeClass) ArchitectureApi {
	noMasterTopology := runtimeClass == nil || runtimeClass.Topology == nil || runtimeClass.Topology.Master == nil
	masterDisabled := runtime == nil || runtime.Spec.Master.Disabled
	if noMasterTopology || masterDisabled {
		return &workersOnlyArchApi{name: name, namespace: namespace, runtimeClass: runtimeClass}
	}
	return &masterWorkerArchApi{name: name, namespace: namespace, runtimeClass: runtimeClass}
}
