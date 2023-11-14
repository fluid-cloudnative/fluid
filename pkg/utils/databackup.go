/*
Copyright 2023 The Fluid Authors.

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

package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDataBackup gets the DataBackup given its name and namespace
func GetDataBackup(client client.Client, name, namespace string) (*datav1alpha1.DataBackup, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var databackup datav1alpha1.DataBackup
	if err := client.Get(context.TODO(), key, &databackup); err != nil {
		return nil, err
	}
	return &databackup, nil
}

// GetAddressOfMaster return the ip and port of engine master
func GetAddressOfMaster(pod *v1.Pod) (nodeName string, ip string, rpcPort int32) {
	// TODO: Get address of master by calling runtime controller interface instead of reading pod object
	nodeName = pod.Spec.NodeName
	ip = pod.Status.HostIP
	for _, container := range pod.Spec.Containers {
		rpcPort = GetRpcPortFromMasterContainer(&container)
		if rpcPort != 0 {
			return
		}
	}
	return
}

// GetDataBackupReleaseName returns DataBackup helm release's name given the DataBackup's name
func GetDataBackupReleaseName(name string) string {
	return fmt.Sprintf("%s-charts", name)
}

// GetDataBackupPodName returns DataBackup pod's name given the DataBackup's name
func GetDataBackupPodName(name string) string {
	return fmt.Sprintf("%s-pod", name)
}

// ParseBackupRestorePath parse the BackupPath in spec of DataBackup or the RestorePath in spec of Dataset
func ParseBackupRestorePath(backupRestorePath string) (pvcName string, path string, err error) {
	if backupRestorePath == "" {
		err = errors.New("DataBackupRestorePath is empty, cannot parse")
		return
	}
	if strings.HasPrefix(backupRestorePath, common.VolumeScheme.String()) {
		path = strings.TrimPrefix(backupRestorePath, common.VolumeScheme.String())
		split := strings.Split(path, "/")
		pvcName = split[0]
		path = strings.TrimPrefix(path, pvcName)
	} else if strings.HasPrefix(backupRestorePath, common.PathScheme.String()) {
		path = strings.TrimPrefix(backupRestorePath, common.PathScheme.String())
	} else {
		err = errors.New("DataBackupRestorePath is not in supported formats, cannot parse")
		return
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return
}

// GetBackupUserDir generate the temp dir of backup user
func GetBackupUserDir(namespace string, name string) string {
	return fmt.Sprintf("/tmp/backupuser/%s/%s", namespace, name)
}

func GetRpcPortFromMasterContainer(container *v1.Container) (rpcPort int32) {
	if container == nil {
		return
	}
	if container.Name == "alluxio-master" || container.Name == "goosefs-master" {
		for _, port := range container.Ports {
			if port.Name == "rpc" {
				rpcPort = port.HostPort
				return
			}
		}
	}
	return
}
