/*
Copyright 2020 The Fluid Authors.

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
	"fmt"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/utils"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	versionutil "github.com/fluid-cloudnative/fluid/pkg/utils/version"
)

const (
	// https://github.com/Alluxio/alluxio/pull/15318/files
	newFuseArgsVersion = "2.8.0"
)

// 4. Transform the fuse
func (e *AlluxioEngine) transformFuse(runtime *datav1alpha1.AlluxioRuntime, dataset *datav1alpha1.Dataset, value *Alluxio) (err error) {
	value.Fuse = Fuse{}

	image := runtime.Spec.Fuse.Image
	tag := runtime.Spec.Fuse.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.ImagePullPolicy
	imagePullSecrets := runtime.Spec.Fuse.ImagePullSecrets

	value.Fuse.Image, value.Fuse.ImageTag, value.Fuse.ImagePullPolicy, value.Fuse.ImagePullSecrets = e.parseFuseImage(image, tag, imagePullPolicy, imagePullSecrets)

	if len(runtime.Spec.Fuse.Properties) > 0 {
		value.Fuse.Properties = runtime.Spec.Fuse.Properties
		runtime.Spec.Properties = utils.UnionMapsWithOverride(runtime.Spec.Properties, runtime.Spec.Fuse.Properties)
	}

	if len(runtime.Spec.Fuse.Env) > 0 {
		value.Fuse.Env = runtime.Spec.Fuse.Env
	} else {
		value.Fuse.Env = map[string]string{}
	}

	value.Fuse.MountPath = e.getMountPoint()

	// If the alluxio version is 2.8.0 or greater, the MOUNT_POINT env is not supported anymore.
	// Instead, it will be put into the fuse args
	// https://github.com/Alluxio/alluxio/pull/15318/files
	isNewFuseArgVersion, err := checkIfNewFuseArgVersion(value.Fuse.ImageTag)
	if err != nil {
		e.Log.Error(err, "Failed to transform fuse")
		return err
	}

	if !isNewFuseArgVersion {
		value.Fuse.Env["MOUNT_POINT"] = value.Fuse.MountPath
	}

	e.Log.Info("Check if the alluxio version not less than 2.8",
		"version", value.Fuse.ImageTag,
		"isNewFuseArgVersion", isNewFuseArgVersion)

	e.optimizeDefaultFuse(runtime, value, isNewFuseArgVersion)

	if len(value.Fuse.Args) > 1 && dataset.Spec.Owner != nil {
		value.Fuse.Args[1] = strings.Join([]string{value.Fuse.Args[1], fmt.Sprintf("uid=%d,gid=%d", *dataset.Spec.Owner.UID, *dataset.Spec.Owner.GID)}, ",")
	} else {
		if len(value.Properties) == 0 {
			value.Properties = map[string]string{}
		}
		value.Properties["alluxio.fuse.user.group.translation.enabled"] = "true"
	}
	// value.Fuse.Args[-1]

	// Allow root: only the RunAs user and root can access fuse
	//if !strings.Contains(value.Fuse.Args[len(value.Fuse.Args)-1], "allow_") {
	//	value.Fuse.Args[len(value.Fuse.Args)-1] = strings.Join([]string{value.Fuse.Args[len(value.Fuse.Args)-1], "allow_root"}, ",")
	//}

	// Allow others: all users(including root) can access fuse
	if len(value.Fuse.Args) > 1 && !strings.Contains(value.Fuse.Args[1], "allow_") {
		value.Fuse.Args[1] = strings.Join([]string{value.Fuse.Args[1], "allow_other"}, ",")
	}

	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	} else {
		value.Fuse.NodeSelector = map[string]string{}
	}
	value.Fuse.NodeSelector[utils.GetFuseLabelName(runtime.Namespace, runtime.Name, e.runtimeInfo.GetOwnerDatasetUID())] = "true"

	// parse fuse container network mode
	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode)
	value.Fuse.HostPID = common.HostPIDEnabled(runtime.Annotations)

	value.Fuse.Enabled = true

	e.transformResourcesForFuse(runtime, value)

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// transform volumes for worker
	err = e.transformFuseVolumes(runtime, value)
	if err != nil {
		e.Log.Error(err, "failed to transform volumes for fuse")
	}

	return

}

func checkIfNewFuseArgVersion(version string) (newFuseVersion bool, err error) {
	compare, err := versionutil.Compare(version, newFuseArgsVersion)
	if err != nil {

		return
	}
	newFuseVersion = compare >= 0
	return newFuseVersion, err
}
