package jindo

import (
	"context"
	"fmt"
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindo/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (e *JindoEngine) getTieredStoreType(runtime *datav1alpha1.JindoRuntime) int {
	var mediumType int
	for _, level := range runtime.Spec.TieredStore.Levels {
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

func (e *JindoEngine) getStatefulset(name string, namespace string) (master *appsv1.StatefulSet, err error) {
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

func (e *JindoEngine) getWorkertName() (dsName string) {
	return e.name + "-jindofs-worker"
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
func (e *JindoEngine) TotalJindoStorageBytes(useStsSecret bool) (value int64, err error) {
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
