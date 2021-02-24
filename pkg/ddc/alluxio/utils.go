/*

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

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/docker"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
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

func (e *AlluxioEngine) getMasterStatefulset(name string, namespace string) (master *appsv1.StatefulSet, err error) {
	master = &appsv1.StatefulSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
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

func (e *AlluxioEngine) getMasterStatefulsetName() (dsName string) {
	return e.name + "-master"
}

func (e *AlluxioEngine) getWorkerDaemonsetName() (dsName string) {
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

func (e *AlluxioEngine) isFluidNativeScheme(mountPoint string) bool {
	return strings.HasPrefix(mountPoint, common.PathScheme) || strings.HasPrefix(mountPoint, common.VolumeScheme)
}

func (e *AlluxioEngine) getLocalStorageDirectory() string {
	return "/underFSStorage"
}

func (e *AlluxioEngine) getInitUserDir() string {
	dir := fmt.Sprintf("/tmp/fluid/%s/%s", e.namespace, e.name)
	e.Log.Info("Generate InitUser dir")
	return dir
}

func (e *AlluxioEngine) getInitUsersArgs(runtime *datav1alpha1.AlluxioRuntime) []string {
	uid := strconv.FormatInt(*runtime.Spec.RunAs.UID, 10)
	gid := strconv.FormatInt(*runtime.Spec.RunAs.GID, 10)
	username := runtime.Spec.RunAs.UserName
	args := []string{uid + ":" + username + ":" + gid,
		gid + ":" + runtime.Spec.RunAs.GroupName}

	// groups := runtime.Spec.RunAs.Groups
	// for _, group := range groups {
	// 	gid = strconv.FormatInt(group.ID, 10)
	// 	var tmp string = " " + gid + ":" + group.Name
	// 	args = append(args, tmp)
	// }
	return args
}

func (e *AlluxioEngine) getInitUserEnv(runtime *datav1alpha1.AlluxioRuntime) string {
	return strings.Join(e.getInitUsersArgs(runtime), ",")
}

// Init tierPaths when running as a non-root user: chmod on each path
// Example: "/dev/shm:/var/lib/docker/alluxio:/dev/ssd"
func (e *AlluxioEngine) getInitTierPathsEnv(runtime *datav1alpha1.AlluxioRuntime) string {
	var tierPaths []string
	for _, level := range runtime.Spec.Tieredstore.Levels {
		paths := strings.Split(level.Path, ",")
		tierPaths = append(tierPaths, paths...)
	}
	return strings.Join(tierPaths, ":")
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.ALLUXIO_RUNTIME
	} else {
		path = path + "/" + common.ALLUXIO_RUNTIME
	}
	// e.Log.Info("Mount root", "path", path)
	return

}

func getK8sClusterUsedPort(client client.Client) ([]int, error) {
	k8sClusterUsedPorts := []int{}
	pods := &v1.PodList{}
	services := &v1.ServiceList{}

	err := client.List(context.TODO(), pods)
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, pod := range pods.Items {
		// fileter pod
		if kubeclient.ExcludeInactivePod(&pod) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				usedHostPort := port.HostPort
				if pod.Spec.HostNetwork {
					usedHostPort = port.ContainerPort
				}

				k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(usedHostPort))
			}
		}
	}

	err = client.List(context.TODO(), services)
	if err != nil {
		return k8sClusterUsedPorts, err
	}
	for _, service := range services.Items {
		if service.Spec.Type == v1.ServiceTypeNodePort || service.Spec.Type == v1.ServiceTypeLoadBalancer {
			for _, port := range service.Spec.Ports {
				k8sClusterUsedPorts = append(k8sClusterUsedPorts, int(port.NodePort))
			}
		}
	}

	fmt.Printf("Get K8S used ports, %++v", k8sClusterUsedPorts)

	return k8sClusterUsedPorts, err
}

func isPortInUsed(port int, usedPorts []int) bool {
	for _, usedPort := range usedPorts {
		if port == usedPort {
			return true
		}
	}
	return false
}

func (e *AlluxioEngine) getAvaliablePort() (allocatedPorts []int, err error) {
	usedPorts, err := getK8sClusterUsedPort(e.Client)

	portNum := PORT_NUM
	// allocate 9 port
	// master: rpc web job-rpc job-web
	// worker: rpc web job-rpc job-data job-web
	// if HA of master should allocate 11 port
	// addtion: master embedded and job-master embedded
	if e.runtime.Spec.Master.Replicas > 1 {
		portNum = PORT_NUM + 2
	}

	for i := 0; i < portNum; i++ {
		found := false
		for port := AUTO_SELECT_PORT_MIN; port <= AUTO_SELECT_PORT_MAX; port++ {
			if !isPortInUsed(port, usedPorts) && !isPortInUsed(port, allocatedPorts) {
				allocatedPorts = append(allocatedPorts, port)
				found = true
				break
			}
		}

		if !found {
			err = fmt.Errorf("all avaliable port from %d to %d are allocated", AUTO_SELECT_PORT_MIN, AUTO_SELECT_PORT_MAX)
		}
	}

	if len(allocatedPorts) != portNum {
		err = fmt.Errorf("can`t allocate enough port, got %d but expect %d", len(allocatedPorts), portNum)
	}

	return
}

func (e *AlluxioEngine) parseRuntimeImage() (image, tag string) {
	var (
		defaultImage = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio"
		defaultTag   = "2.3.0-SNAPSHOT-2c41226"
	)

	image, tag = docker.GetImageRepoTagFromEnv(common.ALLUXIO_RUNTIME_IMAGE_ENV, defaultImage, defaultTag)
	e.Log.Info("Set image", "image", image, "tag", tag)

	// if value, existed := os.LookupEnv(common.ALLUXIO_RUNTIME_IMAGE_ENV); existed {
	// 	if matched, err := regexp.MatchString("^\\S+:\\S+$", value); err == nil && matched {
	// 		k, v := docker.ParseDockerImage(repos)
	// 		if len(k) ==

	// 	} else {
	// 		e.Log.Info("Failed to parse the ALLUXIO_RUNTIME_IMAGE_ENV", "ALLUXIO_RUNTIME_IMAGE_ENV", value, "error", err)
	// 	}
	// 	e.Log.Info("Get INIT_IMAGE from Env", common.ALLUXIO_RUNTIME_IMAGE_ENV, value)
	// } else {
	// 	e.Log.Info("Use Default ALLUXIO_RUNTIME_IMAGE_ENV", "ALLUXIO_RUNTIME_IMAGE_ENV", common.DEFAULT_ALLUXIO_INIT_IMAGE)
	// }

	return
}

func (e *AlluxioEngine) parseFuseImage() (image, tag string) {
	var (
		defaultImage = "registry.cn-huhehaote.aliyuncs.com/alluxio/alluxio-fuse"
		defaultTag   = "2.3.0-SNAPSHOT-2c41226"
	)

	image, tag = docker.GetImageRepoTagFromEnv(common.ALLUXIO_FUSE_IMAGE_ENV, defaultImage, defaultTag)
	e.Log.Info("Set image", "image", image, "tag", tag)

	// if value, existed := os.LookupEnv(common.ALLUXIO_RUNTIME_IMAGE_ENV); existed {
	// 	if matched, err := regexp.MatchString("^\\S+:\\S+$", value); err == nil && matched {
	// 		k, v := docker.ParseDockerImage(repos)
	// 		if len(k) ==

	// 	} else {
	// 		e.Log.Info("Failed to parse the ALLUXIO_RUNTIME_IMAGE_ENV", "ALLUXIO_RUNTIME_IMAGE_ENV", value, "error", err)
	// 	}
	// 	e.Log.Info("Get INIT_IMAGE from Env", common.ALLUXIO_RUNTIME_IMAGE_ENV, value)
	// } else {
	// 	e.Log.Info("Use Default ALLUXIO_RUNTIME_IMAGE_ENV", "ALLUXIO_RUNTIME_IMAGE_ENV", common.DEFAULT_ALLUXIO_INIT_IMAGE)
	// }

	return
}

func (e *AlluxioEngine) GetMetadataInfoFile() string {
	return cdatabackup.BACPUP_PATH_POD + "/" + e.GetMetadataInfoFileName()
}
func (e *AlluxioEngine) GetMetadataFileName() string {
	return "metadata-backup-" + e.name + "-" + e.namespace + ".gz"
}
func (e *AlluxioEngine) GetMetadataInfoFileName() string {
	return e.name + "-" + e.namespace + ".yaml"
}
