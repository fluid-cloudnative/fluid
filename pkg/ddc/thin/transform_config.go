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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	EnvFuseConfigStorage = "THIN_FUSE_CONFIG_STORAGE"
)

func getFuseConfigStorage() string {
	if envVal, exists := os.LookupEnv(EnvFuseConfigStorage); exists {
		return envVal
	}
	// default value
	return "configmap"
}

func (t *ThinEngine) transformFuseConfig(runtime *datav1alpha1.ThinRuntime, dataset *datav1alpha1.Dataset, value *ThinValue) error {
	fuseConfigStorage := getFuseConfigStorage()

	// For cases like dynamic mount where dataset starts without any mount, we still need mounting secret volumes into fuse containers.
	if len(dataset.Spec.Mounts) == 0 {
		_ = t.transformEncryptOptionsWithSecretVolumes(datav1alpha1.Mount{}, dataset.Spec.SharedEncryptOptions, value)
	}

	mounts := []datav1alpha1.Mount{}
	// todo: support passing flexVolume info
	pvAttributes := map[string]*corev1.CSIPersistentVolumeSource{}
	pvMountOptions := map[string][]string{}
	for _, m := range dataset.Spec.Mounts {
		if strings.HasPrefix(m.MountPoint, common.VolumeScheme.String()) {
			pvcName := strings.TrimPrefix(m.MountPoint, common.VolumeScheme.String())
			csiInfo, mountOptions, err := t.extractVolumeInfo(pvcName)
			if err != nil {
				return errors.Wrapf(err, "failed to extract volume info from PersistentVolumeClaim \"%s\"", pvcName)
			}

			pvAttributes[pvcName] = csiInfo
			pvMountOptions[pvcName] = mountOptions
		}

		options, err := t.extractMountOptions(m, dataset, fuseConfigStorage, value)
		if err != nil {
			return err
		}
		m.Options = options
		m.EncryptOptions = nil
		mounts = append(mounts, m)
	}

	config := Config{}
	config.Mounts = mounts
	config.RuntimeOptions = runtime.Spec.Fuse.Options
	config.TargetPath = t.getTargetPath()
	config.PersistentVolumeAttrs = pvAttributes
	config.PersistentVolumeMountOptions = pvMountOptions
	config.AccessModes = dataset.Spec.AccessModes

	if len(config.AccessModes) == 0 {
		config.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadOnlyMany}
	}

	var configStr []byte
	configStr, err := json.Marshal(config)
	if err != nil {
		return errors.Wrapf(err, "failed to dump fuse config to json, runtime: \"%s/%s\"", runtime.Namespace, runtime.Name)
	}
	value.Fuse.ConfigValue = string(configStr)
	value.Fuse.ConfigStorage = fuseConfigStorage
	return nil
}

func (t *ThinEngine) transformEncryptOptionsWithSecretVolumes(m datav1alpha1.Mount, sharedEncryptOptions []datav1alpha1.EncryptOption, value *ThinValue) (options map[string]string) {
	options = make(map[string]string)
	for _, encryptOpt := range append(sharedEncryptOptions, m.EncryptOptions...) {
		secretName := encryptOpt.ValueFrom.SecretKeyRef.Name
		secretMountPath := fmt.Sprintf("/etc/fluid/secrets/%s", secretName)
		volName := fmt.Sprintf("thin-fuseconfig-%s", secretName)

		volumeToAdd := corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		}
		value.Fuse.Volumes = utils.AppendOrOverrideVolume(value.Fuse.Volumes, volumeToAdd)

		volumeMountToAdd := corev1.VolumeMount{
			Name:      volName,
			ReadOnly:  true,
			MountPath: secretMountPath,
		}

		value.Fuse.VolumeMounts = utils.AppendOrOverrideVolumeMounts(value.Fuse.VolumeMounts, volumeMountToAdd)
		options[encryptOpt.Name] = filepath.Join(secretMountPath, encryptOpt.ValueFrom.SecretKeyRef.Key)
	}

	return options
}

func (t *ThinEngine) extractMountOptions(m datav1alpha1.Mount, dataset *datav1alpha1.Dataset, fuseConfigStorage string, value *ThinValue) (options map[string]string, err error) {
	switch strings.ToLower(fuseConfigStorage) {
	case "configmap":
		options, err = t.genFuseMountOptions(m, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions, false)
		if err != nil {
			return options, errors.Wrap(err, "failed to generate FUSE mount options from dataset mount info")
		}
		transformedEncryptOpts := t.transformEncryptOptionsWithSecretVolumes(m, dataset.Spec.SharedEncryptOptions, value)
		for k, v := range transformedEncryptOpts {
			options[k] = v
		}
	case "secret":
		options, err = t.genFuseMountOptions(m, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions, true)
		if err != nil {
			return options, errors.Wrap(err, "failed to generate FUSE mount options from dataset mount info")
		}
	default:
		return options, fmt.Errorf("FUSE config storage \"%s\" is not supported, valid value: \"configmap\" or \"secret\"", fuseConfigStorage)
	}

	return options, nil
}

func (t *ThinEngine) extractVolumeInfo(pvcName string) (csiInfo *corev1.CSIPersistentVolumeSource, mountOptions []string, err error) {
	pvc, err := kubeclient.GetPersistentVolumeClaim(t.Client, pvcName, t.namespace)
	if err != nil {
		return
	}

	if len(pvc.Spec.VolumeName) == 0 || pvc.Status.Phase != corev1.ClaimBound {
		err = fmt.Errorf("persistent volume claim %s not bounded yet", pvcName)
		return
	}

	pv, err := kubeclient.GetPersistentVolume(t.Client, pvc.Spec.VolumeName)
	if err != nil {
		return
	}

	if pv.Spec.CSI != nil {
		csiInfo = pv.Spec.CSI
	}

	mountOptions, err = t.extractVolumeMountOptions(pv)
	if err != nil {
		return
	}

	return
}

func (t *ThinEngine) extractVolumeMountOptions(pv *corev1.PersistentVolume) (mountOptions []string, err error) {
	if len(pv.Spec.MountOptions) != 0 {
		return pv.Spec.MountOptions, nil
	}

	// fallback to check "volume.beta.kubernetes.io/mount-options", see https://kubernetes.io/docs/concepts/storage/persistent-volumes/#mount-options
	// e.g. volume.beta.kubernetes.io/mount-options: rw,nfsvers=4,noexec
	if opts, exists := pv.Annotations[corev1.MountOptionAnnotation]; exists {
		return strings.Split(opts, ","), nil
	}

	return
}
