package mutating

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
)

func CheckIfPVCIsDataset(pvc *corev1.PersistentVolumeClaim) (isDataset bool) {

	return kubeclient.CheckIfPVCIsDataset(pvc)
}
