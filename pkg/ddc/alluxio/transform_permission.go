/*

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
