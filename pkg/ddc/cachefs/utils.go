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

package cachefs

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	options "sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (j *CacheFSEngine) getTieredStoreType(runtime *datav1alpha1.CacheFSRuntime) int {
	var mediumType int
	for _, level := range runtime.Spec.TieredStore.Levels {
		mediumType = common.GetDefaultTieredStoreOrder(level.MediumType)
	}
	return mediumType
}

func (j *CacheFSEngine) hasTieredStore(runtime *datav1alpha1.CacheFSRuntime) bool {
	return len(runtime.Spec.TieredStore.Levels) > 0
}

func (j *CacheFSEngine) getDataSetFileNum() (string, error) {
	fileCount, err := j.TotalFileNums()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(fileCount, 10), err
}

// getRuntime gets the juicefs runtime
func (j *CacheFSEngine) getRuntime() (*datav1alpha1.CacheFSRuntime, error) {

	key := types.NamespacedName{
		Name:      j.name,
		Namespace: j.namespace,
	}

	var runtime datav1alpha1.CacheFSRuntime
	if err := j.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (j *CacheFSEngine) getFuseDaemonsetName() (dsName string) {
	return j.name + "-fuse"
}
func (j *CacheFSEngine) getWorkerName() (dsName string) {
	return j.name + "-worker"
}

func (j *CacheFSEngine) getDaemonset(name string, namespace string) (fuse *appsv1.DaemonSet, err error) {
	fuse = &appsv1.DaemonSet{}
	err = j.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, fuse)

	return fuse, err
}

func (j *CacheFSEngine) GetRunningPodsOfDaemonset(dsName string, namespace string) (pods []corev1.Pod, err error) {
	ds, err := j.getDaemonset(dsName, namespace)
	if err != nil {
		return pods, err
	}

	selector := ds.Spec.Selector.MatchLabels

	pods = []corev1.Pod{}
	podList := &corev1.PodList{}
	err = j.Client.List(context.TODO(), podList, options.InNamespace(namespace), options.MatchingLabels(selector))
	if err != nil {
		return pods, err
	}

	for _, pod := range podList.Items {
		if !podutil.IsPodReady(&pod) {
			j.Log.Info("Skip the pod because it's not ready", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (j *CacheFSEngine) GetRunningPodsOfStatefulSet(stsName string, namespace string) (pods []corev1.Pod, err error) {
	sts, err := kubeclient.GetStatefulSet(j.Client, stsName, namespace)
	if err != nil {
		return pods, err
	}

	selector := sts.Spec.Selector.MatchLabels

	pods = []corev1.Pod{}
	podList := &corev1.PodList{}
	err = j.Client.List(context.TODO(), podList, options.InNamespace(namespace), options.MatchingLabels(selector))
	if err != nil {
		return pods, err
	}

	for _, pod := range podList.Items {
		if !podutil.IsPodReady(&pod) {
			j.Log.Info("Skip the pod because it's not ready", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}
		pods = append(pods, pod)
	}

	return pods, nil
}

func (j *CacheFSEngine) getSecret(name string, namespace string) (fuse *corev1.Secret, err error) {
	fuse = &corev1.Secret{}
	err = j.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, fuse)

	return fuse, err
}

func (j *CacheFSEngine) parseRuntimeImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.JuiceFSFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultJuiceFSRuntimeImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default juicefs runtime image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.JuiceFSFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultJuiceFSRuntimeImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default juicefs runtime image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (j *CacheFSEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.CacheFSFuseImageEnv)
		if len(image) == 0 {
			image = common.CacheFSDefaultImage
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.CacheFSFuseImageEnv)
		if len(tag) == 0 {
			tag = common.CacheFSDefaultImageTag
		}
	}

	return image, tag, imagePullPolicy
}

func (j *CacheFSEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	j.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/cachefs-fuse", mountRoot, j.namespace, j.name)
}

func (j *CacheFSEngine) getHostMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	j.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s", mountRoot, j.namespace, j.name)
}

func (j *CacheFSEngine) GetValuesConfigMap() (cm *corev1.ConfigMap, err error) {
	jfsValues := j.getConfigmapName()

	cm = &corev1.ConfigMap{}
	err = j.Client.Get(context.TODO(), types.NamespacedName{
		Name:      jfsValues,
		Namespace: j.namespace,
	}, cm)

	return cm, err
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.JuiceFSRuntime
	} else {
		path = path + "/" + common.JuiceFSRuntime
	}
	return
}

func parseInt64Size(sizeStr string) (int64, error) {
	size, err := strconv.ParseFloat(sizeStr, 64)
	return int64(size), err
}

func ParseSubPathFromMountPoint(mountPoint string) (string, error) {
	jPath := strings.Split(mountPoint, "cachefs://")
	if len(jPath) != 2 {
		return "", fmt.Errorf("MountPoint error, can not parse cachefs path")
	}
	return jPath[1], nil
}

func GetMetricsPort(options map[string]string) (int, error) {
	port := int64(9567)
	if options == nil {
		return int(port), nil
	}

	for k, v := range options {
		if k == "metrics" {
			re := regexp.MustCompile(`.*:([0-9]{1,6})`)
			match := re.FindStringSubmatch(v)
			if len(match) == 0 {
				return DefaultMetricsPort, fmt.Errorf("invalid metrics port: %s", v)
			}
			port, _ = strconv.ParseInt(match[1], 10, 32)
			break
		}
	}

	return int(port), nil
}
