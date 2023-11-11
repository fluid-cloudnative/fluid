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

package alluxio

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// transformSecurity transforms security configuration
func (e *AlluxioEngine) transformPermission(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {

	if len(value.Properties) == 0 {
		if len(runtime.Spec.Properties) > 0 {
			value.Properties = runtime.Spec.Properties
		} else {
			value.Properties = map[string]string{}
		}
	}
	setDefaultProperties(runtime, value, "alluxio.master.security.impersonation.root.users", "*")
	setDefaultProperties(runtime, value, "alluxio.master.security.impersonation.root.groups", "*")
	setDefaultProperties(runtime, value, "alluxio.security.authorization.permission.enabled", "false")
	// if runtime.Spec.RunAs != nil {
	// 	value.Properties[fmt.Sprintf("alluxio.master.security.impersonation.%d.users", runtime.Spec.RunAs.UID)]
	// 	value.Properties[fmt.Sprintf("alluxio.master.security.impersonation.%d.groups", runtime.Spec.RunAs.GID)]
	// }

}
