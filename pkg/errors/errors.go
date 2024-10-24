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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FluidStatusError is an error intended for consumption by the controller
// it's for Fluid internal error
type FluidStatusError struct {
	message string
	reason  metav1.StatusReason
	details *metav1.StatusDetails
}

func ReasonForError(err error) metav1.StatusReason {
	switch t := err.(type) {
	case StatusError:
		return t.Reason()
	}
	return metav1.StatusReasonUnknown
}

// StatusError is an interface for Fluid internal error
type StatusError interface {
	Reason() metav1.StatusReason
	Details() *metav1.StatusDetails
}
