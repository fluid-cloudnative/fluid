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
)

const (
	StatusReasonNotSupported metav1.StatusReason = "NotSupported"
)

func NewNotSupported(qualifiedResource schema.GroupResource, targetType string) *FluidStatusError {
	return &FluidStatusError{
		reason: StatusReasonNotSupported,
		details: &metav1.StatusDetails{
			Group: qualifiedResource.Group,
			Kind:  qualifiedResource.Resource,
		},
		message: fmt.Sprintf("%s is not supported by %s", qualifiedResource.Resource, targetType),
	}
}

// IsNotSupported returns true if the specified error was created by NewNotSupported.
func IsNotSupported(err error) (notSupported bool) {
	return ReasonForError(err) == StatusReasonNotSupported
}
