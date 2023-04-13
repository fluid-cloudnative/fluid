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
	"path/filepath"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

func (e *EACEngine) getRuntime() (*datav1alpha1.EFCRuntime, error) {

	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.EFCRuntime
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

func (e *EACEngine) getWorkerPodInfo() (podName string, containerName string) {
	podName = e.getWorkerName() + "-0"
	containerName = "eac-worker"
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

func (e *EACEngine) getWorkerRunningPods() (pods []v1.Pod, err error) {
	sts, err := kubeclient.GetStatefulSet(e.Client, e.getWorkerName(), e.namespace)
	if err != nil {
		return pods, err
	}

	selector, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return pods, err
	}

	allpods, err := kubeclient.GetPodsForStatefulSet(e.Client, sts, selector)
	if err != nil {
		return pods, err
	}

	pods = make([]v1.Pod, 0, len(allpods))
	for _, pod := range allpods {
		if podutil.IsPodReady(&pod) {
			pods = append(pods, pod)
		} else {
			e.Log.V(1).Info("Skip the pod because it's not ready", "pod", pod.Name)
		}
	}

	return pods, nil
}

func (e *EACEngine) getConfigmapName() string {
	return e.name + "-" + e.runtimeType + "-values"
}

func (e *EACEngine) getWorkersEndpointsConfigmapName() string {
	return fmt.Sprintf("%s-worker-endpoints", e.name)
}

func parsePortsFromConfigMap(configMap *v1.ConfigMap) (ports []int, err error) {
	var value EAC
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return ports, err
		}
		ports = append(ports, value.Worker.Port.Rpc)
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

func (e *EACEngine) getMountInfo() (info MountInfo, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return info, err
	}
	if len(dataset.Spec.Mounts) == 0 {
		return info, fmt.Errorf("empty mount point for EFCRuntime name:%s, namespace:%s", e.name, e.namespace)
	}
	mount := dataset.Spec.Mounts[0]
	if !strings.HasSuffix(mount.MountPoint, "/") {
		mount.MountPoint = mount.MountPoint + "/"
	}

	if !strings.HasPrefix(mount.MountPoint, MountPointPrefix) {
		return info, fmt.Errorf("invalid mountpoint prefix for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
	} else {
		info.MountPoint = strings.TrimPrefix(mount.MountPoint, MountPointPrefix)
	}

	if len(strings.Split(info.MountPoint, ".")) < 2 {
		return info, fmt.Errorf("fail to parse serviceaddr for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
	} else {
		info.ServiceAddr = strings.Split(info.MountPoint, ".")[1]
	}

	if len(strings.Split(info.MountPoint, "-")) < 1 {
		return info, fmt.Errorf("fail to parse filesystemid for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
	} else {
		info.FileSystemId = strings.Split(info.MountPoint, "-")[0]
	}

	if len(strings.Split(info.MountPoint, "nas.aliyuncs.com:")) < 2 {
		return info, fmt.Errorf("fail to parse dirpath for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
	} else {
		info.DirPath = strings.Split(info.MountPoint, "nas.aliyuncs.com:")[1]
	}

	info.AccessKeyID, info.AccessKeySecret, err = e.getEACSecret(mount)
	if err != nil {
		return info, err
	}

	e.Log.Info("EFCRuntime MountInfo", "mountPoint", info.MountPoint, "ServiceAddr", info.ServiceAddr, "FileSystemId", info.FileSystemId, "DirPath", info.DirPath, "AccessKeyID", info.AccessKeyID, "AccessKeySecret", info.AccessKeySecret)
	return info, nil
}

func (e *EACEngine) getEACSecret(mount datav1alpha1.Mount) (accessKeyID string, accessKeySecret string, err error) {
	for _, encryptOption := range mount.EncryptOptions {
		var result *string
		switch encryptOption.Name {
		case AccessKeyIDName:
			result = &accessKeyID
		case AccessKeySecretName:
			result = &accessKeySecret
		default:
			continue
		}

		secretKeyRef := encryptOption.ValueFrom.SecretKeyRef
		secret, err := kubeclient.GetSecret(e.Client, secretKeyRef.Name, e.namespace)
		if err != nil {
			e.Log.Info("can't get the secret",
				"namespace", e.namespace,
				"name", e.name,
				"secretName", secretKeyRef.Name)
			return "", "", err
		}

		value, ok := secret.Data[secretKeyRef.Key]
		if !ok {
			err = fmt.Errorf("can't get %s from secret %s namespace %s", secretKeyRef.Key, secretKeyRef.Name, e.namespace)
			return "", "", err
		}
		*result = string(value)
	}

	return
}
