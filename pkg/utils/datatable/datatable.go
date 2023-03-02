package datatable

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetOrCreateAlluxio if alluxio is running, return the master ip; or return err
func GetOrCreateAlluxio(client client.Client, name, ns string) (ip string, err error) {
	dataset := datav1alpha1.Dataset{}
	nn := types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
	if err := client.Get(context.TODO(), nn, &dataset); err != nil {
		if utils.IgnoreNotFound(err) == nil {
			// not found
			return "", CreateDatasetAndAlluxio(client, name, ns)
		} else {
			// other err
			return "", err
		}
	} else {
		if dataset.Status.Phase == "Bound" {
			// Alluxio is ready
			return GetAlluxioMasterIP(client, name, ns)
		} else {
			// Alluxio is starting
			return "", nil
		}
	}
}

func CreateDatasetAndAlluxio(client client.Client, name, ns string) (err error) {
	if err = CreateDataset(client, name, ns); err != nil {
		return err
	}
	fmt.Println("--------------------------create dataset", err)
	if err = CreateAlluxioRuntime(client, name, ns); err != nil {
		return err
	}
	fmt.Println("--------------------------create alluxio", err)
	return nil
}

// CreateDataset Create the dataset CRD
func CreateDataset(client client.Client, name, ns string) (err error) {
	dataset := datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		// TODO: support more attribute settings
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "https://mirrors.tuna.tsinghua.edu.cn/apache/db/ddlutils/",
					Name:       "init-point",
				},
			},
		},
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return client.Create(context.TODO(), &dataset)
	})
	return err
}

// CreateAlluxioRuntime Create AlluxioRuntime to deploy the alluxio
func CreateAlluxioRuntime(client client.Client, name, ns string) (err error) {
	quantity := resource.MustParse("2Gi")
	alluxio := datav1alpha1.AlluxioRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		// TODO: support more attribute settings
		Spec: datav1alpha1.AlluxioRuntimeSpec{
			Replicas: 2,
			TieredStore: datav1alpha1.TieredStore{
				Levels: []datav1alpha1.Level{
					{
						MediumType: common.Memory,
						Path:       "/dev/shm",
						Quota:      &quantity,
						High:       "0.95",
						Low:        "0.7",
						VolumeType: "hostPath",
					},
				},
			},
		},
	}
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return client.Create(context.TODO(), &alluxio)
	})
	return err
}

// GetAlluxioMasterIP Get the alluxio master ip
func GetAlluxioMasterIP(client client.Client, name, ns string) (ip string, err error) {
	pod := v1.Pod{}
	nn := types.NamespacedName{
		Namespace: ns,
		Name:      name + "-master-0",
	}
	if err := client.Get(context.TODO(), nn, &pod); err != nil {
		return "", err
	} else {
		return pod.Status.PodIP, nil
	}
}
