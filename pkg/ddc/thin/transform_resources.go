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

package thin

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func (t *ThinEngine) transformResourcesForWorker(resources corev1.ResourceRequirements, value *ThinValue) {
	if value.Worker.Resources.Requests == nil {
		value.Worker.Resources.Requests = common.ResourceList{}
	}

	if value.Worker.Resources.Limits == nil {
		value.Worker.Resources.Limits = common.ResourceList{}
	}

	if resources.Limits != nil {
		t.Log.Info("setting worker Resources limit")
		if quantity, ok := resources.Limits[corev1.ResourceCPU]; ok {
			value.Worker.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if quantity, ok := resources.Limits[corev1.ResourceMemory]; ok {
			value.Worker.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if resources.Requests != nil {
		t.Log.Info("setting worker Resources request")
		if quantity, ok := resources.Requests[corev1.ResourceCPU]; ok {
			value.Worker.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if quantity, ok := resources.Requests[corev1.ResourceMemory]; ok {
			value.Worker.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}
}

func (t *ThinEngine) transformResourcesForFuse(resources corev1.ResourceRequirements, value *ThinValue) {
	if value.Fuse.Resources.Requests == nil {
		value.Fuse.Resources.Requests = common.ResourceList{}
	}

	if value.Fuse.Resources.Limits == nil {
		value.Fuse.Resources.Limits = common.ResourceList{}
	}

	if resources.Limits != nil {
		t.Log.Info("setting fuse Resources limit")
		if quantity, ok := resources.Limits[corev1.ResourceCPU]; ok {
			value.Fuse.Resources.Limits[corev1.ResourceCPU] = quantity.String()
		}
		if quantity, ok := resources.Limits[corev1.ResourceMemory]; ok {
			value.Fuse.Resources.Limits[corev1.ResourceMemory] = quantity.String()
		}
	}

	if resources.Requests != nil {
		t.Log.Info("setting fuse Resources request")
		if quantity, ok := resources.Requests[corev1.ResourceCPU]; ok {
			value.Fuse.Resources.Requests[corev1.ResourceCPU] = quantity.String()
		}
		if quantity, ok := resources.Requests[corev1.ResourceMemory]; ok {
			value.Fuse.Resources.Requests[corev1.ResourceMemory] = quantity.String()
		}
	}
}
