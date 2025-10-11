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
	"fmt"
	"path/filepath"
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

	err = t.parseHostVolumeFromDataset(dataset, value)
	if err != nil {
		t.Log.Error(err, "failed to transform from dataset")
	}

	// 1. image
	t.parseFuseImage(runtime, value)
	if len(value.Fuse.Image) == 0 {
		err = fmt.Errorf("fuse %s image is nil", runtime.Name)
		return
	}

	// 2. labels & annotations
	t.transformFusePodMetadata(runtime.Spec.Fuse.PodMetadata, value)
	// 3. resources
	t.transformResourcesForFuse(runtime.Spec.Fuse.Resources, value)

	// 4. volumes
	err = t.transformFuseVolumes(runtime.Spec.Volumes, runtime.Spec.Fuse.VolumeMounts, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes for fuse")
	}

	// 5. nodeSelector
	value.Fuse.NodeSelector = map[string]string{}
	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	}
	value.Fuse.NodeSelector[utils.GetFuseLabelName(runtime.Namespace, runtime.Name, t.runtimeInfo.GetOwnerDatasetUID())] = "true"

	// 6. ports
	if len(runtime.Spec.Fuse.Ports) != 0 {
		value.Fuse.Ports = append(value.Fuse.Ports, runtime.Spec.Fuse.Ports...)
	}

	// 7. probe
	if runtime.Spec.Fuse.ReadinessProbe != nil {
		value.Fuse.ReadinessProbe = runtime.Spec.Fuse.ReadinessProbe
	}
	if runtime.Spec.Fuse.LivenessProbe != nil {
		value.Fuse.LivenessProbe = runtime.Spec.Fuse.LivenessProbe
	}
	// 8. command
	if len(runtime.Spec.Fuse.Command) != 0 {
		value.Fuse.Command = runtime.Spec.Fuse.Command
	}
	// 9. args
	if len(runtime.Spec.Fuse.Args) != 0 {
		value.Fuse.Args = runtime.Spec.Fuse.Args
	}

	// 10. network
	if runtime.Spec.Fuse.NetworkMode != "" {
		value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode)
	}
	value.Fuse.HostPID = common.HostPIDEnabled(runtime.Annotations)

	// 11. targetPath to mount
	value.Fuse.TargetPath = t.getTargetPath()

	// 12. lifecycle
	if err := t.parseLifecycle(runtime, profile, value); err != nil {
		return err
	}

	// 13. env
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

	// 14. fuse config
	err = t.transformFuseConfig(runtime, dataset, value)
	if err != nil {
		return err
	}

	// 15. critical
	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// 16. cachedir
	if len(runtime.Spec.TieredStore.Levels) > 0 {
		value.Fuse.CacheDir = runtime.Spec.TieredStore.Levels[0].Path
	}

	// 17. mount related node publish secret to fuse if the dataset specifies any mountpoint with pvc type.
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
	if len(runtime.Spec.Fuse.ImagePullSecrets) != 0 {
		value.Fuse.ImagePullSecrets = runtime.Spec.Fuse.ImagePullSecrets
	}
}

func (t *ThinEngine) parseLifecycle(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile, value *ThinValue) error {
	// default lifecycle config
	value.Fuse.Lifecycle = &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"umount",
					t.getTargetPath(),
				},
			},
		},
	}

	// set lifecycle from profile
	if fuseLifecycleInProfile := profile.Spec.Fuse.Lifecycle; fuseLifecycleInProfile != nil {
		if fuseLifecycleInProfile.PostStart != nil {
			return fmt.Errorf("custom fuse.lifecycle.postStart configuration in thinruntimeprofile is not supported")
		}
		if fuseLifecycleInProfile.PreStop != nil {
			value.Fuse.Lifecycle.PreStop = fuseLifecycleInProfile.PreStop
		}

	}

	// set lifecycle from runtime
	if fuseLifecycleInRuntime := runtime.Spec.Fuse.Lifecycle; fuseLifecycleInRuntime != nil {
		if fuseLifecycleInRuntime.PostStart != nil {
			return fmt.Errorf("custom fuse.lifecycle.postStart configuration in thinruntime is not supported")
		}
		if fuseLifecycleInRuntime.PreStop != nil {
			value.Fuse.Lifecycle.PreStop = fuseLifecycleInRuntime.PreStop
		}
	}
	return nil
}

