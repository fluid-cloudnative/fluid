/*
  Copyright 2022 The Fluid Authors.

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

package eac

import (
	"context"
	"errors"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	options "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
)

func (e *EACEngine) getRuntime() (*datav1alpha1.EACRuntime, error) {

	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.EACRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (e *EACEngine) getMasterName() (dsName string) {
	return e.name + "-master"
}

func (e *EACEngine) getWorkerName() (dsName string) {
	return e.name + "-worker"
}

func (e *EACEngine) getFuseName() (dsName string) {
	return e.name + "-fuse"
}

func (e *EACEngine) getMasterPodInfo() (podName string, containerName string) {
	podName = e.getMasterName() + "-0"
	containerName = "eac-master"
	return
}

func (e *EACEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

func (e *EACEngine) getMountPath() (mountPath string) {
	return filepath.Join(e.getHostMountPath(), FuseMountDir)
}

func (e *EACEngine) getHostMountPath() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s", mountRoot, e.namespace, e.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.EACRuntime
	} else {
		path = path + "/" + common.EACRuntime
	}
	return
}

func (e *EACEngine) getDataSetFileNum() (string, error) {
	fileCount, err := e.TotalFileNums()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(fileCount, 10), err
}

func (e *EACEngine) getWorkerPods() (pods []v1.Pod, err error) {
	sts, err := kubeclient.GetStatefulSet(e.Client, e.getWorkerName(), e.namespace)
	if err != nil {
		return pods, err
	}

	selector := sts.Spec.Selector.MatchLabels

	podList := &v1.PodList{}
	err = e.Client.List(context.TODO(), podList, options.InNamespace(e.namespace), options.MatchingLabels(selector))
	if err != nil {
		return pods, err
	}

	return podList.Items, nil
}

func (e *EACEngine) getConfigmapName() string {
	return e.name + "-" + e.runtimeType + "-values"
}

func (e *EACEngine) getWorkersEndpointsConfigmapName() string {
	return fmt.Sprintf("%s-worker-endpoints", e.name)
}

func (e *EACEngine) parseMasterImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACMasterImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACMasterImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac master image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACMasterImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACMasterImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac master image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EACEngine) parseWorkerImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACWorkerImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACWorkerImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac worker image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACWorkerImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACWorkerImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac worker image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EACEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACFuseImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac fuse image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACFuseImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac fuse image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *EACEngine) parseInitFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.EACInitFuseImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACInitFuseImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default eac init alifuse image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.EACInitFuseImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultEACInitFuseImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default eac init alifuse image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func parsePortsFromConfigMap(configMap *v1.ConfigMap) (ports []int, err error) {
	var value EAC
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return nil, err
		}
		ports = append(ports, value.Worker.Port.Rpc)
		ports = append(ports, value.Fuse.Port.Monitor)
	}
	return ports, nil
}

func parseCacheDirFromConfigMap(configMap *v1.ConfigMap) (cacheDir string, cacheType common.VolumeType, err error) {
	var value EAC
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return "", "", err
		}
		cacheDir = value.getTiredStoreLevel0Path()
		cacheType = common.VolumeType(value.getTiredStoreLevel0Type())
		return
	}
	return "", "", errors.New("fail to parseCacheDirFromConfigMap")
}

func parseDirInfoFromConfigMap(configMap *v1.ConfigMap) (serviceAddr string, fileSystemId string, dirPath string, err error) {
	var value EAC
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return "", "", "", err
		}
		mountPoint := value.Fuse.MountPoint
		serviceAddr = strings.Split(mountPoint, ".")[1]
		fileSystemId = strings.Split(mountPoint, "-")[0]
		dirPath = strings.Split(mountPoint, "nas.aliyuncs.com:")[1]
		return
	}
	return "", "", "", errors.New("fail to parseDirInfoFromConfigMap")
}
