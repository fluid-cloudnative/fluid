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
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/utils"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

func (t *ThinEngine) transformFuse(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile, dataset *datav1alpha1.Dataset, value *ThinValue) (err error) {
	value.Fuse = Fuse{
		Enabled: true,
		Ports:   []corev1.ContainerPort{},
		Envs:    []corev1.EnvVar{},
	}

	t.parseFromProfileFuse(profile, value)

	// 1. image
	t.parseFuseImage(runtime, value)
	if len(value.Fuse.Image) == 0 {
		err = fmt.Errorf("fuse %s image is nil", runtime.Name)
		return
	}
	// 2. resources
	t.transformResourcesForFuse(runtime.Spec.Fuse.Resources, value)

	// 3. volumes
	err = t.transformFuseVolumes(runtime.Spec.Volumes, runtime.Spec.Fuse.VolumeMounts, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes for fuse")
	}

	// 4. nodeSelector
	value.Fuse.NodeSelector = map[string]string{}
	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	}
	value.Fuse.NodeSelector[t.getFuseLabelName()] = "true"

	// 5. ports
	if len(runtime.Spec.Fuse.Ports) != 0 {
		value.Fuse.Ports = append(value.Fuse.Ports, runtime.Spec.Fuse.Ports...)
	}

	// 6. probe
	if runtime.Spec.Fuse.ReadinessProbe != nil {
		value.Fuse.ReadinessProbe = runtime.Spec.Fuse.ReadinessProbe
	}
	if runtime.Spec.Fuse.LivenessProbe != nil {
		value.Fuse.LivenessProbe = runtime.Spec.Fuse.LivenessProbe
	}
	// 7. command
	if len(runtime.Spec.Fuse.Command) != 0 {
		value.Fuse.Command = runtime.Spec.Fuse.Command
	}
	// 8. args
	if len(runtime.Spec.Fuse.Args) != 0 {
		value.Fuse.Args = runtime.Spec.Fuse.Args
	}

	// 9. network
	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode)
	value.Fuse.HostPID = common.HostPIDEnabled()

	// 10. targetPath to mount
	value.Fuse.TargetPath = t.getTargetPath()

	if len(dataset.Spec.Mounts) <= 0 {
		return errors.New("do not assign mount point")
	}

	// 11. env
	options, err := t.parseFuseOptions(runtime, profile, dataset)
	if err != nil {
		return
	}
	value.Fuse.Envs = append(value.Fuse.Envs, runtime.Spec.Fuse.Env...)
	value.Fuse.Envs = append(value.Fuse.Envs, corev1.EnvVar{
		Name:  common.ThinFusePointEnvKey,
		Value: value.Fuse.TargetPath,
	})

	// If Fuse options is not set, skip it.
	if len(options) > 0 {
		value.Fuse.Envs = append(value.Fuse.Envs, corev1.EnvVar{
			Name:  common.ThinFuseOptionEnvKey,
			Value: options,
		})
	}

	// 12. fuse config
	err = t.transformFuseConfig(runtime, dataset, value)
	if err != nil {
		return err
	}

	// 13. critical
	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// 14. cachedir
	if len(runtime.Spec.TieredStore.Levels) > 0 {
		value.Fuse.CacheDir = runtime.Spec.TieredStore.Levels[0].Path
	}

	// 15. mount related node publish secret to fuse if the dataset specifies any mountpoint with pvc type.
	err = t.transfromSecretsForPersistentVolumeClaimMounts(dataset, profile.Spec.NodePublishSecretPolicy, value)
	if err != nil {
		return err
	}
	return
}

func (t *ThinEngine) parseFuseImage(runtime *datav1alpha1.ThinRuntime, value *ThinValue) {
	if len(runtime.Spec.Fuse.Image) != 0 {
		value.Fuse.Image = runtime.Spec.Fuse.Image
	}
	if len(runtime.Spec.Fuse.ImageTag) != 0 {
		value.Fuse.ImageTag = runtime.Spec.Fuse.ImageTag
	}
	if len(runtime.Spec.Fuse.ImagePullPolicy) != 0 {
		value.Fuse.ImagePullPolicy = runtime.Spec.Fuse.ImagePullPolicy
	}
}

func (t *ThinEngine) parseFuseOptions(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile, dataset *datav1alpha1.Dataset) (option string, err error) {
	options := make(map[string]string)
	runtimeInfo := t.runtimeInfo
	if runtimeInfo != nil {
		accessModes, err := utils.GetAccessModesOfDataset(t.Client, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			t.Log.Info("Error:", "err", err)
		}
		if len(accessModes) > 0 {
			for _, mode := range accessModes {
				if mode == corev1.ReadOnlyMany {
					options["ro"] = ""
					break
				}
			}
		}
	}
	if profile != nil && profile.Spec.Fuse.Options != nil {
		options = profile.Spec.Fuse.Options
	}
	// option in runtime will cover option in profile
	for k, v := range runtime.Spec.Fuse.Options {
		options[k] = v
	}

	optionList := make([]string, 0, len(options))
	for k, v := range options {
		if len(v) != 0 {
			optionList = append(optionList, fmt.Sprintf("%s=%s", k, v))
		} else {
			optionList = append(optionList, k)
		}
	}
	if len(optionList) != 0 {
		option = strings.Join(optionList, ",")
	}
	return
}

func (t *ThinEngine) parseFromProfileFuse(profile *datav1alpha1.ThinRuntimeProfile, value *ThinValue) {
	if profile == nil {
		return
	}
	// 1. image
	value.Fuse.Image = profile.Spec.Fuse.Image
	value.Fuse.ImageTag = profile.Spec.Fuse.ImageTag
	value.Fuse.ImagePullPolicy = profile.Spec.Fuse.ImagePullPolicy
	if len(profile.Spec.Fuse.Image) != 0 {
		value.Fuse.Image = profile.Spec.Fuse.Image
	}
	if len(profile.Spec.Fuse.ImageTag) != 0 {
		value.Fuse.ImageTag = profile.Spec.Fuse.ImageTag
	}
	if len(profile.Spec.Fuse.ImagePullPolicy) != 0 {
		value.Fuse.ImagePullPolicy = profile.Spec.Fuse.ImagePullPolicy
	}
	// 2. resources
	t.transformResourcesForFuse(profile.Spec.Fuse.Resources, value)

	// 3. volumes
	err := t.transformFuseVolumes(profile.Spec.Volumes, profile.Spec.Fuse.VolumeMounts, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes from profile for worker")
	}

	// 4. nodeSelector
	if len(profile.Spec.Fuse.NodeSelector) != 0 {
		value.Fuse.NodeSelector = profile.Spec.Fuse.NodeSelector
	}
	// 5. ports
	if len(profile.Spec.Fuse.Ports) != 0 {
		value.Fuse.Ports = profile.Spec.Fuse.Ports
	}
	// 6. probe
	value.Fuse.ReadinessProbe = profile.Spec.Fuse.ReadinessProbe
	value.Fuse.LivenessProbe = profile.Spec.Fuse.LivenessProbe
	// 7. command
	if len(profile.Spec.Fuse.Command) != 0 {
		value.Fuse.Command = profile.Spec.Fuse.Command
	}
	// 8. args
	if len(profile.Spec.Fuse.Args) != 0 {
		value.Fuse.Args = profile.Spec.Fuse.Args
	}
	// 9. network
	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(profile.Spec.Fuse.NetworkMode)
	// 10. env
	if len(profile.Spec.Fuse.Env) != 0 {
		value.Fuse.Envs = profile.Spec.Fuse.Env
	}
}
