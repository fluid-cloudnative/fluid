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

package utils

import (
	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewRuntimeCondition creates a new Cache condition.
func NewRuntimeCondition(conditionType data.RuntimeConditionType, reason, message string, status v1.ConditionStatus) data.RuntimeCondition {
	return data.RuntimeCondition{
		Type: conditionType,
		// Status:             v1.ConditionTrue,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		LastProbeTime:      metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// UpdateRuntimeCondition updates the runtime to include the provided condition.
// If the condition that we are about to add already exists
// and has the same status and reason then we are not going to update.
func UpdateRuntimeCondition(conditions []data.RuntimeCondition, condition data.RuntimeCondition) []data.RuntimeCondition {
	// conditions = trimRuntimeConditions(conditions)

	index, oldCondtion := GetRuntimeCondition(conditions, condition.Type)

	if oldCondtion == nil {
		conditions = append(conditions, condition)
		return conditions
	}

	// We define two types of runtime condition: ready type conditions (e.g. WorkerReady) and action type conditions (e.g. WorkerScaledOut).
	// For ready type conditions, recording the earliest transition and probe time is enough and
	// we avoiding update its probe time and transition time in every sync because it needs large amount of status updates.
	if isReadyTypeCondition(condition) {
		// keep old transition time and probe time if condition status is not changed
		if condition.Status == oldCondtion.Status {
			condition.LastTransitionTime = oldCondtion.LastTransitionTime
			condition.LastProbeTime = oldCondtion.LastProbeTime
		}
	}

	conditions[index] = condition
	return conditions
}

// GetRuntimeCondition gets a runtime condition given a runtime condition type.
// If found, return index of the founded condition in the condition array
// and the founded condition itself, otherwise return -1 and nil.
func GetRuntimeCondition(conditions []data.RuntimeCondition,
	condType data.RuntimeConditionType) (index int, condition *data.RuntimeCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == condType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

func isReadyTypeCondition(condition data.RuntimeCondition) bool {
	return condition.Type == data.RuntimeMasterReady ||
		condition.Type == data.RuntimeFusesReady ||
		condition.Type == data.RuntimeWorkersReady
}

// func trimRuntimeConditions(conditions []data.RuntimeCondition) []data.RuntimeCondition {
// 	knownConditions := map[data.RuntimeConditionType]bool{}
// 	newConditions := []data.RuntimeCondition{}
// 	for _, condition := range conditions {
// 		if _, found := knownConditions[condition.Type]; !found {
// 			newConditions = append(newConditions, condition)
// 			knownConditions[condition.Type] = true
// 		}
// 	}

// 	return newConditions
// }
