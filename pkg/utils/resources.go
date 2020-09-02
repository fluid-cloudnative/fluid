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
