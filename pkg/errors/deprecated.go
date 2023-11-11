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
