/*
Copyright 2025 The Fluid Authors.

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

package fileprefetcher

import (
	"fmt"
	stdlog "log"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/api"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Name string = "FilePrefetcher"

var defaultFilePrefetcherImage = ""

func init() {
	imageRepo := docker.GetImageRepoFromEnv(envKeyFilePrefetcherImage)
	imageTag := docker.GetImageTagFromEnv(envKeyFilePrefetcherImage)
	if len(imageRepo) == 0 || len(imageTag) == 0 {
		stdlog.Printf("WARNING: env variable %s is not set, file prefetcher image is required in Pod's annotation", envKeyFilePrefetcherImage)
		return
	}
	defaultFilePrefetcherImage = fmt.Sprintf("%s:%s", imageRepo, imageTag)
	stdlog.Printf("Found %s value %s, using it as defaultFilePrefetcherImage", envKeyFilePrefetcherImage, defaultFilePrefetcherImage)
}

var _ api.MutatingHandler = &FilePrefetcher{}

type FilePrefetcher struct {
	client client.Client
	name   string
	log    logr.Logger
}

func NewPlugin(c client.Client, args string) (api.MutatingHandler, error) {
	return &FilePrefetcher{
		client: c,
		name:   Name,
		log:    ctrl.Log.WithName("FilePrefetcher"),
	}, nil
}

func (p *FilePrefetcher) GetName() string {
	return p.name
}

func (p *FilePrefetcher) Mutate(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	if !common.CheckExpectValue(pod.Annotations, AnnotationFilePrefetcherInject, common.True) {
		return false, nil
	}

	if common.CheckExpectValue(pod.Annotations, AnnotationFilePrefetcherInjectDone, common.True) {
		return false, nil
	}

	config, err := p.buildFilePrefetcherConfig(pod, runtimeInfos)
	if err != nil {
		p.log.Error(err, "failed to build file prefetcher config")
		err = fmt.Errorf("failed to build file prefetcher config: %v", err)
		return true, err
	}

	if len(config.GlobPaths) == 0 {
		p.log.Info("Skipping injecting file prefetcher sidecar container because there's no valid file-list defined in annotation", "annotation", AnnotationFilePrefetcherFileList)
		return false, nil
	}

	containerSpec, statusFileVolume := p.buildFilePrefetcherSidecarContainer(config)

	// Inject file prefetcher container right after fuse sidecar containers, we assume fluid's fuse sidecar container is injected together.
	// e.g. before injection: [C1, FUSE1, FUSE2, C2, C3], after injection: [C1, FUSE1, FUSE2, FILEPREFETCHER, C2, C3]
	pod.Spec.Containers = p.injectFilePrefetcherSidecar(pod.Spec.Containers, containerSpec)
	if config.AsyncPrefetch {
		statusVolumeMount := corev1.VolumeMount{
			Name:      filePrefetcherStatusVolumeName,
			MountPath: filePrefetcherStatusVolumeMountPath,
		}
		var foundPrefetcherSidecar bool = false
		for idx, ctr := range pod.Spec.Containers {
			// Skip injecting status volume until we found a file prefetcher file prefetcher sidecar
			// e.g. if the containers are [C1, FUSE1, FUSE2, FILEPREFETCHER, C2, C3], only C2 and C3 will get this status volume
			if !foundPrefetcherSidecar {
				if strings.HasPrefix(ctr.Name, filePrefetcherContainerName) {
					foundPrefetcherSidecar = true
				}
				continue
			}

			pod.Spec.Containers[idx].VolumeMounts = append(pod.Spec.Containers[idx].VolumeMounts, statusVolumeMount)
		}
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, statusFileVolume)
	pod.Annotations[AnnotationFilePrefetcherInjectDone] = common.True

	return false, nil
}

type filePrefetcherConfig struct {
	// Image is the image of the file prefetcher sidecar container
	Image string
	// AsyncPrefetch indicates whether to use async prefetching, defaulting to false
	AsyncPrefetch bool
	// VolumeMountPaths is a map of volume name to mount path
	VolumeMountPaths map[string]string
	// GlobPaths is a string indicating all the paths to prefetch. It is a semicolon-separated list of paths
	GlobPaths string
	// TimeoutSeconds is a int indicating the timeout for file prefetcher, defined in seconds
	TimeoutSeconds int
	// ExtraEnvs is a map of extra envs to inject into the file prefetcher sidecar container
	ExtraEnvs map[string]string
}

func (p *FilePrefetcher) buildFilePrefetcherConfig(pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (config filePrefetcherConfig, err error) {
	defaultFn := func(keyValues map[string]string, key string, defaultValue string) (value string) {
		if value, ok := keyValues[key]; ok {
			return value
		}
		return defaultValue
	}

	config.Image = defaultFn(pod.Annotations, AnnotationFilePrefetcherImage, defaultFilePrefetcherImage)
	if len(config.Image) == 0 {
		err = fmt.Errorf("file prefetcher's image is required, set it in pod's annotation \"%s=<image>\"", AnnotationFilePrefetcherImage)
		return
	}

	extraEnvs := map[string]string{}
	// extraEnvsStr takes the format like: '<key1>=<value1> <key2>=<value2> <key3>=<value3>'
	extraEnvsStr := defaultFn(pod.Annotations, AnnotationFilePrefetcherExtraEnvs, "")
	if len(extraEnvsStr) > 0 {
		keyValuePairs := strings.Split(extraEnvsStr, " ")
		for _, keyValuePair := range keyValuePairs {
			kvSlice := strings.Split(keyValuePair, "=")
			if len(kvSlice) != 2 {
				err = fmt.Errorf("file prefetcher's extra envs is required to be '<key1>=<value1> <key2>=<value2> <key3>=<value3>', but found unexpected key-value pair: %s", keyValuePair)
				return
			}
			extraEnvs[kvSlice[0]] = kvSlice[1]
		}
	}
	config.ExtraEnvs = extraEnvs

	fileList := defaultFn(pod.Annotations, AnnotationFilePrefetcherFileList, filePrefetcherDefaultFileList)
	if fileList == filePrefetcherDefaultFileList {
		pvcNames := make([]string, 0)
		for pvcName := range runtimeInfos {
			pvcNames = append(pvcNames, fmt.Sprintf("pvc://%s", pvcName))
		}
		fileList = strings.Join(pvcNames, ";")
	}
	volumeMountPaths, globPaths := p.parseGlobPathsFromFileList(fileList, pod, runtimeInfos)
	config.VolumeMountPaths = volumeMountPaths
	config.GlobPaths = strings.Join(globPaths, ";")

	asyncPrefetchStr := defaultFn(pod.Annotations, AnnotationFilePrefetcherAsync, "false")
	if asyncPrefetch, parseErr := strconv.ParseBool(asyncPrefetchStr); parseErr != nil {
		err = fmt.Errorf("invalid value for %s: %s, must either be false or true: %v", AnnotationFilePrefetcherAsync, asyncPrefetchStr, parseErr)
		return
	} else {
		config.AsyncPrefetch = asyncPrefetch
	}

	timeoutSecondsStr := defaultFn(pod.Annotations, AnnotationFilePrefetcherTimeoutSeconds, filePrefetcherDefaultTimeoutSecondsStr)
	if timeoutSeconds, parseErr := strconv.ParseInt(timeoutSecondsStr, 10, 32); parseErr != nil {
		err = fmt.Errorf("invalid value for %s: %s, must be of type integer: %v", AnnotationFilePrefetcherTimeoutSeconds, timeoutSecondsStr, parseErr)
		return
	} else {
		config.TimeoutSeconds = int(timeoutSeconds)
	}
	p.log.V(1).Info("building file prefetcher config", "config", config)

	return
}

func (p *FilePrefetcher) parseGlobPathsFromFileList(fileList string, pod *corev1.Pod, runtimeInfos map[string]base.RuntimeInfoInterface) (volumeMountPaths map[string]string, globPaths []string) {
	volumeMountPaths = map[string]string{}
	globPaths = []string{}

	if len(fileList) == 0 {
		return
	}

	uriPaths := strings.Split(fileList, ";")
	for _, uriPath := range uriPaths {
		if !strings.HasPrefix(uriPath, string(common.VolumeScheme)) {
			p.log.Info("skip adding path to prefetch list because it does not start with pvc://", "path", uriPath)
			continue
		}
		// e.g. uriPath="pvc://mypvc/path/to/myfolder/*.pkl" => items=["mypvc", "path", "to", "myfolder", "*.pkl"]
		items := strings.Split(strings.TrimPrefix(uriPath, string(common.VolumeScheme)), string(filepath.Separator))
		if len(items) == 0 {
			p.log.Info("skip adding path to prefetch list because it does not specify a valid persistentVolumeClaim", "path", uriPath)
			continue
		}

		var pvcName, globPath string
		if len(items) == 1 {
			pvcName = items[0]
			globPath = "**"
		} else {
			pvcName = items[0]
			globPath = filepath.Clean(fmt.Sprintf("%c%s", filepath.Separator, filepath.Join(items[1:]...)))
		}

		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				volumeMountPaths[volume.Name] = path.Join("/data", volume.Name)
				globPaths = append(globPaths, path.Join(volumeMountPaths[volume.Name], globPath))
			}
		}
	}

	return
}

// This function assembles and returns a file prefetcher sidecar container
func (p *FilePrefetcher) buildFilePrefetcherSidecarContainer(config filePrefetcherConfig) (corev1.Container, corev1.Volume) {
	volumeMounts := []corev1.VolumeMount{}
	for volumeName, mountPath := range config.VolumeMountPaths {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}

	statusFileVolume := corev1.Volume{
		Name: filePrefetcherStatusVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      statusFileVolume.Name,
		MountPath: filePrefetcherStatusVolumeMountPath,
	})

	containerSpec := corev1.Container{
		Name:  filePrefetcherContainerName,
		Image: config.Image,
		Env: []corev1.EnvVar{
			{
				Name:  envKeyFilePrefetcherFileList,
				Value: config.GlobPaths,
			},
			{
				Name:  envKeyFilePrefetcherAsyncPrefetch,
				Value: strconv.FormatBool(config.AsyncPrefetch),
			},
			{
				Name:  envKeyFilePrefetcherTimeoutSeconds,
				Value: strconv.FormatInt(int64(config.TimeoutSeconds), 10),
			},
		},
		VolumeMounts: volumeMounts,
	}

	for k, v := range config.ExtraEnvs {
		containerSpec.Env = append(containerSpec.Env, corev1.EnvVar{Name: k, Value: v})
	}

	if !config.AsyncPrefetch {
		containerSpec.Lifecycle = &corev1.Lifecycle{
			PostStart: &corev1.LifecycleHandler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"bash",
						"-c",
						`cnt=0; while [[ $cnt -lt $FILE_PREFETCHER_TIMEOUT_SECONDS ]]; do if [[ -e "/tmp/fluid-file-prefetcher/status/prefetcher.status" ]]; then exit 0; fi; cnt=$(expr $cnt + 1); sleep 1; done; echo "time out waiting for prefetching done"; exit 1`,
					},
				},
			},
		}
	}

	return containerSpec, statusFileVolume
}

func (p *FilePrefetcher) injectFilePrefetcherSidecar(oldContainers []corev1.Container, filePrefetcherCtr corev1.Container) (newContainers []corev1.Container) {
	lastFuseSidecarIndex := -1
	for idx, ctr := range oldContainers {
		if strings.HasPrefix(ctr.Name, common.FuseContainerName) {
			lastFuseSidecarIndex = idx
		}
	}

	// Insert file prefetcher sidecar after
	newContainers = make([]corev1.Container, 0, len(oldContainers)+1)
	newContainers = append(newContainers, oldContainers[:lastFuseSidecarIndex+1]...)
	newContainers = append(newContainers, filePrefetcherCtr)
	newContainers = append(newContainers, oldContainers[lastFuseSidecarIndex+1:]...)

	return newContainers
}
