/*
Copyright 2026 The Fluid Authors.

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

package runtimeconfiginjector

import (
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
   This plugin injects a cache runtime's config into app pods of runtimes that have
   no fuse client, where the app talks to the cache service through its own SDK
   instead of a mount point. It is runtime-agnostic: everything it needs comes from
   the pod's own labels and annotations plus the config map fluid already generates.
*/

const Name string = "RuntimeConfigInjector"

// The injected volume is named after the config map, and a volume name must be a
// DNS-1123 label, so the dataset name cannot use up more than what is left of 63.
const maxDatasetNameLen = 63 - len("fluid-runtime-config-")

var log = ctrl.Log.WithName(Name)

type RuntimeConfigInjector struct {
	client client.Client
	name   string
}

// configMount is everything injected on behalf of a single dataset. The three parts
// share one name and a derived path, so they are built together and stay consistent.
type configMount struct {
	volume      corev1.Volume
	volumeMount corev1.VolumeMount
	env         corev1.EnvVar
}

var _ api.MutatingHandler = &RuntimeConfigInjector{}

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	return &RuntimeConfigInjector{
		client: c,
		name:   Name,
	}, nil
}

func (injector *RuntimeConfigInjector) GetName() string {
	return injector.name
}

func (injector *RuntimeConfigInjector) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	if common.CheckExpectValue(pod.Labels, common.InjectSidecarDone, common.True) {
		return false, nil
	}

	if len(runtimeInfos) == 0 {
		return false, nil
	}

	// Resolve everything before touching the pod. A plain error from a plugin does not
	// stop the pod from being admitted, so a failure halfway through would otherwise
	// leave the user with a partially injected pod and no explanation.
	//
	// The names are sorted because iterating a map has no stable order, and an unstable
	// order would make the injected volumes and env vars come out differently on each
	// admission of the very same pod spec.
	datasetNames := make([]string, 0, len(runtimeInfos))
	for ds := range runtimeInfos {
		datasetNames = append(datasetNames, ds)
	}
	slices.Sort(datasetNames)

	mounts := make([]configMount, 0, len(datasetNames))
	for _, ds := range datasetNames {
		// The dataset name is already known to be a valid object name, since the caller
		// resolved a runtime info for it. What is not guaranteed is that it still fits
		// once "fluid-runtime-config-" is prepended to form the volume name.
		if len(ds) > maxDatasetNameLen {
			return true, fmt.Errorf("dataset name %q is too long (%d > %d): the injected volume name would exceed the 63-character limit",
				ds, len(ds), maxDatasetNameLen)
		}

		cmName := common.GetCacheRuntimeConfigConfigMapName(ds)
		cm, cmErr := kubeclient.GetConfigmapByName(injector.client, cmName, pod.Namespace)
		if cmErr != nil {
			return true, errors.Wrapf(cmErr, "failed to get configmap \"%s/%s\"", pod.Namespace, cmName)
		}
		// GetConfigmapByName reports NotFound as (nil, nil), so a nil config map is the
		// not-found case rather than an error.
		if cm == nil {
			return true, fmt.Errorf("configmap \"%s/%s\" not found, the cache runtime of dataset %q may not be ready yet",
				pod.Namespace, cmName, ds)
		}
		if _, ok := cm.Data[common.RuntimeConfigShellFileName]; !ok {
			return true, fmt.Errorf("configmap \"%s/%s\" has no key %q, the runtime may not support clientless access",
				pod.Namespace, cmName, common.RuntimeConfigShellFileName)
		}

		mounts = append(mounts, buildConfigMount(ds))
	}

	// From here on nothing can fail, so the pod is either fully injected or untouched.
	for _, m := range mounts {
		pod.Spec.Volumes = utils.AppendOrOverrideVolume(pod.Spec.Volumes, m.volume)
	}
	injectToContainers(pod.Spec.InitContainers, mounts)
	injectToContainers(pod.Spec.Containers, mounts)

	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}
	pod.Labels[common.InjectSidecarDone] = common.True

	log.V(1).Info("Injected runtime config into pod",
		"pod", fmt.Sprintf("%s/%s", pod.Namespace, podName(pod)),
		"datasets", len(mounts))

	return false, nil
}

// buildConfigMount assembles the volume, the mount and the env pointer for one dataset.
func buildConfigMount(ds string) configMount {
	name := common.GetCacheRuntimeConfigConfigMapName(ds)
	dir := filepath.Join(common.RuntimeConfigDir, ds)

	return configMount{
		volume: corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: name},
				},
			},
		},
		volumeMount: corev1.VolumeMount{
			// must match the volume name above
			Name:      name,
			MountPath: dir,
			ReadOnly:  true,
		},
		env: corev1.EnvVar{
			Name:  common.EnvRuntimeConfigPathPrefix + envSuffix(ds),
			Value: filepath.Join(dir, common.RuntimeConfigShellFileName),
		},
	}
}

// envSuffix turns a dataset name into a shell-safe variable name suffix. A dataset
// name is a DNS-1035 label, which allows '-' but not '_', so mapping '-' to '_'
// cannot make two different datasets share a suffix.
func envSuffix(ds string) string {
	return strings.ToUpper(strings.ReplaceAll(ds, "-", "_"))
}

func injectToContainers(containers []corev1.Container, mounts []configMount) {
	for i := range containers {
		if strings.HasPrefix(containers[i].Name, common.FuseContainerName) {
			continue
		}
		for _, m := range mounts {
			containers[i].VolumeMounts = utils.AppendOrOverrideVolumeMounts(containers[i].VolumeMounts, m.volumeMount)
			containers[i].Env = appendOrOverrideEnv(containers[i].Env, m.env)
		}
	}
}

// appendOrOverrideEnv appends env to envs, or overrides the existing one with the same name.
func appendOrOverrideEnv(envs []corev1.EnvVar, env corev1.EnvVar) []corev1.EnvVar {
	var existed bool
	for idx, e := range envs {
		if e.Name == env.Name {
			if !reflect.DeepEqual(e, env) {
				envs[idx] = env
			}
			existed = true
			break
		}
	}

	if !existed {
		envs = append(envs, env)
	}

	return envs
}

// podName returns a name usable in logs: a pod created from a workload has no name
// of its own yet at admission time, only a generateName.
func podName(pod *corev1.Pod) string {
	if len(pod.Name) != 0 {
		return pod.Name
	}
	return pod.GenerateName
}
