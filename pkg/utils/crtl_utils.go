/*
Copyright 2023 The Fluid Authors.

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
	"strings"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// IgnoreAlreadyExists ignores already existes error
func IgnoreAlreadyExists(err error) error {
	if apierrs.IsAlreadyExists(err) {
		return nil
	}
	return err
}

// IgnoreNotFound ignores not found
func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func IgnoreNoKindMatchError(err error) error {
	if apimeta.IsNoMatchError(err) {
		return nil
	}
	return err
}

// NoRequeue returns the result of a reconcile invocation and no err
// The Object will not requeue
func NoRequeue() (ctrl.Result, error) {
	return RequeueIfError(nil)
}

// RequeueAfterInterval returns the result of a reconcile invocation with a given requeue interval and no err
// The Object will requeue after the given requeue interval
func RequeueAfterInterval(interval time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: interval}, nil
}

// RequeueImmediately returns the result of a reconciler invocation and no err
// The Object will requeue immediately whether the err is nil or not
func RequeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueIfError returns the result of a reconciler invocation and the err
// The Object will requeue when err is not nil
func RequeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueImmediatelyUnlessGenerationChanged requeues immediately if the object generation has not changed.
// Otherwise, since the generation change will trigger an immediate update anyways, this will not requeue.
// This prevents some cases where two reconciliation loops will occur.
func RequeueImmediatelyUnlessGenerationChanged(prevGeneration, curGeneration int64) (ctrl.Result, error) {
	if prevGeneration == curGeneration {
		return RequeueImmediately()
	} else {
		return NoRequeue()
	}
}

// GetOrDefault returns the default value unless there is a specified value.
func GetOrDefault(str *string, defaultValue string) string {
	if str == nil {
		return defaultValue
	} else {
		return *str
	}
}

// Now returns the current time
func Now() *metav1.Time {
	now := metav1.Now()
	return &now
}

// ContainsString Determine whether the string array contains a specific string
// return true if contains the string and return false if not.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ContainsSubString Determine whether the string array contains a sub string
// return true if contains the string and return false if not.
func ContainsSubString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.Contains(item, s) {
			return true
		}
	}
	return false
}

// ContainsOwners Determine whether the slice of owners contains the owner of a Dataset
// return true if contains the owner and return false if not.
func ContainsOwners(owners []metav1.OwnerReference, dataset *datav1alpha1.Dataset) bool {
	for _, owner := range owners {
		if owner.UID == dataset.UID {
			return true
		}
	}
	return false
}

func ContainsLabel(labels map[string]string, labelKey, labelValue string) bool {
	value, exit := labels[labelKey]
	if !exit {
		return false
	}
	if value == labelValue {
		return true
	}
	return false
}

// ContainsSelector Determine whether the labels contain the selector
func ContainsSelector(labels map[string]string, selector map[string]string) bool {
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

// RemoveString removes strings in a array, which is equal to a given string.
func RemoveString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// HasDeletionTimestamp method that makes logic easier to read.
func HasDeletionTimestamp(obj metav1.ObjectMeta) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// CalculateDuration generates a string of duration from creationTime and finishTime
// if finish time is zero, use current time as default
func CalculateDuration(creationTime time.Time, finishTime time.Time) string {
	if finishTime.IsZero() {
		finishTime = time.Now()
	}
	return finishTime.Sub(creationTime).Round(time.Second).String()
}
