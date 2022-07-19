/*
  Copyright 2022 The Fluid Authors.

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

package thin

import (
	"errors"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func (t *ThinEngine) transformFuse(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinProfile, dataset *datav1alpha1.Dataset, value *ThinValue) (err error) {
	value.Fuse = Fuse{}
	value.Fuse.Enabled = true

	value.Fuse.Image, value.Fuse.ImageTag, value.ImagePullPolicy = t.parseFuseImage(runtime, profile)

	value.Fuse.MountPath = t.getMountPoint()

	if len(dataset.Spec.Mounts) <= 0 {
		return errors.New("do not assign mount point")
	}

	options := t.parseFuseOptions(runtime, profile, dataset)

	envs := t.parseFuseEnv(runtime, profile)
	envs = append(envs, corev1.EnvVar{
		Name:  common.ThinFuseOptionEnvKey,
		Value: options,
	})
	value.Fuse.Envs = envs

	value.Fuse.Args = runtime.Spec.Fuse.Args
	if len(value.Fuse.Args) == 0 && profile != nil {
		value.Fuse.Args = profile.Spec.Args
	}

	value.Fuse.Command = runtime.Spec.Fuse.Command
	if len(value.Fuse.Command) == 0 && profile != nil {
		value.Fuse.Command = profile.Spec.Command
	}

	t.transformResourcesForFuse(runtime, value)

	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	} else {
		value.Fuse.NodeSelector = map[string]string{}
	}
	value.Fuse.NodeSelector[t.getFuseLabelName()] = "true"

	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode)

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// transform volumes for fuse
	err = t.transformFuseVolumes(runtime, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes for fuse")
	}
	return
}

func (t ThinEngine) parseFuseImage(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinProfile) (image string, tag string, imagePullPolicy string) {
	image = runtime.Spec.Fuse.Image
	tag = runtime.Spec.Fuse.ImageTag
	imagePullPolicy = runtime.Spec.Fuse.ImagePullPolicy
	if profile != nil {
		if len(image) == 0 {
			image = profile.Spec.Version.Image
		}
		if len(tag) == 0 {
			tag = profile.Spec.Version.ImageTag
		}
		if len(imagePullPolicy) == 0 {
			imagePullPolicy = profile.Spec.Version.ImagePullPolicy
		}
	}
	return
}

func (t ThinEngine) parseFuseOptions(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinProfile, dataset *datav1alpha1.Dataset) (option string) {
	options := make(map[string]string)
	if profile != nil {
		options = profile.Spec.Options
	}
	// option in runtime will cover option in profile
	for k, v := range runtime.Spec.Fuse.Options {
		options[k] = v
	}
	// option in dataset will cover option in runtime
	for k, v := range dataset.Spec.Mounts[0].Options {
		// support only one mountpoint
		options[k] = v
	}
	optionList := make([]string, len(options))
	for k, v := range options {
		if len(v) != 0 {
			optionList = append(optionList, fmt.Sprintf("%s=%s", k, v))
		} else {
			optionList = append(optionList, fmt.Sprintf("%s", k))
		}
	}
	if len(optionList) != 0 {
		option = strings.Join(optionList, ",")
	}
	return
}

func (t ThinEngine) parseFuseEnv(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinProfile) (envs []corev1.EnvVar) {
	if profile != nil {
		envs = profile.Spec.Env
	}
	envs = append(envs, runtime.Spec.Fuse.Env...)
	return
}
