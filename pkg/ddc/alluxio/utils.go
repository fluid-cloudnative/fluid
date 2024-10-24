/*
Copyright 2020 The Fluid Authors.

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

package alluxio

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fluid-cloudnative/fluid/pkg/types/cacheworkerset"
	openkruise "github.com/openkruise/kruise/apis/apps/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

func (e *AlluxioEngine) getDataSetFileNum() (string, error) {
	fileCount, err := e.TotalFileNums()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(fileCount, 10), err
}

// getRuntime gets the alluxio runtime
func (e *AlluxioEngine) getRuntime() (*datav1alpha1.AlluxioRuntime, error) {
	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.AlluxioRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (e *AlluxioEngine) getMasterPod(name string, namespace string) (pod *v1.Pod, err error) {
	pod = &v1.Pod{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, pod)

	return pod, err
}

func (e *AlluxioEngine) getMasterStatefulset(name string, namespace string) (master *appsv1.StatefulSet, err error) {
	master = &appsv1.StatefulSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}
func (e *AlluxioEngine) getMasterAdvancedStatefulset(name string, namespace string) (master *openkruise.StatefulSet, err error) {
	master = &openkruise.StatefulSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}
func (e *AlluxioEngine) getMasterCacheWorkerset(name string, namespace string) (master *cacheworkerset.CacheWorkerSet, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return nil, err // 确保在获取 runtime 失败时返回错误
	}

	switch runtime.Spec.ScaleConfig.WorkerType {
	case cacheworkerset.StatefulSetType:
		cacheMaster, err := e.getMasterStatefulset(name, namespace)
		if err != nil {
			return nil, err // 返回 nil 和错误
		}
		master = cacheworkerset.StsToCacheWorkerSet(cacheMaster) // 使用返回参数 master
		return master, nil

	case cacheworkerset.DaemonSetType:
		cacheMaster, err := e.getDaemonset(name, namespace)
		if err != nil {
			return nil, err // 返回 nil 和错误
		}
		master = cacheworkerset.DsToCacheWorkerSet(cacheMaster) // 使用返回参数 master
		return master, nil

	case cacheworkerset.AdvancedStatefulSetType:
		cacheMaster, err := e.getMasterAdvancedStatefulset(name, namespace)
		if err != nil {
			return nil, err // 返回 nil 和错误
		}
		master = cacheworkerset.AstsToCacheWorkerSet(cacheMaster) // 使用返回参数 master
		return master, nil
	}

	return nil, fmt.Errorf("unsupported worker type %s", runtime.Spec.ScaleConfig.WorkerType)
}

func (e *AlluxioEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

// func (e *AlluxioEngine) getConfigMap(name string, namespace string) (configMap *corev1.ConfigMap, err error) {
// 	configMap = &corev1.ConfigMap{}
// 	err = e.Client.Get(context.TODO(), types.NamespacedName{
// 		Name:      name,
// 		Namespace: namespace,
// 	}, configMap)

// 	return configMap, err
// }

func (e *AlluxioEngine) getMasterPodInfo() (podName string, containerName string) {
	podName = e.name + "-master-0"
	containerName = "alluxio-master"

	return
}

func (e *AlluxioEngine) getMasterName() (dsName string) {
	return e.name + "-master"
}

func (e *AlluxioEngine) getWorkerName() (dsName string) {
	return e.name + "-worker"
}

func (e *AlluxioEngine) getFuseDaemonsetName() (dsName string) {
	return e.name + "-fuse"
}

// getRunningPodsOfDaemonset gets worker pods
//func (e *AlluxioEngine) getRunningPodsOfDaemonset(dsName, namespace string) (pods []corev1.Pod, err error) {
//
//	ds, err := e.getDaemonset(dsName, namespace)
//	if err != nil {
//		return pods, err
//	}
//
//	selector := ds.Spec.Selector.MatchLabels
//	// labels := selector.MatchLabels
//
//	pods = []corev1.Pod{}
//	podList := &corev1.PodList{}
//	err = e.Client.List(context.TODO(), podList, options.InNamespace(namespace), options.MatchingLabels(selector))
//	if err != nil {
//		return pods, err
//	}
//
//	for _, pod := range podList.Items {
//		if !podutil.IsPodReady(&pod) {
//			e.Log.Info("Skip the pod because it's not ready", "pod", pod.Name, "namespace", pod.Namespace)
//			continue
//		}
//		pods = append(pods, pod)
//	}
//
//	return pods, nil
//
//}

func (e *AlluxioEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/alluxio-fuse", mountRoot, e.namespace, e.name)
}

func (e *AlluxioEngine) getInitUserDir() string {
	dir := fmt.Sprintf("/tmp/fluid/%s/%s", e.namespace, e.name)
	e.Log.Info("Generate InitUser dir")
	return dir
}

// Init tierPaths when running as a non-root user: chmod on each path
// Example: "/dev/shm:/var/lib/docker/alluxio:/dev/ssd"
func (e *AlluxioEngine) getInitTierPathsEnv(runtime *datav1alpha1.AlluxioRuntime) string {
	var tierPaths []string
	for _, level := range runtime.Spec.TieredStore.Levels {
		paths := strings.Split(level.Path, ",")
		tierPaths = append(tierPaths, paths...)
	}
	return strings.Join(tierPaths, ":")
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.AlluxioRuntime
	} else {
		path = path + "/" + common.AlluxioRuntime
	}
	// e.Log.Info("Mount root", "path", path)
	return
}

func isPortInUsed(port int, usedPorts []int) bool {
	for _, usedPort := range usedPorts {
		if port == usedPort {
			return true
		}
	}
	return false
}

func (e *AlluxioEngine) parseRuntimeImage(image string, tag string, imagePullPolicy string, imagePullSecrets []v1.LocalObjectReference) (string, string, string, []v1.LocalObjectReference) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.AlluxioRuntimeImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultAlluxioRuntimeImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default alluxio runtime image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.AlluxioRuntimeImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultAlluxioRuntimeImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default alluxio runtime image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	if len(imagePullSecrets) == 0 {
		// if the environment variable is not set, it is still an empty slice
		imagePullSecrets = docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	}

	return image, tag, imagePullPolicy, imagePullSecrets
}

func (e *AlluxioEngine) parseFuseImage(image string, tag string, imagePullPolicy string, imagePullSecrets []v1.LocalObjectReference) (string, string, string, []v1.LocalObjectReference) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.AlluxioFuseImageEnv)
		if len(image) == 0 {
			fuseImageInfo := strings.Split(common.DefaultAlluxioFuseImage, ":")
			if len(fuseImageInfo) < 1 {
				panic("invalid default alluxio fuse image!")
			} else {
				image = fuseImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.AlluxioFuseImageEnv)
		if len(tag) == 0 {
			fuseImageInfo := strings.Split(common.DefaultAlluxioFuseImage, ":")
			if len(fuseImageInfo) < 2 {
				panic("invalid default init image!")
			} else {
				tag = fuseImageInfo[1]
			}
		}
	}

	if len(imagePullSecrets) == 0 {
		// if the environment variable is not set, it is still an empty slice
		imagePullSecrets = docker.GetImagePullSecretsFromEnv(common.EnvImagePullSecretsKey)
	}

	return image, tag, imagePullPolicy, imagePullSecrets
}

func (e *AlluxioEngine) GetMetadataInfoFile() string {
	return cdatabackup.AlluxioBackupPathPod + "/" + e.GetMetadataInfoFileName()
}

func (e *AlluxioEngine) GetMetadataFileName() string {
	return "metadata-backup-" + e.name + "-" + e.namespace + ".gz"
}

func (e *AlluxioEngine) GetMetadataInfoFileName() string {
	return e.name + "-" + e.namespace + ".yaml"
}

// GetWorkerUsedCapacity gets cache capacity usage for each worker as a map.
// It parses result from stdout when executing `alluxio fsadmin report capacity` command
// and extracts worker name(IP or hostname) along with used capacity for that worker
func (e *AlluxioEngine) GetWorkerUsedCapacity() (map[string]int64, error) {
	// 2. run clean action
	capacityReport, err := e.reportCapacity()
	if err != nil {
		return nil, err
	}

	// An Example of capacityReport:
	/////////////////////////////////////////////////////////////////
	// Capacity information for all workers:
	//    Total Capacity: 4096.00MB
	//        Tier: MEM  Size: 4096.00MB
	//    Used Capacity: 443.89MB
	//        Tier: MEM  Size: 443.89MB
	//    Used Percentage: 10%
	//    Free Percentage: 90%
	//
	// Worker Name      Last Heartbeat   Storage       MEM
	// 192.168.1.147    0                capacity      2048.00MB
	//                                   used          443.89MB (21%)
	// 192.168.1.146    0                capacity      2048.00MB
	//                                   used          0B (0%)
	/////////////////////////////////////////////////////////////////
	lines := strings.Split(capacityReport, "\n")
	startIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "Worker Name") {
			startIdx = i + 1
			break
		}
	}

	if startIdx == -1 {
		return nil, fmt.Errorf("can't parse result form alluxio fsadmin report capacity")
	}

	worker2UsedCapacityMap := make(map[string]int64)
	lenLines := len(lines)
	for lineIdx := startIdx; lineIdx < lenLines; lineIdx += 2 {
		// e.g. ["192.168.1.147", "0", "capacity", "2048.00MB", "used", "443.89MB", "(21%)"]
		workerInfoFields := append(strings.Fields(lines[lineIdx]), strings.Fields(lines[lineIdx+1])...)
		workerName := workerInfoFields[0]
		usedCapacity, _ := utils.FromHumanSize(workerInfoFields[5])
		worker2UsedCapacityMap[workerName] = usedCapacity
	}

	return worker2UsedCapacityMap, nil
}

// lookUpUsedCapacity looks up used capacity for a given node in a map.
func lookUpUsedCapacity(node v1.Node, usedCapacityMap map[string]int64) int64 {
	var ip, hostname string
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			ip = addr.Address
		}
		if addr.Type == v1.NodeInternalDNS {
			hostname = addr.Address
		}
	}

	if len(ip) != 0 {
		if usedCapacity, found := usedCapacityMap[ip]; found {
			return usedCapacity
		}
	}

	if len(hostname) != 0 {
		if usedCapacity, found := usedCapacityMap[hostname]; found {
			return usedCapacity
		}
	}
	// no info stored in Alluxio master. Scale in such node first.
	return 0
}
