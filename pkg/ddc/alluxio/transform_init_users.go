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
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// transform dataset which has ufsPaths and ufsVolumes
func (e *AlluxioEngine) transformInitUsers(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) {

	value.InitUsers = common.InitUsers{
		Enabled: false,
	}

	if runtime.Spec.RunAs != nil {
		value.UserInfo.User = int(*runtime.Spec.RunAs.UID)
		value.UserInfo.Group = int(*runtime.Spec.RunAs.GID)
		value.InitUsers = common.InitUsers{
			Enabled:        true,
			Dir:            e.getInitUserDir(),
			EnvUsers:       utils.GetInitUserEnv(runtime.Spec.RunAs),
			EnvTieredPaths: e.getInitTierPathsEnv(runtime),
		}

		initImageInfo := strings.Split(e.initImage, ":")
		value.InitUsers.ImageInfo = common.ImageInfo{
			Image:           initImageInfo[0],
			ImageTag:        initImageInfo[1],
			ImagePullPolicy: "IfNotPresent",
		}

	}

	// Overwrite ImageInfo.
	// Priority: runtime.Spec.InitUsers > helm chart value > default value
	if len(runtime.Spec.InitUsers.Image) > 0 {
		value.InitUsers.Image = runtime.Spec.InitUsers.Image
	}

	if len(runtime.Spec.InitUsers.ImageTag) > 0 {
		value.InitUsers.ImageTag = runtime.Spec.InitUsers.ImageTag
	}

	if len(runtime.Spec.InitUsers.ImagePullPolicy) > 0 {
		value.InitUsers.ImagePullPolicy = runtime.Spec.InitUsers.ImagePullPolicy
	}

	e.Log.Info("Check InitUsers", "InitUsers", value.InitUsers)

}
