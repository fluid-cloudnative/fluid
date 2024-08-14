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

package jindofsx

import (
	"context"
	"fmt"
	"strconv"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/jindofsx/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (e *JindoFSxEngine) getTieredStoreType(runtime *datav1alpha1.JindoRuntime) int {
	var mediumType int
	for _, level := range runtime.Spec.TieredStore.Levels {
		mediumType = common.GetDefaultTieredStoreOrder(level.MediumType)
	}
	return mediumType
}

func (e *JindoFSxEngine) hasTieredStore(runtime *datav1alpha1.JindoRuntime) bool {
	return len(runtime.Spec.TieredStore.Levels) > 0
}

func (e *JindoFSxEngine) getMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	e.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s/jindofs-fuse", mountRoot, e.namespace, e.name)
}

func (j *JindoFSxEngine) getHostMountPoint() (mountPath string) {
	mountRoot := getMountRoot()
	j.Log.Info("mountRoot", "path", mountRoot)
	return fmt.Sprintf("%s/%s/%s", mountRoot, j.namespace, j.name)
}

// getMountRoot returns the default path, if it's not set
func getMountRoot() (path string) {
	path, err := utils.GetMountRoot()
	if err != nil {
		path = "/" + common.JindoRuntime
	} else {
		path = path + "/" + common.JindoRuntime
	}
	// e.Log.Info("Mount root", "path", path)
	return
}

// getRuntime gets the jindo runtime
func (e *JindoFSxEngine) getRuntime() (*datav1alpha1.JindoRuntime, error) {

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

func (e *JindoFSxEngine) getMasterName() (dsName string) {
	return e.name + "-jindofs-master"
}

func (e *JindoFSxEngine) getWorkerName() (dsName string) {
	return e.name + "-jindofs-worker"
}

func (e *JindoFSxEngine) getFuseName() (dsName string) {
	return e.name + "-jindofs-fuse"
}

func (e *JindoFSxEngine) getDaemonset(name string, namespace string) (daemonset *appsv1.DaemonSet, err error) {
	daemonset = &appsv1.DaemonSet{}
	err = e.Client.Get(context.TODO(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, daemonset)

	return daemonset, err
}

func (e *JindoFSxEngine) getMasterPodInfo() (podName string, containerName string) {
	podName = e.name + "-jindofs-master-0"
	containerName = "jindofs-master"

	return
}

// return total storage size of Jindo in bytes
func (e *JindoFSxEngine) TotalJindoStorageBytes() (value int64, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewJindoFileUtils(podName, containerName, e.namespace, e.Log)
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return 0, err
	}

	ready := fileUtils.Ready()
	if !ready {
		return 0, fmt.Errorf("the UFS is not ready")
	}

	ufsSize := int64(0)
	for _, mount := range dataset.Spec.Mounts {
		// e.g. jindo:// + /mybucket -> jindo:///mybucket
		mountPath := "jindo://" + utils.UFSPathBuilder{}.GenUFSPathInUnifiedNamespace(mount)
		mountPathSize, err := fileUtils.GetUfsTotalSize(mountPath)
		if err != nil {
			e.Log.Error(err, "get total size with path error", "mountPath", mountPath)
			continue
		}
		mountSize, err := strconv.ParseInt(mountPathSize, 10, 64)
		if err != nil {
			e.Log.Error(err, "ParseInt with mount size failed", "mountPathSize", mountPathSize)
			continue
		}
		e.Log.V(1).Info("calculated jindofsx storage ufs size for the mount path", "mountPathSize", utils.BytesSize(float64(mountSize)), "mountPath", mountPath)
		ufsSize += mountSize
	}
	e.Log.Info("calculated total ufs size of jindofsx storage", "total ufs size", utils.BytesSize(float64(ufsSize)))
	return ufsSize, err
}
