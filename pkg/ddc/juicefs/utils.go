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
	"sigs.k8s.io/yaml"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

func (j *JuiceFSEngine) getTieredStoreType(runtime *datav1alpha1.JuiceFSRuntime) int {
	var mediumType int
	for _, level := range runtime.Spec.TieredStore.Levels {
		mediumType = common.GetDefaultTieredStoreOrder(level.MediumType)
	}
	return mediumType
}

func (j *JuiceFSEngine) hasTieredStore(runtime *datav1alpha1.JuiceFSRuntime) bool {
	return len(runtime.Spec.TieredStore.Levels) > 0
}

func (j *JuiceFSEngine) getDataSetFileNum() (string, error) {
	fileCount, err := j.TotalFileNums()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(fileCount, 10), err
}

// getRuntime gets the juicefs runtime
func (j *JuiceFSEngine) getRuntime() (*datav1alpha1.JuiceFSRuntime, error) {

	key := types.NamespacedName{
		Name:      j.name,
		Namespace: j.namespace,
	}

	var runtime datav1alpha1.JuiceFSRuntime
	if err := j.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (j *JuiceFSEngine) getFuseDaemonsetName() (dsName string) {
	return j.name + "-fuse"
}
func (j *JuiceFSEngine) getWorkerName() (dsName string) {
	return j.name + "-worker"
}

func (j *JuiceFSEngine) getDaemonset(name string, namespace string) (fuse *appsv1.DaemonSet, err error) {
	fuse = &appsv1.DaemonSet{}
	err = j.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, fuse)

	return fuse, err
}

func (j *JuiceFSEngine) GetRunningPodsOfDaemonset(dsName string, namespace string) (pods []corev1.Pod, err error) {
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

func (j *JuiceFSEngine) GetRunningPodsOfStatefulSet(stsName string, namespace string) (pods []corev1.Pod, err error) {
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

func (j *JuiceFSEngine) getSecret(name string, namespace string) (fuse *corev1.Secret, err error) {
	fuse = &corev1.Secret{}
	err = j.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, fuse)

	return fuse, err
}

func (j *JuiceFSEngine) parseRuntimeImage(image string, tag string, imagePullPolicy string) (string, string, string) {
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

func (j *JuiceFSEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.JuiceFSFuseImageEnv)
		if len(image) == 0 {
			fuseImageInfo := strings.Split(common.DefaultJuiceFSFuseImage, ":")
			if len(fuseImageInfo) < 1 {
				panic("invalid default juicefs fuse image!")
			} else {
				image = fuseImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.JuiceFSFuseImageEnv)
		if len(tag) == 0 {
			fuseImageInfo := strings.Split(common.DefaultJuiceFSFuseImage, ":")
			if len(fuseImageInfo) < 2 {
				panic("invalid default init image!")
			} else {
				tag = fuseImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (j *JuiceFSEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	j.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/juicefs-fuse", mountRoot, j.namespace, j.name)
}

func (j *JuiceFSEngine) getHostMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	j.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s", mountRoot, j.namespace, j.name)
}

func (j *JuiceFSEngine) GetValuesConfigMap() (cm *corev1.ConfigMap, err error) {
	jfsValues := j.getConfigmapName()

	cm = &corev1.ConfigMap{}
	err = j.Client.Get(context.TODO(), types.NamespacedName{
		Name:      jfsValues,
		Namespace: j.namespace,
	}, cm)

	return cm, err
}

func (j *JuiceFSEngine) GetEdition() (edition string) {
	cm, err := j.GetValuesConfigMap()
	if err != nil {
		return ""
	}

	data := []byte(cm.Data["data"])
	fuseValues := make(map[string]interface{})
	err = yaml.Unmarshal(data, &fuseValues)
	if err != nil {
		return ""
	}

	editionStr, ok := fuseValues["edition"]
	if !ok {
		return ""
	}

	edition = editionStr.(string)
	return
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
	jPath := strings.Split(mountPoint, "juicefs://")
	if len(jPath) != 2 {
		return "", fmt.Errorf("MountPoint error, can not parse jfs path")
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

func parseVersion(version string) (*ClientVersion, error) {
	if version == common.NightlyTag {
		return &ClientVersion{
			Tag: common.NightlyTag,
		}, nil
	}
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)
	matches := re.FindStringSubmatch(strings.TrimSpace(version))
	if matches == nil || len(matches) < 4 {
		return nil, fmt.Errorf("invalid version string: %s", version)
	}
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", matches[1])
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", matches[2])
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", matches[3])
	}
	var tag string
	if len(matches) > 4 {
		tag = matches[4]
	}

	return &ClientVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
		Tag:   tag,
	}, nil
}

type ClientVersion struct {
	Major, Minor, Patch int
	Tag                 string
}

func (v *ClientVersion) LessThan(other *ClientVersion) bool {
	if v.Tag == common.NightlyTag {
		return false
	}
	if v.Major < other.Major {
		return true
	}
	if v.Major > other.Major {
		return false
	}
	if v.Minor < other.Minor {
		return true
	}
	if v.Minor > other.Minor {
		return false
	}
	if v.Patch < other.Patch {
		return true
	}
	return false
}
