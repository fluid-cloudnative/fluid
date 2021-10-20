/*
Copyright 2021 The Fluid Authors.

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

package juicefs

import (
	"errors"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"strings"
)

func (j *JuiceFSEngine) transformFuse(runtime *datav1alpha1.JuiceFSRuntime, dataset *datav1alpha1.Dataset, value *JuiceFS) (err error) {
	value.Fuse = Fuse{}
	value.Fuse.Prepare = Prepare{}

	if len(dataset.Spec.Mounts) <= 0 {
		return errors.New("do not assign mount point")
	}
	mount := dataset.Spec.Mounts[0]

	source := ""
	for _, encryptOption := range mount.EncryptOptions {
		key := encryptOption.Name
		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
		secret, err := utils.GetSecret(j.Client, secretKeyRef.Name, j.namespace)
		if err != nil {
			j.Log.Info("can't get the secret",
				"namespace", j.namespace,
				"name", j.name,
				"secretName", secretKeyRef.Name)
			return err
		}

		switch key {
		case "name":
			value.Fuse.Prepare.NameSecret = secretKeyRef.Name
		case "metaurl":
			value.Fuse.Prepare.MetaUrlSecret = secretKeyRef.Name
			v, ok := secret.Data[secretKeyRef.Key]
			if !ok {
				return errors.New(fmt.Sprintf("can't get metaurl from secret %s", secret.Name))
			}
			source = string(v)
		case "access-key":
			value.Fuse.Prepare.AccessKeySecret = secretKeyRef.Name
		case "secret-key":
			value.Fuse.Prepare.SecretKeySecret = secretKeyRef.Name
		case "storage":
			value.Fuse.Prepare.StorageSecret = secretKeyRef.Name
		case "bucket":
			value.Fuse.Prepare.BucketSecret = secretKeyRef.Name
		}
	}

	if value.Fuse.Prepare.NameSecret == "" {
		return errors.New("can't get name in secret")
	}
	if source == "" {
		return errors.New("can't get metaurl in secret")
	}
	if !strings.Contains(source, "://") {
		source = "redis://" + source
	}
	value.Fuse.MetaUrl = source

	image := runtime.Spec.Fuse.Image
	tag := runtime.Spec.Fuse.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.ImagePullPolicy

	subPath := ""
	if mount.Path != "" {
		subPath = mount.Path
	} else {
		subPath = mount.Name
	}
	value.Fuse.Image, value.Fuse.ImageTag, value.ImagePullPolicy = j.parseFuseImage(image, tag, imagePullPolicy)
	value.Fuse.MountPath = j.getMountPoint()
	value.Fuse.NodeSelector = map[string]string{}
	value.Fuse.HostMountPath = j.getMountPoint()
	value.Fuse.Prepare.SubPath = subPath

	mountArgs := []string{common.JuiceFSMountPath, source, value.Fuse.MountPath}
	options := []string{"metrics=0.0.0.0:9567"}
	for k, v := range mount.Options {
		options = append(options, fmt.Sprintf("%s=%s", k, v))
	}
	options = append(options, fmt.Sprintf("subdir=%s", subPath))
	if len(runtime.Spec.TieredStore.Levels) > 0 {
		cacheDir := runtime.Spec.TieredStore.Levels[0].Path
		if runtime.Spec.TieredStore.Levels[0].MediumType == common.Memory {
			cacheDir = "memory"
		}
		cacheSize := runtime.Spec.TieredStore.Levels[0].Quota.String()
		cacheRatio := runtime.Spec.TieredStore.Levels[0].Low
		options = append(options, fmt.Sprintf("cache-dir=%s", cacheDir))
		options = append(options, fmt.Sprintf("cache-size=%s", cacheSize))
		options = append(options, fmt.Sprintf("free-space-ratio=%s", cacheRatio))
	}

	mountArgs = append(mountArgs, "-o", strings.Join(options, ","))

	value.Fuse.Command = strings.Join(mountArgs, " ")
	value.Fuse.StatCmd = "stat -c %i " + value.Fuse.MountPath

	if runtime.Spec.Fuse.Global {
		if len(runtime.Spec.Fuse.NodeSelector) > 0 {
			value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
		}
		value.Fuse.NodeSelector[common.FluidFuseBalloonKey] = common.FluidBalloonValue
		j.Log.Info("Enable Fuse's global mode")
	} else {
		labelName := j.getCommonLabelName()
		value.Fuse.NodeSelector[labelName] = "true"
		j.Log.Info("Disable Fuse's global mode")
	}

	value.Fuse.Enabled = true

	j.transformResourcesForFuse(runtime, value)
	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	return
}
