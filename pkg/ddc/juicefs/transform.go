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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base/portallocator"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/security"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transformer"
)

func (j *JuiceFSEngine) transform(runtime *datav1alpha1.JuiceFSRuntime) (value *JuiceFS, err error) {
	if runtime == nil {
		err = fmt.Errorf("the juicefsRuntime is null")
		return
	}
	defer utils.TimeTrack(time.Now(), "JuiceFSRuntime.Transform", "name", runtime.Name)

	dataset, err := utils.GetDataset(j.Client, j.name, j.namespace)
	if err != nil {
		return value, err
	}

	value = &JuiceFS{
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: runtime.Namespace,
			Name:      runtime.Name,
		},
	}

	// TODO: Handle cases that FullnameOverride is too long (> 63 chars)
	value.FullnameOverride = j.name
	value.Owner = transformer.GenerateOwnerReferenceFromObject(runtime)

	// transform toleration
	j.transformTolerations(dataset, value)

	value.Fuse = Fuse{
		Privileged: true,
	}
	value.Worker = Worker{
		Privileged: true,
	}

	// generate edition
	j.genEdition(dataset.Spec.Mounts[0], value, dataset.Spec.Mounts[0].EncryptOptions)

	// allocate ports
	err = j.allocatePorts(runtime, value)
	if err != nil {
		return
	}

	// transform the fuse
	err = j.transformFuse(runtime, dataset, value)
	if err != nil {
		return
	}

	// transform the workers
	err = j.transformWorkers(runtime, dataset, value)
	if err != nil {
		return
	}

	// transform runtime pod metadata
	err = j.transformPodMetadata(runtime, value)
	if err != nil {
		return
	}

	// set the placementMode
	j.transformPlacementMode(dataset, value)
	return
}

func (j *JuiceFSEngine) genEdition(mount datav1alpha1.Mount, value *JuiceFS, SharedEncryptOptions []datav1alpha1.EncryptOption) {
	value.Edition = EnterpriseEdition

	for _, encryptOption := range SharedEncryptOptions {
		key := encryptOption.Name

		if key == JuiceMetaUrl {
			value.Edition = CommunityEdition
		}
	}

	for _, encryptOption := range mount.EncryptOptions {
		key := encryptOption.Name

		if key == JuiceMetaUrl {
			value.Edition = CommunityEdition
		}
	}
}

func (j *JuiceFSEngine) transformWorkers(runtime *datav1alpha1.JuiceFSRuntime, dataset *datav1alpha1.Dataset, value *JuiceFS) (err error) {

	image := runtime.Spec.JuiceFSVersion.Image
	imageTag := runtime.Spec.JuiceFSVersion.ImageTag
	imagePullPolicy := runtime.Spec.JuiceFSVersion.ImagePullPolicy

	value.Worker.Envs = runtime.Spec.Worker.Env

	value.Image, value.ImageTag, value.ImagePullPolicy, err = j.parseJuiceFSImage(value.Edition, image, imageTag, imagePullPolicy)
	if err != nil {
		return
	}

	// add ImagePullSecrets
	imagePullSecrets := docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	if len(imagePullSecrets) > 0 {
		value.ImagePullSecrets = imagePullSecrets
	}

	// nodeSelector
	value.Worker.NodeSelector = map[string]string{}
	if len(runtime.Spec.Worker.NodeSelector) > 0 {
		value.Worker.NodeSelector = runtime.Spec.Worker.NodeSelector
	}

	// options
	mount := dataset.Spec.Mounts[0]
	var tieredStoreLevel *datav1alpha1.Level
	if len(runtime.Spec.TieredStore.Levels) != 0 {
		tieredStoreLevel = &runtime.Spec.TieredStore.Levels[0]
	}
	option, err := j.genMountOptions(mount, tieredStoreLevel)
	if err != nil {
		return err
	}
	for k, v := range runtime.Spec.Worker.Options {
		option[k] = v
	}
	if runtime.Spec.Worker.Options["cache-size"] != "" || runtime.Spec.Worker.Options["cache-dir"] != "" {
		// cache-size & cache-dir in worker.options will be deprecated in the future
		// send an event in runtime
		msg := "cache-size & cache-dir in worker.options will be deprecated in the future, please use tieredStore.levels instead"
		j.Log.Info(msg)
		j.Recorder.Eventf(runtime, corev1.EventTypeWarning, common.RuntimeDeprecated, msg)
	}

	// transform mount cmd & stat cmd
	j.genWorkerMount(value, option)

	// transform resources for worker
	err = j.transformResourcesForWorker(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform resource for worker")
		return
	}

	// transform volumes for worker
	err = j.transformWorkerVolumes(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform volumes for worker")
	}
	// transform cache volumes for worker
	err = j.transformWorkerCacheVolumes(runtime, value)
	if err != nil {
		j.Log.Error(err, "failed to transform cache volumes for worker")
		return err
	}

	// parse work pod network mode
	value.Worker.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode)
	return
}

