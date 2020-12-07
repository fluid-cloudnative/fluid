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
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"time"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// IgnoreNotFound ignores not found
func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

// NoRequeue returns the result of a reconciler invocation and won't requeue.
func NoRequeue() (ctrl.Result, error) {
	return RequeueIfError(nil)
}

// RequeueAfterInterval returns the result of a reconciler invocation with a given requeue interval.
func RequeueAfterInterval(interval time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: interval}, nil
}

// RequeueImmediately returns the result of a reconciler invocation and requeue immediately.
func RequeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueIfError returns the result of a reconciler invocation and requeue immediately if err is not nil.
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

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ContainsOwners(owners []metav1.OwnerReference, dataset *datav1alpha1.Dataset) bool {
	for _, owner := range owners {
		if owner.UID == dataset.UID {
			return true
		}
	}
	return false
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
