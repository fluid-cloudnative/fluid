/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
		mountPath := "jindo:///"
		if mount.Path != "/" {
			mountPath += mount.Name
		}
		mountPathSize, err := fileUtils.GetUfsTotalSize(mountPath)
		e.Log.Info("jindofsx storage ufsMount size", "ufsSize", mountPath)
		if err != nil {
			e.Log.Error(err, "get total size with path error", mountPath)
		}
		mountSize, err := strconv.ParseInt(mountPathSize, 10, 64)
		if err != nil {
			e.Log.Error(err, "ParseInt with mount size failed")
		}
		ufsSize += mountSize
	}
	return ufsSize, err
}
