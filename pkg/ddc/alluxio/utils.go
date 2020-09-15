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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	options "sigs.k8s.io/controller-runtime/pkg/client"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

// For passwd and groups father dir consistency
var timestamp = time.Now().Format("20060102150405")

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

func (e *AlluxioEngine) getWorkerContainerName() (name string) {
	return "alluxio-worker"
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
func (e *AlluxioEngine) getRunningPodsOfDaemonset(dsName, namespace string) (pods []corev1.Pod, err error) {

	ds, err := e.getDaemonset(dsName, namespace)
	if err != nil {
		return pods, err
	}

	selector := ds.Spec.Selector.MatchLabels
	// labels := selector.MatchLabels

	pods = []corev1.Pod{}
	podList := &corev1.PodList{}
	err = e.Client.List(context.TODO(), podList, options.InNamespace(namespace), options.MatchingLabels(selector))
	if err != nil {
		return pods, err
	}

	for _, pod := range podList.Items {
		if !podutil.IsPodReady(&pod) {
			e.Log.Info("Skip the pod because it's not ready", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}
		pods = append(pods, pod)
	}

	return pods, nil

}

func (e *AlluxioEngine) getMountPoint() (mountPath string) {
	return fmt.Sprintf("%s/%s/%s/alluxio-fuse", ALLUXIO_MOUNT, e.namespace, e.name)
}

func (e *AlluxioEngine) isFluidNativeScheme(mountPoint string) bool {
	return strings.HasPrefix(mountPoint, pathScheme) || strings.HasPrefix(mountPoint, volumeScheme)
}

func (e *AlluxioEngine) getLocalStorageDirectory() string {
	return "/underFSStorage"
}

func (e *AlluxioEngine) getPasswdPath() string {
	//timestamp := time.Now().Format("20060102150405")
	passwd := "/tmp/" + timestamp + "_passwd"
	e.Log.Info("Generate passwd file")
	return passwd
}

func (e *AlluxioEngine) getGroupsPath() string {
	//timestamp := time.Now().Format("20060102150405")
	group := "/tmp/" + timestamp + "_group"
	e.Log.Info("Generate group file")
	return group
}

func (e *AlluxioEngine) getCreateArgs(runtime *datav1alpha1.AlluxioRuntime) []string {
	uid := strconv.FormatInt(*runtime.Spec.RunAs.UID, 10)
	gid := strconv.FormatInt(*runtime.Spec.RunAs.GID, 10)
	username := runtime.Spec.RunAs.UserName
	var user string = uid + ":" + username + ":" + gid
	args := []string{user}
	groups := runtime.Spec.RunAs.Groups
	for _, group := range groups {
		gid = strconv.FormatInt(group.ID, 10)
		var tmp string = " " + gid + ":" + group.Name
		args = append(args, tmp)
	}
	return args
}

func (e *AlluxioEngine) getPersistVolumeClainDirectory() string {
	return "/pvcs"
}

func (e *AlluxioEngine) containsPersistVolumeClaimSubdir(mountPoint string) bool {
	pvcAndSubdir := strings.TrimPrefix(mountPoint, volumeScheme)
	pvcAndSubdir = strings.Trim(pvcAndSubdir, "/")
	return strings.Contains(pvcAndSubdir, "/")
}