// genMount: generate mount args
func (j *JuiceFSEngine) genWorkerMount(value *JuiceFS, workerOptionMap map[string]string) {
	var mountArgsWorker []string
	if workerOptionMap == nil {
		workerOptionMap = map[string]string{}
	}
	if value.Edition == CommunityEdition {
		if _, ok := workerOptionMap["metrics"]; !ok {
			metricsPort := DefaultMetricsPort
			if value.Worker.MetricsPort != nil {
				metricsPort = *value.Worker.MetricsPort
			}
			workerOptionMap["metrics"] = fmt.Sprintf("0.0.0.0:%d", metricsPort)
		}
		mountArgsWorker = []string{
			"exec",
			common.JuiceFSCeMountPath,
			value.Source,
			security.EscapeBashStr(value.Worker.MountPath),
			"-o",
			security.EscapeBashStr(strings.Join(genArgs(workerOptionMap), ",")),
		}
	} else {
		workerOptionMap["foreground"] = ""
		// do not update config again
		workerOptionMap["no-update"] = ""

		// start independent cache cluster, refer to [juicefs cache sharing](https://juicefs.com/docs/cloud/cache/#client_cache_sharing)
		// fuse and worker use the same cache-group, fuse use no-sharing
		cacheGroup := fmt.Sprintf("%s-%s", j.namespace, security.EscapeBashStr(value.FullnameOverride))
		if _, ok := workerOptionMap["cache-group"]; ok {
			cacheGroup = workerOptionMap["cache-group"]
		}
		workerOptionMap["cache-group"] = cacheGroup
		delete(workerOptionMap, "no-sharing")

		mountArgsWorker = []string{
			"exec",
			common.JuiceFSMountPath,
			value.Source,
			security.EscapeBashStr(value.Worker.MountPath),
			"-o",
			security.EscapeBashStr(strings.Join(genArgs(workerOptionMap), ",")),
		}
	}

	value.Worker.Command = strings.Join(mountArgsWorker, " ")
	value.Worker.StatCmd = "stat -c %i " + security.EscapeBashStr(value.Worker.MountPath)
}

func (j *JuiceFSEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *JuiceFS) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (j *JuiceFSEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *JuiceFS) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}

func (j *JuiceFSEngine) transformPodMetadata(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) (err error) {
	commonLabels := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Labels)
	value.Worker.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Worker.PodMetadata.Labels)
	value.Fuse.Labels = utils.UnionMapsWithOverride(commonLabels, runtime.Spec.Fuse.PodMetadata.Labels)

	commonAnnotations := utils.UnionMapsWithOverride(map[string]string{}, runtime.Spec.PodMetadata.Annotations)
	value.Worker.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Worker.PodMetadata.Annotations)
	value.Fuse.Annotations = utils.UnionMapsWithOverride(commonAnnotations, runtime.Spec.Fuse.PodMetadata.Annotations)

	return nil
}

func (j *JuiceFSEngine) allocatePorts(runtime *datav1alpha1.JuiceFSRuntime, value *JuiceFS) error {
	if value.Edition == EnterpriseEdition {
		// enterprise edition do not need metrics port
		return nil
	}
	fuseMetricsPort, err := GetMetricsPort(runtime.Spec.Fuse.Options)
	if err != nil {
		return err
	}
	workerMetricsPort, err := GetMetricsPort(runtime.Spec.Worker.Options)
	if err != nil {
		return err
	}

	// if not use hostnetwork then use default port
	// use hostnetwork to choose port from port allocator

	expectWorkerPodNum, expectFusePodNum := 1, 1
	if !datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode) {
		value.Worker.MetricsPort = &workerMetricsPort
		expectWorkerPodNum--
	}
	if !datav1alpha1.IsHostNetwork(runtime.Spec.Fuse.NetworkMode) {
		value.Fuse.MetricsPort = &fuseMetricsPort
		expectFusePodNum--
	}
	if expectWorkerPodNum+expectFusePodNum == 0 {
		return nil
	}

	allocator, err := portallocator.GetRuntimePortAllocator()
	if err != nil {
		j.Log.Error(err, "can't get runtime port allocator")
		return err
	}

	allocatedPorts, err := allocator.GetAvailablePorts(expectFusePodNum + expectWorkerPodNum)
	if err != nil {
		j.Log.Error(err, "can't get available ports", "expected port num", expectFusePodNum+expectWorkerPodNum)
		return err
	}

	index := 0
	if expectWorkerPodNum > 0 {
		value.Worker.MetricsPort = &allocatedPorts[index]
		index++
	}
	if expectFusePodNum > 0 {
		value.Fuse.MetricsPort = &allocatedPorts[index]
	}
	return nil
}
