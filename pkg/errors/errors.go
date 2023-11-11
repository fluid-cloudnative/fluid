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
