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
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	webhookutils "github.com/fluid-cloudnative/fluid/pkg/webhook/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
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

var _ api.ReaderAwareHandler = &RuntimeConfigInjector{}

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
	return injector.MutateWithReader(pod, runtimeInfos, injector.client)
}

func (injector *RuntimeConfigInjector) MutateWithReader(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface, reader client.Reader) (shouldStop bool, err error) {
	if common.CheckExpectValue(pod.Labels, common.InjectSidecarDone, common.True) {
		return false, nil
	}

	if len(runtimeInfos) == 0 {
		return false, nil
	}

	// Resolve everything before touching the pod, so that a failure halfway through
	// cannot leave a partially injected pod behind.
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
		// Only a cache runtime renders the config this plugin mounts. A dataset of any other
		// runtime listed in the annotation would otherwise be reported as a config map that
		// is not there yet, which points the user at the wrong cause.
		if runtimeType := runtimeInfos[ds].GetRuntimeType(); runtimeType != common.CacheRuntime {
			return true, webhookutils.NewRejectPodError(fmt.Errorf(
				"dataset %q is served by a %q runtime, while runtime config injection only supports cache runtimes; mount the dataset through its PVC instead",
				ds, runtimeType))
		}

		// The dataset name is already known to be a valid object name, since the caller
		// resolved a runtime info for it. What is not guaranteed for a dataset that comes
		// from a PVC is that it still fits once "fluid-runtime-config-" is prepended to
		// form the volume name.
		if len(ds) > common.MaxClientlessDatasetNameLength {
			return true, webhookutils.NewRejectPodError(fmt.Errorf(
				"dataset name %q is too long (%d > %d): the injected volume name would exceed the 63-character limit",
				ds, len(ds), common.MaxClientlessDatasetNameLength))
		}

		mount := buildConfigMount(ds)
		// A dataset object name may contain dots, which a volume name may not. The annotation
		// path rejects such a name up front; a dataset resolved from a PVC has not been checked.
		if nameErrs := validation.IsDNS1123Label(mount.volume.Name); len(nameErrs) > 0 {
			return true, webhookutils.NewRejectPodError(fmt.Errorf(
				"dataset name %q cannot form the volume name %q of its runtime config: %s",
				ds, mount.volume.Name, strings.Join(nameErrs, "; ")))
		}
		if conflictErr := checkVolumeNameConflict(pod, mount); conflictErr != nil {
			return true, webhookutils.NewRejectPodError(conflictErr)
		}
		if conflictErr := checkMountPathConflict(pod, mount.volumeMount); conflictErr != nil {
			return true, webhookutils.NewRejectPodError(conflictErr)
		}

		cmName := common.GetCacheRuntimeConfigConfigMapName(ds)
		cm := &corev1.ConfigMap{}
		// A missing config map or key is returned as NeedRetryWithApiReaderError: the cached
		// client may lag behind a dataset that has just been bound, so the handler retries
		// with the API reader, and rejects the pod if the config is still not there.
		getErr := reader.Get(context.TODO(), types.NamespacedName{Namespace: pod.Namespace, Name: cmName}, cm)
		if apierrors.IsNotFound(getErr) {
			return true, webhookutils.NewNeedRetryWithApiReaderError(fmt.Errorf(
				"configmap \"%s/%s\" not found, the cache runtime of dataset %q may not be ready yet",
				pod.Namespace, cmName, ds))
		}
		if getErr != nil {
			return true, webhookutils.NewNeedRetryWithApiReaderError(
				errors.Wrapf(getErr, "failed to get configmap \"%s/%s\"", pod.Namespace, cmName))
		}
		if _, ok := cm.Data[common.RuntimeConfigShellFileName]; !ok {
			return true, webhookutils.NewNeedRetryWithApiReaderError(fmt.Errorf(
				"configmap \"%s/%s\" has no key %q, the runtime may not support clientless access",
				pod.Namespace, cmName, common.RuntimeConfigShellFileName))
		}

		mounts = append(mounts, mount)
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

// checkVolumeNameConflict fails if the pod already uses the config's volume name for
// something else: the injection overrides a volume or mount of the same name, which would
// silently swap the app's own volume for the config, drop the files and modes it projects,
// or move its mount elsewhere. What is left alone is a spec the injection has already
// produced, possibly with api server defaults filled in.
func checkVolumeNameConflict(pod *corev1.Pod, mount configMount) error {
	for _, existing := range pod.Spec.Volumes {
		if existing.Name != mount.volume.Name {
			continue
		}
		if existing.ConfigMap == nil || existing.ConfigMap.Name != mount.volume.ConfigMap.Name {
			return fmt.Errorf("the pod already has a volume named %q that is not the config map %q, so the runtime config cannot be injected",
				existing.Name, mount.volume.ConfigMap.Name)
		}
		if !isDefaultConfigMapProjection(existing.ConfigMap) {
			return fmt.Errorf("the pod already has a volume named %q that projects the config map %q with its own items, mode or optional setting, which the runtime config injection would replace",
				existing.Name, existing.ConfigMap.Name)
		}
	}

	for _, containers := range [][]corev1.Container{pod.Spec.InitContainers, pod.Spec.Containers} {
		for _, container := range containers {
			if isFuseSidecar(container.Name) {
				continue
			}
			for _, existing := range container.VolumeMounts {
				if existing.Name != mount.volumeMount.Name {
					continue
				}
				if filepath.Clean(existing.MountPath) != mount.volumeMount.MountPath || len(existing.SubPath) != 0 || len(existing.SubPathExpr) != 0 {
					return fmt.Errorf("container %q already mounts volume %q at %s, while the runtime config would be mounted at %s",
						container.Name, existing.Name, existing.MountPath, mount.volumeMount.MountPath)
				}
			}
		}
	}
	return nil
}

// isDefaultConfigMapProjection reports whether a config map volume projects every key with
// the default mode and is not optional, i.e. whether it is what the injection builds once
// the api server has defaulted it.
func isDefaultConfigMapProjection(source *corev1.ConfigMapVolumeSource) bool {
	return len(source.Items) == 0 &&
		(source.DefaultMode == nil || *source.DefaultMode == corev1.ConfigMapVolumeSourceDefaultMode) &&
		(source.Optional == nil || !*source.Optional)
}

// checkMountPathConflict fails if a container the config would be injected into already
// mounts another volume at the config's mount path: the injection only replaces a mount
// of the same volume, so the pod would end up with two mounts at one path, which the API
// server rejects with an error that does not point at the injection.
func checkMountPathConflict(pod *corev1.Pod, volumeMount corev1.VolumeMount) error {
	for _, containers := range [][]corev1.Container{pod.Spec.InitContainers, pod.Spec.Containers} {
		for _, container := range containers {
			if isFuseSidecar(container.Name) {
				continue
			}
			for _, existing := range container.VolumeMounts {
				if existing.Name != volumeMount.Name && filepath.Clean(existing.MountPath) == volumeMount.MountPath {
					return fmt.Errorf("container %q already mounts volume %q at %s, where the runtime config would be injected",
						container.Name, existing.Name, volumeMount.MountPath)
				}
			}
		}
	}
	return nil
}

// isFuseSidecar reports whether a container is a fuse sidecar fluid injects, which is named
// fluid-fuse-<index>. Matching the prefix alone would also skip an app container whose name
// merely starts with it.
func isFuseSidecar(name string) bool {
	suffix, ok := strings.CutPrefix(name, common.FuseContainerName)
	if !ok {
		return false
	}
	if suffix == "" {
		return true
	}
	index, ok := strings.CutPrefix(suffix, "-")
	if !ok || index == "" {
		return false
	}
	for _, r := range index {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// envSuffix turns a dataset name into a shell-safe variable name suffix. A dataset is
// only injected once its volume name passes the DNS-1123 label check, so an injected
// name holds only lower case letters, digits and '-'. Mapping '-' to '_' therefore cannot
// make two injected datasets share a suffix, and a leading digit is harmless behind the
// fixed variable prefix. A name rejected by that check may still pass through here first,
// but nothing built from it is injected.
func envSuffix(ds string) string {
	return strings.ToUpper(strings.ReplaceAll(ds, "-", "_"))
}

func injectToContainers(containers []corev1.Container, mounts []configMount) {
	for i := range containers {
		if isFuseSidecar(containers[i].Name) {
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
