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

package errors

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// StatusReasonDeprecated means the mode has been deprecated
	StatusReasonDeprecated metav1.StatusReason = "Deprecated"
)

func (e FluidStatusError) Error() string {
	return e.message
}

func (e FluidStatusError) Reason() metav1.StatusReason {
	return e.reason
}

func (e FluidStatusError) Details() *metav1.StatusDetails {
	return e.details
}

func NewDeprecated(qualifiedResource schema.GroupResource, key types.NamespacedName) *FluidStatusError {
	return &FluidStatusError{
		reason: StatusReasonDeprecated,
		details: &metav1.StatusDetails{
			Group: qualifiedResource.Group,
			Kind:  qualifiedResource.Resource,
		},
		message: fmt.Sprintf("%s in namespace %s %q is deprecated", qualifiedResource.String(), key.Namespace, key.Name),
	}
}

// IsDeprecated returns true if the specified error was created by NewDeprecated.
func IsDeprecated(err error) (deprecated bool) {
	return ReasonForError(err) == StatusReasonDeprecated
}
