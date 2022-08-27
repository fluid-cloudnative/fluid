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
	"encoding/json"
	"errors"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"strings"
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

	// 10. mountpath
	value.Fuse.MountPath = t.getMountPoint()

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
		Name:  common.ThinFuseOptionEnvKey,
		Value: options,
	}, corev1.EnvVar{
		Name:  common.ThinFusePointEnvKey,
		Value: value.Fuse.MountPath,
	})

	// 12. config
	config := make(map[string]string)
	config[value.Fuse.MountPath] = options
	var configStr []byte
	configStr, err = json.Marshal(config)
	if err != nil {
		return
	}
	value.Fuse.ConfigValue = string(configStr)

	// 13. critical
	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// 14. cachedir
	if len(runtime.Spec.TieredStore.Levels) > 0 {
		value.Fuse.CacheDir = runtime.Spec.TieredStore.Levels[0].Path
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

func (t ThinEngine) parseFuseOptions(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile, dataset *datav1alpha1.Dataset) (option string, err error) {
	options := make(map[string]string)
	if profile != nil && profile.Spec.Fuse.Options != nil {
		options = profile.Spec.Fuse.Options
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
	for _, encryptOption := range dataset.Spec.Mounts[0].EncryptOptions {
		key := encryptOption.Name
		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(t.Client, secretKeyRef.Name, t.namespace)
		if err != nil {
			t.Log.Info("can't get the secret",
				"namespace", t.namespace,
				"name", t.name,
				"secretName", secretKeyRef.Name)
			return "", err
		}
		val := secret.Data[secretKeyRef.Key]
		options[key] = string(val)
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
	value.Fuse.Image = profile.Spec.Version.Image
	value.Fuse.ImageTag = profile.Spec.Version.ImageTag
	value.Fuse.ImagePullPolicy = profile.Spec.Version.ImagePullPolicy
	if len(profile.Spec.Fuse.Image) != 0 {
		value.Fuse.Image = profile.Spec.Version.Image
	}
	if len(profile.Spec.Fuse.ImageTag) != 0 {
		value.Fuse.ImageTag = profile.Spec.Version.ImageTag
	}
	if len(profile.Spec.Fuse.ImagePullPolicy) != 0 {
		value.Fuse.ImagePullPolicy = profile.Spec.Version.ImagePullPolicy
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
	return
}
