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

package v1alpha1

import corev1 "k8s.io/api/core/v1"

// InitUsersSpec is a description of the initialize the users for runtime
type InitUsersSpec struct {

	// Image for initialize the users for runtime(e.g. alluxio/alluxio-User init)
	Image string `json:"image,omitempty"`

	// Image Tag for initialize the users for runtime(e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Environment variables that will be used by initialize the users for runtime
	Env map[string]string `json:"env,omitempty"`

	// Resources that will be requested by initialize the users for runtime. <br>
	// <br>
	// Resources are not allowed for ephemeral containers. Ephemeral containers use spare resources
	// already allocated to the pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// HCFS Endpoint info
type HCFSStatus struct {
	// Endpoint for accessing
	Endpoint string `json:"endpoint,omitempty"`

	// Underlayer HCFS Compatible Version
	UnderlayerFileSystemVersion string `json:"underlayerFileSystemVersion,omitempty"`
}

// VersionSpec represents the settings for the  version that fluid is orchestrating.
type VersionSpec struct {
	// Image (e.g. alluxio/alluxio)
	Image string `json:"image,omitempty"`

	// Image tag (e.g. 2.3.0-SNAPSHOT)
	ImageTag string `json:"imageTag,omitempty"`

	// One of the three policies: `Always`, `IfNotPresent`, `Never`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`
}
