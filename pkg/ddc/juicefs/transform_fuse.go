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
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"strings"
)

func (j *JuiceFSEngine) transformFuse(runtime *datav1alpha1.JuiceFSRuntime, dataset *datav1alpha1.Dataset, value *JuiceFS) (err error) {
	value.Fuse = Fuse{}

	if len(dataset.Spec.Mounts) <= 0 {
		return errors.New("do not assign mount point")
	}
	mount := dataset.Spec.Mounts[0]

	value.Fuse.Name = mount.Name

	image := runtime.Spec.Fuse.Image
	tag := runtime.Spec.Fuse.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.ImagePullPolicy

	value.Fuse.Image, value.Fuse.ImageTag, value.ImagePullPolicy = j.parseFuseImage(image, tag, imagePullPolicy)
	value.Fuse.NodeSelector = map[string]string{}
	value.Fuse.Envs = runtime.Spec.Fuse.Env

	var tiredStoreLevel *datav1alpha1.Level
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		tiredStoreLevel = &runtime.Spec.TieredStore.Levels[0]
	}
	option, err := j.genValue(mount, tiredStoreLevel, value)
	if err != nil {
		return err
	}
	j.genFormatCmd(value)
	err = j.genMount(value, option)
	if err != nil {
		return err
	}

	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	} else {
		value.Fuse.NodeSelector = map[string]string{}
	}
	value.Fuse.NodeSelector[j.getFuseLabelName()] = "true"
	value.Fuse.Enabled = true

	j.transformResourcesForFuse(runtime, value)
	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	return
}

func (j *JuiceFSEngine) genValue(mount datav1alpha1.Mount, tiredStoreLevel *datav1alpha1.Level, value *JuiceFS) ([]string, error) {
	value.Fuse.Name = mount.Name
	opts := make(map[string]string)
	source := ""
	for k, v := range mount.Options {
		switch k {
		case JuiceStorage:
			value.Fuse.Storage = v
			continue
		case JuiceBucket:
			value.Fuse.Bucket = v
			continue
		default:
			opts[k] = v
		}
	}
	options := []string{}
	for k, v := range opts {
		options = append(options, fmt.Sprintf("%s=%s", k, v))
	}
	for _, encryptOption := range mount.EncryptOptions {
		key := encryptOption.Name
		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(j.Client, secretKeyRef.Name, j.namespace)
		if err != nil {
			j.Log.Info("can't get the secret",
				"namespace", j.namespace,
				"name", j.name,
				"secretName", secretKeyRef.Name)
			return nil, err
		}

		switch key {
		case JuiceMetaUrl:
			source = "${METAURL}"
			value.Fuse.MetaUrlSecret = secretKeyRef.Name
			_, ok := secret.Data[secretKeyRef.Key]
			if !ok {
				return nil, fmt.Errorf("can't get metaurl from secret %s", secret.Name)
			}
			value.IsCE = true
		case JuiceAccessKey:
			value.Fuse.AccessKeySecret = secretKeyRef.Name
		case JuiceSecretKey:
			value.Fuse.SecretKeySecret = secretKeyRef.Name
		case JuiceToken:
			value.Fuse.TokenSecret = secretKeyRef.Name
		}
	}

	if source == "" {
		source = mount.Name
	}

	value.Source = source
	subPath, err := ParseSubPathFromMountPoint(mount.MountPoint)
	if err != nil {
		return nil, err
	}
	value.Fuse.MountPath = j.getMountPoint()
	value.Fuse.HostMountPath = j.getHostMountPoint()
	value.Fuse.SubPath = subPath
	options = append(options, fmt.Sprintf("subdir=%s", subPath))

	var cacheDir = DefaultCacheDir
	if tiredStoreLevel != nil {
		cacheDir = tiredStoreLevel.Path
		if tiredStoreLevel.MediumType == common.Memory {
			cacheDir = "memory"
		}
		if tiredStoreLevel.Quota != nil {
			options = append(options, fmt.Sprintf("cache-size=%s", tiredStoreLevel.Quota.String()))
		}
		if tiredStoreLevel.Low != "" {
			options = append(options, fmt.Sprintf("free-space-ratio=%s", tiredStoreLevel.Low))
		}
	}
	value.Fuse.CacheDir = cacheDir
	options = append(options, fmt.Sprintf("cache-dir=%s", cacheDir))

	return options, nil
}

