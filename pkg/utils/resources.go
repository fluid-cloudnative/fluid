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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func TransformRequirementsToResources(res corev1.ResourceRequirements) (cRes common.Resources) {

	cRes = common.Resources{}

	if len(res.Requests) > 0 {
		cRes.Requests = make(common.ResourceList)
		for k, v := range res.Requests {
			cRes.Requests[k] = v.String()
		}
	}

	if len(res.Limits) > 0 {
		cRes.Limits = make(common.ResourceList)
		for k, v := range res.Limits {
			cRes.Limits[k] = v.String()
		}
	}

	return
}

func ResourceRequirementsEqual(source corev1.ResourceRequirements,
	target corev1.ResourceRequirements) bool {
	return resourceListsEqual(source.Requests, target.Requests) &&
		resourceListsEqual(source.Limits, target.Limits)
}

func resourceListsEqual(a corev1.ResourceList, b corev1.ResourceList) bool {
	a = withoutZeroElems(a)
	b = withoutZeroElems(b)
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		vb, found := b[k]
		if !found {
			return false
		}
		if v.Cmp(vb) != 0 {
			return false
		}
	}
	return true
}

func withoutZeroElems(input corev1.ResourceList) (output corev1.ResourceList) {
	output = corev1.ResourceList{}
	for k, v := range input {
		if !v.IsZero() {
			output[k] = v
		}
	}
	return
}
