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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (j *JuiceFSEngine) transformFuse(runtime *datav1alpha1.JuiceFSRuntime, dataset *datav1alpha1.Dataset, value *JuiceFS) (err error) {
	if len(dataset.Spec.Mounts) <= 0 {
		return errors.New("do not assign mount point")
	}
	mount := dataset.Spec.Mounts[0]

	value.Configs.Name = mount.Name

	// transform image
	image := runtime.Spec.Fuse.Image
	tag := runtime.Spec.Fuse.ImageTag
	imagePullPolicy := runtime.Spec.Fuse.ImagePullPolicy
	value.Fuse.Image, value.Fuse.ImageTag, value.Fuse.ImagePullPolicy = j.parseFuseImage(image, tag, imagePullPolicy)

	// transform envs
	value.Fuse.Envs = runtime.Spec.Fuse.Env

	// transform options
	var tiredStoreLevel *datav1alpha1.Level
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		tiredStoreLevel = &runtime.Spec.TieredStore.Levels[0]
	}
	option, err := j.genValue(mount, tiredStoreLevel, value)
	if err != nil {
		return err
	}

	// transform format cmd
	j.genFormatCmd(value, runtime.Spec.Configs)

	// transform mount cmd & stat cmd
	err = j.genMount(value, runtime, option)
	if err != nil {
		return err
	}

	// transform nodeSelector
	j.transformFuseNodeSelector(runtime, value)
	value.Fuse.Enabled = true

	// transform resource
	err = j.transformResourcesForFuse(runtime, value)
	if err != nil {
		return err
	}
	// transform volumes for fuse
	err = j.transformFuseVolumes(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform volumes for fuse")
		return err
	}

	// set critical fuse pod to avoid eviction
	value.Fuse.CriticalPod = common.CriticalFusePodEnabled()

	// parse fuse container network mode
	value.Fuse.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode)
	return
}

func (j *JuiceFSEngine) transformFuseNodeSelector(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) {
	value.Fuse.NodeSelector = map[string]string{}
	if len(runtime.Spec.Fuse.NodeSelector) > 0 {
		value.Fuse.NodeSelector = runtime.Spec.Fuse.NodeSelector
	}

	// The label will be added by CSI Plugin when any workload pod is scheduled on the node.
	value.Fuse.NodeSelector[j.getFuseLabelName()] = "true"
}

func (j *JuiceFSEngine) genValue(mount datav1alpha1.Mount, tiredStoreLevel *datav1alpha1.Level, value *JuiceFS) (map[string]string, error) {
	value.Configs.Name = mount.Name
	options := make(map[string]string)
	source := ""
	value.Edition = EnterpriseEdition
	for k, v := range mount.Options {
		switch k {
		case JuiceStorage:
			value.Configs.Storage = v
			continue
		case JuiceBucket:
			value.Configs.Bucket = v
			continue
		default:
			options[k] = v
		}
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
			value.Configs.MetaUrlSecret = secretKeyRef.Name
			_, ok := secret.Data[secretKeyRef.Key]
			if !ok {
				return nil, fmt.Errorf("can't get metaurl from secret %s", secret.Name)
			}
			value.Edition = CommunityEdition
		case JuiceAccessKey:
			value.Configs.AccessKeySecret = secretKeyRef.Name
		case JuiceSecretKey:
			value.Configs.SecretKeySecret = secretKeyRef.Name
		case JuiceToken:
			value.Configs.TokenSecret = secretKeyRef.Name
		}
	}

	if source == "" {
		source = mount.Name
	}

	// transform source
	value.Source = source

	// transform mountPath & subPath
	subPath, err := ParseSubPathFromMountPoint(mount.MountPoint)
	if err != nil {
		return nil, err
	}
	value.Fuse.MountPath = j.getMountPoint()
	value.Worker.MountPath = j.getMountPoint()
	value.Fuse.HostMountPath = j.getHostMountPoint()
	if subPath != "/" {
		value.Fuse.SubPath = subPath
		options["subdir"] = subPath
	}

	var storagePath = DefaultCacheDir
	var volumeType = common.VolumeTypeHostPath
	if tiredStoreLevel != nil {
		storagePath = tiredStoreLevel.Path // /mnt/disk1/bigboot or /mnt/disk1/bigboot,/mnt/disk2/bigboot
		if tiredStoreLevel.Quota != nil {
			q := tiredStoreLevel.Quota
			// juicefs cache-size should be integer in MiB
			// community doc: https://juicefs.com/docs/community/command_reference/#juicefs-mount
			// enterprise doc: https://juicefs.com/docs/cloud/commands_reference#mount
			cacheSize := q.Value() >> 20
			options["cache-size"] = strconv.FormatInt(cacheSize, 10)
		}
		if tiredStoreLevel.Low != "" {
			options["free-space-ratio"] = tiredStoreLevel.Low
		}
		volumeType = tiredStoreLevel.VolumeType
	}
	originPath := strings.Split(storagePath, ",")
	options["cache-dir"] = storagePath

	// transform cacheDir
	value.CacheDirs = make(map[string]cache)
	for i, v := range originPath {
		value.CacheDirs[strconv.Itoa(i+1)] = cache{
			Path: v,
			Type: string(volumeType),
		}
	}

	return options, nil
}

