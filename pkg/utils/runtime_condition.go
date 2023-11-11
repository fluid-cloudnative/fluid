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

	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondtion.Status {
		condition.LastTransitionTime = oldCondtion.LastTransitionTime
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
