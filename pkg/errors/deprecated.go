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
