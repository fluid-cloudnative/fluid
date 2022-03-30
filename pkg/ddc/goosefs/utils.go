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

package goosefs

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
)

func (e *GooseFSEngine) getDataSetFileNum() (string, error) {
	fileCount, err := e.TotalFileNums()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(fileCount, 10), err
}

// getRuntime gets the goosefs runtime
func (e *GooseFSEngine) getRuntime() (*datav1alpha1.GooseFSRuntime, error) {

	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.GooseFSRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (e *GooseFSEngine) getMasterStatefulset(name string, namespace string) (master *appsv1.StatefulSet, err error) {
	master = &appsv1.StatefulSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}

func (e *GooseFSEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

// func (e *GooseFSEngine) getConfigMap(name string, namespace string) (configMap *corev1.ConfigMap, err error) {
// 	configMap = &corev1.ConfigMap{}
// 	err = e.Client.Get(context.TODO(), types.NamespacedName{
// 		Name:      name,
// 		Namespace: namespace,
// 	}, configMap)

// 	return configMap, err
// }

func (e *GooseFSEngine) getMasterPodInfo() (podName string, containerName string) {
	podName = e.name + "-master-0"
	containerName = "goosefs-master"

	return
}

func (e *GooseFSEngine) getMasterName() (dsName string) {
	return e.name + "-master"
}

func (e *GooseFSEngine) getWorkerName() (dsName string) {
	return e.name + "-worker"
}

func (e *GooseFSEngine) getFuseDaemonsetName() (dsName string) {
	return e.name + "-fuse"
}

// getRunningPodsOfDaemonset gets worker pods
//func (e *GooseFSEngine) getRunningPodsOfDaemonset(dsName, namespace string) (pods []corev1.Pod, err error) {
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

func (e *GooseFSEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/goosefs-fuse", mountRoot, e.namespace, e.name)
}

func (e *GooseFSEngine) getInitUserDir() string {
	dir := fmt.Sprintf("/tmp/fluid/%s/%s", e.namespace, e.name)
	e.Log.Info("Generate InitUser dir")
	return dir
}

// Init tierPaths when running as a non-root user: chmod on each path
// Example: "/dev/shm:/var/lib/docker/goosefs:/dev/ssd"
func (e *GooseFSEngine) getInitTierPathsEnv(runtime *datav1alpha1.GooseFSRuntime) string {
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
		path = "/" + common.GooseFSRuntime
	} else {
		path = path + "/" + common.GooseFSRuntime
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

func (e *GooseFSEngine) parseRuntimeImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.GooseFSRuntimeImageEnv)
		if len(image) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultGooseFSRuntimeImage, ":")
			if len(runtimeImageInfo) < 1 {
				panic("invalid default goosefs runtime image!")
			} else {
				image = runtimeImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.GooseFSRuntimeImageEnv)
		if len(tag) == 0 {
			runtimeImageInfo := strings.Split(common.DefaultGooseFSRuntimeImage, ":")
			if len(runtimeImageInfo) < 2 {
				panic("invalid default goosefs runtime image!")
			} else {
				tag = runtimeImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *GooseFSEngine) parseFuseImage(image string, tag string, imagePullPolicy string) (string, string, string) {
	if len(imagePullPolicy) == 0 {
		imagePullPolicy = common.DefaultImagePullPolicy
	}

	if len(image) == 0 {
		image = docker.GetImageRepoFromEnv(common.GooseFSFuseImageEnv)
		if len(image) == 0 {
			fuseImageInfo := strings.Split(common.DefaultGooseFSFuseImage, ":")
			if len(fuseImageInfo) < 1 {
				panic("invalid default goosefs fuse image!")
			} else {
				image = fuseImageInfo[0]
			}
		}
	}

	if len(tag) == 0 {
		tag = docker.GetImageTagFromEnv(common.GooseFSFuseImageEnv)
		if len(tag) == 0 {
			fuseImageInfo := strings.Split(common.DefaultGooseFSFuseImage, ":")
			if len(fuseImageInfo) < 2 {
				panic("invalid default init image!")
			} else {
				tag = fuseImageInfo[1]
			}
		}
	}

	return image, tag, imagePullPolicy
}

func (e *GooseFSEngine) GetMetadataInfoFile() string {
	return cdatabackup.GooseFSBackupPathPod + "/" + e.GetMetadataInfoFileName()
}
func (e *GooseFSEngine) GetMetadataFileName() string {
	return "metadata-backup-" + e.name + "-" + e.namespace + ".gz"
}
func (e *GooseFSEngine) GetMetadataInfoFileName() string {
	return e.name + "-" + e.namespace + ".yaml"
}

// GetWorkerUsedCapacity gets cache capacity usage for each worker as a map.
// It parses result from stdout when executing `goosefs fsadmin report capacity` command
// and extracts worker name(IP or hostname) along with used capacity for that worker
func (e *GooseFSEngine) GetWorkerUsedCapacity() (map[string]int64, error) {
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
		return nil, fmt.Errorf("can't parse result form goosefs fsadmin report capacity")
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
	// no info stored in GooseFS master. Scale in such node first.
	return 0
}