func (j *JuiceFSEngine) genMount(value *JuiceFS, runtime *datav1alpha1.JuiceFSRuntime, optionMap map[string]string) (err error) {
	var mountArgs, mountArgsWorker []string
	workerOptionMap := make(map[string]string)
	if optionMap == nil {
		optionMap = map[string]string{}
	}
	readonly := false
	runtimeInfo := j.runtimeInfo
	if runtimeInfo != nil {
		accessModes, err := utils.GetAccessModesOfDataset(j.Client, runtimeInfo.GetName(), runtimeInfo.GetNamespace())
		if err != nil {
			j.Log.Info("Error:", "err", err)
		}
		if len(accessModes) > 0 {
			for _, mode := range accessModes {
				if mode == corev1.ReadOnlyMany {
					optionMap["ro"] = ""
					readonly = true
					break
				}
			}
		}
	}
	// gen worker option
	for k, v := range optionMap {
		workerOptionMap[k] = v
	}
	if runtime != nil {
		// if runtime.worker option is set, take it
		for k, v := range runtime.Spec.Worker.Options {
			workerOptionMap[k] = v
		}
	}
	if value.Edition == CommunityEdition {
		if readonly {
			optionMap["attr-cache"] = "7200"
			optionMap["entry-cache"] = "7200"
		}
		if _, ok := optionMap["metrics"]; !ok {
			optionMap["metrics"] = "0.0.0.0:9567"
		}
		if _, ok := workerOptionMap["metrics"]; !ok {
			workerOptionMap["metrics"] = "0.0.0.0:9567"
		}
		mountArgs = []string{common.JuiceFSCeMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(genOption(optionMap), ",")}
		mountArgsWorker = []string{common.JuiceFSCeMountPath, value.Source, value.Worker.MountPath, "-o", strings.Join(genOption(workerOptionMap), ",")}
	} else {
		if readonly {
			optionMap["attrcacheto"] = "7200"
			optionMap["entrycacheto"] = "7200"
		}
		optionMap["foreground"] = ""
		workerOptionMap["foreground"] = ""

		// start independent cache cluster, refer to [juicefs cache sharing](https://juicefs.com/docs/cloud/cache/#client_cache_sharing)
		// fuse and worker use the same cache-group, fuse use no-sharing
		cacheGroup := fmt.Sprintf("%s-%s", j.namespace, value.FullnameOverride)
		if _, ok := optionMap["cache-group"]; ok {
			cacheGroup = optionMap["cache-group"]
		}
		optionMap["cache-group"] = cacheGroup
		workerOptionMap["cache-group"] = cacheGroup
		optionMap["no-sharing"] = ""
		delete(workerOptionMap, "no-sharing")

		mountArgs = []string{common.JuiceFSMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(genOption(optionMap), ",")}
		mountArgsWorker = []string{common.JuiceFSMountPath, value.Source, value.Worker.MountPath, "-o", strings.Join(genOption(workerOptionMap), ",")}
	}

	value.Worker.Command = strings.Join(mountArgsWorker, " ")
	value.Fuse.Command = strings.Join(mountArgs, " ")
	value.Fuse.StatCmd = "stat -c %i " + value.Fuse.MountPath
	value.Worker.StatCmd = "stat -c %i " + value.Worker.MountPath
	return nil
}

func genOption(optionMap map[string]string) []string {
	options := []string{}
	for k, v := range optionMap {
		if v != "" {
			k = fmt.Sprintf("%s=%s", k, v)
		}
		options = append(options, k)
	}
	return options
}

func (j *JuiceFSEngine) genFormatCmd(value *JuiceFS, config *[]string) {
	args := make([]string, 0)
	if config != nil {
		for _, option := range *config {
			o := strings.TrimSpace(option)
			if o != "" {
				args = append(args, fmt.Sprintf("--%s", o))
			}
		}
	}
	if value.Edition == CommunityEdition {
		// ce
		if value.Configs.AccessKeySecret != "" {
			args = append(args, "--access-key=${ACCESS_KEY}")
		}
		if value.Configs.SecretKeySecret != "" {
			args = append(args, "--secret-key=${SECRET_KEY}")
		}
		if value.Configs.Storage == "" || value.Configs.Bucket == "" {
			args = append(args, "--no-update")
		}
		if value.Configs.Storage != "" {
			args = append(args, fmt.Sprintf("--storage=%s", value.Configs.Storage))
		}
		if value.Configs.Bucket != "" {
			args = append(args, fmt.Sprintf("--bucket=%s", value.Configs.Bucket))
		}
		args = append(args, value.Source, value.Configs.Name)
		cmd := append([]string{common.JuiceCeCliPath, "format"}, args...)
		value.Configs.FormatCmd = strings.Join(cmd, " ")
		return
	}
	// ee
	if value.Configs.TokenSecret == "" {
		// skip juicefs auth
		return
	}
	args = append(args, "--token=${TOKEN}")
	if value.Configs.AccessKeySecret != "" {
		args = append(args, "--accesskey=${ACCESS_KEY}")
	}
	if value.Configs.SecretKeySecret != "" {
		args = append(args, "--secretkey=${SECRET_KEY}")
	}
	if value.Configs.Bucket != "" {
		args = append(args, fmt.Sprintf("--bucket=%s", value.Configs.Bucket))
	}
	args = append(args, value.Source)
	cmd := append([]string{common.JuiceCliPath, "auth"}, args...)
	value.Configs.FormatCmd = strings.Join(cmd, " ")
}
