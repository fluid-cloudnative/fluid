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

package lifecycle

import v1 "k8s.io/api/core/v1"

// toleratesTaints tolerates the taints in node
func toleratesTaints(taints []v1.Taint, tolerations []v1.Toleration) bool {
	filteredTaints := []v1.Taint{}
	for _, taint := range taints {
		if taint.Effect == v1.TaintEffectNoExecute || taint.Effect == v1.TaintEffectNoSchedule {
			filteredTaints = append(filteredTaints, taint)
		}
	}

	if len(tolerations) == 0 {
		return false
	}

	toleratesTaint := func(taint v1.Taint) bool {
		for _, toleration := range tolerations {
			if toleration.ToleratesTaint(&taint) {
				return true
			}
		}

		return false
	}

	for _, taint := range filteredTaints {
		if !toleratesTaint(taint) {
			return false
		}
	}

	return true
}
