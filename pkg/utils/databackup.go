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
