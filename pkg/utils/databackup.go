package utils

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cdatabackup "github.com/fluid-cloudnative/fluid/pkg/databackup"
	"k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
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
func GetAddressOfMaster(pod *v1.Pod)(nodeName string, ip string, rpcPort int32){
	nodeName = pod.Spec.NodeName
	ip = pod.Status.HostIP
	for _, container := range pod.Spec.Containers {
		if container.Name == "alluxio-master"{
			for _, port := range container.Ports {
				if port.Name == "rpc" {
					rpcPort = port.HostPort
				}
			}
		}
	}
	return
}

// CreateBackupPod create the pod to backup
func CreateBackupPod(client client.Client, masterPod *v1.Pod, dataset *datav1alpha1.Dataset, databackup *datav1alpha1.DataBackup) error {
	nodeName, ip , rpcPort := GetAddressOfMaster(masterPod)
	env := "-Dalluxio.master.hostname=" + ip + " -Dalluxio.master.rpc.port=" + strconv.Itoa(int(rpcPort))

	hostPathType := v1.HostPathDirectoryOrCreate
	volumes := []v1.Volume{
		{
			Name: "backup",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: cdatabackup.BACPUP_PATH_HOST,
					Type: &hostPathType,
				},
			},
		},
	}
	volumeMounts:= []v1.VolumeMount{
		{
			Name: "backup",
			MountPath: cdatabackup.BACPUP_PATH_POD,
		},
	}
	args := []string{
		"--namespace=" + databackup.Namespace,
		"--dataset=" + dataset.Name,
	}

	path := databackup.Spec.BackupPath
	if strings.HasPrefix(path, common.PathScheme){
		path = strings.TrimPrefix(path, common.PathScheme)
	}else if strings.HasPrefix(path, common.VolumeScheme) {
		path = strings.TrimPrefix(path, common.VolumeScheme)
		split := strings.Split(path, "/")
		path = strings.TrimPrefix(path, split[0])
		args = append(args, "--pvc")
		volumes = append(volumes, v1.Volume{
			Name: "pvc",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1. PersistentVolumeClaimVolumeSource{
					ClaimName: split[0],
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
				Name: "pvc",
				MountPath: cdatabackup.PVC_PATH_POD,
			})
	} else {
		err := fmt.Errorf("PathNotSupported")
		log.Error(err, "don't support path in this form")
		return err
	}
	args = append(args, "--subpath=" + path)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      databackup.Name,     // name of Pod will be same of databackup
			Namespace: databackup.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: databackup.APIVersion,
					Kind:       databackup.Kind,
					Name:       databackup.Name,
					UID:        databackup.UID,
				},
			},
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
			RestartPolicy: v1.RestartPolicyNever,
			Volumes: volumes,
			Containers : []v1.Container{
				{
					Name: cdatabackup.BACKUP_CONTAINER_NAME,
					Image: cdatabackup.DATABACKUP_IMA_URL + cdatabackup.DATABACKUP_IMA_TAG,
					ImagePullPolicy: v1.PullIfNotPresent,
					Args: args,
					Command: []string{
						"databackup",
						"start",
					},
					Env : []v1.EnvVar{
						{
							Name: "ALLUXIO_JAVA_OPTS",
							Value: env,
						},
					},
					VolumeMounts: volumeMounts,
				},
			},

		},

	}
	err := client.Create(context.TODO(), pod)
	if apierrs.IsNotFound(err) {
		err = nil
	}
	return err
}
