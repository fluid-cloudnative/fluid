package utils

import (
	"context"
	"errors"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"time"
)

// GetDataBackupRef returns the identity of the Backup by combining its namespace and name.
// The identity is used for identifying current lock holder on the target dataset.
func GetDataBackupRef(name, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, name)
}

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

// GetAddressOfMaster return the ip and port of alluxio-master
func GetAddressOfMaster(pod *v1.Pod) (nodeName string, ip string, rpcPort int32) {
	// TODO: Get address of master by calling runtime controller interface instead of reading pod object
	nodeName = pod.Spec.NodeName
	ip = pod.Status.HostIP
	for _, container := range pod.Spec.Containers {
		if container.Name == "alluxio-master" {
			for _, port := range container.Ports {
				if port.Name == "rpc" {
					rpcPort = port.HostPort
				}
			}
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
	if strings.HasPrefix(backupRestorePath, common.VolumeScheme) {
		path = strings.TrimPrefix(backupRestorePath, common.VolumeScheme)
		split := strings.Split(path, "/")
		pvcName = split[0]
		path = strings.TrimPrefix(path, pvcName)
	} else if strings.HasPrefix(backupRestorePath, common.PathScheme) {
		path = strings.TrimPrefix(backupRestorePath, common.PathScheme)
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

// CalTimeCost calculate the time cost of DataBackup
// the result is a string in the form of **h**m**s
func CalTimeCost(startTime int64) (timeCost string) {
	var (
		hour   int64
		minute int64
		second int64
	)
	diff := time.Now().Unix() - startTime
	hour = diff / 3600
	minute = (diff - hour*3600) / 60
	second = diff - hour*3600 - minute*60
	if hour != 0 {
		timeCost = strconv.FormatInt(hour, 10) + "h" + strconv.FormatInt(minute, 10) + "m" + strconv.FormatInt(second, 10) + "s"
	} else {
		if minute != 0 {
			timeCost = strconv.FormatInt(minute, 10) + "m" + strconv.FormatInt(second, 10) + "s"
		} else {
			timeCost = strconv.FormatInt(second, 10) + "s"
		}
	}
	return
}
