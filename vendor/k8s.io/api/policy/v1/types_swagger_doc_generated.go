/*
Copyright The Kubernetes Authors.

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

package v1

// This file contains a collection of methods that can be used from go-restful to
// generate Swagger API documentation for its models. Please read this PR for more
// information on the implementation: https://github.com/emicklei/go-restful/pull/215
//
// TODOs are ignored from the parser (e.g. TODO(andronat):... || TODO:...) if and only if
// they are on one line! For multiple line or blocks that you want to ignore use ---.
// Any context after a --- is ignored.
//
// Those methods can be generated by using hack/update-generated-swagger-docs.sh

// AUTO-GENERATED FUNCTIONS START HERE. DO NOT EDIT.
var map_Eviction = map[string]string{
	"":              "Eviction evicts a pod from its node subject to certain policies and safety constraints. This is a subresource of Pod.  A request to cause such an eviction is created by POSTing to .../pods/<pod name>/evictions.",
	"metadata":      "ObjectMeta describes the pod that is being evicted.",
	"deleteOptions": "DeleteOptions may be provided",
}

func (Eviction) SwaggerDoc() map[string]string {
	return map_Eviction
}

var map_PodDisruptionBudget = map[string]string{
	"":         "PodDisruptionBudget is an object to define the max disruption that can be caused to a collection of pods",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"spec":     "Specification of the desired behavior of the PodDisruptionBudget.",
	"status":   "Most recently observed status of the PodDisruptionBudget.",
}

func (PodDisruptionBudget) SwaggerDoc() map[string]string {
	return map_PodDisruptionBudget
}

var map_PodDisruptionBudgetList = map[string]string{
	"":         "PodDisruptionBudgetList is a collection of PodDisruptionBudgets.",
	"metadata": "Standard object's metadata. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata",
	"items":    "Items is a list of PodDisruptionBudgets",
}

func (PodDisruptionBudgetList) SwaggerDoc() map[string]string {
	return map_PodDisruptionBudgetList
}

var map_PodDisruptionBudgetSpec = map[string]string{
	"":               "PodDisruptionBudgetSpec is a description of a PodDisruptionBudget.",
	"minAvailable":   "An eviction is allowed if at least \"minAvailable\" pods selected by \"selector\" will still be available after the eviction, i.e. even in the absence of the evicted pod.  So for example you can prevent all voluntary evictions by specifying \"100%\".",
	"selector":       "Label query over pods whose evictions are managed by the disruption budget. A null selector will match no pods, while an empty ({}) selector will select all pods within the namespace.",
	"maxUnavailable": "An eviction is allowed if at most \"maxUnavailable\" pods selected by \"selector\" are unavailable after the eviction, i.e. even in absence of the evicted pod. For example, one can prevent all voluntary evictions by specifying 0. This is a mutually exclusive setting with \"minAvailable\".",
}

func (PodDisruptionBudgetSpec) SwaggerDoc() map[string]string {
	return map_PodDisruptionBudgetSpec
}

var map_PodDisruptionBudgetStatus = map[string]string{
	"":                   "PodDisruptionBudgetStatus represents information about the status of a PodDisruptionBudget. Status may trail the actual state of a system.",
	"observedGeneration": "Most recent generation observed when updating this PDB status. DisruptionsAllowed and other status information is valid only if observedGeneration equals to PDB's object generation.",
	"disruptedPods":      "DisruptedPods contains information about pods whose eviction was processed by the API server eviction subresource handler but has not yet been observed by the PodDisruptionBudget controller. A pod will be in this map from the time when the API server processed the eviction request to the time when the pod is seen by PDB controller as having been marked for deletion (or after a timeout). The key in the map is the name of the pod and the value is the time when the API server processed the eviction request. If the deletion didn't occur and a pod is still there it will be removed from the list automatically by PodDisruptionBudget controller after some time. If everything goes smooth this map should be empty for the most of the time. Large number of entries in the map may indicate problems with pod deletions.",
	"disruptionsAllowed": "Number of pod disruptions that are currently allowed.",
	"currentHealthy":     "current number of healthy pods",
	"desiredHealthy":     "minimum desired number of healthy pods",
	"expectedPods":       "total number of pods counted by this disruption budget",
	"conditions":         "Conditions contain conditions for PDB. The disruption controller sets the DisruptionAllowed condition. The following are known values for the reason field (additional reasons could be added in the future): - SyncFailed: The controller encountered an error and wasn't able to compute\n              the number of allowed disruptions. Therefore no disruptions are\n              allowed and the status of the condition will be False.\n- InsufficientPods: The number of pods are either at or below the number\n                    required by the PodDisruptionBudget. No disruptions are\n                    allowed and the status of the condition will be False.\n- SufficientPods: There are more pods than required by the PodDisruptionBudget.\n                  The condition will be True, and the number of allowed\n                  disruptions are provided by the disruptionsAllowed property.",
}

func (PodDisruptionBudgetStatus) SwaggerDoc() map[string]string {
	return map_PodDisruptionBudgetStatus
}

// AUTO-GENERATED FUNCTIONS END HERE