func (j *JuiceFSEngine) genMount(value *JuiceFS, options []string) (err error) {
	var mountArgs, mountArgsWorker []string
	if options == nil {
		options = []string{}
	}
	if value.IsCE {
		if !utils.ContainsSubString(options, "metrics") {
			options = append(options, "metrics=0.0.0.0:9567")
		}
		mountArgs = []string{common.JuiceFSCeMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(options, ",")}
		mountArgsWorker = []string{common.JuiceFSCeMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(options, ",")}
	} else {
		if !utils.ContainsString(options, "foreground") {
			options = append(options, "foreground")
		}
		fuseOption := make([]string, len(options))
		copy(fuseOption, options)
		if !utils.ContainsSubString(options, "cache-group") {
			// start independent cache cluster, refer to [juicefs cache sharing](https://juicefs.com/docs/cloud/cache/#client_cache_sharing)
			// fuse and worker use the same cache-group, fuse use no-sharing
			options = append(options, fmt.Sprintf("cache-group=%s", value.FullnameOverride))
			fuseOption = append(fuseOption, fmt.Sprintf("cache-group=%s", value.FullnameOverride))
			fuseOption = append(fuseOption, "no-sharing")
		}
		mountArgs = []string{common.JuiceFSMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(fuseOption, ",")}
		mountArgsWorker = []string{common.JuiceFSMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(options, ",")}
	}

	value.Worker.Command = strings.Join(mountArgsWorker, " ")
	value.Fuse.Command = strings.Join(mountArgs, " ")
	value.Fuse.StatCmd = "stat -c %i " + value.Fuse.MountPath
	return nil
}

func (j *JuiceFSEngine) genFormatCmd(value *JuiceFS) {
	args := make([]string, 0)
	if value.IsCE {
		// ce
		if value.Fuse.AccessKeySecret != "" {
			args = append(args, "--access-key=${ACCESS_KEY}")
		}
		if value.Fuse.SecretKeySecret != "" {
			args = append(args, "--secret-key=${SECRET_KEY}")
		}
		if value.Fuse.Storage == "" || value.Fuse.Bucket == "" {
			args = append(args, "--no-update")
		}
		if value.Fuse.Storage != "" {
			args = append(args, fmt.Sprintf("--storage=%s", value.Fuse.Storage))
		}
		if value.Fuse.Bucket != "" {
			args = append(args, fmt.Sprintf("--bucket=%s", value.Fuse.Bucket))
		}
		args = append(args, value.Source, value.Fuse.Name)
		cmd := append([]string{common.JuiceCeCliPath, "format"}, args...)
		value.Fuse.FormatCmd = strings.Join(cmd, " ")
	} else {
		// ee
		if value.Fuse.TokenSecret == "" {
			// skip juicefs auth
			return
		}
		args = append(args, "--token=${TOKEN}")
		if value.Fuse.AccessKeySecret != "" {
			args = append(args, "--accesskey=${ACCESS_KEY}")
		}
		if value.Fuse.SecretKeySecret != "" {
			args = append(args, "--secretkey=${SECRET_KEY}")
		}
		if value.Fuse.Bucket != "" {
			args = append(args, fmt.Sprintf("--bucket=%s", value.Fuse.Bucket))
		}
		args = append(args, value.Source)
		cmd := append([]string{common.JuiceCliPath, "auth"}, args...)
		value.Fuse.FormatCmd = strings.Join(cmd, " ")
	}
}
