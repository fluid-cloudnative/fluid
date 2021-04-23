package jindo

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

func (e *JindoEngine) getTieredStoreType(runtime *datav1alpha1.JindoRuntime) int {
	var mediumType int
	for _, level := range runtime.Spec.Tieredstore.Levels {
		mediumType = common.GetDefaultTieredStoreOrder(level.MediumType)
	}
	return mediumType
}

func (e *JindoEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/jindofs-fuse", mountRoot, e.namespace, e.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.JINDO_RUNTIME
	} else {
		path = path + "/" + common.JINDO_RUNTIME
	}
	// e.Log.Info("Mount root", "path", path)
	return
}

// getRuntime gets the jindo runtime
func (e *JindoEngine) getRuntime() (*datav1alpha1.JindoRuntime, error) {

	key := types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}

	var runtime datav1alpha1.JindoRuntime
	if err := e.Get(context.TODO(), key, &runtime); err != nil {
		return nil, err
	}
	return &runtime, nil
}

func (e *JindoEngine) getMasterStatefulset(name string, namespace string) (master *appsv1.StatefulSet, err error) {
	master = &appsv1.StatefulSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, master)

	return master, err
}

func (e *JindoEngine) getMasterStatefulsetName() (dsName string) {
	return e.name + "-jindofs-master"
}

func (e *JindoEngine) getWorkerDaemonsetName() (dsName string) {
	return e.name + "-jindofs-worker"
}

func (e *JindoEngine) getFuseDaemonsetName() (dsName string) {
	return e.name + "-jindofs-fuse"
}

func (e *JindoEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

func (e *JindoEngine) getMasterPodInfo() (podName string, containerName string) {
	podName = e.name + "-jindofs-master-0"
	containerName = "jindofs-master"

	return
}

// return total storage size of Jindo in bytes
func (e *JindoEngine) TotalJindoStorageBytes(name string, useStsSecret bool) (value int64, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	url := "jfs://jindo/"
	ufsSize, err := fileUtils.GetUfsTotalSize(url, useStsSecret)
	e.Log.Info("jindo storage ufsSize", "ufsSize", ufsSize)
	if err != nil {
		e.Log.Error(err, "get total size")
	}
	return strconv.ParseInt(ufsSize, 10, 64)
}

func (e *JindoEngine) getAvaliablePort() (masterRpcPort int, clientRpcPort int, err error) {
	masterRpcPort = 8101
	clientRpcPort = 6101

	usedPorts, err := getK8sClusterUsedPort(e.Client)

	// lookup masterRpcPort
	found := false
	for port := DEFAULT_MASTER_RPC_PORT; port <= JINDO_MASTER_PORT_MAX; port++ {
		if !isPortInUsed(port, usedPorts) {
			masterRpcPort = port
			found = true
			break
		}
	}

	if !found {
		err = fmt.Errorf("no free pod find to assign for master from %d to %d", DEFAULT_MASTER_RPC_PORT, JINDO_MASTER_PORT_MAX)
	}

	// lookup masterRpcPort
	found = false
	for port := DEFAULT_WORKER_RPC_PORT; port <= JINDO_WORKER_PORT_MAX; port++ {
		if !isPortInUsed(port, usedPorts) {
			clientRpcPort = port
			found = true
			break
		}
	}

	if !found {
		err = fmt.Errorf("no free pod find to assign for client from %d to %d", DEFAULT_WORKER_RPC_PORT, JINDO_WORKER_PORT_MAX)
	}

	return masterRpcPort, clientRpcPort, err
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