func (t *ThinEngine) parseHostVolumeFromDataset(dataset *datav1alpha1.Dataset, thinValue *ThinValue) error {
	index := 0
	for _, mount := range dataset.Spec.Mounts {
		for key, value := range mount.Options {
			if key != common.DatasetOptionFluidFuseHostVolume {
				continue
			}

			var hostPath, mountPath string
			subStrings := strings.Split(value, ":")
			if len(subStrings) > 2 {
				return fmt.Errorf("invalid dataset option, %s: %s", key, value)
			}

			hostPath = subStrings[0]
			if !filepath.IsAbs(hostPath) {
				return fmt.Errorf("invalid dataset option, %s: %s, should be a absolute hostPath", key, value)
			}

			mountPath = hostPath
			if len(subStrings) == 2 {
				mountPath = subStrings[1]
				if !filepath.IsAbs(mountPath) {
					return fmt.Errorf("invalid dataset options, %s: %s, should be a absolute mountPath", key, value)
				}

				for _, volumeMount := range thinValue.Fuse.VolumeMounts {
					if volumeMount.MountPath == mountPath {
						return fmt.Errorf("invalid dataset options, %s: %s, mountPath %s is conflicted", key, value, volumeMount.MountPath)
					}
				}
			}

			readOnly := false
			for _, mode := range dataset.Spec.AccessModes {
				if mode == corev1.ReadOnlyMany {
					readOnly = true
					break
				}
			}

			volume, volumeMount := createVolumeAndMount(index, hostPath, mountPath, readOnly)
			thinValue.Fuse.Volumes = append(thinValue.Fuse.Volumes, volume)
			thinValue.Fuse.VolumeMounts = append(thinValue.Fuse.VolumeMounts, volumeMount)
			index++
			break
		}
	}

	return nil
}

func createVolumeAndMount(index int, hostPath, mountPath string, readOnly bool) (corev1.Volume, corev1.VolumeMount) {
	volumeName := fmt.Sprintf("fluid-fuse-hostvolume-%v", index)
	return corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: hostPath,
				},
			},
		}, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  readOnly,
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

func (t *ThinEngine) transformFusePodMetadata(podMetadata datav1alpha1.PodMetadata, value *ThinValue) {
	value.Fuse.Labels = utils.UnionMapsWithOverride(value.Fuse.Labels, podMetadata.Labels)
	value.Fuse.Annotations = utils.UnionMapsWithOverride(value.Fuse.Annotations, podMetadata.Annotations)
}

func (t *ThinEngine) parseFromProfileFuse(profile *datav1alpha1.ThinRuntimeProfile, value *ThinValue) {
	if profile == nil {
		return
	}
	// 1. image
	value.Fuse.Image = profile.Spec.Fuse.Image
	value.Fuse.ImageTag = profile.Spec.Fuse.ImageTag
	value.Fuse.ImagePullPolicy = profile.Spec.Fuse.ImagePullPolicy
	value.Fuse.ImagePullSecrets = profile.Spec.Fuse.ImagePullSecrets
	if len(profile.Spec.Fuse.Image) != 0 {
		value.Fuse.Image = profile.Spec.Fuse.Image
	}
	if len(profile.Spec.Fuse.ImageTag) != 0 {
		value.Fuse.ImageTag = profile.Spec.Fuse.ImageTag
	}
	if len(profile.Spec.Fuse.ImagePullPolicy) != 0 {
		value.Fuse.ImagePullPolicy = profile.Spec.Fuse.ImagePullPolicy
	}

	// 2. labels & annotations
	t.transformFusePodMetadata(profile.Spec.Fuse.PodMetadata, value)

	// 3. resources
	t.transformResourcesForFuse(profile.Spec.Fuse.Resources, value)

	// 4. volumes
	err := t.transformFuseVolumes(profile.Spec.Volumes, profile.Spec.Fuse.VolumeMounts, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes from profile for worker")
	}

	// 5. nodeSelector
	if len(profile.Spec.Fuse.NodeSelector) != 0 {
		value.Fuse.NodeSelector = profile.Spec.Fuse.NodeSelector
	}
	// 6. ports
	if len(profile.Spec.Fuse.Ports) != 0 {
		value.Fuse.Ports = profile.Spec.Fuse.Ports
	}
	// 7. probe
	value.Fuse.ReadinessProbe = profile.Spec.Fuse.ReadinessProbe
	value.Fuse.LivenessProbe = profile.Spec.Fuse.LivenessProbe
	// 8. command
	if len(profile.Spec.Fuse.Command) != 0 {
		value.Fuse.Command = profile.Spec.Fuse.Command
	}
	// 9. args
	if len(profile.Spec.Fuse.Args) != 0 {
		value.Fuse.Args = profile.Spec.Fuse.Args
	}
	// 10. network
	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(profile.Spec.Fuse.NetworkMode)
	// 11. env
	if len(profile.Spec.Fuse.Env) != 0 {
		value.Fuse.Envs = profile.Spec.Fuse.Env
	}
}
