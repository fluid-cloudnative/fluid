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
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
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
	value.Fuse.Image, value.Fuse.ImageTag, value.Fuse.ImagePullPolicy, err = j.parseJuiceFSImage(value.Edition, image, tag, imagePullPolicy)
	if err != nil {
		return
	}

	// transform envs
	value.Fuse.Envs = runtime.Spec.Fuse.Env

	// transform options
	var tiredStoreLevel *datav1alpha1.Level
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		tiredStoreLevel = &runtime.Spec.TieredStore.Levels[0]
	}
	optionsFromDataset, err := j.genValue(mount, tiredStoreLevel, value, dataset.Spec.SharedOptions, dataset.Spec.SharedEncryptOptions)
	if err != nil {
		return err
	}

	// transform format cmd
	j.genFormatCmd(value, runtime.Spec.Configs, optionsFromDataset)

	// transform quota cmd
	err = j.genQuotaCmd(value, mount)
	if err != nil {
		return err
	}

	// transform mount cmd & stat cmd
	options, err := j.genMountOptions(mount, tiredStoreLevel)
	if err != nil {
		return err
	}

	// Keep mount options in dataset still work, but it can be overwrited by fuse speicifed option
	options = utils.UnionMapsWithOverride(optionsFromDataset, options)
	for k, v := range runtime.Spec.Fuse.Options {
		options[k] = v
	}
	err = j.genFuseMount(value, options)
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
	// transform cache volumes for fuse
	err = j.transformFuseCacheVolumes(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform cache volumes for fuse")
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

// genValue: generate the value of juicefs
func (j *JuiceFSEngine) genValue(mount datav1alpha1.Mount, tiredStoreLevel *datav1alpha1.Level, value *JuiceFS,
	sharedOptions map[string]string, sharedEncryptOptions []datav1alpha1.EncryptOption) (map[string]string, error) {
	options := make(map[string]string)
	value.Configs.Name = mount.Name
	value.Configs.EncryptEnvOptions = make([]EncryptEnvOption, 0)
	source := ""

	for k, v := range sharedOptions {
		switch k {
		case JuiceStorage:
			value.Configs.Storage = v
			continue
		case JuiceBucket:
			value.Configs.Bucket = v
			continue
		default:
			if k != "quota" {
				options[k] = v
			}
		}
	}

	for k, v := range mount.Options {
		switch k {
		case JuiceStorage:
			value.Configs.Storage = v
			continue
		case JuiceBucket:
			value.Configs.Bucket = v
			continue
		default:
			if k != "quota" {
				options[k] = v
			}
		}
	}

	for _, encryptOption := range sharedEncryptOptions {
		key := encryptOption.Name
		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef

		switch key {
		case JuiceMetaUrl:
			source = "${METAURL}"
			value.Configs.MetaUrlSecret = secretKeyRef.Name
			value.Configs.MetaUrlSecretKey = secretKeyRef.Key
		case JuiceAccessKey:
			value.Configs.AccessKeySecret = secretKeyRef.Name
			value.Configs.AccessKeySecretKey = secretKeyRef.Key
		case JuiceSecretKey:
			value.Configs.SecretKeySecret = secretKeyRef.Name
			value.Configs.SecretKeySecretKey = secretKeyRef.Key
		case JuiceToken:
			value.Configs.TokenSecret = secretKeyRef.Name
			value.Configs.TokenSecretKey = secretKeyRef.Key
		default:
			if key != "quota" {
				envName := utils.ConvertDashToUnderscore(key)
				err := utils.CheckValidateEnvName(envName)
				if err != nil {
					return options, err
				}
				// options[key] = "${" + envName + "}"
				value.Configs.EncryptEnvOptions = append(value.Configs.EncryptEnvOptions,
					EncryptEnvOption{
						Name:             key,
						EnvName:          envName,
						SecretKeyRefName: secretKeyRef.Name,
						SecretKeyRefKey:  secretKeyRef.Key,
					})
			}
		}
	}

	for _, encryptOption := range mount.EncryptOptions {
		key := encryptOption.Name
		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef

		switch key {
		case JuiceMetaUrl:
			source = "${METAURL}"
			value.Configs.MetaUrlSecret = secretKeyRef.Name
			value.Configs.MetaUrlSecretKey = secretKeyRef.Key
		case JuiceAccessKey:
			value.Configs.AccessKeySecret = secretKeyRef.Name
			value.Configs.AccessKeySecretKey = secretKeyRef.Key
		case JuiceSecretKey:
			value.Configs.SecretKeySecret = secretKeyRef.Name
			value.Configs.SecretKeySecretKey = secretKeyRef.Key
		case JuiceToken:
			value.Configs.TokenSecret = secretKeyRef.Name
			value.Configs.TokenSecretKey = secretKeyRef.Key
		default:
			if key != "quota" {
				envName := utils.ConvertDashToUnderscore(key)
				err := utils.CheckValidateEnvName(envName)
				if err != nil {
					return options, err
				}
				// options[key] = "${" + envName + "}"
				value.Configs.EncryptEnvOptions = append(value.Configs.EncryptEnvOptions,
					EncryptEnvOption{
						Name:             key,
						EnvName:          envName,
						SecretKeyRefName: secretKeyRef.Name,
						SecretKeyRefKey:  secretKeyRef.Key,
					})
			}
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
		return options, err
	}
	value.Fuse.MountPath = j.getMountPoint()
	value.Worker.MountPath = j.getMountPoint()
	value.Fuse.HostMountPath = j.getHostMountPoint()
	if subPath != "/" {
		value.Fuse.SubPath = subPath
		// options["subdir"] = subPath
	}

	var storagePath = DefaultCacheDir
	var volumeType = common.VolumeTypeHostPath
	var volumeSource datav1alpha1.VolumeSource
	if tiredStoreLevel != nil {
		// juicefs cache-dir use colon (:) to separate multiple paths
		// community doc: https://juicefs.com/docs/community/command_reference/#juicefs-mount
		// enterprise doc: https://juicefs.com/docs/cloud/commands_reference#mount
		// /mnt/disk1/bigboot or /mnt/disk1/bigboot:/mnt/disk2/bigboot
		storagePath = tiredStoreLevel.Path
		volumeType = tiredStoreLevel.VolumeType
		volumeSource = tiredStoreLevel.VolumeSource
	}
	originPath := strings.Split(storagePath, ":")

	// transform cacheDir
	value.CacheDirs = make(map[string]cache)
	for i, v := range originPath {
		value.CacheDirs[strconv.Itoa(i+1)] = cache{
			Path:         v,
			Type:         string(volumeType),
			VolumeSource: &volumeSource,
		}
	}

	return options, nil
}

func (j *JuiceFSEngine) genMountOptions(mount datav1alpha1.Mount, tiredStoreLevel *datav1alpha1.Level) (options map[string]string, err error) {
	options = map[string]string{}
	var subPath string
	subPath, err = ParseSubPathFromMountPoint(mount.MountPoint)
	if subPath != "/" {
		options["subdir"] = subPath
	}

	var storagePath = DefaultCacheDir
	if tiredStoreLevel != nil {
		// juicefs cache-dir use colon (:) to separate multiple paths
		// community doc: https://juicefs.com/docs/community/command_reference/#juicefs-mount
		// enterprise doc: https://juicefs.com/docs/cloud/commands_reference#mount
		// /mnt/disk1/bigboot or /mnt/disk1/bigboot:/mnt/disk2/bigboot
		storagePath = tiredStoreLevel.Path
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
	}
	options["cache-dir"] = storagePath
	return
}

// genFuseMount: generate fuse mount args
func (j *JuiceFSEngine) genFuseMount(value *JuiceFS, optionMap map[string]string) (err error) {
	var mountArgs []string
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
	if value.Edition == CommunityEdition {
		if readonly {
			optionMap["attr-cache"] = "7200"
			optionMap["entry-cache"] = "7200"
		}

		// set metrics port
		if _, ok := optionMap["metrics"]; !ok {
			metricsPort := DefaultMetricsPort
			if value.Fuse.MetricsPort != nil {
				metricsPort = *value.Fuse.MetricsPort
			}
			optionMap["metrics"] = fmt.Sprintf("0.0.0.0:%d", metricsPort)
		}
		mountArgs = []string{common.JuiceFSCeMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(genArgs(optionMap), ",")}
	} else {
		if readonly {
			optionMap["attrcacheto"] = "7200"
			optionMap["entrycacheto"] = "7200"
		}
		optionMap["foreground"] = ""
		// do not update config again
		optionMap["no-update"] = ""

		// start independent cache cluster, refer to [juicefs cache sharing](https://juicefs.com/docs/cloud/cache/#client_cache_sharing)
		// fuse and worker use the same cache-group, fuse use no-sharing
		cacheGroup := fmt.Sprintf("%s-%s", j.namespace, value.FullnameOverride)
		if _, ok := optionMap["cache-group"]; ok {
			cacheGroup = optionMap["cache-group"]
		}
		optionMap["cache-group"] = cacheGroup
		optionMap["no-sharing"] = ""

		mountArgs = []string{common.JuiceFSMountPath, value.Source, value.Fuse.MountPath, "-o", strings.Join(genArgs(optionMap), ",")}
	}

	value.Fuse.Command = strings.Join(mountArgs, " ")
	value.Fuse.StatCmd = "stat -c %i " + value.Fuse.MountPath
	return nil
}

// genArgs: generate mount option as `a=b` format
func genArgs(optionMap map[string]string) []string {
	options := []string{}
	for k, v := range optionMap {
		if v != "" {
			k = fmt.Sprintf("%s=%s", k, v)
		}
		options = append(options, k)
	}
	return options
}

func (j *JuiceFSEngine) genFormatCmd(value *JuiceFS, config *[]string, options map[string]string) {

	var (
		// allow Options for CommunityEdition
		ceAllowOptions []string = []string{}
		// allow Options for CommunityEdition
		eeAllowOptions []string = []string{"bucket2", "access-key2", "secret-key2"}
	)
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
		formatOpts := filterOptionsWithKeys(options, ceAllowOptions)
		for k, v := range formatOpts {
			args = append(args, fmt.Sprintf("--%s=%s", k, v))
		}
		encryptOptions := filterEncryptEnvOptionsWithKeys(value.Configs.EncryptEnvOptions,
			ceAllowOptions)
		for _, v := range encryptOptions {
			args = append(args, fmt.Sprintf("--%s=${%s}", v.Name, v.EnvName))
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
	formatOpts := filterOptionsWithKeys(options, eeAllowOptions)
	for k, v := range formatOpts {
		args = append(args, fmt.Sprintf("--%s=%s", k, v))
	}
	encryptOptions := filterEncryptEnvOptionsWithKeys(value.Configs.EncryptEnvOptions,
		eeAllowOptions)
	for _, v := range encryptOptions {
		args = append(args, fmt.Sprintf("--%s=${%s}", v.Name, v.EnvName))
	}
	args = append(args, value.Source)
	cmd := append([]string{common.JuiceCliPath, "auth"}, args...)
	value.Configs.FormatCmd = strings.Join(cmd, " ")
}

// getQuota: get quota from string
func (j *JuiceFSEngine) getQuota(v string) (int64, error) {
	q, err := resource.ParseQuantity(v)
	if err != nil {
		return 0, fmt.Errorf("invalid quota %s: %v", v, err)
	}
	qs := q.Value() / 1024 / 1024 / 1024
	if qs <= 0 {
		return 0, fmt.Errorf("quota %s is too small, at least 1GiB for quota", v)
	}

	return qs, nil
}

// genQuotaCmd: generate command for set quota of subpath
func (j *JuiceFSEngine) genQuotaCmd(value *JuiceFS, mount datav1alpha1.Mount) error {
	options := mount.Options
	for k, v := range options {
		if k == "quota" {
			qs, err := j.getQuota(v)
			if err != nil {
				return errors.Wrapf(err, "invalid quota %s", v)
			}
			if value.Fuse.SubPath == "" {
				return fmt.Errorf("subPath must be set when quota is enabled")
			}
			if value.Edition == CommunityEdition {
				// ce
				// juicefs quota set ${metaurl} --path ${path} --capacity ${capacity}
				value.Configs.QuotaCmd = fmt.Sprintf("%s quota set %s --path %s --capacity %d", common.JuiceCeCliPath, value.Source, value.Fuse.SubPath, qs)
				return nil
			}
			// ee
			// juicefs quota set ${metaurl} --path ${path} --capacity ${capacity}
			cli := common.JuiceCliPath
			value.Configs.QuotaCmd = fmt.Sprintf("%s quota set %s --path %s --capacity %d", cli, value.Source, value.Fuse.SubPath, qs)
			return nil
		}
	}
	return nil
}

func ParseImageTag(imageTag string) (*ClientVersion, *ClientVersion, error) {
	if imageTag == common.NightlyTag {
		return &ClientVersion{0, 0, 0, common.NightlyTag}, &ClientVersion{0, 0, 0, common.NightlyTag}, nil
	}
	versions := strings.Split(imageTag, "-")
	if len(versions) < 2 {
		return nil, nil, fmt.Errorf("can not parse version from image tag: %s", imageTag)
	}

	ceVersion, err := parseVersion(versions[0])
	if err != nil {
		return nil, nil, err
	}
	eeVersion, err := parseVersion(versions[1])
	if err != nil {
		return nil, nil, err
	}
	return ceVersion, eeVersion, nil
}
