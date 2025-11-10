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

package efc

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
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

func (e *EFCEngine) getRuntime() (*datav1alpha1.EFCRuntime, error) {

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

func (e *EFCEngine) getMasterName() (dsName string) {
	return e.name + "-master"
}

func (e *EFCEngine) getWorkerName() (dsName string) {
	return e.name + "-worker"
}

func (e *EFCEngine) getFuseName() (dsName string) {
	return e.name + "-fuse"
}

func (e *EFCEngine) getMasterPodInfo() (podName string, containerName string) {
	podName = e.getMasterName() + "-0"
	containerName = "efc-master"
	return
}

func (e *EFCEngine) getWorkerPodInfo() (podName string, containerName string) {
	podName = e.getWorkerName() + "-0"
	containerName = "efc-worker"
	return
}

func (e *EFCEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

func (e *EFCEngine) getMountPath() (mountPath string) {
	return filepath.Join(e.getHostMountPath(), FuseMountDir)
}

func (e *EFCEngine) getHostMountPath() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s", mountRoot, e.namespace, e.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.EFCRuntime
	} else {
		path = path + "/" + common.EFCRuntime
	}
	return
}

func (e *EFCEngine) getWorkerRunningPods() (pods []v1.Pod, err error) {
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

func (e *EFCEngine) getHelmValuesConfigMapName() string {
	return e.name + "-" + e.engineImpl + "-values"
}

func (e *EFCEngine) getWorkersEndpointsConfigmapName() string {
	return fmt.Sprintf("%s-worker-endpoints", e.name)
}

func parsePortsFromConfigMap(configMap *v1.ConfigMap) (ports []int, err error) {
	var value EFC
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return ports, err
		}
		ports = append(ports, value.Worker.Port.Rpc)
	}
	return ports, nil
}

func parseCacheDirFromConfigMap(configMap *v1.ConfigMap) (cacheDir string, cacheType common.VolumeType, err error) {
	var value EFC
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

// getMountInfo retrieves and parses mount configuration from the associated Dataset resource.
// The function performs the following operations:
//  1. Fetches the Dataset resource using the controller's client
//  2. Validates the existence of mount configuration
//  3. Normalizes the mount path format (trims whitespace, ensures trailing slash)
//  4. Parses filesystem-specific metadata based on protocol prefix:
//     - NAS (nfs://): Extracts filesystem ID, service endpoint, and directory path
//     - CPFS (cpfs://): Extracts filesystem ID, cluster address, and share path
//
// Returns:
//  - MountInfo struct containing parsed mount details:
//      MountPoint        : Cleaned mount path without protocol prefix
//      MountPointPrefix  : Protocol identifier (e.g., "nfs://" or "cpfs://")
//      FileSystemId      : Unique filesystem identifier (NAS IDs are truncated to prefix)
//      ServiceAddr       : Regional service endpoint (e.g., "cn-hangzhou.nas.aliyuncs.com")
//      DirPath           : Absolute path within the filesystem
//  - error in these cases:
//      * Dataset retrieval failure
//      * Missing mount configuration in Dataset
//      * Invalid mount path format (doesn't match expected patterns)
//      * Unsupported protocol prefix
//      * Regex parsing failures
//
// Notes:
//  - Only processes the FIRST mount entry in Dataset.Spec.Mounts
//  - Enforces strict path normalization: trims whitespace and appends trailing slash
//  - NAS filesystem IDs are truncated at first hyphen (e.g., "fs-12345" becomes "fs")
//  - Supported prefixes:
//      NasMountPointPrefix = "nfs://"
//      CpfsMountPointPrefix = "cpfs://"
//  - All errors include contextual identifiers (Runtime name/namespace/mountpoint) for diagnostics
//  - Successful parsing emits structured log with all extracted mount parameters

func (e *EFCEngine) getMountInfo() (info MountInfo, err error) {
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return info, err
	}
	if len(dataset.Spec.Mounts) == 0 {
		return info, fmt.Errorf("empty mount point for EFCRuntime name:%s, namespace:%s", e.name, e.namespace)
	}

	mount := dataset.Spec.Mounts[0]
	mount.MountPoint = strings.TrimSpace(mount.MountPoint)
	if !strings.HasSuffix(mount.MountPoint, "/") {
		mount.MountPoint = mount.MountPoint + "/"
	}

	if strings.HasPrefix(mount.MountPoint, NasMountPointPrefix) {
		reg, err := regexp.Compile(`^(nfs://)([a-z0-9-]+)\.([a-z0-9-]+)\.nas\.aliyuncs\.com:`)
		if err != nil {
			return info, fmt.Errorf("error regexp nas mount point for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
		}

		result := reg.FindAllStringSubmatch(mount.MountPoint, -1)
		if len(result) == 0 || len(result[0]) != 4 {
			return info, fmt.Errorf("error nas mount point format for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
		}

		info.MountPoint = strings.TrimPrefix(mount.MountPoint, NasMountPointPrefix)
		info.MountPointPrefix = result[0][1]
		info.FileSystemId = strings.Split(result[0][2], "-")[0]
		info.ServiceAddr = result[0][3]
		info.DirPath = strings.TrimPrefix(mount.MountPoint, result[0][0])
	} else if strings.HasPrefix(mount.MountPoint, CpfsMountPointPrefix) {
		reg, err := regexp.Compile(`^(cpfs://)([a-z0-9-]+)\.([a-z0-9-]+)\.cpfs\.aliyuncs\.com:/share`)
		if err != nil {
			return info, fmt.Errorf("error regexp cpfs mount point for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
		}

		result := reg.FindAllStringSubmatch(mount.MountPoint, -1)
		if len(result) == 0 || len(result[0]) != 4 {
			return info, fmt.Errorf("error cpfs mount point format for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
		}

		info.MountPoint = strings.TrimPrefix(mount.MountPoint, CpfsMountPointPrefix)
		info.MountPointPrefix = result[0][1]
		info.FileSystemId = result[0][2]
		info.ServiceAddr = result[0][3]
		info.DirPath = strings.TrimPrefix(mount.MountPoint, result[0][0])
	} else {
		return info, fmt.Errorf("invalid mountpoint format for EFCRuntime name:%s, namespace:%s, mountpoint:%s", e.name, e.namespace, mount.MountPoint)
	}

	e.Log.Info("EFCRuntime MountInfo", "mountPoint", info.MountPoint, "mountPointPrefix", info.MountPointPrefix, "ServiceAddr", info.ServiceAddr, "FileSystemId", info.FileSystemId, "DirPath", info.DirPath)

	return info, nil
}
