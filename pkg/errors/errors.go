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
